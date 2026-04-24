package pace

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

type Protocol string

const (
	ProtocolRS485   Protocol = "rs485"
	ProtocolRS232   Protocol = "rs232"
	ProtocolConsole Protocol = "console"
)

type Command string

const (
	CommandPackNumber      Command = "pack_number"
	CommandAnalog          Command = "analog"
	CommandWarningInfo     Command = "warning_info"
	CommandProductInfo     Command = "product_info"
	CommandSoftwareVersion Command = "software_version"
)

type Request struct {
	Protocol Protocol
	Command  Command
	Pack     uint8
}

func BuildRequest(req Request) ([]byte, error) {
	cid2, ok := commandCodes[req.Command]
	if !ok {
		return nil, fmt.Errorf("unknown command %q", req.Command)
	}
	lenID, ok := commandLengths[req.Command]
	if !ok {
		return nil, fmt.Errorf("missing length for command %q", req.Command)
	}

	addr := fmt.Sprintf("%02X", req.Pack)
	if req.Protocol == ProtocolRS232 {
		addr = "00"
	} else if req.Protocol == ProtocolConsole {
		addr = "01"
	}

	var frame bytes.Buffer
	frame.WriteByte('~')
	frame.WriteString("25")
	frame.WriteString(addr)
	frame.WriteString("46")
	frame.WriteString(cid2)
	frame.WriteString(lengthChecksum(lenID))
	frame.WriteString(lenID)
	if lenID != "000" {
		frame.WriteString(fmt.Sprintf("%02X", req.Pack))
	}
	frame.WriteString(payloadChecksum(frame.Bytes()))
	frame.WriteByte('\r')
	return frame.Bytes(), nil
}

func ReadFrame(r io.Reader, timeout time.Duration) ([]byte, error) {
	deadline := time.Now().Add(timeout)
	buf := make([]byte, 0, 512)
	one := make([]byte, 1)
	for time.Now().Before(deadline) {
		n, err := r.Read(one)
		if n > 0 {
			if one[0] == '\n' {
				continue
			}
			buf = append(buf, one[0])
			if expected, ok := expectedFrameLength(buf); ok && len(buf) >= expected {
				return bytes.TrimSpace(buf), nil
			}
			if one[0] == '\r' && !looksLikePaceHeader(buf) {
				return bytes.TrimSpace(buf), nil
			}
		}
		if err != nil {
			if len(buf) > 0 && isTimeoutish(err) {
				continue
			}
			if isTimeoutish(err) {
				continue
			}
			return nil, err
		}
	}
	if len(buf) > 0 {
		return bytes.TrimSpace(buf), nil
	}
	return nil, fmt.Errorf("timeout waiting for PACE frame")
}

func expectedFrameLength(buf []byte) (int, bool) {
	if len(buf) < 13 || buf[0] != '~' {
		return 0, false
	}
	infoChars, err := strconv.ParseInt(string(buf[10:13]), 16, 32)
	if err != nil {
		return 0, false
	}
	return 18 + int(infoChars), true
}

func looksLikePaceHeader(buf []byte) bool {
	if len(buf) < 13 || buf[0] != '~' {
		return false
	}
	_, err := strconv.ParseInt(string(buf[10:13]), 16, 32)
	return err == nil
}

func ParsePackNumber(response []byte) (uint8, error) {
	data, err := parseEnvelope(response)
	if err != nil {
		return 0, err
	}
	if len(data.Info) < 1 {
		return 0, fmt.Errorf("pack number response has empty data")
	}
	return data.Info[0], nil
}

func ParseAnalog(response []byte, pack uint8) (Pack, error) {
	packs, err := ParseAnalogPacks(response, pack)
	if err != nil {
		return Pack{}, err
	}
	if len(packs) == 0 {
		return Pack{}, fmt.Errorf("analog response contains no packs")
	}
	return packs[0], nil
}

