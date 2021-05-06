package backup

import (
	"archive/zip"
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const ZipContentType string = "application/zip"

type BackupClient interface {
	Backup(*os.File, *bytes.Buffer) error
	Close() error
}

func Zipit(source, target string) (*os.File, *bytes.Buffer, error) {
	zipfile, err := os.Create(target)
	buff := &bytes.Buffer{}
	if err != nil {
		return nil, nil, err
	}

	archive := zip.NewWriter(io.MultiWriter(zipfile, buff))
	defer archive.Close()

	var baseDir string
	err = filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		if baseDir != "" {
			header.Name = filepath.Join(baseDir, strings.TrimPrefix(path, source))
		}

		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}

		writer, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}

		defer file.Close()

		_, err = io.Copy(writer, file)
		return err
	})

	return zipfile, buff, err
}
