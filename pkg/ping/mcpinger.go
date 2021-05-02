package ping

import (
	"time"

	mcpinger "github.com/Raqbit/mc-pinger"
)

// Minecraft server pinger using mc-pinger
type McPinger struct {
	Port    uint16
	Host    string
	Timeout time.Duration
}

// Pings minecraft server and returns info obj. Returns error on failed ping
func (p *McPinger) Ping() (*ServerInfo, error) {
	res, err := mcpinger.New(p.Host, p.Port).Ping()
	if err != nil {
		return nil, err
	}
	return &ServerInfo{
		Protocol:      res.Version.Protocol,
		Version:       res.Version.Name,
		MaxPlayers:    res.Players.Max,
		OnlinePlayers: res.Players.Online,
	}, nil
}

// Pings minecraft server and return info obj. Return error on failed ping or timed out context
func (p *McPinger) PingWithTimeout() (*ServerInfo, error) {
	res, err := mcpinger.NewTimed(p.Host, p.Port, p.Timeout).Ping()
	if err != nil {
		return nil, err
	}
	return &ServerInfo{
		Protocol:      res.Version.Protocol,
		Version:       res.Version.Name,
		MaxPlayers:    res.Players.Max,
		OnlinePlayers: res.Players.Online,
	}, nil
}

// Checks if the current timeout duration is zero
func (p *McPinger) IsTimeoutZero() bool {
	return p.Timeout == 0
}