func ParseAnalogPacks(response []byte, requestedPack uint8) ([]Pack, error) {
	data, err := parseEnvelope(response)
	if err != nil {
		return nil, err
	}
	fields := data.Info
	if len(fields) < 2 {
		return nil, fmt.Errorf("analog payload too short")
	}
	offset := 0
	infoFlag := fields[offset]
	offset++
	reportedPackCount := fields[offset]
	offset++
	if reportedPackCount == 0 {
		return nil, fmt.Errorf("analog response reports zero packs")
	}

	packs := make([]Pack, 0, int(reportedPackCount))
	for packIndex := 0; packIndex < int(reportedPackCount); packIndex++ {
		address := requestedPack
		if reportedPackCount > 1 {
			address = uint8(packIndex + 1)
		}
		p := Pack{Address: address, InfoFlag: infoFlag, ReportedPackCount: reportedPackCount, UpdatedAt: time.Now().UTC()}

		if offset >= len(fields) {
			return nil, fmt.Errorf("analog payload missing cell count")
		}
		cellCount := int(fields[offset])
		offset++
		if !plausibleCellCount(byte(cellCount)) {
			return nil, fmt.Errorf("analog payload has implausible cell count %d", cellCount)
		}
		p.CellsMV = make([]int, 0, cellCount)
		for i := 0; i < cellCount; i++ {
			value, ok := readU16(fields, &offset)
			if !ok {
				return nil, fmt.Errorf("analog payload ended in cell voltages")
			}
			if !plausibleCellMillivolts(value) {
				return nil, fmt.Errorf("analog payload has implausible cell voltage %d mV", value)
			}
			p.CellsMV = append(p.CellsMV, int(value))
		}

		if offset >= len(fields) {
			return nil, fmt.Errorf("analog payload missing temperature count")
		}
		tempCount := int(fields[offset])
		offset++
		if !plausibleTemperatureCount(byte(tempCount)) {
			return nil, fmt.Errorf("analog payload has implausible temperature count %d", tempCount)
		}
		p.TemperaturesC = make([]float64, 0, tempCount)
		for i := 0; i < tempCount; i++ {
			value, ok := readU16(fields, &offset)
			if !ok {
				return nil, fmt.Errorf("analog payload ended in temperatures")
			}
			p.TemperaturesC = append(p.TemperaturesC, round(float64(value)/10-273.15, 2))
		}

		currentRaw, ok := readU16(fields, &offset)
		if !ok {
			return nil, fmt.Errorf("analog payload missing current")
		}
		p.CurrentA = round(float64(int16(currentRaw))/100, 2)

		voltageRaw, ok := readU16(fields, &offset)
		if !ok {
			return nil, fmt.Errorf("analog payload missing voltage")
		}
		p.VoltageV = round(float64(voltageRaw)/1000, 2)
		p.PowerKW = round(p.VoltageV*p.CurrentA/1000, 4)

		remainingRaw, ok := readU16(fields, &offset)
		if !ok {
			return nil, fmt.Errorf("analog payload missing remaining capacity")
		}
		p.RemainingCapacityAh = round(float64(remainingRaw)/100, 2)

		if offset < len(fields) {
			p.DefinedNumberP = fields[offset]
			offset++
		}

		fullRaw, ok := readU16(fields, &offset)
		if !ok {
			return nil, fmt.Errorf("analog payload missing full capacity")
		}
		p.FullCapacityAh = round(float64(fullRaw)/100, 2)
		if p.FullCapacityAh > 0 {
			p.SOC = round(p.RemainingCapacityAh/p.FullCapacityAh*100, 1)
		}

		cycles, ok := readU16(fields, &offset)
		if ok {
			p.CycleCount = int(cycles)
		}
		designRaw, ok := readU16(fields, &offset)
		if ok {
			p.DesignCapacityAh = round(float64(designRaw)/100, 2)
			if p.DesignCapacityAh > 0 {
				p.SOH = round(p.FullCapacityAh/p.DesignCapacityAh*100, 0)
			}
		}
		parsedV2Tail := false
		if hasAnalogV2Tail(fields, offset, packIndex, int(reportedPackCount)) {
			p.SOC = float64(fields[offset])
			offset++
			offset += 8
			p.SOH = float64(fields[offset])
			offset++
			offset += 4
			parsedV2Tail = true
		}
		packs = append(packs, p)

		// Multi-pack RS232 responses include one trailing flag byte per pack.
		if !parsedV2Tail && offset < len(fields) {
			offset++
		}
		// Some firmwares include an additional INFOFLAG byte before the next pack.
		if !parsedV2Tail && packIndex+1 < int(reportedPackCount) && offset < len(fields) && fields[offset] != byte(cellCount) {
			offset++
		}
	}

	return packs, nil
}

func hasAnalogV2Tail(fields []byte, offset, packIndex, reportedPackCount int) bool {
	const tailLength = 14
	if offset+tailLength > len(fields) {
		return false
	}
	if packIndex+1 == reportedPackCount {
		return true
	}
	nextOffset := offset + tailLength
	return nextOffset < len(fields) && plausibleCellCount(fields[nextOffset])
}

func ParseProductInfo(response []byte) (string, error) {
	data, err := parseEnvelope(response)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data.Info)), nil
}

