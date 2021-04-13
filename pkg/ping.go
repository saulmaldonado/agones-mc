package ping

import (
	"errors"
	"time"

	sdk "agones.dev/agones/sdks/go"
	mcpinger "github.com/Raqbit/mc-pinger"
)

type AgonesPinger struct {
	host    string
	port    uint16
	timeout time.Duration
	Sdk     *sdk.SDK
}

// Creates a new AgonesPinger with that will ping the minecraft server at the given host and on the given port.
// Also initializes a connection with the local Agones server on localhost port 9357.
// Blocks until connection and handshake is made. Timesout and returns an error after 30 seconds
func New(host string, port uint16) (*AgonesPinger, error) {
	sdk, err := sdk.NewSDK()

	if err != nil {
		return nil, err
	}

	return &AgonesPinger{host: host, port: port, Sdk: sdk}, nil
}

// Creates a new AgonesPinger with that will ping the minecraft server at the given host and on the given port.
// Ping will timeout after the give timeout duration.
// Also initializes a connection with the local Agones server on localhost port 9357.
// Blocks until connection and handshake is made. Timesout and returns an error after 30 seconds
func NewTimed(host string, port uint16, timeout time.Duration) (*AgonesPinger, error) {
	sdk, err := sdk.NewSDK()

	if err != nil {
		return nil, err
	}

	return &AgonesPinger{host, port, timeout, sdk}, nil
}

// Pings the minecraft server and sends Health() signal to the local Agones server on localhost port 9357
// Returns an error if the ping is unsuccessful
func (p *AgonesPinger) HealthPing() error {
	_, err := p.ping()

	if err != nil {
		return err
	}

	return p.Sdk.Health()
}

// Pings the minecraft server and sends Ready() signal to the local Agones server on localhost port 9357
// Returns an error if the ping is unsuccessful
func (p *AgonesPinger) ReadyPing() error {
	info, err := p.ping()

	if err != nil {
		return err
	}

	if info.Players.Max == 0 {
		return StartingUpErr{}
	}

	return p.Sdk.Ready()
}

// Pings the minecraft server and sends Health() signal to the local Agones server on localhost port 9357
// Returns an error if the ping is unsuccessful or timeouts
func (p *AgonesPinger) HealthPingWithTimeout() error {
	if p.timeout == 0 {
		return errors.New("ping timeout is set to 0s")
	}

	_, err := p.pingWithTimeout()

	if err != nil {
		return err
	}

	err = p.Sdk.Health()
	return err
}

// Pings the minecraft server and sends Ready() signal to the local Agones server on localhost port 9357
// Returns an error if the ping is unsuccessful or timeouts
func (p *AgonesPinger) ReadyPingWithTimeout() error {
	if p.timeout == 0 {
		return errors.New("ping timeout is set to 0s")
	}

	info, err := p.pingWithTimeout()

	if err != nil {
		return err
	}

	if info.Players.Max == 0 {
		return StartingUpErr{}
	}

	return p.Sdk.Ready()
}

func (p *AgonesPinger) ping() (*mcpinger.ServerInfo, error) {
	return mcpinger.New(p.host, p.port).Ping()
}

func (p *AgonesPinger) pingWithTimeout() (*mcpinger.ServerInfo, error) {
	return mcpinger.NewTimed(p.host, p.port, p.timeout).Ping()
}

type StartingUpErr struct{}

func (e StartingUpErr) Error() string {
	return "server starting up..."
}
