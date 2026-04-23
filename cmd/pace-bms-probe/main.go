package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/SPDG/pace-bms-mqtt-bridge/internal/config"
	"github.com/SPDG/pace-bms-mqtt-bridge/internal/pace"
)

func main() {
	port := flag.String("port", "/dev/pace-bms", "Serial port")
	baud := flag.Int("baud", 115200, "Serial baud rate")
	protocol := flag.String("protocol", "rs485", "PACE protocol flavor: rs485 or rs232")
	first := flag.Int("first", 0, "First pack address to probe")
	last := flag.Int("last", 16, "Last pack address to probe")
	flag.Parse()

	cfg := config.Default()
	cfg.Serial.Port = *port
	cfg.Serial.BaudRate = *baud
	cfg.Device.Protocol = *protocol
	client, err := pace.Open(cfg)
	if err != nil {
		log.Fatalf("open serial: %v", err)
	}
	defer client.Close()

	for address := *first; address <= *last; address++ {
		if address < 0 || address > 255 {
			log.Fatalf("address out of range: %d", address)
		}
		pack := uint8(address)
		got, err := client.PackNumber(pack)
		if err != nil {
			fmt.Printf("addr=%d pack_number error=%v\n", address, err)
			continue
		}
		fmt.Printf("addr=%d pack_number=%d\n", address, got)
		analog, err := client.Analog(pack)
		if err != nil {
			fmt.Printf("addr=%d analog error=%v\n", address, err)
			continue
		}
		fmt.Printf("addr=%d voltage=%.2fV current=%.2fA soc=%.1f%% cells=%d temps=%d\n",
			address,
			analog.VoltageV,
			analog.CurrentA,
			analog.SOC,
			len(analog.CellsMV),
			len(analog.TemperaturesC),
		)
	}
}
