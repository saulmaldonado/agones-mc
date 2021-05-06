package cmd

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/saulmaldonado/agones-mc/pkg/ping"
	"github.com/saulmaldonado/agones-mc/pkg/signal"
)

var (
	host        string
	port        uint
	attempts    uint
	interval    time.Duration
	timeout     time.Duration
	intialDelay time.Duration
	edition     string
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
	monitorCmd.PersistentFlags().StringVar(&edition, "edition", "java", "Minecraft server edition. java or bedrock")
	monitorCmd.PersistentFlags().StringVar(&host, "host", "localhost", "Minecraft server host")
	monitorCmd.PersistentFlags().UintVar(&port, "port", 25565, "Minecraft server port")
	monitorCmd.PersistentFlags().DurationVar(&interval, "interval", time.Second*10, "Server ping interval")
	monitorCmd.PersistentFlags().UintVar(&attempts, "attempts", 5, "Ping attempt limit. Process will end after failing the last attempt")
	monitorCmd.PersistentFlags().DurationVar(&timeout, "timeout", interval, "Ping timeout")
	monitorCmd.PersistentFlags().DurationVar(&intialDelay, "initial-delay", time.Minute, "Initial startup delay before first ping")

	zLogger, _ = zap.NewProduction()
	logger = zLogger.Sugar()

	RootCmd.AddCommand(&monitorCmd)
}

// Main monitor function
func RunMonitor(cmd *cobra.Command, args []string) {
	defer logger.Desugar().Sync()
	stop := signal.SetupSignalHandler(logger)

	// Create new timed pinger
	pinger, err := ping.NewTimed(host, uint16(port), timeout, edition)

	if err != nil {
		logger.Fatalw("Error creating ping client", "error", err)
	}

	// Startup delay before the first ping (initial-delay)
	logger.Info("Starting up...")
	time.Sleep(intialDelay)

	// Ping server until startup
	err = pingUntilStartup(attempts, interval, pinger, stop)

	// Exit in case of unsuccessful startup
	if err != nil {
		if errors.Is(err, &ProcessStopped{}) {
			os.Exit(0)
		}
		logger.Fatalw("Fatal Mincraft server. Exiting...", "error", err)
	}

	// delay before next ping cycle
	time.Sleep(interval)

	// Ping infinitely or until after a series of unsuccessful pings
	err = pingUntilFatal(attempts, interval, pinger, stop)

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
func pingUntilStartup(attempts uint, interval time.Duration, pinger *ping.ServerPinger, stop chan bool) error {
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
func pingUntilFatal(attempts uint, interval time.Duration, pinger *ping.ServerPinger, stop chan bool) error {
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
func retryPing(attempts uint, interval time.Duration, stop chan bool, p func() error) error {
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
