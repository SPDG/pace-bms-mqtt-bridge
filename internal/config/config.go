package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type Duration struct {
	time.Duration
}

func (d *Duration) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind != yaml.ScalarNode {
		return fmt.Errorf("duration must be a scalar value")
	}
	parsed, err := time.ParseDuration(strings.TrimSpace(value.Value))
	if err != nil {
		return fmt.Errorf("invalid duration %q: %w", value.Value, err)
	}
	d.Duration = parsed
	return nil
}

func (d Duration) MarshalYAML() (any, error) {
	return d.String(), nil
}

func (d *Duration) UnmarshalJSON(data []byte) error {
	var raw string
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("duration must be a JSON string: %w", err)
	}
	parsed, err := time.ParseDuration(strings.TrimSpace(raw))
	if err != nil {
		return fmt.Errorf("invalid duration %q: %w", raw, err)
	}
	d.Duration = parsed
	return nil
}

func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

type Config struct {
	Device  DeviceConfig  `yaml:"device" json:"device"`
	Serial  SerialConfig  `yaml:"serial" json:"serial"`
	Polling PollingConfig `yaml:"polling" json:"polling"`
	MQTT    MQTTConfig    `yaml:"mqtt" json:"mqtt"`
	HTTP    HTTPConfig    `yaml:"http" json:"http"`
	Logging LoggingConfig `yaml:"logging" json:"logging"`
}

type DeviceConfig struct {
	Name              string `yaml:"name" json:"name"`
	Protocol          string `yaml:"protocol" json:"protocol"`
	FirstPackAddress  uint8  `yaml:"first_pack_address" json:"firstPackAddress"`
	MaxParallelPacks  uint8  `yaml:"max_parallel_packs" json:"maxParallelPacks"`
	Manufacturer      string `yaml:"manufacturer" json:"manufacturer"`
	Model             string `yaml:"model" json:"model"`
	DiscoverOnStartup bool   `yaml:"discover_on_startup" json:"discoverOnStartup"`
}

type SerialConfig struct {
	Port     string   `yaml:"port" json:"port"`
	BaudRate int      `yaml:"baud_rate" json:"baudRate"`
	DataBits int      `yaml:"data_bits" json:"dataBits"`
	Parity   string   `yaml:"parity" json:"parity"`
	StopBits int      `yaml:"stop_bits" json:"stopBits"`
	Timeout  Duration `yaml:"timeout" json:"timeout"`
}

type PollingConfig struct {
	Interval       Duration `yaml:"interval" json:"interval"`
	ReconnectDelay Duration `yaml:"reconnect_delay" json:"reconnectDelay"`
}

type MQTTConfig struct {
	Broker          string `yaml:"broker" json:"broker"`
	Username        string `yaml:"username" json:"username"`
	Password        string `yaml:"password" json:"password"`
	ClientID        string `yaml:"client_id" json:"clientId"`
	TopicPrefix     string `yaml:"topic_prefix" json:"topicPrefix"`
	DiscoveryPrefix string `yaml:"discovery_prefix" json:"discoveryPrefix"`
	Retain          bool   `yaml:"retain" json:"retain"`
}

type HTTPConfig struct {
	Listen string `yaml:"listen" json:"listen"`
}

type LoggingConfig struct {
	Level string `yaml:"level" json:"level"`
}

func Default() Config {
	return Config{
		Device: DeviceConfig{
			Name:              "pace-main",
			Protocol:          "rs485",
			FirstPackAddress:  0,
			MaxParallelPacks:  16,
			Manufacturer:      "PACE",
			Model:             "PACE BMS",
			DiscoverOnStartup: true,
		},
		Serial: SerialConfig{
			Port:     "/dev/pace-bms",
			BaudRate: 115200,
			DataBits: 8,
			Parity:   "N",
			StopBits: 1,
			Timeout:  Duration{Duration: 3 * time.Second},
		},
		Polling: PollingConfig{
			Interval:       Duration{Duration: 15 * time.Second},
			ReconnectDelay: Duration{Duration: 5 * time.Second},
		},
		MQTT: MQTTConfig{
			Broker:          "tcp://127.0.0.1:1883",
			ClientID:        "pace-bms-hive-nuc",
			TopicPrefix:     "pace/pace-main",
			DiscoveryPrefix: "homeassistant",
			Retain:          true,
		},
		HTTP: HTTPConfig{
			Listen: "127.0.0.1:8080",
		},
		Logging: LoggingConfig{
			Level: "info",
		},
	}
}

func Load(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}
	cfg := Default()
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse config: %w", err)
	}
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func LoadOrCreate(path string) (Config, bool, error) {
	cfg, err := Load(path)
	if err == nil {
		return cfg, false, nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return Config{}, false, err
	}
	cfg = Default()
	if err := Save(path, cfg); err != nil {
		return Config{}, false, err
	}
	return cfg, true, nil
}

func Save(path string, cfg Config) error {
	if err := cfg.Validate(); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}
	data, err := yaml.Marshal(&cfg)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0o600); err != nil {
		return fmt.Errorf("write temp config: %w", err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("replace config: %w", err)
	}
	return nil
}

func (c Config) Validate() error {
	if strings.TrimSpace(c.Device.Name) == "" {
		return errors.New("device.name is required")
	}
	switch c.Device.Protocol {
	case "rs485", "rs232", "console":
	default:
		return errors.New("device.protocol must be rs485, rs232 or console")
	}
	if c.Device.MaxParallelPacks == 0 {
		return errors.New("device.max_parallel_packs must be greater than 0")
	}
	if strings.TrimSpace(c.Serial.Port) == "" {
		return errors.New("serial.port is required")
	}
	if c.Serial.BaudRate <= 0 {
		return errors.New("serial.baud_rate must be greater than 0")
	}
	if c.Serial.DataBits < 5 || c.Serial.DataBits > 8 {
		return errors.New("serial.data_bits must be between 5 and 8")
	}
	if c.Serial.StopBits < 1 || c.Serial.StopBits > 2 {
		return errors.New("serial.stop_bits must be 1 or 2")
	}
	if c.Serial.Timeout.Duration <= 0 {
		return errors.New("serial.timeout must be greater than 0")
	}
	if c.Polling.Interval.Duration <= 0 {
		return errors.New("polling.interval must be greater than 0")
	}
	if c.Polling.ReconnectDelay.Duration <= 0 {
		return errors.New("polling.reconnect_delay must be greater than 0")
	}
	if strings.TrimSpace(c.MQTT.TopicPrefix) == "" {
		return errors.New("mqtt.topic_prefix is required")
	}
	if strings.TrimSpace(c.MQTT.ClientID) == "" {
		return errors.New("mqtt.client_id is required")
	}
	if strings.TrimSpace(c.MQTT.DiscoveryPrefix) == "" {
		return errors.New("mqtt.discovery_prefix is required")
	}
	if strings.TrimSpace(c.HTTP.Listen) == "" {
		return errors.New("http.listen is required")
	}
	return nil
}
