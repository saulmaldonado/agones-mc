package main

import (
	"flag"
	"log"
	"os"
	"time"

	ping "github.com/saulmaldonado/agones-mc-monitor/pkg"
)

var host string
var port uint
var retry uint
var interval time.Duration
var timeout time.Duration
var logger *log.Logger

func init() {
	flag.StringVar(&host, "host", "localhost", "Minecraft server host")
	flag.UintVar(&port, "port", 25565, "Minecraft server port")
	flag.DurationVar(&interval, "interval", time.Second*10, "Server ping interval")
	flag.UintVar(&retry, "retry", 5, "Ping retry attempt limit")
	flag.DurationVar(&timeout, "timeout", interval, "Ping timeout")

	flag.Parse()
	logger = log.New(os.Stdout, "[agones-mc-monitor] ", log.Ltime|log.Ldate)
}

func main() {
	pinger, err := ping.NewTimed(host, uint16(port), timeout)

	if err != nil {
		log.Fatal(err)
	}

	// Ping server until startup
	err = pingUntilStartup(retry, interval, pinger)

	// Exit in case of unsuccessful startup
	if err != nil {
		logger.Println(err)
		logger.Fatal("Fatal Mincraft server. Exiting...")
	}

	// Pause before starting next ping cycle
	time.Sleep(interval)

	// Ping infinitely or until after a series of unsuccessful pings
	err = pingUntilFatal(retry, interval, pinger)

	// Exit in case of fatal server
	if err != nil {
		logger.Println(err)
		logger.Fatal("Fatal Mincraft server. Exiting...")
	}
}

// Pings server with the specified retries until the server returns a complete response
// Will also signal the local Agones server with Ready()
// Returns an error if the pings or singaling local Agones server fails
func pingUntilStartup(retry uint, interval time.Duration, pinger *ping.AgonesPinger) error {
	err := retryPing(retry, interval, pinger.ReadyPingWithTimeout)

	if err != nil {
		return err
	}

	logger.Println("Server ready")
	return nil
}

// Pings the server infinitely or the server fails to reposnd after a series of retries
// Signals to local SDK that server is healthy
func pingUntilFatal(retry uint, interval time.Duration, pinger *ping.AgonesPinger) error {
	for {
		err := retryPing(retry, interval, pinger.HealthPingWithTimeout)

		if err != nil {
			return err
		}

		logger.Printf("Server healthy")
		time.Sleep(interval)
	}
}

// Retry wrapper function that will retry the given ping function with the specified attempts and intervals
// until it dosen't return an error or until all attempts have been made
func retryPing(attempts uint, interval time.Duration, p func() error) error {
	err := p()

	for err != nil && attempts > 0 {
		logger.Println(err)
		logger.Printf("Unsuccessful ping. retrying in %v...", interval)
		time.Sleep(interval)

		err = p()
		attempts--
	}

	return err
}
