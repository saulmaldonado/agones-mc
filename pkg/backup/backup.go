package backup

import (
	"bytes"
	"os"
)

const ZipContentType string = "application/zip"

type BackupClient interface {
	Backup(*os.File, *bytes.Buffer) error
}
