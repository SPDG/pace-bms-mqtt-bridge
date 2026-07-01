package pace

import (
	"fmt"
	"math"
	"strconv"
	"time"
)

type Pack struct {
	Address             uint8       `json:"address"`
	InfoFlag            uint8       `json:"infoFlag"`
	ReportedPackCount   uint8       `json:"reportedPackCount"`
	CellsMV             []int       `json:"cellsMv"`
	TemperaturesC       []float64   `json:"temperaturesC"`
	Status              *PackStatus `json:"status,omitempty"`
	CurrentA            float64     `json:"currentA"`
	VoltageV            float64     `json:"voltageV"`
	PowerKW             float64     `json:"powerKw"`
	RemainingCapacityAh float64     `json:"remainingCapacityAh"`
	FullCapacityAh      float64     `json:"fullCapacityAh"`
	DesignCapacityAh    float64     `json:"designCapacityAh"`
	SOC                 float64     `json:"soc"`
	SOH                 float64     `json:"soh"`
	CycleCount          int         `json:"cycleCount"`
	DefinedNumberP      uint8       `json:"definedNumberP"`
	UpdatedAt           time.Time   `json:"updatedAt"`
}

type PackStatus struct {
	Address                 uint8             `json:"address"`
	InfoFlag                uint8             `json:"infoFlag"`
	CellCount               int               `json:"cellCount"`
	TemperatureCount        int               `json:"temperatureCount"`
	CellWarnings            []string          `json:"cellWarnings,omitempty"`
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
	UpdatedAt               time.Time         `json:"updatedAt"`
}

type CurrentLimitParameter struct {
	Address    uint8     `json:"address"`
	CurrentA   float64   `json:"currentA"`
	SourceCID2 string    `json:"sourceCid2"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

type ProtectionStatus struct {
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
	ChargerAvailable    bool `json:"chargerAvailable"`
	ReverseConnected    bool `json:"reverseConnected"`
	DischargeEnabled    bool `json:"dischargeEnabled"`
	ChargeEnabled       bool `json:"chargeEnabled"`
	CurrentLimitEnabled bool `json:"currentLimitEnabled"`
}

type ControlStatus struct {
	LEDWarnFunction      bool `json:"ledWarnFunction"`
	CurrentLimitFunction bool `json:"currentLimitFunction"`
	CurrentLimitGear     bool `json:"currentLimitGear"`
	BuzzerWarnFunction   bool `json:"buzzerWarnFunction"`
}

type FaultStatus struct {
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
		UpdatedAt:                 p.UpdatedAt,
	}
}

func intv(p Pack, id, name string, value int, unit, deviceClass, stateClass, icon string) Telemetry {
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
		UpdatedAt:   p.UpdatedAt,
	}
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
