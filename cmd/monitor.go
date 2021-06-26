package cmd

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/saulmaldonado/agones-mc/internal/config"
	"github.com/saulmaldonado/agones-mc/pkg/ping"
	"github.com/saulmaldonado/agones-mc/pkg/signal"
)

var (
	logger  *zap.SugaredLogger
	zLogger *zap.Logger
)

var monitorCmd = cobra.Command{
	Use:   "monitor",
	Short: "Agones minecraft server monitor",
	Long:  "Monitor process thats pings a minecraft server and reports statues to a local Agones SDK server",
	Run:   RunMonitor,
}

func init() {
	zLogger, _ = zap.NewProduction()
	logger = zLogger.Sugar()

	RootCmd.AddCommand(&monitorCmd)
}

// Main monitor function
func RunMonitor(cmd *cobra.Command, args []string) {
	defer logger.Desugar().Sync()
	stop := signal.SetupSignalHandler(logger)

	cfg := config.NewMonitorConfig()

	// Create new timed pinger
	pinger, err := ping.NewTimed(cfg.GetHost(), uint16(cfg.GetPort()), cfg.GetTimeout(), cfg.GetEdition())

	if err != nil {
		logger.Fatalw("Error creating ping client", "error", err)
	}

	// Startup delay before the first ping (initial-delay)
	logger.Info("Starting up...")
	time.Sleep(cfg.GetInitialDelay())

	// Ping server until startup
	err = pingUntilStartup(cfg.GetAttempts(), cfg.GetInterval(), pinger, stop)

	// Exit in case of unsuccessful startup
	if err != nil {
		if errors.Is(err, &ProcessStopped{}) {
			os.Exit(0)
		}
		logger.Fatalw("Fatal Mincraft server. Exiting...", "error", err)
	}

	// delay before next ping cycle
	time.Sleep(cfg.GetInterval())

	// Ping infinitely or until after a series of unsuccessful pings
	err = pingUntilFatal(cfg.GetAttempts(), cfg.GetInterval(), pinger, stop)

	// Exit in case of fatal server
	if err != nil {
		if errors.Is(err, &ProcessStopped{}) {
			os.Exit(0)
		}
		logger.Fatalw("Fatal Mincraft server. Exiting...", "error", err)
	}
}

// Pings server with the specified retries until the server returns a complete response
// Will also signal the local Agones server with Ready()
// Returns an error if the pings or singaling local Agones server fails
func pingUntilStartup(attempts int, interval time.Duration, pinger *ping.ServerPinger, stop chan bool) error {
	for {
		var err error
		if err = retryPing(attempts, interval, stop, pinger.ReadyPingWithTimeout); err == nil {
			break
		}

		if _, ok := err.(ping.StartingUpErr); ok {
			logger.Errorw("Server still starting...", "attemptsLeft", attempts, "errorMessage", err.Error())
		} else {
			return err
		}

		select {
		case <-stop:
			return &ProcessStopped{}
		case <-time.After(interval):
		}

	}

	logger.Info("Server ready")
	return nil
}

// Pings the server infinitely or the server fails to reposnd after a series of retries
// Signals to local SDK that server is healthy
func pingUntilFatal(attempts int, interval time.Duration, pinger *ping.ServerPinger, stop chan bool) error {
	for {
		err := retryPing(attempts, interval, stop, pinger.HealthPingWithTimeout)

		if err != nil {
			return err
		}

		logger.Info("Server healthy")
		select {
		case <-stop:
			return &ProcessStopped{}
		case <-time.After(interval):
		}
	}
}

// Retry wrapper function that will retry the given ping function with the specified attempts and intervals
// until it dosen't return an error or until all attempts have been made
func retryPing(attempts int, interval time.Duration, stop chan bool, p func() error) error {
	for {

		err := p()
		if err != nil {
			if attempts-1 > 0 {
				logger.Errorw(fmt.Sprintf("Unsuccessful ping. retrying in %v...", interval), "attemptsLeft", attempts-1, "errorMessage", err.Error())
				attempts--
			} else {
				return err
			}
		}

		if err == nil {
			return nil
		}

		select {
		case <-stop:
			return &ProcessStopped{}
		case <-time.After(interval):
		}

	}
}

type ProcessStopped struct{}

func (e *ProcessStopped) Error() string {
	return "process stopped"
}
