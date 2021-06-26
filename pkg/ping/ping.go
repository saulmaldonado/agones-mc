package ping

import (
	"errors"
	"strings"
	"time"

	sdk "agones.dev/agones/sdks/go"

	"github.com/saulmaldonado/agones-mc/internal/config"
)

// Minecraft server pinger and SDK state manager
type ServerPinger struct {
	host   string
	port   uint16
	sdk    *sdk.SDK
	pinger Pinger
}

type ServerInfo struct {
	Protocol      int32
	Version       string
	MaxPlayers    int32
	OnlinePlayers int32
}

// Interface for pinger implementation
type Pinger interface {
	Ping() (*ServerInfo, error)
	PingWithTimeout() (*ServerInfo, error)
	IsTimeoutZero() bool
}

const (
	JavaEdition    string = "java"
	BedrockEdition string = "bedrock"
)

// Creates a new AgonesPinger with that will ping the minecraft server at the given host and on the given port.
// Also initializes a connection with the local Agones server on localhost port 9357.
// Blocks until connection and handshake is made. Timesout and returns an error after 30 seconds
func New(host string, port uint16, edition string) (*ServerPinger, error) {
	sdk, err := sdk.NewSDK()

	if err != nil {
		return nil, err
	}

	if strings.ToLower(edition) == "bedrock" {
		return &ServerPinger{host, port, sdk, &BedrockPinger{Port: port, Host: host, Timeout: 0}}, nil
	}
	return &ServerPinger{host, port, sdk, &McPinger{Port: port, Host: host, Timeout: 0}}, nil
}

// Creates a new AgonesPinger with that will ping the minecraft server at the given host and on the given port.
// Ping will timeout after the give timeout duration.
// Also initializes a connection with the local Agones server on localhost port 9357.
// Blocks until connection and handshake is made. Timesout and returns an error after 30 seconds
func NewTimed(host string, port uint16, timeout time.Duration, edition config.Edition) (*ServerPinger, error) {
	sdk, err := sdk.NewSDK()

	if err != nil {
		return nil, err
	}

	if strings.ToLower(string(edition)) == "bedrock" {
		return &ServerPinger{host, port, sdk, &BedrockPinger{Port: port, Host: host, Timeout: timeout}}, nil
	}
	return &ServerPinger{host, port, sdk, &McPinger{Port: port, Host: host, Timeout: timeout}}, nil
}

// Pings the minecraft server and sends Health() signal to the local Agones server on localhost port 9357
// Returns an error if the ping is unsuccessful
func (p *ServerPinger) HealthPing() error {
	_, err := p.pinger.Ping()

	if err != nil {
		return err
	}

	return p.sdk.Health()
}

// Pings the minecraft server and sends Ready() signal to the local Agones server on localhost port 9357
// Returns an error if the ping is unsuccessful
func (p *ServerPinger) ReadyPing() error {
	info, err := p.pinger.Ping()

	if err != nil {
		return err
	}

	if info.MaxPlayers == 0 {
		return StartingUpErr{}
	}

	return p.sdk.Ready()
}

// Pings the minecraft server and sends Health() signal to the local Agones server on localhost port 9357
// Returns an error if the ping is unsuccessful or timeouts
func (p *ServerPinger) HealthPingWithTimeout() error {
	if p.pinger.IsTimeoutZero() {
		return errors.New("ping timeout is set to 0s")
	}

	_, err := p.pinger.PingWithTimeout()

	if err != nil {
		return err
	}

	err = p.sdk.Health()
	return err
}

// Pings the minecraft server and sends Ready() signal to the local Agones server on localhost port 9357
// Returns an error if the ping is unsuccessful or timeouts
func (p *ServerPinger) ReadyPingWithTimeout() error {
	if p.pinger.IsTimeoutZero() {
		return errors.New("ping timeout is set to 0s")
	}

	info, err := p.pinger.PingWithTimeout()

	if err != nil {
		return err
	}

	if info.MaxPlayers == 0 {
		return StartingUpErr{}
	}

	return p.sdk.Ready()
}

// Custom Error for failed pings due to server startup
type StartingUpErr struct{}

func (e StartingUpErr) Error() string {
	return "server starting up..."
}
