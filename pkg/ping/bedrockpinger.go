package ping

import (
	"net"
	"strconv"
	"time"

	"github.com/ZeroErrors/go-bedrockping"
)

// Minecraft server pinger using go-bedrockping
type BedrockPinger struct {
	Port    uint16
	Host    string
	Timeout time.Duration
}

const (
	// Default ping timeout
	DefaultTimeout = time.Second * 10
)

// Pings the bedrock minecraft server and returns server info. Returns error on failed ping
func (p *BedrockPinger) Ping() (*ServerInfo, error) {
	res, err := bedrockping.Query(net.JoinHostPort(p.Host, strconv.Itoa(int(p.Port))), DefaultTimeout, time.Second)
	if err != nil {
		return nil, err
	}
	return &ServerInfo{
		Protocol:      int32(res.ProtocolVersion),
		Version:       res.MCPEVersion,
		MaxPlayers:    int32(res.MaxPlayers),
		OnlinePlayers: int32(res.PlayerCount),
	}, nil
}

// Pings the bedrock minecraft server and returns server info. Returns error on failed ping or timeout
func (p *BedrockPinger) PingWithTimeout() (*ServerInfo, error) {
	res, err := bedrockping.Query(net.JoinHostPort(p.Host, strconv.Itoa(int(p.Port))), p.Timeout, time.Second)
	if err != nil {
		return nil, err
	}
	return &ServerInfo{
		Protocol:      int32(res.ProtocolVersion),
		Version:       res.MCPEVersion,
		MaxPlayers:    int32(res.MaxPlayers),
		OnlinePlayers: int32(res.PlayerCount),
	}, nil
}

// Checks if the current timeout duration is zero
func (p *BedrockPinger) IsTimeoutZero() bool {
	return p.Timeout == 0
}
