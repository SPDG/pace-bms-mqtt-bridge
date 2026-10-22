# PACE BMS Serial Protocol Reference (`HS_PACE_232_BP_V29` / `RS485`)

This document describes the PACE ASCII serial protocol used by `pace-bms-mqtt-bridge`, including both the **implemented read-only commands** and the **documented writable/control commands** (analyzed from `PBmsTools V3.0.13` for future write support).

---

## 1. Frame Format

All requests and responses are ASCII hexadecimal frames delimited by `~` (`0x7E`) and `\r` (`0x0D`):

```text
~ [VER:2] [ADR:2] [CID1:2] [CID2/RTN:2] [LENGTH:4] [INFO:LENID] [CHKSUM:4] \r
```

| Field | Hex Chars | Description |
| :--- | :---: | :--- |
| `SOI` | 1 (`~`) | Start of frame (`0x7E`) |
| `VER` | 2 | Protocol version (`25` for RS232 / RS485 V2.5, `20` on some older units) |
| `ADR` | 2 | Pack address (`00` or `01` on RS232, `01`..`10` or `FF` broadcast on RS485) |
| `CID1` | 2 | Device type (`46` = Lithium iron phosphate BMS) |
| `CID2` / `RTN` | 2 | Command ID (`CID2`) in requests; Return Code (`RTN`, `00` = OK) in responses |
| `LENGTH` | 4 | `[LCHKSUM:1][LENID:3]` where `LENID` is the number of ASCII hex characters in `INFO` |
| `INFO` | `LENID` | Uppercase hex-encoded payload bytes (`2 * byte_count` ASCII chars) |
| `CHKSUM` | 4 | 16-bit two's complement checksum of all ASCII bytes from `VER` through `INFO` |
| `EOI` | 1 (`\r`) | End of frame (`0x0D`) |

### Checksum Formulas

1. **Length Checksum (`LCHKSUM`, 1 hex nibble)**:
   - Sum the three 4-bit nibbles of `LENID`: `sum = (n0 + n1 + n2) & 0xF`
   - `LCHKSUM = ((^sum) + 1) & 0xF`
   - *Note*: When `LENID == "000"`, `sum == 0` and `LCHKSUM == 0` (`"0000"`).
2. **Payload Checksum (`CHKSUM`, 4 hex chars)**:
   - Sum the ASCII byte values of `VER + ADR + CID1 + CID2/RTN + LENGTH + INFO`
   - `CHKSUM = ((^sum) + 1) & 0xFFFF`

---

## 2. Implemented Read-Only Commands

### 2.1 Fast & Status Polling

| `CID2` | Name | Request `INFO` | Response Payload Layout |
| :---: | :--- | :---: | :--- |
| `0x42` | Read Pack Analog Telemetry | `FF` or `[packAddr]` | `[infoFlag:1][packCount:1]` followed by per-pack analog blocks: `[cellCount:1][cellMv:2 * N][tempCount:1][tempRaw:2 * M][currentRaw:2 (int16 * 0.01A)][voltageRaw:2 (uint16 * 0.01V)][remainCap:2 (uint16 * 0.01Ah)][defineP:1][fullCap:2 (uint16 * 0.01Ah)][cycleCount:2][designCap:2 (uint16 * 0.01Ah)` *(when `defineP >= 3`)*`][soc:1` *(when `defineP >= 5`)*`][soh:1` *(when `defineP >= 8`)*`]`. Temperatures are `(raw - 2731) / 10.0 °C`. |
| `0x44` | Read Pack Warnings & Status | `FF` or `[packAddr]` | `[infoFlag:1][packCount:1]` followed by per-pack blocks: `[cellCount:1][cellWarn:1 * N][tempCount:1][tempWarn:1 * M][chgCurrWarn:1][totalVoltWarn:1][dsgCurrWarn:1][protect1:1][protect2:1][instruction:1][control:1][fault:1][balance1:1][balance2:1][warn1:1][warn2:1]`. `balance1` (cells 1–8) and `balance2` (cells 9–16) bitmasks indicate active cell balancing. |
| `0x90` | Read Pack Number | `[addr]` | `[uint16 packNumber]` |
| `0xB1` | Read BMS RTC Date/Time | *(empty)* | 6 bytes: `[year-2000:1][month:1][day:1][hour:1][minute:1][second:1]` |
| `0xED` | Read Current Limit Parameter | *(empty)* | 2 bytes: `[uint16 currentLimitA]` |

