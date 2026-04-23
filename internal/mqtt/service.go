package mqtt

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	paho "github.com/eclipse/paho.mqtt.golang"

	"github.com/SPDG/pace-bms-mqtt-bridge/internal/buildinfo"
	"github.com/SPDG/pace-bms-mqtt-bridge/internal/config"
	"github.com/SPDG/pace-bms-mqtt-bridge/internal/pace"
	"github.com/SPDG/pace-bms-mqtt-bridge/internal/state"
)

type ConfigProvider interface {
	GetConfig() config.Config
}

type Service struct {
	provider ConfigProvider
	state    *state.Store
	build    buildinfo.Info

	mu             sync.Mutex
	key            string
	client         paho.Client
	discovered     map[string]struct{}
	lastPublished  map[string]string
	unhealthySince time.Time
}

func NewService(provider ConfigProvider, runtimeState *state.Store, build buildinfo.Info) *Service {
	return &Service{
		provider:      provider,
		state:         runtimeState,
		build:         build,
		discovered:    make(map[string]struct{}),
		lastPublished: make(map[string]string),
	}
}

func (s *Service) Run(ctx context.Context) error {
	s.state.SetServiceStatus("mqtt", "starting", false, "", time.Time{})
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	s.sync()
	for {
		select {
		case <-ctx.Done():
			s.disconnect()
			return nil
		case <-ticker.C:
			s.sync()
		}
	}
}

func (s *Service) sync() {
	cfg := s.provider.GetConfig()
	if strings.TrimSpace(cfg.MQTT.Broker) == "" {
		s.disconnect()
		s.state.SetServiceStatus("mqtt", "disabled", false, "mqtt broker is empty", time.Time{})
		return
	}
	client := s.ensureClient(cfg)
	if client == nil {
		return
	}
	if !client.IsConnectionOpen() {
		s.state.SetServiceStatus("mqtt", "connecting", false, "", time.Time{})
		return
	}
	if err := s.publishAvailability(cfg, "online"); err != nil {
		s.recordError(err)
		return
	}
	if err := s.publishTelemetry(cfg); err != nil {
		s.recordError(err)
		return
	}
	s.markHealthy()
	s.state.SetServiceStatus("mqtt", "connected", true, "", time.Now().UTC())
}

func (s *Service) ensureClient(cfg config.Config) paho.Client {
	key := mqttKey(cfg)
	s.mu.Lock()
	if s.client != nil && s.key == key {
		client := s.client
		s.mu.Unlock()
		return client
	}
	old := s.client
	s.client = nil
	s.key = ""
	s.discovered = make(map[string]struct{})
	s.lastPublished = make(map[string]string)
	s.unhealthySince = time.Time{}
	s.mu.Unlock()
	if old != nil {
		old.Disconnect(250)
	}

	client := paho.NewClient(s.clientOptions(cfg))
	s.mu.Lock()
	s.client = client
	s.key = key
	s.unhealthySince = time.Now().UTC()
	s.mu.Unlock()
	token := client.Connect()
	if ok := token.WaitTimeout(5 * time.Second); !ok {
		s.state.SetServiceStatus("mqtt", "error", false, "mqtt connect timeout", time.Time{})
		s.resetClient(client)
		return nil
	}
	if err := token.Error(); err != nil {
		s.state.SetServiceStatus("mqtt", "error", false, err.Error(), time.Time{})
		s.resetClient(client)
		return nil
	}
	return client
}

func (s *Service) clientOptions(cfg config.Config) *paho.ClientOptions {
	opts := paho.NewClientOptions().
		AddBroker(cfg.MQTT.Broker).
		SetClientID(cfg.MQTT.ClientID).
		SetAutoReconnect(true).
		SetConnectRetry(false).
		SetConnectTimeout(3 * time.Second).
		SetOrderMatters(false).
		SetKeepAlive(30 * time.Second).
		SetPingTimeout(10 * time.Second).
		SetCleanSession(true)
	if cfg.MQTT.Username != "" {
		opts.SetUsername(cfg.MQTT.Username)
	}
	if cfg.MQTT.Password != "" {
		opts.SetPassword(cfg.MQTT.Password)
	}
	opts.OnConnectionLost = func(_ paho.Client, err error) {
		s.markUnhealthy(time.Now().UTC())
		s.state.SetServiceStatus("mqtt", "error", false, err.Error(), time.Time{})
	}
	return opts
}

func (s *Service) publishTelemetry(cfg config.Config) error {
	snapshot := s.state.Snapshot()
	for _, value := range snapshot.Telemetry {
		if err := s.ensureDiscovery(cfg, value); err != nil {
			return err
		}
		topic := stateTopic(cfg, value.ID)
		if last, ok := s.getLastPublished(topic); ok && last == value.Rendered {
			continue
		}
		if err := s.publish(cfg, topic, value.Rendered, cfg.MQTT.Retain); err != nil {
			return err
		}
		s.setLastPublished(topic, value.Rendered)
	}
	return nil
}

