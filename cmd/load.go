package cmd

import (
	"context"
	"os"

	"github.com/saulmaldonado/agones-mc/pkg/backup/google"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var loadLog *zap.SugaredLogger

type LoadConfig struct {
	GcpBucketName string
	ServerName    string
	Volume        string
	BackupName    string
}

var loadCmd = cobra.Command{
	Use:   "load",
	Short: "Loads minecraft world from Google Cloud Storage",
	Long:  "Load is an init container process that will load a minecraft world save/backup from Google Cloud Storage and load it into a volume",
	Run: func(cmd *cobra.Command, args []string) {
		zLogger, _ := zap.NewProduction()
		loadLog = zLogger.Sugar().Named("agones-mc-load")
		defer zLogger.Sync()

		backup := os.Getenv("BACKUP")
		name := os.Getenv("NAME")

		if backup == "" {
			loadLog.Infow("No backup annotation. Creating a new world.")
			return
		}

		bkt, _ := cmd.Flags().GetString("gcp-bucket-name")
		vol, _ := cmd.Flags().GetString("volume")
		cfg := &LoadConfig{bkt, name, vol, backup}

		loadLog.Infow("loading saved world", "serverName", cfg.ServerName, "backupName", cfg.BackupName)

		if err := RunLoad(cfg); err != nil {
			loadLog.Fatalw("world loading failed", "serverName", cfg.ServerName, "backupName", cfg.BackupName)
		}
		loadLog.Infow("world loading succeeded", "serverName", cfg.ServerName, "backupName", cfg.BackupName)
	},
}

func init() {
	loadCmd.PersistentFlags().String("gcp-bucket-name", "", "Cloud storage bucket name for storing backups")
	loadCmd.PersistentFlags().String("volume", "/data", "Path to minecraft server data volume")

	RootCmd.AddCommand(&loadCmd)
}

func RunLoad(cfg *LoadConfig) error {
	client, err := google.New(context.Background(), cfg.GcpBucketName)
	if err != nil {
		loadLog.Error(err)
		return err
	}

	if err := client.Load(cfg.BackupName, cfg.Volume); err != nil {
		loadLog.Error(err)
		return err
	}

	return nil
}
