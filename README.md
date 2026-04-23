# PACE BMS MQTT Bridge

Linux-first single-binary bridge for polling PACE BMS battery packs over local serial communication and publishing read-only telemetry to MQTT and Home Assistant.

## Status

Early implementation. The bridge currently focuses on safe reads only:

- PACE ASCII serial frames used by common PACE/Gobel packs,
- pack discovery over RS485-style addressing,
- analog telemetry parsing for cells, temperatures, voltage, current, capacity, SOC, SOH and cycle count,
- MQTT state topics and Home Assistant MQTT Discovery,
- embedded HTTP status page and JSON API.

No BMS writes are implemented.

## Quick Start

Run locally with the example config:

```bash
go run ./cmd/pace-bms-mqtt-bridge --config ./configs/config.example.yaml
```

Probe a serial adapter:

```bash
go run ./cmd/pace-bms-probe --port /dev/pace-bms --baud 115200 --protocol rs485 --first 0 --last 16
go run ./cmd/pace-bms-probe --port /dev/pace-bms --baud 9600 --protocol console --first 1 --last 1
```

By default the web panel listens on `http://127.0.0.1:8080`.

## Configuration

See [`configs/config.example.yaml`](configs/config.example.yaml).

For Linux deployments, prefer a stable udev symlink such as `/dev/pace-bms` instead of raw `/dev/ttyUSB*`.

Deployment notes and a `systemd` unit are in [`docs/DEPLOYMENT.md`](docs/DEPLOYMENT.md).

## Credits

Protocol behavior was cross-checked against:

- <https://github.com/fancyui/Gobel-Battery-HA-Addon>
- <https://github.com/Tertiush/bmspace>

## License

Apache-2.0. See [`LICENSE`](LICENSE).