func (s *Service) ensureDiscovery(cfg config.Config, value pace.Telemetry) error {
	s.mu.Lock()
	_, ok := s.discovered[value.ID]
	s.mu.Unlock()
	if ok {
		return nil
	}
	deviceID := sanitizeID(cfg.Device.Name)
	payload := map[string]any{
		"name":               value.Name,
		"unique_id":          fmt.Sprintf("%s_%s", deviceID, value.ID),
		"state_topic":        stateTopic(cfg, value.ID),
		"availability_topic": availabilityTopic(cfg),
		"icon":               value.Icon,
		"device": map[string]any{
			"identifiers":  []string{deviceID},
			"name":         cfg.Device.Name,
			"manufacturer": cfg.Device.Manufacturer,
			"model":        cfg.Device.Model,
			"sw_version":   s.build.Version,
		},
	}
	if value.Unit != "" {
		payload["unit_of_measurement"] = value.Unit
	}
	if value.DeviceClass != "" {
		payload["device_class"] = value.DeviceClass
	}
	if value.StateClass != "" {
		payload["state_class"] = value.StateClass
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	if err := s.publish(cfg, discoveryTopic(cfg, deviceID, value.ID), string(body), true); err != nil {
		return err
	}
	s.mu.Lock()
	s.discovered[value.ID] = struct{}{}
	s.mu.Unlock()
	return nil
}

func (s *Service) publishAvailability(cfg config.Config, payload string) error {
	return s.publish(cfg, availabilityTopic(cfg), payload, true)
}

func (s *Service) publish(cfg config.Config, topic, payload string, retained bool) error {
	s.mu.Lock()
	client := s.client
	s.mu.Unlock()
	if client == nil {
		return fmt.Errorf("mqtt client is not initialized")
	}
	token := client.Publish(topic, 0, retained, payload)
	if ok := token.WaitTimeout(10 * time.Second); !ok {
		return fmt.Errorf("mqtt publish timeout for %s", topic)
	}
	return token.Error()
}

func (s *Service) disconnect() {
	cfg := s.provider.GetConfig()
	s.mu.Lock()
	client := s.client
	s.client = nil
	s.key = ""
	s.discovered = make(map[string]struct{})
	s.lastPublished = make(map[string]string)
	s.unhealthySince = time.Time{}
	s.mu.Unlock()
	if client != nil && client.IsConnectionOpen() {
		token := client.Publish(availabilityTopic(cfg), 0, true, "offline")
		token.WaitTimeout(3 * time.Second)
		client.Disconnect(250)
	}
}

func (s *Service) recordError(err error) {
	s.markUnhealthy(time.Now().UTC())
	s.state.SetServiceStatus("mqtt", "error", false, err.Error(), time.Time{})
	log.Printf("mqtt sync failed: %v", err)
}

func (s *Service) markHealthy() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.unhealthySince = time.Time{}
}

func (s *Service) markUnhealthy(now time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.unhealthySince.IsZero() {
		s.unhealthySince = now
	}
}

func (s *Service) resetClient(expected paho.Client) {
	s.mu.Lock()
	client := s.client
	if expected != nil && client != expected {
		s.mu.Unlock()
		expected.Disconnect(250)
		return
	}
	s.client = nil
	s.key = ""
	s.discovered = make(map[string]struct{})
	s.lastPublished = make(map[string]string)
	s.unhealthySince = time.Now().UTC()
	s.mu.Unlock()
	if client != nil {
		client.Disconnect(250)
	}
}

func mqttKey(cfg config.Config) string {
	return fmt.Sprintf("%s|%s|%s|%s|%s|%t",
		cfg.MQTT.Broker,
		cfg.MQTT.ClientID,
		cfg.MQTT.Username,
		cfg.MQTT.TopicPrefix,
		cfg.MQTT.DiscoveryPrefix,
		cfg.MQTT.Retain,
	)
}

func availabilityTopic(cfg config.Config) string {
	return fmt.Sprintf("%s/availability", strings.TrimSuffix(cfg.MQTT.TopicPrefix, "/"))
}

func stateTopic(cfg config.Config, entityID string) string {
	return fmt.Sprintf("%s/state/%s", strings.TrimSuffix(cfg.MQTT.TopicPrefix, "/"), entityID)
}

func discoveryTopic(cfg config.Config, deviceID, objectID string) string {
	return fmt.Sprintf("%s/sensor/%s/%s/config", cfg.MQTT.DiscoveryPrefix, deviceID, objectID)
}

func sanitizeID(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	var b strings.Builder
	lastUnderscore := false
	for _, r := range value {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			lastUnderscore = false
			continue
		}
		if !lastUnderscore {
			b.WriteByte('_')
			lastUnderscore = true
		}
	}
	return strings.Trim(b.String(), "_")
}

func (s *Service) getLastPublished(topic string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	value, ok := s.lastPublished[topic]
	return value, ok
}

func (s *Service) setLastPublished(topic, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lastPublished[topic] = value
}
