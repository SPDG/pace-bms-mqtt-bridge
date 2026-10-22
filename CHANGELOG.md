# Changelog

## v0.1.0 - 2026-10-22

- Fixed PACE ASCII `LENID` length checksum calculation for zero-length payloads (`LENID=000`).
- Added cell balancing bitmask decoding (`CellsBalancing`, `BalancingCellCount`) and extended status flags (`ChargerHighVoltage`, `HeaterOn`, `SystemLock`, `Heater`, `CurrentLimit`) in `0x44`.
- Added read-only polling and parsing for BMS metadata: firmware/hardware versions (`0xC1`, `0xC6`), BMS/Pack serial numbers (`0xC2`), RTC date/time (`0xB1`), and CAN/RS485 inverter protocol selection (`0xEB`/`0xEC`/`0x62`).
- Added read-only polling and parsing for BMS protection & equalization parameters: current limit (`0xED`), equalization (`0xB6`), full-charge & low-SOC alarm (`0xAF`), single hibernate (`0xA0`), cell/pack OVP & UVP (`0xD1`, `0xD3`, `0xD5`, `0xD7`), charge/discharge OCP & short-circuit delay (`0xD9`, `0xDB`, `0xE3`, `0xE5`), and cell/MOS/ambient temperature thresholds (`0xDD`, `0xDF`, `0xE1`, `0xE7`).
- Exposed new read-only sensors via MQTT and Home Assistant discovery, and updated the Web UI (`Live data` and `Controls` tabs).
- Added responsive mobile navigation bar with collapsible hamburger menu in the Web UI.
- Added full protocol and writable command reference in `docs/PROTOCOL.md`.

## v0.0.1 - 2026-04-24

- Initial read-only PACE BMS bridge for RS232/RS485 serial telemetry.
- MQTT state publishing and Home Assistant MQTT Discovery.
- Embedded UniFi-inspired HTTP status dashboard.
- Linux systemd deployment example and release artifacts.
