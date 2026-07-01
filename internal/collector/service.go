package collector

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/SPDG/pace-bms-mqtt-bridge/internal/config"
	"github.com/SPDG/pace-bms-mqtt-bridge/internal/pace"
	"github.com/SPDG/pace-bms-mqtt-bridge/internal/state"
)

type ConfigProvider interface {
	GetConfig() config.Config
}

type Service struct {
	provider                 ConfigProvider
	state                    *state.Store
	packs                    []uint8
	lastStatusPoll           time.Time
	currentLimitPollDisabled bool
	lastStatusSignature      map[uint8]string
}

func NewService(provider ConfigProvider, runtimeState *state.Store) *Service {
	return &Service{
		provider:            provider,
		state:               runtimeState,
		lastStatusSignature: make(map[uint8]string),
	}
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
	statusDue := s.lastStatusPoll.IsZero() || time.Since(s.lastStatusPoll) >= cfg.Polling.StatusInterval.Duration
	statusAttempted := false
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
		if !statusDue {
			continue
		}
		statusAttempted = true
		statuses, err := client.WarningPacks(pack)
		if err != nil {
			lastErr = fmt.Errorf("pack %d warning: %w", pack, err)
			log.Printf("pace warning poll failed: %v", lastErr)
			continue
		}
		for _, status := range statuses {
			s.logStatusChange(status)
			s.state.UpsertPackStatus(status)
		}
		s.pollCurrentLimitParameter(client, pack)
	}
	if statusAttempted {
		s.lastStatusPoll = time.Now()
	}
	if success == 0 && lastErr != nil {
		return lastErr
	}
	s.state.SetServiceStatus("pace", "connected", true, "", time.Now().UTC())
	return nil
}

func (s *Service) pollCurrentLimitParameter(client *pace.Client, pack uint8) {
	if s.currentLimitPollDisabled {
		return
	}
	limit, err := client.CurrentLimitParameter(pack)
	if err != nil {
		s.currentLimitPollDisabled = true
		log.Printf("pace current limit parameter poll disabled after unsupported response: pack %d: %v", pack, err)
		return
	}
	s.state.UpsertCurrentLimitParameter(limit)
}

func (s *Service) logStatusChange(status pace.PackStatus) {
	signature := statusSignature(status)
	if s.lastStatusSignature[status.Address] == signature {
		return
	}
	s.lastStatusSignature[status.Address] = signature
	active := strings.Join(activeStatusFields(status), ",")
	if active == "" {
		active = "none"
	}
	log.Printf(
		"pace bms status changed pack=%02d charge_current=%s discharge_current=%s total_voltage=%s charger_available=%t charge_enabled=%t discharge_enabled=%t current_limit=%t current_limit_function=%t current_limit_gear=%t active=%s raw_warning_frame=%q",
		status.Address,
		status.ChargeCurrentWarning,
		status.DischargeCurrentWarning,
		status.TotalVoltageWarning,
		status.Instruction.ChargerAvailable,
		status.Instruction.ChargeEnabled,
		status.Instruction.DischargeEnabled,
		status.Instruction.CurrentLimitEnabled,
		status.Control.CurrentLimitFunction,
		status.Control.CurrentLimitGear,
		active,
		status.RawFrame,
	)
}

func statusSignature(status pace.PackStatus) string {
	return fmt.Sprintf(
		"cc=%s;dc=%s;tv=%s;instruction=%+v;control=%+v;protection=%+v;warning=%+v;fault=%+v;cell=%v;temp=%v",
		status.ChargeCurrentWarning,
		status.DischargeCurrentWarning,
		status.TotalVoltageWarning,
		status.Instruction,
		status.Control,
		status.Protection,
		status.Warning,
		status.Fault,
		status.CellWarnings,
		status.TemperatureWarnings,
	)
}

