package integcov

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"cloud.google.com/go/storage"
	"golang.org/x/sync/errgroup"
)

type GoogleCloudStorage struct {
	client *storage.Client
	bucket *storage.BucketHandle
}

func NewGoogleCloudStorage(ctx context.Context, bucketName string) (*GoogleCloudStorage, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCS client: %w", err)
	}

	bucket := client.Bucket(bucketName)

	return &GoogleCloudStorage{
		client: client,
		bucket: bucket,
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

			w := gcs.bucket.Object(f.Name()).NewWriter(ctx)
			if _, err := io.Copy(w, f); err != nil {
				return fmt.Errorf("failed to copy file %q to remote: %w", f.Name(), err)
			}

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return fmt.Errorf("failed to copy all files: %w", err)
	}

	return nil
}
