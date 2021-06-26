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

var backupCmd = cobra.Command{
	Use:   "backup",
	Short: "Saves and backsup minecraft world",
	Long:  "backup is for saving and backup up current minecraft world",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.NewBackupConfig()

		dur := cfg.GetInitialDelay()
		if dur > 0 {
			logger.Info("initial delay...", zap.Duration("duration", dur))
			time.Sleep(dur)
		}

		if cron := cfg.GetBackupCron(); cron != "" {
			stop := signal.SetupSignalHandler(logger)

			s := gocron.NewScheduler(time.UTC)

			s.Cron(cron).Do(func() {
				if err := RunBackup(cfg); err != nil {
					logger.Error("backup failed", zap.String("serverName", cfg.GetPodName()), zap.Error(err))
				} else {
					logger.Info("backup successful", zap.String("serverName", cfg.GetPodName()))
				}
			})

			s.StartAsync()
			<-stop // SIGTERM
			s.Clear()
			s.Stop()
			// attempt a final backup before terminating
		}

		if err := RunBackup(cfg); err != nil {
			logger.Fatal("backup failed", zap.String("serverName", cfg.GetPodName()))
		}

		logger.Info("backup successful", zap.String("serverName", cfg.GetPodName()))

	},
}

func init() {
	RootCmd.AddCommand(&backupCmd)
}

func RunBackup(cfg config.BackupConfig) error {
	// Run save-all on minecraft server to force save-all before backup
	if err := saveAll(cfg.GetHost(), cfg.GetPort(), cfg.GetRCONPassword()); err != nil {
		logger.Warn("error saving world. skipping save-all", zap.Error(err))
	}

	// Authenticate and create Google Cloud Storage client
	cloudStorageClient, err := google.New(context.Background(), cfg.GetBucketName())
	if err != nil {
		logger.Error("error connecting to bucket", zap.Error(err))
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
		logger.Error("error creating zip backup", zap.Error(err))
		return err
	}

	file, err := os.Open(backupName)
	if err != nil {
		logger.Error("error creating zip backup", zap.Error(err))
		return err
	}

	defer file.Close()

	// Backup to Google Cloud Storage
	if err := cloudStorageClient.Backup(file); err != nil {
		logger.Error("error backup up to bucket", zap.Error(err))
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
		logger.Warn("mismatch RCON request and response id", zap.Int("reqId", reqId), zap.Int("resId", resId))
	}

	logger.Info(res)

	return nil
}
