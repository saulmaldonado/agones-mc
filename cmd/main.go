package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	ping "github.com/saulmaldonado/agones-mc-monitor/pkg/ping"
	"go.uber.org/zap"
)

var host string
var port uint
var attempts uint
var interval time.Duration
var timeout time.Duration
var intialDelay time.Duration

var logger *zap.SugaredLogger
var zLogger *zap.Logger

func init() {
	flag.StringVar(&host, "host", "localhost", "Minecraft server host")
	flag.UintVar(&port, "port", 25565, "Minecraft server port")
	flag.DurationVar(&interval, "interval", time.Second*10, "Server ping interval")
	flag.UintVar(&attempts, "attempts", 5, "Ping attempt limit. Process will end after failing the last attempt")
	flag.DurationVar(&timeout, "timeout", interval, "Ping timeout")
	flag.DurationVar(&intialDelay, "initial-delay", time.Minute, "Initial startup delay before first ping")

	flag.Parse()
	zLogger, _ := zap.NewProduction()
	logger = zLogger.Sugar()
}

func main() {
	defer zLogger.Sync()
	stop := setupSignalHandler()

	// Create new timed pinger
	pinger, err := ping.NewTimed(host, uint16(port), timeout)

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
		logger.Fatalw("Fatal Mincraft server. Exiting...", "error", err)
	}

	// delay before next ping cycle
	time.Sleep(interval)

	// Ping infinitely or until after a series of unsuccessful pings
	err = pingUntilFatal(attempts, interval, pinger, stop)

	// Exit in case of fatal server
	if err != nil {
		logger.Fatalw("Fatal Mincraft server. Exiting...", "error", err)
	}
}

// Pings server with the specified retries until the server returns a complete response
// Will also signal the local Agones server with Ready()
// Returns an error if the pings or singaling local Agones server fails
func pingUntilStartup(attempts uint, interval time.Duration, pinger *ping.ServerPinger, stop chan struct{}) error {
	for {
		err := retryPing(attempts, interval, stop, pinger.ReadyPingWithTimeout)

		if err == nil {
			break
		}

		if _, ok := err.(ping.StartingUpErr); ok {
			logger.Errorw("Server still starting...", "attemptsLeft", attempts, "errorMessage", err.Error())
		} else {
			return err
		}

		time.Sleep(interval)
	}

	logger.Info("Server ready")
	return nil
}

// Pings the server infinitely or the server fails to reposnd after a series of retries
// Signals to local SDK that server is healthy
func pingUntilFatal(attempts uint, interval time.Duration, pinger *ping.ServerPinger, stop chan struct{}) error {
	for {
		err := retryPing(attempts, interval, stop, pinger.HealthPingWithTimeout)

		if err != nil {
			return err
		}

		logger.Info("Server healthy")
		time.Sleep(interval)
	}
}

// Retry wrapper function that will retry the given ping function with the specified attempts and intervals
// until it dosen't return an error or until all attempts have been made
func retryPing(attempts uint, interval time.Duration, stop chan struct{}, p func() error) error {
	for {

		if err := p(); err != nil && attempts-1 > 0 {
			logger.Errorw(fmt.Sprintf("Unsuccessful ping. retrying in %v...", interval), "attemptsLeft", attempts-1, "errorMessage", err.Error())
			attempts--
		} else {
			return err
		}

		select {
		case <-stop:
			return nil
		case <-time.After(interval):
		}

	}
}

func setupSignalHandler() chan struct{} {
	c := make(chan os.Signal, 1)
	stop := make(chan struct{})

	signal.Notify(c, syscall.SIGTERM)

	go func() {
		<-c
		logger.Info("Received SIGTERM. Terminating...")
		close(stop)
	}()

	return stop
}
