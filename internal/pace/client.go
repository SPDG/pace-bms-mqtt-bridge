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

func (c *Client) WarningPacks(pack uint8) ([]PackStatus, error) {
	response, err := c.Query(CommandWarningInfo, pack)
	if err != nil {
		return nil, err
	}
	return ParseWarningPacks(response, pack)
}

func (c *Client) CurrentLimitParameter(pack uint8) (CurrentLimitParameter, error) {
	response, err := c.Query(CommandCurrentLimit, pack)
	if err != nil {
		return CurrentLimitParameter{}, err
	}
	return ParseCurrentLimitParameter(response, pack)
}

func (c *Client) SoftwareVersion(pack uint8) (string, error) {
	response, err := c.Query(CommandSoftwareVersion, pack)
	if err != nil {
		return "", err
	}
	return ParseSoftwareVersion(response)
}

func (c *Client) PaceSoftwareVersion(pack uint8) (string, string, error) {
	response, err := c.Query(CommandPaceSoftwareVersion, pack)
	if err != nil {
		return "", "", err
	}
	return ParsePaceSoftwareVersion(response)
}

func (c *Client) ProductInfoDetails(pack uint8) (string, string, error) {
	response, err := c.Query(CommandProductInfo, pack)
	if err != nil {
		return "", "", err
	}
	return ParseProductInfoDetails(response)
}

func (c *Client) BMSDateTime(pack uint8) (string, error) {
	response, err := c.Query(CommandBMSDateTime, pack)
	if err != nil {
		return "", err
	}
	return ParseBMSDateTime(response)
}

func (c *Client) InverterProtocol(pack uint8) (InverterProtocolInfo, error) {
	response, err := c.Query(CommandInverterProtocol, pack)
	if err == nil {
		if info, parseErr := ParseInverterProtocol(response, "EB"); parseErr == nil {
			return info, nil
		}
	}
	responseV2, errV2 := c.Query(CommandInverterProtocolV2, pack)
	if errV2 != nil {
		if err != nil {
			return InverterProtocolInfo{}, err
		}
		return InverterProtocolInfo{}, errV2
	}
	return ParseInverterProtocol(responseV2, "62")
}

func (c *Client) EqualizationParameters(pack uint8) (EqualizationParameters, error) {
	response, err := c.Query(CommandEqualizationParams, pack)
	if err != nil {
		return EqualizationParameters{}, err
	}
	return ParseEqualizationParameters(response)
}

func (c *Client) FullChargeParameters(pack uint8) (FullChargeParameters, error) {
	response, err := c.Query(CommandFullChargeParams, pack)
	if err != nil {
		return FullChargeParameters{}, err
	}
	return ParseFullChargeParameters(response)
}

func (c *Client) SingleHibernateParameters(pack uint8) (SingleHibernateParameters, error) {
	response, err := c.Query(CommandSingleHibernateParams, pack)
	if err != nil {
		return SingleHibernateParameters{}, err
	}
	return ParseSingleHibernateParameters(response)
}

func (c *Client) CellOVPParameters(pack uint8) (VoltageProtectionParameters, error) {
	response, err := c.Query(CommandCellOVPParams, pack)
	if err != nil {
		return VoltageProtectionParameters{}, err
	}
	return ParseVoltageProtectionParameters(response)
}

func (c *Client) CellUVPParameters(pack uint8) (VoltageProtectionParameters, error) {
	response, err := c.Query(CommandCellUVPParams, pack)
	if err != nil {
		return VoltageProtectionParameters{}, err
	}
	return ParseVoltageProtectionParameters(response)
}

func (c *Client) PackOVPParameters(pack uint8) (PackVoltageProtectionParameters, error) {
	response, err := c.Query(CommandPackOVPParams, pack)
	if err != nil {
		return PackVoltageProtectionParameters{}, err
	}
	return ParsePackVoltageProtectionParameters(response)
}

func (c *Client) PackUVPParameters(pack uint8) (PackVoltageProtectionParameters, error) {
	response, err := c.Query(CommandPackUVPParams, pack)
	if err != nil {
		return PackVoltageProtectionParameters{}, err
	}
	return ParsePackVoltageProtectionParameters(response)
}

func (c *Client) ChargeOCPParameters(pack uint8) (CurrentProtectionParameters, error) {
	response, err := c.Query(CommandChargeOCPParams, pack)
	if err != nil {
		return CurrentProtectionParameters{}, err
	}
	return ParseCurrentProtectionParameters(response, false)
}

func (c *Client) DischargeOCP1Parameters(pack uint8) (CurrentProtectionParameters, error) {
	response, err := c.Query(CommandDischargeOCP1Params, pack)
	if err != nil {
		return CurrentProtectionParameters{}, err
	}
	return ParseCurrentProtectionParameters(response, true)
}

func (c *Client) DischargeOCP2Parameters(pack uint8) (DischargeOCP2Parameters, error) {
	response, err := c.Query(CommandDischargeOCP2Params, pack)
	if err != nil {
		return DischargeOCP2Parameters{}, err
	}
	return ParseDischargeOCP2Parameters(response)
}

func (c *Client) ShortCircuitDelay(pack uint8) (int, error) {
	response, err := c.Query(CommandShortCircuitDelay, pack)
	if err != nil {
		return 0, err
	}
	return ParseShortCircuitDelay(response)
}

func (c *Client) HighTempParameters(pack uint8) (TemperatureProtectionParameters, TemperatureProtectionParameters, error) {
	response, err := c.Query(CommandHighTempParams, pack)
	if err != nil {
		return TemperatureProtectionParameters{}, TemperatureProtectionParameters{}, err
	}
	return ParseDualTemperatureProtectionParameters(response)
}

func (c *Client) LowTempParameters(pack uint8) (TemperatureProtectionParameters, TemperatureProtectionParameters, error) {
	response, err := c.Query(CommandLowTempParams, pack)
	if err != nil {
		return TemperatureProtectionParameters{}, TemperatureProtectionParameters{}, err
	}
	return ParseDualTemperatureProtectionParameters(response)
}

func (c *Client) MOSHighTempParameters(pack uint8) (TemperatureProtectionParameters, error) {
	response, err := c.Query(CommandMOSHighTempParams, pack)
	if err != nil {
		return TemperatureProtectionParameters{}, err
	}
	return ParseSingleTemperatureProtectionParameters(response)
}

func (c *Client) EnvTempParameters(pack uint8) (TemperatureProtectionParameters, TemperatureProtectionParameters, error) {
	response, err := c.Query(CommandEnvTempParams, pack)
	if err != nil {
		return TemperatureProtectionParameters{}, TemperatureProtectionParameters{}, err
	}
	return ParseDualTemperatureProtectionParameters(response)
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
