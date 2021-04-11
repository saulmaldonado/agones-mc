package main

import (
	"flag"
	"log"
	"os"
	"time"

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
	var pinger mcpinger.Pinger

	for {
		pinger = mcpinger.NewTimed(host, uint16(port), timeout)
		info, err := pinger.Ping()

		if err != nil {
			retry := retry

			for retry > 0 {
				logger.Println(err)
				logger.Printf("Unsuccessful ping. retrying in %v...", interval)
				time.Sleep(interval)

				pinger = mcpinger.NewTimed(host, uint16(port), timeout)
				info, err = pinger.Ping()
				if err != nil {
					retry--
				} else {
					break
				}
			}

			if err != nil {
				logger.Println(err)
				logger.Fatalf("Pings unsuccessful. Exiting...")
			}
		}

		logger.Println(info)
		time.Sleep(interval)
	}
}
