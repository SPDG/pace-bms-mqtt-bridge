package state

import (
	"sort"
	"sync"
	"time"

	"github.com/SPDG/pace-bms-mqtt-bridge/internal/pace"
)

type ServiceStatus struct {
	Name        string    `json:"name"`
	Status      string    `json:"status"`
	Connected   bool      `json:"connected"`
	LastError   string    `json:"lastError,omitempty"`
	LastSuccess time.Time `json:"lastSuccess,omitempty"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type Snapshot struct {
	Services  map[string]ServiceStatus `json:"services"`
	Packs     []pace.Pack              `json:"packs"`
	Telemetry []pace.Telemetry         `json:"telemetry"`
}

type Store struct {
	mu        sync.RWMutex
	services  map[string]ServiceStatus
	packs     map[uint8]pace.Pack
	telemetry map[string]pace.Telemetry
}

func New() *Store {
	return &Store{
		services:  make(map[string]ServiceStatus),
		packs:     make(map[uint8]pace.Pack),
		telemetry: make(map[string]pace.Telemetry),
	}
}

func (s *Store) SetServiceStatus(name, status string, connected bool, lastError string, lastSuccess time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	existing := s.services[name]
	if lastSuccess.IsZero() {
		lastSuccess = existing.LastSuccess
	}
	s.services[name] = ServiceStatus{
		Name:        name,
		Status:      status,
		Connected:   connected,
		LastError:   lastError,
		LastSuccess: lastSuccess,
		UpdatedAt:   time.Now().UTC(),
	}
}

func (s *Store) UpsertPack(pack pace.Pack) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.packs[pack.Address] = pack
	for _, value := range pace.TelemetryForPack(pack) {
		s.telemetry[value.ID] = value
	}
}

func (s *Store) Snapshot() Snapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	services := make(map[string]ServiceStatus, len(s.services))
	for key, value := range s.services {
		services[key] = value
	}
	packs := make([]pace.Pack, 0, len(s.packs))
	for _, pack := range s.packs {
		packs = append(packs, pack)
	}
	sort.Slice(packs, func(i, j int) bool {
		return packs[i].Address < packs[j].Address
	})
	telemetry := make([]pace.Telemetry, 0, len(s.telemetry))
	for _, value := range s.telemetry {
		telemetry = append(telemetry, value)
	}
	telemetry = append(telemetry, pace.AggregateTelemetry(packs)...)
	sort.Slice(telemetry, func(i, j int) bool {
		return telemetry[i].ID < telemetry[j].ID
	})
	return Snapshot{
		Services:  services,
		Packs:     packs,
		Telemetry: telemetry,
	}
}
