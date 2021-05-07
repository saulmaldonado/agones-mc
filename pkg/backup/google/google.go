package google

import (
	"context"
	"io"
	"os"
	"path"

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

func (g *GoogleClient) Backup(file *os.File) error {
	ctx := context.Background()
	bkt := g.client.Bucket(g.bktName)

	obj := bkt.Object(file.Name())

	w := obj.NewWriter(ctx)
	w.ContentType = backup.ZipContentType

	if _, err := io.Copy(w, file); err != nil {
		return err
	}

	if err := w.Close(); err != nil {
		return err
	}

	return nil
}

func (g *GoogleClient) Load(name, targetVol string) error {
	ctx := context.Background()
	bkt := g.client.Bucket(g.bktName)

	obj := bkt.Object(name)

	r, err := obj.NewReader(ctx)
	if err != nil {
		return err
	}

	zipfile, err := os.Create(path.Join(targetVol, "world.zip"))
	if err != nil {
		return err
	}

	defer zipfile.Close()

	if _, err := io.Copy(zipfile, r); err != nil {
		return err
	}

	if err := r.Close(); err != nil {
		return err
	}

	return nil
}

func (g *GoogleClient) Close() error {
	return g.client.Close()
}
