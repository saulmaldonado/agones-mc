package cmd

import (
	"context"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/saulmaldonado/agones-mc/internal/config"
	"github.com/saulmaldonado/agones-mc/pkg/backup/google"
)

var loadCmd = cobra.Command{
	Use:   "load",
	Short: "Loads minecraft world from Google Cloud Storage",
	Long:  "Load is an init container process that will load a minecraft world save/backup from Google Cloud Storage and load it into a volume",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.NewLoadConfig()

		if cfg.GetBackupName() == "" {
			logger.Info("no backup annotation. creating a new world")
			return
		}

		logger.Info("loading saved world", zap.String("serverName", cfg.GetPodName()), zap.String("backupName", cfg.GetBackupName()))

		if err := RunLoad(cfg); err != nil {
			logger.Fatal("world loading failed", zap.String("serverName", cfg.GetPodName()), zap.String("backupName", cfg.GetBackupName()))
		}
		logger.Info("world loading succeeded", zap.String("serverName", cfg.GetPodName()), zap.String("backupName", cfg.GetBackupName()))
	},
}

func init() {
	RootCmd.AddCommand(&loadCmd)
}

func RunLoad(cfg config.LoadConfig) error {
	client, err := google.New(context.Background(), cfg.GetBucketName())
	if err != nil {
		logger.Error("error connecting to bucket", zap.Error(err))
		return err
	}

	if err := client.Load(cfg.GetBackupName(), cfg.GetVolume()); err != nil {
		logger.Error("error loading world", zap.Error(err))
		return err
	}

	return nil
}
