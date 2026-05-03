package pace

import (
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	"github.com/goburrow/serial"

	"github.com/SPDG/pace-bms-mqtt-bridge/internal/config"
)

type Client struct {
	port     io.ReadWriteCloser
	protocol Protocol
	timeout  time.Duration
}

func Open(cfg config.Config) (*Client, error) {
	port, err := openTransport(cfg)
	if err != nil {
		return nil, err
	}
	return &Client{
		port:     port,
		protocol: Protocol(cfg.Device.Protocol),
		timeout:  cfg.Serial.Timeout.Duration,
	}, nil
}

func (c *Client) Close() error {
	if c.port == nil {
		return nil
	}
	return c.port.Close()
}

func (c *Client) Query(command Command, pack uint8) ([]byte, error) {
	request, err := BuildRequest(Request{Protocol: c.protocol, Command: command, Pack: pack})
	if err != nil {
		return nil, err
	}
	if conn, ok := c.port.(interface{ SetDeadline(time.Time) error }); ok {
		_ = conn.SetDeadline(time.Now().Add(c.timeout))
	}
	if _, err := c.port.Write(request); err != nil {
		return nil, fmt.Errorf("write request: %w", err)
	}
	response, err := ReadFrame(c.port, c.timeout)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}
	return response, nil
}

func (c *Client) PackNumber(pack uint8) (uint8, error) {
	response, err := c.Query(CommandPackNumber, pack)
	if err != nil {
		return 0, err
	}
	return ParsePackNumber(response)
}

func (c *Client) Analog(pack uint8) (Pack, error) {
	packs, err := c.AnalogPacks(pack)
	if err != nil {
		return Pack{}, err
	}
	if len(packs) == 0 {
		return Pack{}, fmt.Errorf("analog response contains no packs")
	}
	return packs[0], nil
}

func (c *Client) AnalogPacks(pack uint8) ([]Pack, error) {
	response, err := c.Query(CommandAnalog, pack)
	if err != nil {
		return nil, err
	}
	return ParseAnalogPacks(response, pack)
}

func openTransport(cfg config.Config) (io.ReadWriteCloser, error) {
	if address, ok := tcpAddress(cfg.Serial.Port); ok {
		timeout := cfg.Serial.Timeout.Duration
		if timeout <= 0 {
			timeout = 3 * time.Second
		}
		return net.DialTimeout("tcp", address, timeout)
	}
	return serial.Open(&serial.Config{
		Address:  cfg.Serial.Port,
		BaudRate: cfg.Serial.BaudRate,
		DataBits: cfg.Serial.DataBits,
		Parity:   cfg.Serial.Parity,
		StopBits: cfg.Serial.StopBits,
		Timeout:  cfg.Serial.Timeout.Duration,
	})
}

func tcpAddress(port string) (string, bool) {
	port = strings.TrimSpace(port)
	switch {
	case strings.HasPrefix(port, "tcp://"):
		return strings.TrimPrefix(port, "tcp://"), true
	case strings.HasPrefix(port, "tcp:"):
		return strings.TrimPrefix(port, "tcp:"), true
	default:
		return "", false
	}
}
