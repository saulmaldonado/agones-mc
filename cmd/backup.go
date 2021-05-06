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

	"github.com/saulmaldonado/agones-mc/pkg/backup"
	"github.com/saulmaldonado/agones-mc/pkg/backup/google"
	"github.com/saulmaldonado/agones-mc/pkg/signal"
)

type BackupConfig struct {
	Host          string
	Port          uint
	Volume        string
	RCONPassword  string
	GcpBucketName string
	InitialDelay  time.Duration
	ServerName    string
}

var backupCmd = cobra.Command{
	Use:   "backup",
	Short: "Saves and backsup minecraft world",
	Long:  "backup is for saving and backup up current minecraft world",
	Run:   Run,
}

var backupLog *zap.SugaredLogger

func init() {
	backupCmd.PersistentFlags().String("host", "localhost", "Minecraft server host")
	backupCmd.PersistentFlags().Uint("rcon-port", 25575, "Minecraft server rcon port")
	backupCmd.PersistentFlags().String("volume", "/data", "Path to minecraft server data volume")
	backupCmd.PersistentFlags().String("gcp-bucket-name", "", "Cloud storage bucket name for storing backups")
	backupCmd.PersistentFlags().Duration("initial-delay", 0, "Initial delay in duration.")
	backupCmd.PersistentFlags().String("backup-cron", "", "crin")

	RootCmd.AddCommand(&backupCmd)
}

func Run(cmd *cobra.Command, args []string) {

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
	cron, _ := cmd.Flags().GetString("backup-cron")

	cfg := &BackupConfig{host, port, vol, pw, bucket, dur, name}

	if cron != "" {
		stop := signal.SetupSignalHandler(backupLog)

		s := gocron.NewScheduler(time.UTC)

		s.Cron(cron).Do(func() {
			if err := RunBackup(cfg); err != nil {
				backupLog.Errorw("backup failed", "serverName", cfg.ServerName)
			} else {
				backupLog.Infow("backup successful", "serverName", cfg.ServerName)
			}
		})

		s.StartAsync()
		<-stop // SIGTERM
		s.Clear()
		s.Stop()
		// attempt a final backup before terminating
	}

	if err := RunBackup(cfg); err != nil {
		backupLog.Fatalw("backup failed", "serverName", cfg.ServerName)
	}

	backupLog.Infow("backup successful", "serverName", cfg.ServerName)
}

func RunBackup(cfg *BackupConfig) error {
	// Run save-all on minecraft server to force save-all before backup
	if err := saveAll(cfg.Host, cfg.Port, cfg.RCONPassword); err != nil {
		backupLog.Warn(err)
		backupLog.Warn("Skipping save-all")
	}

	// Authenticate and create Google Cloud Storage client
	cloudStorageClient, err := google.New(context.Background(), cfg.GcpBucketName)
	if err != nil {
		backupLog.Error(err)
	}

	defer cloudStorageClient.Close()

	backupName := fmt.Sprintf("%s-%v.zip", cfg.ServerName, time.Now().Format(time.RFC3339))

	// Create zip backup
	file, buff, err := backup.Zipit(path.Join(cfg.Volume, "world"), backupName)
	if err != nil {
		backupLog.Error(err)
	}

	defer file.Close()

	// Backup to Google Cloud Storage
	if err := cloudStorageClient.Backup(file, buff); err != nil {
		backupLog.Error(err)
	}

	return nil
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