### 2.2 Device Metadata & Inverter Protocol

| `CID2` | Name | Request `INFO` | Response Payload Layout |
| :---: | :--- | :---: | :--- |
| `0xC1` | Read Software Version | *(empty)* | 20 bytes ASCII software version (e.g. `P16S300A-32030-1.03`) |
| `0xC6` | Read PACE SW & HW Version | *(empty)* | 40 bytes ASCII: bytes `0..19` = PACE Software Version, bytes `20..39` = Hardware Version (`31390V120`) |
| `0xC2` | Read Serial Numbers | *(empty)* | 20 or 40 bytes ASCII: bytes `0..19` = BMS Serial Number, bytes `20..39` = Pack Serial Number |
| `0xEB` / `0xEC` / `0x62` | Read Inverter Protocol | *(empty)* | `0xEB`: `[mode:1][canCode:1][rs485Code:1]` (`mode`: `0`=auto, `1`=manual, `2`=can, `3`=rs485)<br>`0xEC` / `0x62`: `[canCode:1][rs485Code:1]` |

### 2.3 Protection, Equalization & Hibernate Parameters (Read-Only)

| `CID2` | Name | Request `INFO` | Response Payload Layout |
| :---: | :--- | :---: | :--- |
| `0xB6` | Read Equalization Parameters | *(empty)* | 4 bytes: `[uint16 startVoltageMv][uint16 deltaVoltageMv]` |
| `0xAF` | Read Full Charge & Low SOC Alarm | *(empty)* | 5 bytes: `[uint16 fullVoltage (0.01 V)][uint16 cutOffCurrent (0.01 A)][uint8 lowSocAlarmPercent]` |
| `0xA0` | Read Single Hibernate Parameter | *(empty)* | 4 bytes: `[uint16 sleepVoltageMv][uint16 delayMinutes]` |
| `0xA6` | Read Pack Capacity Config | *(empty)* | 6 bytes: `[uint16 remainCap (0.1 Ah)][uint16 fullCap (0.1 Ah)][uint16 designCap (0.1 Ah)]` |
| `0xD1` | Read Cell OVP Thresholds | *(empty)* | 8 bytes: `[enable:1][uint16 warnMv][uint16 protectMv][uint16 recoveryMv][uint8 delay (100 ms)]` (`enable`: `0`=enabled, `1`=disabled) |
| `0xD3` | Read Cell UVP Thresholds | *(empty)* | 8 bytes: `[enable:1][uint16 warnMv][uint16 protectMv][uint16 recoveryMv][uint8 delay (100 ms)]` |
| `0xD5` | Read Pack OVP Thresholds | *(empty)* | 8 bytes: `[enable:1][uint16 warn (0.01 V)][uint16 protect (0.01 V)][uint16 recovery (0.01 V)][uint8 delay (100 ms)]` |
| `0xD7` | Read Pack UVP Thresholds | *(empty)* | 8 bytes: `[enable:1][uint16 warn (0.01 V)][uint16 protect (0.01 V)][uint16 recovery (0.01 V)][uint8 delay (100 ms)]` |
| `0xD9` | Read Charge OCP Thresholds | *(empty)* | 6 bytes: `[enable:1][uint16 warnA][uint16 protectA][uint8 delay (100 ms)]` |
| `0xDB` | Read Discharge OCP1 Thresholds | *(empty)* | 6 bytes: `[enable:1][uint16 warnA][uint16 protectA][uint8 delay (100 ms)]` |
| `0xE3` | Read Discharge OCP2 Thresholds | *(empty)* | 3 bytes: `[uint16 protectA][uint8 delay (100 ms)]` |
| `0xE5` | Read Short Circuit Delay | *(empty)* | 1 byte: `[uint8 delayUs]` (in `PBmsTools`, raw value `< 50` is multiplied by `20 µs`; otherwise raw `µs`) |
| `0xDD` | Read High Temp Thresholds (CHG/DSG) | *(empty)* | 13 bytes: `[enable:1][uint16 chgWarn][uint16 chgProtect][uint16 chgRecovery][uint16 dsgWarn][uint16 dsgProtect][uint16 dsgRecovery]` where `°C = (raw - 2731) / 10.0` |
| `0xDF` | Read Low Temp Thresholds (CHG/DSG) | *(empty)* | 13 bytes: `[enable:1][uint16 chgWarn][uint16 chgProtect][uint16 chgRecovery][uint16 dsgWarn][uint16 dsgProtect][uint16 dsgRecovery]` where `°C = (raw - 2731) / 10.0` |
| `0xE1` | Read MOS High Temp Thresholds | *(empty)* | 7 bytes: `[enable:1][uint16 warn][uint16 protect][uint16 recovery]` where `°C = (raw - 2731) / 10.0` |
| `0xE7` | Read Ambient (ENV) Temp Thresholds | *(empty)* | 13 bytes: `[enable:1][uint16 lowWarn][uint16 lowProtect][uint16 lowRecovery][uint16 highWarn][uint16 highProtect][uint16 highRecovery]` where `°C = (raw - 2731) / 10.0` |

