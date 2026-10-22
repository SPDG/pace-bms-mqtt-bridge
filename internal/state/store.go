package state

import (
	"fmt"
	"sort"
	"strings"
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
	s.upsertPackLocked(pack)
}

func (s *Store) ReplacePacks(packs []pace.Pack) {
	s.mu.Lock()
	defer s.mu.Unlock()
	visible := make(map[uint8]struct{}, len(packs))
	for _, pack := range packs {
		visible[pack.Address] = struct{}{}
	}
	for address := range s.packs {
		if _, ok := visible[address]; ok {
			continue
		}
		delete(s.packs, address)
		s.deletePackTelemetryLocked(address)
	}
	for _, pack := range packs {
		s.upsertPackLocked(pack)
	}
}

func (s *Store) UpsertPackStatus(status pace.PackStatus) {
	s.mu.Lock()
	defer s.mu.Unlock()
	pack, ok := s.packs[status.Address]
	if !ok {
		return
	}
	pack.Address = status.Address
	pack.Status = &status
	s.packs[status.Address] = pack
	for _, value := range pace.TelemetryForPack(pack) {
		s.telemetry[value.ID] = value
	}
}

func (s *Store) UpsertPackInfo(info pace.PackInfo) {
	s.mu.Lock()
	defer s.mu.Unlock()
	pack, ok := s.packs[info.Address]
	if !ok {
		return
	}
	if pack.Info != nil {
		merged := *pack.Info
		if info.SoftwareVersion != "" {
			merged.SoftwareVersion = info.SoftwareVersion
		}
		if info.PaceSoftwareVersion != "" {
			merged.PaceSoftwareVersion = info.PaceSoftwareVersion
		}
		if info.HardwareVersion != "" {
			merged.HardwareVersion = info.HardwareVersion
		}
		if info.BMSSerialNumber != "" {
			merged.BMSSerialNumber = info.BMSSerialNumber
		}
		if info.PackSerialNumber != "" {
			merged.PackSerialNumber = info.PackSerialNumber
		}
		if info.BMSDateTime != "" {
			merged.BMSDateTime = info.BMSDateTime
		}
		if info.InverterProtocol != nil {
			merged.InverterProtocol = info.InverterProtocol
		}
		merged.UpdatedAt = info.UpdatedAt
		pack.Info = &merged
	} else {
		pack.Info = &info
	}
	s.packs[info.Address] = pack
	for _, value := range pace.TelemetryForPack(pack) {
		s.telemetry[value.ID] = value
	}
}

func (s *Store) UpsertPackParameters(params pace.PackParameters) {
	s.mu.Lock()
	defer s.mu.Unlock()
	pack, ok := s.packs[params.Address]
	if !ok {
		return
	}
	if pack.Parameters != nil && params.CurrentLimitA == nil {
		params.CurrentLimitA = pack.Parameters.CurrentLimitA
	}
	pack.Parameters = &params
	s.packs[params.Address] = pack
	for _, value := range pace.TelemetryForPack(pack) {
		s.telemetry[value.ID] = value
	}
}

func (s *Store) UpsertCurrentLimitParameter(value pace.CurrentLimitParameter) {
	s.mu.Lock()
	defer s.mu.Unlock()
	telemetry := pace.TelemetryForCurrentLimitParameter(value)
	s.telemetry[telemetry.ID] = telemetry
	if pack, ok := s.packs[value.Address]; ok {
		if pack.Parameters == nil {
			pack.Parameters = &pace.PackParameters{Address: value.Address, UpdatedAt: value.UpdatedAt}
		}
		limit := value.CurrentA
		pack.Parameters.CurrentLimitA = &limit
		s.packs[value.Address] = pack
	}
}

func (s *Store) upsertPackLocked(pack pace.Pack) {
	if existing, ok := s.packs[pack.Address]; ok {
		if pack.Status == nil {
			pack.Status = existing.Status
		}
		if pack.Info == nil {
			pack.Info = existing.Info
		}
		if pack.Parameters == nil {
			pack.Parameters = existing.Parameters
		}
	}
	s.packs[pack.Address] = pack
	for _, value := range pace.TelemetryForPack(pack) {
		s.telemetry[value.ID] = value
	}
}

func (s *Store) deletePackTelemetryLocked(address uint8) {
	prefix := fmt.Sprintf("pack_%02d_", address)
	for id := range s.telemetry {
		if strings.HasPrefix(id, prefix) {
			delete(s.telemetry, id)
		}
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
