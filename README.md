# PACE BMS MQTT Bridge

Linux-first bridge for polling a PACE BMS over local serial communication and publishing battery telemetry to MQTT and Home Assistant.

## Goal

This project will follow the same general shape as `srne-inverter-to-mqtt`, but target PACE battery management systems:

- single Go binary,
- YAML-backed configuration,
- local serial polling,
- MQTT telemetry publishing,
- Home Assistant MQTT Discovery,
- optional embedded web status and configuration panel.

## Status

Initial repository scaffold. Protocol mapping and implementation are not started yet.

## License

Apache-2.0. See `LICENSE`.