type envelope struct {
	Version string
	Address uint8
	CID1    uint8
	Return  uint8
	Info    []byte
}

func parseEnvelope(response []byte) (envelope, error) {
	raw := strings.TrimSpace(string(response))
	raw = strings.TrimPrefix(raw, "~")
	if len(raw) < 16 {
		return envelope{}, fmt.Errorf("short response %q", raw)
	}
	cid1, err := parseHexByte(raw[4:6])
	if err != nil {
		return envelope{}, fmt.Errorf("decode CID1: %w", err)
	}
	rtn, err := parseHexByte(raw[6:8])
	if err != nil {
		return envelope{}, fmt.Errorf("decode return code: %w", err)
	}
	if cid1 != 0x46 {
		return envelope{}, fmt.Errorf("unexpected CID1 0x%02X", cid1)
	}
	if rtn != 0x00 {
		return envelope{}, fmt.Errorf("BMS returned code 0x%02X", rtn)
	}
	lchk := raw[8:9]
	lenID := raw[9:12]
	if got := lengthChecksum(lenID); got != lchk {
		return envelope{}, fmt.Errorf("length checksum mismatch got %s want %s", lchk, got)
	}
	infoChars, err := strconv.ParseInt(lenID, 16, 32)
	if err != nil {
		return envelope{}, fmt.Errorf("decode LENID: %w", err)
	}
	infoStart := 12
	infoEnd := infoStart + int(infoChars)
	if len(raw) < infoEnd+4 {
		return envelope{}, fmt.Errorf("response shorter than declared info length %d", infoChars)
	}
	if got, want := raw[infoEnd:infoEnd+4], payloadChecksum([]byte("~"+raw[:infoEnd])); got != want {
		return envelope{}, fmt.Errorf("payload checksum mismatch got %s want %s", got, want)
	}
	infoASCII := raw[infoStart:infoEnd]
	if len(infoASCII)%2 != 0 {
		return envelope{}, fmt.Errorf("odd INFO hex length %d", len(infoASCII))
	}
	info, err := hex.DecodeString(infoASCII)
	if err != nil {
		return envelope{}, fmt.Errorf("decode INFO: %w", err)
	}
	address, err := parseHexByte(raw[2:4])
	if err != nil {
		return envelope{}, fmt.Errorf("decode address: %w", err)
	}
	return envelope{
		Version: raw[0:2],
		Address: address,
		CID1:    cid1,
		Return:  rtn,
		Info:    info,
	}, nil
}

func parseHexByte(value string) (uint8, error) {
	parsed, err := strconv.ParseUint(value, 16, 8)
	if err != nil {
		return 0, err
	}
	return uint8(parsed), nil
}

func readU16(data []byte, offset *int) (uint16, bool) {
	if *offset+1 >= len(data) {
		return 0, false
	}
	value := uint16(data[*offset])<<8 | uint16(data[*offset+1])
	*offset += 2
	return value, true
}

func peekU16(data []byte, offset int) uint16 {
	if offset+1 >= len(data) {
		return 0
	}
	return uint16(data[offset])<<8 | uint16(data[offset+1])
}

func plausibleCellMillivolts(value uint16) bool {
	return value >= 2000 && value <= 4500
}

func plausibleCellCount(value byte) bool {
	return value >= 1 && value <= 32
}

func plausibleTemperatureCount(value byte) bool {
	return value >= 1 && value <= 12
}

func lengthChecksum(lenID string) string {
	var sum int64
	for _, r := range lenID {
		n, _ := strconv.ParseInt(string(r), 16, 64)
		sum += n
	}
	return fmt.Sprintf("%X", ((^sum)&0xF)+1&0xF)
}

func payloadChecksum(data []byte) string {
	var sum int
	for _, b := range data[1:] {
		sum += int(b)
	}
	return fmt.Sprintf("%04X", ((^sum)&0xFFFF)+1&0xFFFF)
}

func isTimeoutish(err error) bool {
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "timeout") || strings.Contains(msg, "temporarily unavailable")
}

var commandCodes = map[Command]string{
	CommandPackNumber:      "90",
	CommandAnalog:          "42",
	CommandWarningInfo:     "44",
	CommandProductInfo:     "C2",
	CommandSoftwareVersion: "C1",
}

var commandLengths = map[Command]string{
	CommandPackNumber:      "000",
	CommandAnalog:          "002",
	CommandWarningInfo:     "002",
	CommandProductInfo:     "000",
	CommandSoftwareVersion: "000",
}
