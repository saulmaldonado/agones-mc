package cmd

import (
	"context"
	"fmt"
	"net"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/james4k/rcon"
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/saulmaldonado/agones-mc/internal/config"
	"github.com/saulmaldonado/agones-mc/pkg/backup"
	"github.com/saulmaldonado/agones-mc/pkg/backup/google"
	"github.com/saulmaldonado/agones-mc/pkg/signal"
)

var backupLog *zap.SugaredLogger

var backupCmd = cobra.Command{
	Use:   "backup",
	Short: "Saves and backsup minecraft world",
	Long:  "backup is for saving and backup up current minecraft world",
	Run: func(cmd *cobra.Command, args []string) {
		zLogger, _ := zap.NewProduction()
		backupLog = zLogger.Sugar().Named("agones-mc-backup")
		defer zLogger.Sync()

		cfg := config.NewBackupConfig()

		dur := cfg.GetInitialDelay()
		if dur > 0 {
			backupLog.Infow("Initial delay...", "duration", dur.String())
			time.Sleep(dur)
		}

		if cron := cfg.GetBackupCron(); cron != "" {
			stop := signal.SetupSignalHandler(backupLog)

			s := gocron.NewScheduler(time.UTC)

			s.Cron(cron).Do(func() {
				if err := RunBackup(cfg); err != nil {
					backupLog.Errorw("backup failed", "serverName", cfg.GetPodName())
				} else {
					backupLog.Infow("backup successful", "serverName", cfg.GetPodName())
				}
			})

			s.StartAsync()
			<-stop // SIGTERM
			s.Clear()
			s.Stop()
			// attempt a final backup before terminating
		}

		if err := RunBackup(cfg); err != nil {
			backupLog.Fatalw("backup failed", "serverName", cfg.GetPodName())
		}

		backupLog.Infow("backup successful", "serverName", cfg.GetPodName())

	},
}

func init() {
	RootCmd.AddCommand(&backupCmd)
}

func RunBackup(cfg config.BackupConfig) error {
	// Run save-all on minecraft server to force save-all before backup
	if err := saveAll(cfg.GetHost(), cfg.GetPort(), cfg.GetRCONPassword()); err != nil {
		backupLog.Warn(err)
		backupLog.Warn("Skipping save-all")
	}

	// Authenticate and create Google Cloud Storage client
	cloudStorageClient, err := google.New(context.Background(), cfg.GetBucketName())
	if err != nil {
		backupLog.Error(err)
		return err
	}

	defer cloudStorageClient.Close()

	backupName := fmt.Sprintf("%s-%v.zip", cfg.GetPodName(), time.Now().Format(time.RFC3339))

	var worldPath string
	if cfg.GetEdition() == config.JavaEdition {
		worldPath = path.Join(cfg.GetVolume(), "worlds", "Bedrock level")
	} else {
		worldPath = path.Join(cfg.GetVolume(), "world")
	}

	// Create zip backup
	err = backup.Zipit(worldPath, backupName)
	if err != nil {
		backupLog.Error(err)
		return err
	}

	file, err := os.Open(backupName)
	if err != nil {
		backupLog.Error(err)
		return err
	}

	defer file.Close()

	// Backup to Google Cloud Storage
	if err := cloudStorageClient.Backup(file); err != nil {
		backupLog.Error(err)
		return err
	}

	os.Remove(backupName)

	return nil
}

func saveAll(host string, port int, password string) error {
	if password == "" {
		return fmt.Errorf("password env var is empty")
	}

	hostport := net.JoinHostPort(host, strconv.Itoa(port))

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