func activeStatusFields(status pace.PackStatus) []string {
	var fields []string
	if status.Instruction.CurrentLimitEnabled {
		fields = append(fields, "instruction.current_limit")
	}
	if status.Instruction.ReverseConnected {
		fields = append(fields, "instruction.reverse_connected")
	}
	if status.Control.CurrentLimitFunction {
		fields = append(fields, "control.current_limit_function")
	}
	if status.Control.CurrentLimitGear {
		fields = append(fields, "control.current_limit_gear")
	}
	if status.Protection.ShortCircuit {
		fields = append(fields, "protection.short_circuit")
	}
	if status.Protection.HighDischargeCurrent {
		fields = append(fields, "protection.high_discharge_current")
	}
	if status.Protection.HighChargeCurrent {
		fields = append(fields, "protection.high_charge_current")
	}
	if status.Protection.LowTotalVoltage {
		fields = append(fields, "protection.low_total_voltage")
	}
	if status.Protection.HighTotalVoltage {
		fields = append(fields, "protection.high_total_voltage")
	}
	if status.Protection.LowCellVoltage {
		fields = append(fields, "protection.low_cell_voltage")
	}
	if status.Protection.HighCellVoltage {
		fields = append(fields, "protection.high_cell_voltage")
	}
	if status.Protection.FullyCharged {
		fields = append(fields, "protection.fully_charged")
	}
	if status.Protection.LowEnvironmentTemp {
		fields = append(fields, "protection.low_environment_temp")
	}
	if status.Protection.HighEnvironmentTemp {
		fields = append(fields, "protection.high_environment_temp")
	}
	if status.Protection.HighMOSTemp {
		fields = append(fields, "protection.high_mos_temp")
	}
	if status.Protection.LowDischargeTemp {
		fields = append(fields, "protection.low_discharge_temp")
	}
	if status.Protection.LowChargeTemp {
		fields = append(fields, "protection.low_charge_temp")
	}
	if status.Protection.HighDischargeTemp {
		fields = append(fields, "protection.high_discharge_temp")
	}
	if status.Protection.HighChargeTemp {
		fields = append(fields, "protection.high_charge_temp")
	}
	if status.Warning.HighDischargeCurrent {
		fields = append(fields, "warning.high_discharge_current")
	}
	if status.Warning.HighChargeCurrent {
		fields = append(fields, "warning.high_charge_current")
	}
	if status.Warning.LowTotalVoltage {
		fields = append(fields, "warning.low_total_voltage")
	}
	if status.Warning.HighTotalVoltage {
		fields = append(fields, "warning.high_total_voltage")
	}
	if status.Warning.LowCellVoltage {
		fields = append(fields, "warning.low_cell_voltage")
	}
	if status.Warning.HighCellVoltage {
		fields = append(fields, "warning.high_cell_voltage")
	}
	if status.Warning.LowSOC {
		fields = append(fields, "warning.low_soc")
	}
	if status.Warning.HighMOSTemp {
		fields = append(fields, "warning.high_mos_temp")
	}
	if status.Warning.LowEnvironmentTemp {
		fields = append(fields, "warning.low_environment_temp")
	}
	if status.Warning.HighEnvironmentTemp {
		fields = append(fields, "warning.high_environment_temp")
	}
	if status.Warning.LowDischargeTemp {
		fields = append(fields, "warning.low_discharge_temp")
	}
	if status.Warning.LowChargeTemp {
		fields = append(fields, "warning.low_charge_temp")
	}
	if status.Warning.HighDischargeTemp {
		fields = append(fields, "warning.high_discharge_temp")
	}
	if status.Warning.HighChargeTemp {
		fields = append(fields, "warning.high_charge_temp")
	}
	if status.Fault.Sampling {
		fields = append(fields, "fault.sampling")
	}
	if status.Fault.Cell {
		fields = append(fields, "fault.cell")
	}
	if status.Fault.NTC {
		fields = append(fields, "fault.ntc")
	}
	if status.Fault.DischargeMOS {
		fields = append(fields, "fault.discharge_mos")
	}
	if status.Fault.ChargeMOS {
		fields = append(fields, "fault.charge_mos")
	}
	return fields
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
