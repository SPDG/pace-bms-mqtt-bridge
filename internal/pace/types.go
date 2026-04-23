package pace

import (
	"fmt"
	"math"
	"strconv"
	"time"
)

type Pack struct {
	Address             uint8     `json:"address"`
	InfoFlag            uint8     `json:"infoFlag"`
	ReportedPackCount   uint8     `json:"reportedPackCount"`
	CellsMV             []int     `json:"cellsMv"`
	TemperaturesC       []float64 `json:"temperaturesC"`
	CurrentA            float64   `json:"currentA"`
	VoltageV            float64   `json:"voltageV"`
	PowerKW             float64   `json:"powerKw"`
	RemainingCapacityAh float64   `json:"remainingCapacityAh"`
	FullCapacityAh      float64   `json:"fullCapacityAh"`
	DesignCapacityAh    float64   `json:"designCapacityAh"`
	SOC                 float64   `json:"soc"`
	SOH                 float64   `json:"soh"`
	CycleCount          int       `json:"cycleCount"`
	DefinedNumberP      uint8     `json:"definedNumberP"`
	UpdatedAt           time.Time `json:"updatedAt"`
}

type Telemetry struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	PackAddress uint8     `json:"packAddress"`
	Unit        string    `json:"unit,omitempty"`
	DeviceClass string    `json:"deviceClass,omitempty"`
	StateClass  string    `json:"stateClass,omitempty"`
	Icon        string    `json:"icon,omitempty"`
	Value       any       `json:"value"`
	Rendered    string    `json:"rendered"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

func TelemetryForPack(p Pack) []Telemetry {
	values := []Telemetry{
		num(p, "soc", "SOC", p.SOC, 1, "%", "battery", "measurement", "mdi:battery"),
		num(p, "voltage", "Voltage", p.VoltageV, 2, "V", "voltage", "measurement", "mdi:car-battery"),
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

	for i, mv := range p.CellsMV {
		id := fmt.Sprintf("cell_%02d_voltage", i+1)
		values = append(values, Telemetry{
			ID:          packID(p.Address, id),
			Name:        fmt.Sprintf("Pack %02d Cell %02d Voltage", p.Address, i+1),
			PackAddress: p.Address,
			Unit:        "V",
			DeviceClass: "voltage",
			StateClass:  "measurement",
			Icon:        "mdi:battery",
			Value:       float64(mv) / 1000,
			Rendered:    strconv.FormatFloat(float64(mv)/1000, 'f', 3, 64),
			UpdatedAt:   p.UpdatedAt,
		})
	}
	for i, temp := range p.TemperaturesC {
		id := fmt.Sprintf("temperature_%02d", i+1)
		values = append(values, Telemetry{
			ID:          packID(p.Address, id),
			Name:        fmt.Sprintf("Pack %02d Temperature %02d", p.Address, i+1),
			PackAddress: p.Address,
			Unit:        "C",
			DeviceClass: "temperature",
			StateClass:  "measurement",
			Icon:        "mdi:thermometer",
			Value:       temp,
			Rendered:    strconv.FormatFloat(temp, 'f', 2, 64),
			UpdatedAt:   p.UpdatedAt,
		})
	}
	return values
}

func num(p Pack, id, name string, value float64, precision int, unit, deviceClass, stateClass, icon string) Telemetry {
	return Telemetry{
		ID:          packID(p.Address, id),
		Name:        fmt.Sprintf("Pack %02d %s", p.Address, name),
		PackAddress: p.Address,
		Unit:        unit,
		DeviceClass: deviceClass,
		StateClass:  stateClass,
		Icon:        icon,
		Value:       value,
		Rendered:    strconv.FormatFloat(value, 'f', precision, 64),
		UpdatedAt:   p.UpdatedAt,
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

func round(v float64, precision int) float64 {
	pow := math.Pow10(precision)
	return math.Round(v*pow) / pow
}
