package cmd

import (
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/saulmaldonado/agones-mc/internal/config"
	"github.com/saulmaldonado/agones-mc/pkg/fileserver"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var fileServerCmd = cobra.Command{
	Use:   "fileserver",
	Short: "Minecraft GameServer pod file server",
	Long:  "Pod file server for viewing and editing minecraft world data and config files in the minecraft server's data directory",
	Run: func(cmd *cobra.Command, args []string) {
		if err := Run(); err != nil {
			logger.Fatal("file server error", zap.Error(err))
		}
	},
}

func init() {
	RootCmd.AddCommand(&fileServerCmd)
}

func Run() error {

	cfg := config.NewFileServerConfig()
	vol := cfg.GetVolume()

	http.Handle("/", http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {

		switch r.Method {
		case http.MethodGet:

			if err := fileserver.GetFile(rw, r); err != nil {
				logger.Error("error getting file", zap.Error(err))
			}

		case http.MethodPost:

			if err := fileserver.UploadFile(rw, r, vol); err != nil {
				logger.Error("error uploading new file", zap.Error(err))
			}

		case http.MethodPut:

			r.URL.Path = strings.TrimPrefix(r.URL.Path, "/")
			if _, err := os.Stat(path.Join(vol, r.URL.Path)); os.IsNotExist(err) {
				http.Error(rw, err.Error(), http.StatusNotFound)
				return
			}

			if err := fileserver.UploadFile(rw, r, vol); err != nil {
				logger.Error("error editing file", zap.Error(err))
			}

		case http.MethodDelete:

			r.URL.Path = strings.TrimPrefix(r.URL.Path, "/")
			if err := fileserver.DeleteFile(rw, r, path.Join(r.URL.Path, vol), vol); err != nil {
				logger.Error("error deleting file", zap.Error(err))
			}

		}
	}))

	logger.Info("starting server on :8080")
	return http.ListenAndServe(":8080", nil)
}
