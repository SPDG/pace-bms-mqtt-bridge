package pace

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

type Pack struct {
	Address             uint8           `json:"address"`
	InfoFlag            uint8           `json:"infoFlag"`
	ReportedPackCount   uint8           `json:"reportedPackCount"`
	CellsMV             []int           `json:"cellsMv"`
	TemperaturesC       []float64       `json:"temperaturesC"`
	Status              *PackStatus     `json:"status,omitempty"`
	Info                *PackInfo       `json:"info,omitempty"`
	Parameters          *PackParameters `json:"parameters,omitempty"`
	CurrentA            float64         `json:"currentA"`
	VoltageV            float64         `json:"voltageV"`
	PowerKW             float64         `json:"powerKw"`
	RemainingCapacityAh float64         `json:"remainingCapacityAh"`
	FullCapacityAh      float64         `json:"fullCapacityAh"`
	DesignCapacityAh    float64         `json:"designCapacityAh"`
	SOC                 float64         `json:"soc"`
	SOH                 float64         `json:"soh"`
	CycleCount          int             `json:"cycleCount"`
	DefinedNumberP      uint8           `json:"definedNumberP"`
	UpdatedAt           time.Time       `json:"updatedAt"`
}

type PackStatus struct {
	Address                 uint8             `json:"address"`
	InfoFlag                uint8             `json:"infoFlag"`
	CellCount               int               `json:"cellCount"`
	TemperatureCount        int               `json:"temperatureCount"`
	CellWarnings            []string          `json:"cellWarnings,omitempty"`
	CellsBalancing          []bool            `json:"cellsBalancing,omitempty"`
	BalancingCellCount      int               `json:"balancingCellCount"`
	TemperatureWarnings     []string          `json:"temperatureWarnings,omitempty"`
	ChargeCurrentWarning    string            `json:"chargeCurrentWarning,omitempty"`
	TotalVoltageWarning     string            `json:"totalVoltageWarning,omitempty"`
	DischargeCurrentWarning string            `json:"dischargeCurrentWarning,omitempty"`
	Protection              ProtectionStatus  `json:"protection"`
	Instruction             InstructionStatus `json:"instruction"`
	Control                 ControlStatus     `json:"control"`
	Fault                   FaultStatus       `json:"fault"`
	Warning                 WarningStatus     `json:"warning"`
	BalanceState1           uint8             `json:"balanceState1"`
	BalanceState2           uint8             `json:"balanceState2"`
	RawFrame                string            `json:"-"`
	UpdatedAt               time.Time         `json:"updatedAt"`
}

type PackInfo struct {
	Address             uint8                 `json:"address"`
	SoftwareVersion     string                `json:"softwareVersion,omitempty"`
	PaceSoftwareVersion string                `json:"paceSoftwareVersion,omitempty"`
	HardwareVersion     string                `json:"hardwareVersion,omitempty"`
	BMSSerialNumber     string                `json:"bmsSerialNumber,omitempty"`
	PackSerialNumber    string                `json:"packSerialNumber,omitempty"`
	BMSDateTime         string                `json:"bmsDateTime,omitempty"`
	InverterProtocol    *InverterProtocolInfo `json:"inverterProtocol,omitempty"`
	UpdatedAt           time.Time             `json:"updatedAt"`
}

type InverterProtocolInfo struct {
	CANCode    uint8  `json:"canCode"`
	CANName    string `json:"canName"`
	RS485Code  uint8  `json:"rs485Code"`
	RS485Name  string `json:"rs485Name"`
	Mode       uint8  `json:"mode"`
	ModeName   string `json:"modeName,omitempty"`
	SourceCID2 string `json:"sourceCid2,omitempty"`
}

