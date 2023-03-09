package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/sethvargo/go-envconfig"
	"golang.org/x/exp/slog"
)

type Config struct {
	GoCoverDir string `env:"GOCOVERDIR,default=/tmp/integcov"`
	Storage    string `env:"INTEGCOV_STORAGE"`
}

func main() {
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	if err := realMain(ctx); err != nil {
		done()
		log.Fatal(err)
	}
	slog.Info("integcov exited")
}

func realMain(ctx context.Context) error {
	if len(os.Args) < 2 {
		return fmt.Errorf("missing entrypoint")
	}

	entrypoint := os.Args[1]
	remainingFlags := os.Args[2:]

	var cfg Config
	if err := envconfig.Process(ctx, &cfg); err != nil {
		return fmt.Errorf("failed to process config: %w", err)
	}

	if err := os.MkdirAll(cfg.GoCoverDir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create GOCOVERDIR %q: %w", cfg.GoCoverDir, err)
	}

	storage, err := NewGoogleCloudStorage(ctx, cfg.Storage)
	if err != nil {
		return fmt.Errorf("failed to create GCS: %w", err)
	}

	cmd := exec.CommandContext(ctx, entrypoint, remainingFlags...)
	cmd.Env = append(cmd.Env, "GOCOVERDIR="+cfg.GoCoverDir)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Cancel = func() error {
		// By default, Cancel will call cmd.Process.Kill, which would not trigger
		// graceful shutdown of the actual program.
		return cmd.Process.Signal(os.Interrupt)
	}

	if err := cmd.Run(); err != nil {
		slog.Error("target exited with error", err, "entrypoint", entrypoint)
		// We don't really care.
	}

	// Don't reuse the context since it could be "done" already. It seems cloud
	// run only has 10s max for graceful shutdown. That would be all the time we
	// have to upload coverage files. But hopefully we only have a couple of files
	// to write.
	uploadCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := storage.Upload(uploadCtx, cfg.GoCoverDir); err != nil {
		return fmt.Errorf("failed to upload coverage files: %w", err)
	}

	return nil
}
