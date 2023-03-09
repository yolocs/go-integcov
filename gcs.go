package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"cloud.google.com/go/storage"
	"golang.org/x/sync/errgroup"
)

type GoogleCloudStorage struct {
	client *storage.Client
	bucket *storage.BucketHandle
	path   string
}

func NewGoogleCloudStorage(ctx context.Context, path string) (*GoogleCloudStorage, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCS client: %w", err)
	}

	path = strings.TrimPrefix(path, "gs://")
	parts := strings.SplitN(path, "/", 2)
	if len(parts) < 2 {
		return nil, fmt.Errorf("got invalid GCS path %q, want gs://{bucket}/{dir_prefix}", path)
	}
	bucketName, p := parts[0], parts[1]

	return &GoogleCloudStorage{
		client: client,
		bucket: client.Bucket(bucketName),
		path:   p,
	}, nil
}

func (gcs *GoogleCloudStorage) Upload(ctx context.Context, dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to read dir %q: %w", dir, err)
	}

	g, ctx := errgroup.WithContext(ctx)
	for _, entry := range entries {
		entry := entry
		if entry.IsDir() {
			continue
		}
		g.Go(func() error {
			f, err := os.Open(filepath.Join(dir, entry.Name()))
			if err != nil {
				return fmt.Errorf("failed to read file %q: %w", f.Name(), err)
			}

			w := gcs.bucket.Object(gcs.path + "/" + entry.Name()).NewWriter(ctx)
			if _, err := io.Copy(w, f); err != nil {
				return fmt.Errorf("failed to copy file %q to remote: %w", f.Name(), err)
			}
			if err := w.Close(); err != nil {
				return fmt.Errorf("failed to close object: %w", err)
			}

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return fmt.Errorf("failed to copy all files: %w", err)
	}

	return nil
}
