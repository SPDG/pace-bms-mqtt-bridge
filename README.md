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
go run ./cmd/pace-bms-probe --port /dev/pace-bms --baud 9600 --protocol rs232 --first 255 --last 255
go run ./cmd/pace-bms-probe --port /dev/pace-bms --baud 115200 --protocol rs485 --first 0 --last 16
```

By default the web panel listens on `http://127.0.0.1:8080`.

## Configuration

See [`configs/config.example.yaml`](configs/config.example.yaml).

For Linux deployments, prefer a stable udev symlink such as `/dev/pace-bms` instead of raw `/dev/ttyUSB*`.

Serial-over-Ethernet converters can be used by setting `serial.port` to a TCP endpoint such as `tcp://192.168.6.134:4196`. The converter should run in transparent TCP server mode with the same serial parameters configured in `serial`.

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
