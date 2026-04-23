package collector

import (
	"context"
	"fmt"
	"log"
	"time"

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
	packs    []uint8
}

func NewService(provider ConfigProvider, runtimeState *state.Store) *Service {
	return &Service{provider: provider, state: runtimeState}
}

func (s *Service) Run(ctx context.Context) error {
	s.state.SetServiceStatus("pace", "starting", false, "", time.Time{})
	var client *pace.Client
	defer func() {
		if client != nil {
			_ = client.Close()
		}
	}()
	for {
		cfg := s.provider.GetConfig()
		if client == nil {
			opened, err := pace.Open(cfg)
			if err != nil {
				err = fmt.Errorf("open %s: %w", cfg.Serial.Port, err)
				s.state.SetServiceStatus("pace", "error", false, err.Error(), time.Time{})
				log.Printf("pace poll failed: %v", err)
				if !sleep(ctx, cfg.Polling.ReconnectDelay.Duration) {
					return nil
				}
				continue
			}
			client = opened
		}

		if err := s.pollOnce(ctx, client, cfg); err != nil {
			_ = client.Close()
			client = nil
			s.state.SetServiceStatus("pace", "error", false, err.Error(), time.Time{})
			log.Printf("pace poll failed: %v", err)
			if !sleep(ctx, cfg.Polling.ReconnectDelay.Duration) {
				return nil
			}
			continue
		}
		if !sleep(ctx, cfg.Polling.Interval.Duration) {
			return nil
		}
	}
}

func (s *Service) pollOnce(ctx context.Context, client *pace.Client, cfg config.Config) error {
	if len(s.packs) == 0 {
		packs := s.discoverPacks(ctx, client, cfg)
		if len(packs) > 0 {
			s.packs = packs
		}
	}
	if len(s.packs) == 0 {
		s.packs = []uint8{cfg.Device.FirstPackAddress}
	}

	var lastErr error
	success := 0
	for _, pack := range s.packs {
		select {
		case <-ctx.Done():
			return nil
		default:
		}
		packs, err := client.AnalogPacks(pack)
		if err != nil {
			lastErr = fmt.Errorf("pack %d analog: %w", pack, err)
			continue
		}
		for _, analog := range packs {
			s.state.UpsertPack(analog)
			success++
		}
	}
	if success == 0 && lastErr != nil {
		return lastErr
	}
	s.state.SetServiceStatus("pace", "connected", true, "", time.Now().UTC())
	return nil
}

func (s *Service) discoverPacks(ctx context.Context, client *pace.Client, cfg config.Config) []uint8 {
	packs := make([]uint8, 0)
	if cfg.Device.Protocol == string(pace.ProtocolRS232) {
		return []uint8{255}
	}
	if cfg.Device.Protocol == string(pace.ProtocolConsole) {
		got, err := client.PackNumber(cfg.Device.FirstPackAddress)
		if err != nil {
			return packs
		}
		for i := uint8(1); i <= got; i++ {
			packs = append(packs, i)
		}
		log.Printf("discovered PACE console pack count=%d", got)
		return packs
	}
	start := cfg.Device.FirstPackAddress
	end := start + cfg.Device.MaxParallelPacks - 1
	for pack := start; pack <= end; pack++ {
		select {
		case <-ctx.Done():
			return packs
		default:
		}
		got, err := client.PackNumber(pack)
		if err != nil {
			continue
		}
		if got == pack {
			packs = append(packs, pack)
			log.Printf("discovered PACE pack address=%d", pack)
		}
		if pack == 255 {
			break
		}
	}
	return packs
}

func sleep(ctx context.Context, duration time.Duration) bool {
	timer := time.NewTimer(duration)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}
