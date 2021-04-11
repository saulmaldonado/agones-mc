package main

import (
	"flag"
	"log"
	"os"
	"time"

	sdk "agones.dev/agones/sdks/go"
	mcpinger "github.com/Raqbit/mc-pinger"
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
	// connect to local SDK server
	client, err := sdk.NewSDK()

	if err != nil {
		log.Fatal(err)
	}

	// Ping server until startup
	_, err = pingUntilSuccessful(retry)

	// Exit in case of unsuccessful startup
	if err != nil {
		logger.Println(err)
		logger.Fatal("Fatal Mincraft server. Exiting...")
	}

	// Signal that server is ready for players
	err = client.Ready()

	if err != nil {
		log.Fatal(err)
	}

	logger.Println("Server Ready")

	// Pause before starting next ping cycle
	time.Sleep(interval)

	// Ping infinitely or until after a series of unsuccessful pings
	err = pingUntilFatal(retry, client)

	// Exit in case of fatal server
	if err != nil {
		logger.Println(err)
		logger.Fatal("Fatal Mincraft server. Exiting...")
	}
}

// Creates a pinger with Context containing timeout
func ping() (*mcpinger.ServerInfo, error) {
	pinger := mcpinger.NewTimed(host, uint16(port), timeout)
	return pinger.Ping()
}

// Creates pings the server on an interval until the first sucessful ping which then will return the server info response
// After a series of failed retries an error will be returned indicating a fatal server
func pingUntilSuccessful(retry uint) (*mcpinger.ServerInfo, error) {
	info, err := ping()

	for err != nil && retry > 0 {
		logger.Println(err)
		logger.Printf("Unsuccessful ping. retrying in %v...", interval)
		time.Sleep(interval)

		info, err = ping()
		retry--
	}

	// Server will respond with an empty JSON object if its starting up
	// Pings will continue to happen until the server returns the expected response
	for info.Players.Max == 0 {
		logger.Println("Minecraft server starting up...")
		time.Sleep(interval)
		info, err = ping()
	}

	return info, err
}

// Pings the server infinitely or the server fails to reposnd after a series of retries
// Signals to local SDK that server is healthy
func pingUntilFatal(retry uint, client *sdk.SDK) error {
	for {
		_, err := pingUntilSuccessful(retry)

		if err != nil {
			return err
		}

		// Signal that server is healthy
		err = client.Health()

		if err != nil {
			logger.Fatal(err)
		}

		logger.Printf("Server healthy")
		time.Sleep(interval)
	}
}
