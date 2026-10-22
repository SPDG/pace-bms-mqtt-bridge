# PACE BMS MQTT Bridge

Linux-first single-binary bridge for polling PACE BMS battery packs over local serial communication and publishing read-only telemetry to MQTT and Home Assistant.

## Status

The bridge currently focuses on safe reads only:

- PACE ASCII serial frames (`RS232` and `RS485`) over local serial ports or transparent TCP (`tcp://host:port`),
- pack discovery over RS485-style addressing,
- analog telemetry parsing (`0x42`) for cells, temperatures, voltage, current, capacity, SOC, SOH and cycle count,
- warning, protection, fault, MOSFET, and per-cell balancing status parsing (`0x44`),
- device metadata readback: firmware/hardware versions (`0xC1`, `0xC6`), BMS/Pack serial numbers (`0xC2`), RTC date/time (`0xB1`), and CAN/RS485 inverter protocol selection (`0xEB`/`0xEC`/`0x62`),
- full readback of BMS protection, equalization (`0xB6`), full-charge (`0xAF`), hibernate (`0xA0`), current limit (`0xED`), OVP/UVP (`0xD1`–`0xD7`), OCP/SCP (`0xD9`, `0xDB`, `0xE3`, `0xE5`), and temperature thresholds (`0xDD`, `0xDF`, `0xE1`, `0xE7`),
- MQTT state topics and Home Assistant MQTT Discovery,
- responsive embedded HTTP status page (desktop & mobile) and JSON API.

No BMS writes are enabled in the runtime bridge yet. Full protocol documentation—including all writable, control (`0x99`, `0x9A`, `0x9B`, `0x9C`), equalization (`0xB5`), and calibration commands—is available in [`docs/PROTOCOL.md`](docs/PROTOCOL.md).

## Quick Start

Run locally with the example config:

```bash
go run ./cmd/pace-bms-mqtt-bridge --config ./configs/config.example.yaml
```

Probe a serial adapter:

```bash
go run ./cmd/pace-bms-probe --port /dev/pace-bms --baud 9600 --protocol rs232 --first 255 --last 255
go run ./cmd/pace-bms-probe --port /dev/pace-bms --baud 115200 --protocol rs485 --first 0 --last 16
```

By default the web panel listens on `http://127.0.0.1:8080`.

## Configuration

See [`configs/config.example.yaml`](configs/config.example.yaml).

For Linux deployments, prefer a stable udev symlink such as `/dev/pace-bms` instead of raw `/dev/ttyUSB*`.

Serial-over-Ethernet converters can be used by setting `serial.port` to a TCP endpoint such as `tcp://192.168.6.134:4196`. The converter should run in transparent TCP server mode with the same serial parameters configured in `serial`.

`polling.interval` controls the fast analog telemetry loop. `polling.status_interval` controls the slower warning/status loop used for MOS, fault and protection readback.

Deployment notes and a `systemd` unit are in [`docs/DEPLOYMENT.md`](docs/DEPLOYMENT.md).

## Releases

GitHub Actions builds and tests every push to `main` and every pull request.

Pushing a tag such as `v0.0.1` creates a GitHub release with Linux `amd64` and `arm64` archives. After CI succeeds on `main`, the repository also creates the next patch tag automatically and dispatches the release workflow.

## Credits

Protocol behavior was cross-checked against:

- <https://github.com/fancyui/Gobel-Battery-HA-Addon>
- <https://github.com/Tertiush/bmspace>

## License

Apache-2.0. See [`LICENSE`](LICENSE).