---

## 3. Writable & Control Commands (Documented Only — Not Enabled Yet)

> **Note:** Write commands are intentionally disabled in the current bridge release. All commands below were extracted from `PBmsTools V3.0.13` (`HS_PACE_232_BP_V29`) for future implementation.

### 3.1 Switch & Control Commands

| `CID2` | Method in `PBmsTools` | Request Payload (`INFO` bytes) | Description |
| :---: | :--- | :--- | :--- |
| `0x99` | `ControlCommand` | 1 byte (`[uint8 controlCode]`) | Executes a control switch action (see table 3.2 below). |
| `0x9A` | `ChargingMOSFETControl` | 1 byte (`0x00` = Enable ON, `0x01` = Disable OFF) | Controls the Charge MOSFET state (`PACKCOMMAND = (enable == false) ? 1 : 0`). |
| `0x9B` | `DisChargingMOSFETControl` | 1 byte (`0x00` = Enable ON, `0x01` = Disable OFF) | Controls the Discharge MOSFET state (`PACKCOMMAND = (enable == false) ? 1 : 0`). |
| `0x9C` | `Hibernate` (`ForcedDormancy`) | 0 bytes (empty `INFO`) | Forces the BMS into sleep / standby mode. |
| `0xA4` | `ResetDefault` | 0 bytes (empty `INFO`) | Resets BMS protection/configuration parameters to factory defaults. |
| `0xB3` | `ClearHistoryRecord` | 0 bytes (empty `INFO`) | Clears internal BMS history log records. |

### 3.2 `0x99` (`ControlCommand`) Sub-Codes