type PackParameters struct {
	Address             uint8                            `json:"address"`
	CurrentLimitA       *float64                         `json:"currentLimitA,omitempty"`
	Equalization        *EqualizationParameters          `json:"equalization,omitempty"`
	FullCharge          *FullChargeParameters            `json:"fullCharge,omitempty"`
	SingleHibernate     *SingleHibernateParameters       `json:"singleHibernate,omitempty"`
	CellOVP             *VoltageProtectionParameters     `json:"cellOvp,omitempty"`
	CellUVP             *VoltageProtectionParameters     `json:"cellUvp,omitempty"`
	PackOVP             *PackVoltageProtectionParameters `json:"packOvp,omitempty"`
	PackUVP             *PackVoltageProtectionParameters `json:"packUvp,omitempty"`
	ChargeOCP           *CurrentProtectionParameters     `json:"chargeOcp,omitempty"`
	DischargeOCP1       *CurrentProtectionParameters     `json:"dischargeOcp1,omitempty"`
	DischargeOCP2       *DischargeOCP2Parameters         `json:"dischargeOcp2,omitempty"`
	ShortCircuitDelayUS *int                             `json:"shortCircuitDelayUs,omitempty"`
	ChargeHighTemp      *TemperatureProtectionParameters `json:"chargeHighTemp,omitempty"`
	DischargeHighTemp   *TemperatureProtectionParameters `json:"dischargeHighTemp,omitempty"`
	ChargeLowTemp       *TemperatureProtectionParameters `json:"chargeLowTemp,omitempty"`
	DischargeLowTemp    *TemperatureProtectionParameters `json:"dischargeLowTemp,omitempty"`
	MOSHighTemp         *TemperatureProtectionParameters `json:"mosHighTemp,omitempty"`
	EnvLowTemp          *TemperatureProtectionParameters `json:"envLowTemp,omitempty"`
	EnvHighTemp         *TemperatureProtectionParameters `json:"envHighTemp,omitempty"`
	UpdatedAt           time.Time                        `json:"updatedAt"`
}

type EqualizationParameters struct {
	StartVoltageMV int `json:"startVoltageMv"`
	DeltaVoltageMV int `json:"deltaVoltageMv"`
}

type FullChargeParameters struct {
	FullVoltageV       float64 `json:"fullVoltageV"`
	CutOffCurrentA     float64 `json:"cutOffCurrentA"`
	LowSOCAlarmPercent int     `json:"lowSocAlarmPercent"`
}

type SingleHibernateParameters struct {
	SleepVoltageMV int `json:"sleepVoltageMv"`
	DelayMinutes   int `json:"delayMinutes"`
}

type VoltageProtectionParameters struct {
	Enabled    bool `json:"enabled"`
	WarningMV  int  `json:"warningMv"`
	ProtectMV  int  `json:"protectMv"`
	RecoveryMV int  `json:"recoveryMv"`
	DelayMS    int  `json:"delayMs"`
}

type PackVoltageProtectionParameters struct {
	Enabled   bool    `json:"enabled"`
	WarningV  float64 `json:"warningV"`
	ProtectV  float64 `json:"protectV"`
	RecoveryV float64 `json:"recoveryV"`
	DelayMS   int     `json:"delayMs"`
}

type CurrentProtectionParameters struct {
	Enabled  bool    `json:"enabled"`
	WarningA float64 `json:"warningA"`
	ProtectA float64 `json:"protectA"`
	DelayMS  int     `json:"delayMs"`
}

type DischargeOCP2Parameters struct {
	ProtectA float64 `json:"protectA"`
	DelayMS  int     `json:"delayMs"`
}

type TemperatureProtectionParameters struct {
	Enabled   bool    `json:"enabled"`
	WarningC  float64 `json:"warningC"`
	ProtectC  float64 `json:"protectC"`
	RecoveryC float64 `json:"recoveryC"`
}

