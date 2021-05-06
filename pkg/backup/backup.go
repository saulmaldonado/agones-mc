package backup

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const ZipContentType string = "application/zip"

type BackupClient interface {
	Backup(file *os.File) error
	Close() error
}

func Zipit(source, target string) error {
	zipfile, err := os.Create(target)
	if err != nil {
		return err
	}

	defer zipfile.Close()

	archive := zip.NewWriter(zipfile)
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

	return err
}