| Sub-Code (`uint8`) | Enum / Handler Name | Meaning |
| :---: | :--- | :--- |
| `0x04` (4) | `SystemLockDisable` / `FireExSwitch(off)` | Disable System Lock (or Fire Extinguisher switch off on supported models) |
| `0x05` (5) | `SystemLockEnable` / `FireExSwitch(on)` | Enable System Lock (or Fire Extinguisher switch on) |
| `0x06` (6) | `LEDAlarmEnable` | Enable LED warning indicator function (`ControlState` bit 0) |
| `0x07` (7) | `LEDAlarmDisable` | Disable LED warning indicator function |
| `0x08` (8) | `SelectHighLimitGears` | Select High Current Limit Gear (`ControlState` bit 2 = 0) |
| `0x09` (9) | `SelectLowLimitGears` | Select Low Current Limit Gear (`ControlState` bit 2 = 1) |
| `0x0A` (10) | `LimitEnable` | Enable Charge Current Limit function (`ControlState` bit 1) |
| `0x0B` (11) | `LimitDisable` | Disable Charge Current Limit function |
| `0x0C` (12) | `BeepAlarmEnable` | Enable Buzzer audible alarm function (`ControlState` bit 3) |
| `0x0D` (13) | `BeepAlarmDisable` | Disable Buzzer audible alarm function |
| `0x0E` (14) | `HeaterSwitch(off)` / `CommCHGSwitch(off)` | Turn off Heater / CommCHG switch |
| `0x0F` (15) | `HeaterSwitch(on)` / `CommCHGSwitch(on)` | Turn on Heater / CommCHG switch |
| `0x10` (16) | `VolDiffCHGSwitch(off)` | Turn off Voltage-Difference Charge switch |
| `0x11` (17) | `VolDiffCHGSwitch(on)` | Turn on Voltage-Difference Charge switch |
| `0x13` (19) | `CommLock_Enable` / `Soldershts1Switch(off)` | Enable Communication Lock / Soldershts1 off |
| `0x14` (20) | `CommLock_Disable` / `Soldershts1Switch(on)` | Disable Communication Lock / Soldershts1 on |
| `0x15` (21) | `Soldershts2_Enable` | Soldershts2 switch off/enable |
| `0x16` (22) | `Soldershts2_Disable` | Soldershts2 switch on/disable |

---

### 3.3 Parameter Write Commands

All multi-byte integer fields are **big-endian** (`uint16`). Enable flags use `0x00` = protection enabled, `0x01` = protection disabled. Temperatures are encoded as `uint16((tempC * 10) + 2731)`.

| Write `CID2` | Read `CID2` | Method in `PBmsTools` | Payload Size | Payload Byte Layout |
| :---: | :---: | :--- | :---: | :--- |
| `0xB5` | `0xB6` | `SetEqualizationParameters` | 4 B | `[uint16 startVoltageMv][uint16 deltaVoltageMv]` *(Note: passive balancing activates only during charging when `maxCell >= startVoltageMv` and `maxCell - minCell >= deltaVoltageMv`)* |
| `0xAE` | `0xAF` | `SetFullParametersAndBatteryAlarm` | 5 B | `[uint16 fullVoltage (0.01 V)][uint16 cutOffCurrent (0.01 A)][uint8 lowSocAlarmPercent]` *(Note: in `HS_PACE_232_BP_V29`, `PBmsTools` passes `cutOffCurrent` in mA or `0.01 A` depending on UI unit)* |
| `0xA8` | `0xA0` | `SetSingleHibernateParameter` | 4 B | `[uint16 sleepVoltageMv][uint16 delayMinutes]` |
| `0xEE` | `0xED` | `SetCurrentLimitParameters` | 2 B | `[uint16 currentLimitA]` |
| `0xEC` / `0x63` | `0xEB` / `0x62` | `SetInverterProtocal` | 2 or 3 B | `0xEC` (`HS_PACE_232_BP_V29` override): `[mode:1][canCode:1][rs485Code:1]` (or 2 B `[canCode:1][rs485Code:1]` on `0x63`) |
| `0xD0` | `0xD1` | `SetSingleOverChargeParameters` (Cell OVP) | 8 B | `[enable:1][uint16 warnMv][uint16 protectMv][uint16 recoveryMv][uint8 delay (100 ms)]` |
| `0xD2` | `0xD3` | `SetSingleOverDisChargeParameters` (Cell UVP) | 8 B | `[enable:1][uint16 warnMv][uint16 protectMv][uint16 recoveryMv][uint8 delay (100 ms)]` |
| `0xD4` | `0xD5` | `SetTotalOverChargeParameters` (Pack OVP) | 8 B | `[enable:1][uint16 warn (0.01 V)][uint16 protect (0.01 V)][uint16 recovery (0.01 V)][uint8 delay (100 ms)]` |
| `0xD6` | `0xD7` | `SetTotalOverDisChargeParameters` (Pack UVP) | 8 B | `[enable:1][uint16 warn (0.01 V)][uint16 protect (0.01 V)][uint16 recovery (0.01 V)][uint8 delay (100 ms)]` |
| `0xD8` | `0xD9` | `SetChargeOverCurrentParameters` (CHG OCP) | 6 B | `[enable:1][uint16 warnA][uint16 protectA][uint8 delay (100 ms)]` |
| `0xDA` | `0xDB` | `SetDisChargeOverCurrent1Parameters` (DSG OCP1) | 6 B | `[enable:1][uint16 warnA][uint16 protectA][uint8 delay (100 ms)]` |
| `0xE2` | `0xE3` | `SetDisChargeOverCurrent2Parameters` (DSG OCP2) | 3 B | `[uint16 protectA][uint8 delay (100 ms)]` |
| `0xE4` | `0xE5` | `SetShortCircuitDelayParameters` (SCP) | 1 B | `[uint8 delayUs]` (`delayUs / 20` on `HS_PACE_232_BP_V29` when scaled) |
| `0xDC` | `0xDD` | `SetBatteryHighTemperatureParameters` | 13 B | `[enable:1][uint16 chgWarnK10][uint16 chgProtectK10][uint16 chgRecoveryK10][uint16 dsgWarnK10][uint16 dsgProtectK10][uint16 dsgRecoveryK10]` |
| `0xDE` | `0xDF` | `SetBatteryLowTemperatureParameters` | 13 B | `[enable:1][uint16 chgWarnK10][uint16 chgProtectK10][uint16 chgRecoveryK10][uint16 dsgWarnK10][uint16 dsgProtectK10][uint16 dsgRecoveryK10]` |
| `0xE0` | `0xE1` | `SetMOSFETHighTemperatureParameters` | 7 B | `[enable:1][uint16 warnK10][uint16 protectK10][uint16 recoveryK10]` |
| `0xE6` | `0xE7` | `SetAmbientTemperatureParameters` (ENV) | 13 B | `[enable:1][uint16 lowWarnK10][uint16 lowProtectK10][uint16 lowRecoveryK10][uint16 highWarnK10][uint16 highProtectK10][uint16 highRecoveryK10]` |

