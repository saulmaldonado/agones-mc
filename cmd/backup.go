package cmd

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/james4k/rcon"
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/saulmaldonado/agones-mc/pkg/backup/google"
)

var backupCmd = cobra.Command{
	Use:   "backup",
	Short: "Saves and backsup minecraft world",
	Long:  "backup is for saving and backup up current minecraft world",
	Run:   RunBackup,
}

var backupLog *zap.SugaredLogger

func init() {
	backupCmd.PersistentFlags().String("host", "localhost", "Minecraft server host")
	backupCmd.PersistentFlags().Uint("rcon-port", 25575, "Minecraft server rcon port")
	backupCmd.PersistentFlags().String("volume", "/data", "Path to minecraft server data volume")
	backupCmd.PersistentFlags().String("gcp-bucket-name", "", "Cloud storage bucket name for storing backups")
	backupCmd.PersistentFlags().Duration("initial-delay", 0, "Initial delay in duration. ")

	RootCmd.AddCommand(&backupCmd)
}

func RunBackup(cmd *cobra.Command, args []string) {
	zLogger, _ := zap.NewProduction()
	backupLog = zLogger.Sugar().Named("agones-mc-backup")
	defer zLogger.Sync()

	dur, _ := cmd.Flags().GetDuration("initial-delay")
	if dur > 0 {
		backupLog.Infow("Initial delay...", "duration", dur.String())
		time.Sleep(dur)
	}

	pw := os.Getenv("RCON_PASSWORD")
	name := os.Getenv("NAME")

	host, _ := cmd.Flags().GetString("host")
	port, _ := cmd.Flags().GetUint("rcon-port")
	vol, _ := cmd.Flags().GetString("volume")
	bucket, _ := cmd.Flags().GetString("gcp-bucket-name")

	// Run save-all on minecraft server to force save-all before backup
	if err := saveAll(host, port, pw); err != nil {
		backupLog.Error(err)
		backupLog.Warn("Skipping save-all")
	}

	backupName := fmt.Sprintf("%s-%v.zip", name, time.Now().Format(time.RFC3339))

	// Create zip backup
	file, buff, err := zipit(path.Join(vol, "world"), backupName)
	if err != nil {
		backupLog.Error(err)
		backupLog.Fatal("backup failed")
	}

	defer file.Close()

	// Authenticate and create Google Cloud Storage client
	cloudStorageClient, err := google.New(context.Background(), bucket)
	if err != nil {
		backupLog.Error(err)
		backupLog.Fatal("Error with Google Cloud Storage")
	}

	// Backup to Google Cloud Storage
	if err := cloudStorageClient.Backup(file, buff); err != nil {
		backupLog.Error(err)
		backupLog.Fatal("Error with Google Cloud Storage")
	}

	backupLog.Infow("backup successful", "backupName", backupName)
}

func saveAll(host string, port uint, password string) error {
	if password == "" {
		return fmt.Errorf("password env var is empty")
	}

	hostport := net.JoinHostPort(host, strconv.Itoa(int(port)))

	rc, err := rcon.Dial(hostport, password)
	if err != nil {
		return err
	}

	defer rc.Close()

	reqId, err := rc.Write("save-all")
	if err != nil {
		return err
	}

	res, resId, err := rc.Read()
	if err != nil {
		return err
	}

	if reqId != resId {
		backupLog.Warnf("Mismatch in request ids", "reqId", reqId, "resId", resId)
	}

	backupLog.Info(res)

	return nil
}

func zipit(source, target string) (*os.File, *bytes.Buffer, error) {
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
