package google

import (
	"bytes"
	"context"
	"io"
	"os"

	"cloud.google.com/go/storage"

	"github.com/saulmaldonado/agones-mc/pkg/backup"
)

type GoogleClient struct {
	client  *storage.Client
	bktName string
}

func New(ctx context.Context, bucketName string) (backup.BackupClient, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, err
	}

	return &GoogleClient{client, bucketName}, nil
}

func (g *GoogleClient) Backup(file *os.File, buff *bytes.Buffer) error {
	ctx := context.Background()
	bkt := g.client.Bucket(g.bktName)

	obj := bkt.Object(file.Name())

	w := obj.NewWriter(ctx)
	w.ContentType = backup.ZipContentType

	if _, err := io.Copy(w, buff); err != nil {
		return err
	}

	if err := w.Close(); err != nil {
		return err
	}

	return nil
}

func (g *GoogleClient) Close() error {
	return g.client.Close()
}