---

### 3.4 System & Calibration Write Commands

| Write `CID2` | Read `CID2` | Method in `PBmsTools` | Payload Size | Payload Byte Layout |
| :---: | :---: | :--- | :---: | :--- |
| `0xB2` | `0xB1` | `SetBMSDateTime` | 6 B | `[uint8 year-2000][uint8 month][uint8 day][uint8 hour][uint8 minute][uint8 second]` |
| `0xA7` | `0xA6` | `SetPACKCapacityInfo` | 6 B | `[uint16 remainCap (0.1 Ah)][uint16 fullCap (0.1 Ah)][uint16 designCap (0.1 Ah)]` |
| `0xA5` | — | `SetCapacityReleasePercentage` | 1 B | `[uint8 percent]` (`1..96%`, intermittent charge limit) |
| `0xB4` | — | `SetNumberOfBattery` | 1 B | `[uint8 packNumber]` |
| `0xBD` | — | `SetNumberOfCharges` | 2 B | `[uint16 cycleCount]` |
| `0xC3` | `0xC2` | `SetBMSProductInformation` | 20 B | 20-byte ASCII BMS Serial Number (right-padded with `0x00` or spaces) |
| `0xC4` | `0xC2` | `SetPACKProductInformation` | 20 B | 20-byte ASCII Pack Serial Number |
| `0xA9` | — | `CalibrationReferenceVoltage` | 2 B | `[uint16 referenceVoltageMv]` |
| `0xAD` | — | `CalibrationTotalVoltage` | 2 B | `[uint16 totalVoltage (0.01 V or mV)]` |
| `0xAC` | — | `CalibrationChargingCurrent` | 2 B | `[uint16 chargeCurrent (0.1 A)]` |
| `0xAA` | — | `CurrentZeroCalibration` | 2 B | `[uint16 zeroCurrent (0.1 A)]` (`0x0000`) |
| `0xAB` | — | `CalibrationDisChargingCurrent` | 2 B | `[uint16 dischargeCurrent (0.1 A)]` |
