# Deployment

Build the binary:

```bash
go build -o dist/pace-bms-mqtt-bridge ./cmd/pace-bms-mqtt-bridge
```

Install on Linux:

```bash
sudo install -d /opt/pace-bms-mqtt-bridge /etc/pace-bms-mqtt-bridge
sudo install -m 0755 dist/pace-bms-mqtt-bridge /opt/pace-bms-mqtt-bridge/
sudo install -m 0644 configs/config.example.yaml /etc/pace-bms-mqtt-bridge/config.yaml
sudo install -m 0644 deploy/systemd/pace-bms-mqtt-bridge.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable --now pace-bms-mqtt-bridge
```

Prefer a stable udev symlink such as `/dev/pace-bms` instead of raw `/dev/ttyUSB*`.