type CurrentLimitParameter struct {
	Address    uint8     `json:"address"`
	CurrentA   float64   `json:"currentA"`
	SourceCID2 string    `json:"sourceCid2"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

type ProtectionStatus struct {
	ChargerHighVoltage   bool `json:"chargerHighVoltage"`
	ShortCircuit         bool `json:"shortCircuit"`
	HighDischargeCurrent bool `json:"highDischargeCurrent"`
	HighChargeCurrent    bool `json:"highChargeCurrent"`
	LowTotalVoltage      bool `json:"lowTotalVoltage"`
	HighTotalVoltage     bool `json:"highTotalVoltage"`
	LowCellVoltage       bool `json:"lowCellVoltage"`
	HighCellVoltage      bool `json:"highCellVoltage"`
	FullyCharged         bool `json:"fullyCharged"`
	LowEnvironmentTemp   bool `json:"lowEnvironmentTemp"`
	HighEnvironmentTemp  bool `json:"highEnvironmentTemp"`
	HighMOSTemp          bool `json:"highMosTemp"`
	LowDischargeTemp     bool `json:"lowDischargeTemp"`
	LowChargeTemp        bool `json:"lowChargeTemp"`
	HighDischargeTemp    bool `json:"highDischargeTemp"`
	HighChargeTemp       bool `json:"highChargeTemp"`
}

type InstructionStatus struct {
	HeaterOn            bool `json:"heaterOn"`
	ChargerAvailable    bool `json:"chargerAvailable"`
	ReverseConnected    bool `json:"reverseConnected"`
	DischargeEnabled    bool `json:"dischargeEnabled"`
	ChargeEnabled       bool `json:"chargeEnabled"`
	CurrentLimitEnabled bool `json:"currentLimitEnabled"`
}

type ControlStatus struct {
	SystemLock           bool `json:"systemLock"`
	LEDWarnFunction      bool `json:"ledWarnFunction"`
	CurrentLimitFunction bool `json:"currentLimitFunction"`
	CurrentLimitGear     bool `json:"currentLimitGear"`
	BuzzerWarnFunction   bool `json:"buzzerWarnFunction"`
}

type FaultStatus struct {
	Heater       bool `json:"heater"`
	CurrentLimit bool `json:"currentLimit"`
	Sampling     bool `json:"sampling"`
	Cell         bool `json:"cell"`
	NTC          bool `json:"ntc"`
	DischargeMOS bool `json:"dischargeMos"`
	ChargeMOS    bool `json:"chargeMos"`
}

type WarningStatus struct {
	HighDischargeCurrent bool `json:"highDischargeCurrent"`
	HighChargeCurrent    bool `json:"highChargeCurrent"`
	LowTotalVoltage      bool `json:"lowTotalVoltage"`
	HighTotalVoltage     bool `json:"highTotalVoltage"`
	LowCellVoltage       bool `json:"lowCellVoltage"`
	HighCellVoltage      bool `json:"highCellVoltage"`
	LowSOC               bool `json:"lowSoc"`
	HighMOSTemp          bool `json:"highMosTemp"`
	LowEnvironmentTemp   bool `json:"lowEnvironmentTemp"`
	HighEnvironmentTemp  bool `json:"highEnvironmentTemp"`
	LowDischargeTemp     bool `json:"lowDischargeTemp"`
	LowChargeTemp        bool `json:"lowChargeTemp"`
	HighDischargeTemp    bool `json:"highDischargeTemp"`
	HighChargeTemp       bool `json:"highChargeTemp"`
}

type Telemetry struct {
	ID                        string    `json:"id"`
	Name                      string    `json:"name"`
	PackAddress               uint8     `json:"packAddress"`
	Unit                      string    `json:"unit,omitempty"`
	DeviceClass               string    `json:"deviceClass,omitempty"`
	StateClass                string    `json:"stateClass,omitempty"`
	Icon                      string    `json:"icon,omitempty"`
	SuggestedDisplayPrecision *int      `json:"suggestedDisplayPrecision,omitempty"`
	Value                     any       `json:"value"`
	Rendered                  string    `json:"rendered"`
	UpdatedAt                 time.Time `json:"updatedAt"`
}

func TelemetryForPack(p Pack) []Telemetry {
	values := []Telemetry{
		num(p, "soc", "SOC", p.SOC, 1, "%", "battery", "measurement", "mdi:battery"),
		num(p, "voltage", "Voltage", p.VoltageV, 2, "V", "", "measurement", "mdi:car-battery"),
		num(p, "current", "Current", p.CurrentA, 2, "A", "current", "measurement", "mdi:current-dc"),
		num(p, "power", "Power", p.PowerKW*1000, 0, "W", "power", "measurement", "mdi:flash"),
		num(p, "remaining_capacity", "Remaining Capacity", p.RemainingCapacityAh, 2, "Ah", "", "measurement", "mdi:battery-clock"),
		num(p, "full_capacity", "Full Capacity", p.FullCapacityAh, 2, "Ah", "", "measurement", "mdi:battery-high"),
		num(p, "design_capacity", "Design Capacity", p.DesignCapacityAh, 2, "Ah", "", "measurement", "mdi:battery-high"),
		num(p, "soh", "SOH", p.SOH, 0, "%", "", "measurement", "mdi:battery-heart"),
		intv(p, "cycle_count", "Cycle Count", p.CycleCount, "", "", "total_increasing", "mdi:counter"),
		intv(p, "cell_count", "Cell Count", len(p.CellsMV), "", "", "measurement", "mdi:counter"),
		intv(p, "temperature_count", "Temperature Count", len(p.TemperaturesC), "", "", "measurement", "mdi:counter"),
	}
	if minMV, maxMV, ok := cellVoltageRange(p.CellsMV); ok {
		diffMV := maxMV - minMV
		values = append(values,
			cellSummary(p, "cell_voltage_min", "Cell Voltage Min", minMV, "mdi:align-vertical-bottom"),
			cellSummary(p, "cell_voltage_max", "Cell Voltage Max", maxMV, "mdi:align-vertical-top"),
			cellSummary(p, "cell_voltage_diff", "Cell Voltage Diff", diffMV, "mdi:format-align-middle"),
		)
	}

	for i, mv := range p.CellsMV {
		id := fmt.Sprintf("cell_%02d_voltage", i+1)
		values = append(values, Telemetry{
			ID:                        packID(p.Address, id),
			Name:                      fmt.Sprintf("Pack %02d Cell %02d Voltage", p.Address, i+1),
			PackAddress:               p.Address,
			Unit:                      "mV",
			DeviceClass:               "voltage",
			StateClass:                "measurement",
			Icon:                      "mdi:battery",
			SuggestedDisplayPrecision: displayPrecision(0),
			Value:                     mv,
			Rendered:                  strconv.Itoa(mv),
			UpdatedAt:                 p.UpdatedAt,
		})
	}
	for i, temp := range p.TemperaturesC {
		id := fmt.Sprintf("temperature_%02d", i+1)
		values = append(values, Telemetry{
			ID:                        packID(p.Address, id),
			Name:                      fmt.Sprintf("Pack %02d Temperature %02d", p.Address, i+1),
			PackAddress:               p.Address,
			Unit:                      "\u00b0C",
			DeviceClass:               "temperature",
			StateClass:                "measurement",
			Icon:                      "mdi:thermometer",
			SuggestedDisplayPrecision: displayPrecision(2),
			Value:                     temp,
			Rendered:                  strconv.FormatFloat(temp, 'f', 2, 64),
			UpdatedAt:                 p.UpdatedAt,
		})
	}

	if p.Status != nil {
		st := p.Status
		statusTime := st.UpdatedAt
		if statusTime.IsZero() {
			statusTime = p.UpdatedAt
		}
		values = append(values,
			intvAt(p, "balancing_cell_count", "Balancing Cell Count", st.BalancingCellCount, "", "", "measurement", "mdi:scale-balance", statusTime),
			boolvAt(p, "balancing_active", "Balancing Active", st.BalancingCellCount > 0, "mdi:scale-balance", statusTime),
			strvAt(p, "balancing_cells", "Balancing Cells", balancingCellsSummary(st.CellsBalancing), "mdi:scale-balance", statusTime),
			boolvAt(p, "charge_mos", "Charge MOS", st.Instruction.ChargeEnabled, "mdi:electric-switch", statusTime),
			boolvAt(p, "discharge_mos", "Discharge MOS", st.Instruction.DischargeEnabled, "mdi:electric-switch", statusTime),
			boolvAt(p, "charger_available", "Charger Available", st.Instruction.ChargerAvailable, "mdi:power-plug", statusTime),
			boolvAt(p, "heater_active", "Heater Active", st.Instruction.HeaterOn, "mdi:radiator", statusTime),
		)
	}

	if p.Info != nil {
		info := p.Info
		infoTime := info.UpdatedAt
		if infoTime.IsZero() {
			infoTime = p.UpdatedAt
		}
		if info.SoftwareVersion != "" {
			values = append(values, strvAt(p, "software_version", "Firmware Version", info.SoftwareVersion, "mdi:chip", infoTime))
		}
		if info.PaceSoftwareVersion != "" {
			values = append(values, strvAt(p, "pace_software_version", "PACE Firmware Version", info.PaceSoftwareVersion, "mdi:chip", infoTime))
		}
		if info.HardwareVersion != "" {
			values = append(values, strvAt(p, "hardware_version", "Hardware Version", info.HardwareVersion, "mdi:expansion-card", infoTime))
		}
		if info.BMSSerialNumber != "" {
			values = append(values, strvAt(p, "bms_serial_number", "BMS Serial Number", info.BMSSerialNumber, "mdi:identifier", infoTime))
		}
		if info.PackSerialNumber != "" {
			values = append(values, strvAt(p, "pack_serial_number", "Pack Serial Number", info.PackSerialNumber, "mdi:identifier", infoTime))
		}
		if info.BMSDateTime != "" {
			values = append(values, strvAt(p, "bms_datetime", "BMS Clock", info.BMSDateTime, "mdi:clock-outline", infoTime))
		}
		if info.InverterProtocol != nil {
			if info.InverterProtocol.CANName != "" {
				values = append(values, strvAt(p, "can_protocol", "CAN Protocol", info.InverterProtocol.CANName, "mdi:serial-port", infoTime))
			}
			if info.InverterProtocol.RS485Name != "" {
				values = append(values, strvAt(p, "rs485_protocol", "RS485 Protocol", info.InverterProtocol.RS485Name, "mdi:serial-port", infoTime))
			}
		}
	}

	if p.Parameters != nil {
		params := p.Parameters
		paramTime := params.UpdatedAt
		if paramTime.IsZero() {
			paramTime = p.UpdatedAt
		}
		if params.Equalization != nil {
			values = append(values,
				intvAt(p, "equalization_start_voltage", "Balancing Start Voltage", params.Equalization.StartVoltageMV, "mV", "voltage", "measurement", "mdi:scale-balance", paramTime),
				intvAt(p, "equalization_diff_voltage", "Balancing Trigger Diff", params.Equalization.DeltaVoltageMV, "mV", "voltage", "measurement", "mdi:scale-balance", paramTime),
			)
		}
		if params.FullCharge != nil {
			values = append(values,
				numAt(p, "full_charge_voltage", "Full Charge Voltage", params.FullCharge.FullVoltageV, 2, "V", "voltage", "measurement", "mdi:battery-charging-100", paramTime),
				numAt(p, "full_charge_current", "Full Charge Cut-off Current", params.FullCharge.CutOffCurrentA, 2, "A", "current", "measurement", "mdi:current-dc", paramTime),
				intvAt(p, "low_soc_alarm", "Low SOC Alarm Threshold", params.FullCharge.LowSOCAlarmPercent, "%", "", "measurement", "mdi:battery-alert", paramTime),
			)
		}
		if params.SingleHibernate != nil {
			values = append(values,
				intvAt(p, "hibernate_voltage", "Cell Sleep Voltage", params.SingleHibernate.SleepVoltageMV, "mV", "voltage", "measurement", "mdi:power-sleep", paramTime),
				intvAt(p, "hibernate_delay", "Cell Sleep Delay", params.SingleHibernate.DelayMinutes, "min", "duration", "measurement", "mdi:timer-outline", paramTime),
			)
		}
		if params.CellOVP != nil {
			values = append(values, intvAt(p, "cell_ovp_protect", "Cell OVP Protect", params.CellOVP.ProtectMV, "mV", "voltage", "measurement", "mdi:shield-alert", paramTime))
		}
		if params.CellUVP != nil {
			values = append(values, intvAt(p, "cell_uvp_protect", "Cell UVP Protect", params.CellUVP.ProtectMV, "mV", "voltage", "measurement", "mdi:shield-alert", paramTime))
		}
		if params.PackOVP != nil {
			values = append(values, numAt(p, "pack_ovp_protect", "Pack OVP Protect", params.PackOVP.ProtectV, 2, "V", "voltage", "measurement", "mdi:shield-alert", paramTime))
		}
		if params.PackUVP != nil {
			values = append(values, numAt(p, "pack_uvp_protect", "Pack UVP Protect", params.PackUVP.ProtectV, 2, "V", "voltage", "measurement", "mdi:shield-alert", paramTime))
		}
		if params.ChargeOCP != nil {
			values = append(values, numAt(p, "charge_ocp_protect", "Charge OCP Protect", params.ChargeOCP.ProtectA, 0, "A", "current", "measurement", "mdi:shield-alert", paramTime))
		}
		if params.DischargeOCP1 != nil {
			values = append(values, numAt(p, "discharge_ocp1_protect", "Discharge OCP1 Protect", params.DischargeOCP1.ProtectA, 0, "A", "current", "measurement", "mdi:shield-alert", paramTime))
		}
		if params.DischargeOCP2 != nil {
			values = append(values, numAt(p, "discharge_ocp2_protect", "Discharge OCP2 Protect", params.DischargeOCP2.ProtectA, 0, "A", "current", "measurement", "mdi:shield-alert", paramTime))
		}
		if params.ShortCircuitDelayUS != nil {
			values = append(values, intvAt(p, "short_circuit_delay", "Short Circuit Delay", *params.ShortCircuitDelayUS, "\u00b5s", "", "measurement", "mdi:timer-alert", paramTime))
		}
	}
	return values
}

func AggregateTelemetry(packs []Pack) []Telemetry {
	if len(packs) == 0 {
		return nil
	}

	var totalPowerW float64
	updatedAt := packs[0].UpdatedAt
	for _, pack := range packs {
		// PACE reports discharge current as negative on the tested packs.
		// Home Assistant battery power in "Standard" mode expects positive
		// values while discharging and negative values while charging.
		totalPowerW -= pack.PowerKW * 1000
		if pack.UpdatedAt.After(updatedAt) {
			updatedAt = pack.UpdatedAt
		}
	}

	dischargePowerW := math.Max(totalPowerW, 0)
	chargePowerW := math.Max(-totalPowerW, 0)
	return []Telemetry{
		aggregatePower("battery_power", "Battery Power", totalPowerW, "mdi:battery-charging", updatedAt),
		aggregatePower("battery_discharge_power", "Battery Discharge Power", dischargePowerW, "mdi:battery-arrow-down", updatedAt),
		aggregatePower("battery_charge_power", "Battery Charge Power", chargePowerW, "mdi:battery-arrow-up", updatedAt),
	}
}

func TelemetryForCurrentLimitParameter(value CurrentLimitParameter) Telemetry {
	id := "current_limit_parameter"
	name := "Current Limit Parameter"
	if value.Address != 0 && value.Address != 255 {
		id = packID(value.Address, id)
		name = fmt.Sprintf("Pack %02d Current Limit Parameter", value.Address)
	}
	return Telemetry{
		ID:                        id,
		Name:                      name,
		PackAddress:               value.Address,
		Unit:                      "A",
		DeviceClass:               "current",
		StateClass:                "measurement",
		Icon:                      "mdi:current-dc",
		SuggestedDisplayPrecision: displayPrecision(0),
		Value:                     value.CurrentA,
		Rendered:                  strconv.FormatFloat(value.CurrentA, 'f', 0, 64),
		UpdatedAt:                 value.UpdatedAt,
	}
}

func cellVoltageRange(cells []int) (int, int, bool) {
	if len(cells) == 0 {
		return 0, 0, false
	}
	minMV := cells[0]
	maxMV := cells[0]
	for _, mv := range cells[1:] {
		if mv < minMV {
			minMV = mv
		}
		if mv > maxMV {
			maxMV = mv
		}
	}
	return minMV, maxMV, true
}

func aggregatePower(id, name string, value float64, icon string, updatedAt time.Time) Telemetry {
	rounded := math.Round(value)
	return Telemetry{
		ID:                        id,
		Name:                      name,
		Unit:                      "W",
		DeviceClass:               "power",
		StateClass:                "measurement",
		Icon:                      icon,
		SuggestedDisplayPrecision: displayPrecision(0),
		Value:                     rounded,
		Rendered:                  strconv.FormatFloat(rounded, 'f', 0, 64),
		UpdatedAt:                 updatedAt,
	}
}

func cellSummary(p Pack, id, name string, valueMV int, icon string) Telemetry {
	return Telemetry{
		ID:                        packID(p.Address, id),
		Name:                      fmt.Sprintf("Pack %02d %s", p.Address, name),
		PackAddress:               p.Address,
		Unit:                      "mV",
		DeviceClass:               "voltage",
		StateClass:                "measurement",
		Icon:                      icon,
		SuggestedDisplayPrecision: displayPrecision(0),
		Value:                     valueMV,
		Rendered:                  strconv.Itoa(valueMV),
		UpdatedAt:                 p.UpdatedAt,
	}
}

func num(p Pack, id, name string, value float64, precisionValue int, unit, deviceClass, stateClass, icon string) Telemetry {
	return numAt(p, id, name, value, precisionValue, unit, deviceClass, stateClass, icon, p.UpdatedAt)
}

func numAt(p Pack, id, name string, value float64, precisionValue int, unit, deviceClass, stateClass, icon string, updatedAt time.Time) Telemetry {
	return Telemetry{
		ID:                        packID(p.Address, id),
		Name:                      fmt.Sprintf("Pack %02d %s", p.Address, name),
		PackAddress:               p.Address,
		Unit:                      unit,
		DeviceClass:               deviceClass,
		StateClass:                stateClass,
		Icon:                      icon,
		SuggestedDisplayPrecision: displayPrecision(precisionValue),
		Value:                     value,
		Rendered:                  strconv.FormatFloat(value, 'f', precisionValue, 64),
		UpdatedAt:                 updatedAt,
	}
}

func intv(p Pack, id, name string, value int, unit, deviceClass, stateClass, icon string) Telemetry {
	return intvAt(p, id, name, value, unit, deviceClass, stateClass, icon, p.UpdatedAt)
}

func intvAt(p Pack, id, name string, value int, unit, deviceClass, stateClass, icon string, updatedAt time.Time) Telemetry {
	return Telemetry{
		ID:          packID(p.Address, id),
		Name:        fmt.Sprintf("Pack %02d %s", p.Address, name),
		PackAddress: p.Address,
		Unit:        unit,
		DeviceClass: deviceClass,
		StateClass:  stateClass,
		Icon:        icon,
		Value:       value,
		Rendered:    strconv.Itoa(value),
		UpdatedAt:   updatedAt,
	}
}

func boolvAt(p Pack, id, name string, value bool, icon string, updatedAt time.Time) Telemetry {
	rendered := "off"
	if value {
		rendered = "on"
	}
	return Telemetry{
		ID:          packID(p.Address, id),
		Name:        fmt.Sprintf("Pack %02d %s", p.Address, name),
		PackAddress: p.Address,
		Icon:        icon,
		Value:       value,
		Rendered:    rendered,
		UpdatedAt:   updatedAt,
	}
}

func strvAt(p Pack, id, name, value, icon string, updatedAt time.Time) Telemetry {
	return Telemetry{
		ID:          packID(p.Address, id),
		Name:        fmt.Sprintf("Pack %02d %s", p.Address, name),
		PackAddress: p.Address,
		Icon:        icon,
		Value:       value,
		Rendered:    value,
		UpdatedAt:   updatedAt,
	}
}

func balancingCellsSummary(balancing []bool) string {
	var active []string
	for i, b := range balancing {
		if b {
			active = append(active, fmt.Sprintf("%02d", i+1))
		}
	}
	if len(active) == 0 {
		return "none"
	}
	return strings.Join(active, ", ")
}

func packID(address uint8, id string) string {
	return fmt.Sprintf("pack_%02d_%s", address, id)
}

func displayPrecision(value int) *int {
	return &value
}

func round(v float64, precision int) float64 {
	pow := math.Pow10(precision)
	return math.Round(v*pow) / pow
}
