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
func (p *McPinger) Ping() (*mcpinger.ServerInfo, error) {
	return mcpinger.New(p.Host, p.Port).Ping()
}

// Pings minecraft server and return info obj. Return error on failed ping or timed out context
func (p *McPinger) PingWithTimeout() (*mcpinger.ServerInfo, error) {
	return mcpinger.NewTimed(p.Host, p.Port, p.Timeout).Ping()
}

// Checks if the current timeout duration is zero
func (p *McPinger) IsTimeoutZero() bool {
	return p.Timeout == 0
}
