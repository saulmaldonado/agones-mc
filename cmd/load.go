package cmd

import (
	"context"

	"github.com/saulmaldonado/agones-mc/internal/config"
	"github.com/saulmaldonado/agones-mc/pkg/backup/google"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var loadLog *zap.SugaredLogger

var loadCmd = cobra.Command{
	Use:   "load",
	Short: "Loads minecraft world from Google Cloud Storage",
	Long:  "Load is an init container process that will load a minecraft world save/backup from Google Cloud Storage and load it into a volume",
	Run: func(cmd *cobra.Command, args []string) {
		zLogger, _ := zap.NewProduction()
		loadLog = zLogger.Sugar().Named("agones-mc-load")
		defer zLogger.Sync()

		cfg := config.NewLoadConfig()

		if cfg.GetBackupName() == "" {
			loadLog.Infow("No backup annotation. Creating a new world.")
			return
		}

		loadLog.Infow("loading saved world", "serverName", cfg.GetPodName(), "backupName", cfg.GetBackupName())

		if err := RunLoad(cfg); err != nil {
			loadLog.Fatalw("world loading failed", "serverName", cfg.GetPodName(), "backupName", cfg.GetBackupName())
		}
		loadLog.Infow("world loading succeeded", "serverName", cfg.GetPodName(), "backupName", cfg.GetBackupName())
	},
}

func init() {
	RootCmd.AddCommand(&loadCmd)
}

func RunLoad(cfg config.LoadConfig) error {
	client, err := google.New(context.Background(), cfg.GetBucketName())
	if err != nil {
		loadLog.Error(err)
		return err
	}

	if err := client.Load(cfg.GetBackupName(), cfg.GetVolume()); err != nil {
		loadLog.Error(err)
		return err
	}

	return nil
}
