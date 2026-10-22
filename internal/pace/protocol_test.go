package pace

import (
	"fmt"
	"testing"
)

func TestBuildRequestRS485Analog(t *testing.T) {
	got, err := BuildRequest(Request{Protocol: ProtocolRS485, Command: CommandAnalog, Pack: 1})
	if err != nil {
		t.Fatal(err)
	}
	want := "~25014642E00201FD30\r"
	if string(got) != want {
		t.Fatalf("request mismatch\nwant %q\n got %q", want, string(got))
	}
}

func TestBuildRequestRS232CurrentLimitParameter(t *testing.T) {
	got, err := BuildRequest(Request{Protocol: ProtocolRS232, Command: CommandCurrentLimit, Pack: 255})
	if err != nil {
		t.Fatal(err)
	}
	want := "~250046ED0000FD86\r"
	if string(got) != want {
		t.Fatalf("request mismatch\nwant %q\n got %q", want, string(got))
	}
}

func TestParsePackNumber(t *testing.T) {
	body := "~25004600E00200"
	got, err := ParsePackNumber([]byte(body + payloadChecksum([]byte(body)) + "\r"))
	if err != nil {
		t.Fatal(err)
	}
	if got != 0 {
		t.Fatalf("got %d", got)
	}
}

func TestParseCurrentLimitParameter(t *testing.T) {
	raw := frame([]byte{0x00, 0x64})
	got, err := ParseCurrentLimitParameter([]byte(raw), 255)
	if err != nil {
		t.Fatal(err)
	}
	if got.CurrentA != 100 {
		t.Fatalf("current limit = %.1f, want 100", got.CurrentA)
	}
	if got.Address != 1 {
		t.Fatalf("address = %d, want 1", got.Address)
	}
}

func TestParseAnalogTwoPacks(t *testing.T) {
	raw := "~2501460051280002100C9B0C9C0C9E0C9C0C9D0C9D0C9F0C9D0C9D0C9E0C9F0C9F0C9F0C9F0C9D0C9C060BC20BAE0B990B8C0BDC0BD6FBC4C9DC2CA6097C14002A714824714871487148714864CA120000100C950C970C980C9B0C9B0C970C980C9A0C980C990C8F0C980C950C990C9A0C95060B8C0B850B7C0B7C0B910B82FB9BC9782C8B037AA8000676C02476C076C076C076C064C9F70000BB68\r"
	packs, err := ParseAnalogPacks([]byte(raw), 255)
	if err != nil {
		t.Fatal(err)
	}
	if len(packs) != 2 {
		t.Fatalf("pack count = %d", len(packs))
	}
	for _, pack := range packs {
		if len(pack.CellsMV) != 16 {
			t.Fatalf("pack %d cell count = %d", pack.Address, len(pack.CellsMV))
		}
	}
}

func TestParseAnalogTwoPacksWhenFirstPackSOCLooksLikeCellCount(t *testing.T) {
	info := []byte{
		0x00, 0x02,
		0x10, 0x0C, 0x9B, 0x0C, 0x9C, 0x0C, 0x9E, 0x0C, 0x9C, 0x0C, 0x9D, 0x0C, 0x9D, 0x0C, 0x9F, 0x0C, 0x9D, 0x0C, 0x9D, 0x0C, 0x9E, 0x0C, 0x9F, 0x0C, 0x9F, 0x0C, 0x9F, 0x0C, 0x9F, 0x0C, 0x9D, 0x0C, 0x9C,
		0x06, 0x0B, 0xC2, 0x0B, 0xAE, 0x0B, 0x99, 0x0B, 0x8C, 0x0B, 0xDC, 0x0B, 0xD6,
		0xFB, 0xC4, 0xC9, 0xDC, 0x2C, 0xA6, 0x09, 0x7C, 0x14, 0x00, 0x2A, 0x71, 0x48,
		0x20, 0x71, 0x48, 0x71, 0x48, 0x71, 0x48, 0x71, 0x48, 0x64, 0xCA, 0x12, 0x00, 0x00,
		0x10, 0x0C, 0x95, 0x0C, 0x97, 0x0C, 0x98, 0x0C, 0x9B, 0x0C, 0x9B, 0x0C, 0x97, 0x0C, 0x98, 0x0C, 0x9A, 0x0C, 0x98, 0x0C, 0x99, 0x0C, 0x8F, 0x0C, 0x98, 0x0C, 0x95, 0x0C, 0x99, 0x0C, 0x9A, 0x0C, 0x95,
		0x06, 0x0B, 0x8C, 0x0B, 0x85, 0x0B, 0x7C, 0x0B, 0x7C, 0x0B, 0x91, 0x0B, 0x82,
		0xFB, 0x9B, 0xC9, 0x78, 0x2C, 0x8B, 0x03, 0x7A, 0xA8, 0x00, 0x06, 0x76, 0xC0,
		0x24, 0x76, 0xC0, 0x76, 0xC0, 0x76, 0xC0, 0x76, 0xC0, 0x64, 0xC9, 0xF7, 0x00, 0x00,
	}
	raw := analogFrame(info)
	packs, err := ParseAnalogPacks([]byte(raw), 255)
	if err != nil {
		t.Fatal(err)
	}
	if len(packs) != 2 {
		t.Fatalf("pack count = %d", len(packs))
	}
	if got := len(packs[0].CellsMV); got != 16 {
		t.Fatalf("pack 1 cell count = %d", got)
	}
	if got := len(packs[1].CellsMV); got != 16 {
		t.Fatalf("pack 2 cell count = %d", got)
	}
	if packs[0].SOC != 32 {
		t.Fatalf("pack 1 SOC = %v, want 32", packs[0].SOC)
	}
}

func TestParseAnalogRejectsImplausibleCellVoltage(t *testing.T) {
	info := []byte{
		0x00, 0x01,
		0x01, 0x71, 0x48,
		0x01, 0x0B, 0xC2,
		0x00, 0x00, 0xC9, 0xDC, 0x00, 0x00, 0x00, 0x01,
	}
	raw := analogFrame(info)
	if _, err := ParseAnalogPacks([]byte(raw), 255); err == nil {
		t.Fatal("expected implausible cell voltage error")
	}
}

func TestParseAnalogBroadcastSinglePackUsesPhysicalAddress(t *testing.T) {
	info := []byte{
		0x00, 0x01,
		0x01, 0x0C, 0xE4,
		0x01, 0x0B, 0xC2,
		0x00, 0x00, 0xCE, 0x4C, 0x00, 0x64, 0x00, 0x00, 0x64,
	}
	raw := analogFrame(info)
	packs, err := ParseAnalogPacks([]byte(raw), 255)
	if err != nil {
		t.Fatal(err)
	}
	if len(packs) != 1 {
		t.Fatalf("pack count = %d", len(packs))
	}
	if packs[0].Address != 1 {
		t.Fatalf("pack address = %d, want 1", packs[0].Address)
	}
}

func TestParseWarningPacks(t *testing.T) {
	info := []byte{
		0x00, 0x01,
		0x02, 0x00, 0x02,
		0x03, 0x00, 0x01, 0x00,
		0x00, 0x00, 0x02,
		0x70, 0xB0,
		0x27,
		0x39,
		0x07,
		0x12, 0x34,
		0x33, 0xF5,
		0x00,
	}
	raw := analogFrame(info)
	statuses, err := ParseWarningPacks([]byte(raw), 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(statuses) != 1 {
		t.Fatalf("status count = %d", len(statuses))
	}
	status := statuses[0]
	if !status.Instruction.ChargeEnabled || !status.Instruction.DischargeEnabled || !status.Instruction.ChargerAvailable {
		t.Fatalf("instruction bits not decoded: %+v", status.Instruction)
	}
	if !status.Protection.ShortCircuit || !status.Protection.FullyCharged || !status.Protection.HighEnvironmentTemp {
		t.Fatalf("protection bits not decoded: %+v", status.Protection)
	}
	if !status.Fault.ChargeMOS || !status.Fault.DischargeMOS || !status.Fault.NTC {
		t.Fatalf("fault bits not decoded: %+v", status.Fault)
	}
	if status.DischargeCurrentWarning != "above upper limit" {
		t.Fatalf("discharge current warning = %q", status.DischargeCurrentWarning)
	}
	if status.BalancingCellCount != 1 || len(status.CellsBalancing) != 2 || !status.CellsBalancing[1] {
		t.Fatalf("balancing cells not decoded: count=%d cells=%v", status.BalancingCellCount, status.CellsBalancing)
	}
}

func TestParseLiveBMSMetadataAndParameters(t *testing.T) {
	sw, err := ParseSoftwareVersion([]byte("~25014600602850313653333030412D33323033302D312E303320F58A\r"))
	if err != nil {
		t.Fatal(err)
	}
	if sw != "P16S300A-32030-1.03" {
		t.Fatalf("SoftwareVersion = %q, want P16S300A-32030-1.03", sw)
	}

	c6Info := []byte("P16S300A-32030-1.03-00131390V120\x00")
	paceSW, hwVer, err := ParsePaceSoftwareVersion([]byte(frame(c6Info)))
	if err != nil {
		t.Fatal(err)
	}
	if paceSW != "P16S300A-32030-1.03-001" || hwVer != "31390V120" {
		t.Fatalf("PaceSoftwareVersion = (%q, %q)", paceSW, hwVer)
	}

	c2Info := []byte("320301161690054P                        ")
	bmsSN, packSN, err := ParseProductInfoDetails([]byte(frame(c2Info)))
	if err != nil {
		t.Fatal(err)
	}
	if bmsSN != "320301161690054P" || packSN != "" {
		t.Fatalf("ProductInfoDetails = (%q, %q)", bmsSN, packSN)
	}

	dt, err := ParseBMSDateTime([]byte("~25014600400C1A0A16173200FB20\r"))
	if err != nil {
		t.Fatal(err)
	}
	if dt != "2026-10-22 23:50:00" {
		t.Fatalf("BMSDateTime = %q, want 2026-10-22 23:50:00", dt)
	}

	proto, err := ParseInverterProtocol([]byte("~25014600A006010D01FC61\r"), "EB")
	if err != nil {
		t.Fatal(err)
	}
	if proto.CANName != "PYLON_CAN" || proto.RS485Name != "WOW_485" || proto.ModeName != "manual" {
		t.Fatalf("InverterProtocol = %+v", proto)
	}

	eq, err := ParseEqualizationParameters([]byte("~2501460080080D48001EFBE8\r"))
	if err != nil {
		t.Fatal(err)
	}
	if eq.StartVoltageMV != 3400 || eq.DeltaVoltageMV != 30 {
		t.Fatalf("Equalization = %+v", eq)
	}

	fc, err := ParseFullChargeParameters([]byte(frame([]byte{0xDA, 0xC0, 0x13, 0x88, 0x05})))
	if err != nil {
		t.Fatal(err)
	}
	if fc.FullVoltageV != 56.0 || fc.CutOffCurrentA != 5.0 || fc.LowSOCAlarmPercent != 5 {
		t.Fatalf("FullCharge = %+v", fc)
	}

	sh, err := ParseSingleHibernateParameters([]byte(frame([]byte{0x0C, 0x4E, 0x00, 0x05})))
	if err != nil {
		t.Fatal(err)
	}
	if sh.SleepVoltageMV != 3150 || sh.DelayMinutes != 5 {
		t.Fatalf("SingleHibernate = %+v", sh)
	}

	cellOVP, err := ParseVoltageProtectionParameters([]byte("~25014600F010010E100E420D340AFA39\r"))
	if err != nil {
		t.Fatal(err)
	}
	if !cellOVP.Enabled || cellOVP.WarningMV != 3600 || cellOVP.ProtectMV != 3650 || cellOVP.RecoveryMV != 3380 || cellOVP.DelayMS != 1000 {
		t.Fatalf("CellOVP = %+v", cellOVP)
	}

	packOVP, err := ParsePackVoltageProtectionParameters([]byte(frame([]byte{0x01, 0xE1, 0x00, 0xE4, 0x20, 0xD2, 0xF0, 0x0A})))
	if err != nil {
		t.Fatal(err)
	}
	if !packOVP.Enabled || packOVP.WarningV != 57.6 || packOVP.ProtectV != 58.4 || packOVP.RecoveryV != 54.0 || packOVP.DelayMS != 1000 {
		t.Fatalf("PackOVP = %+v", packOVP)
	}

	chgOCP, err := ParseCurrentProtectionParameters([]byte(frame([]byte{0x01, 0x01, 0x31, 0x01, 0x36, 0x0A})), false)
	if err != nil {
		t.Fatal(err)
	}
	if !chgOCP.Enabled || chgOCP.WarningA != 305 || chgOCP.ProtectA != 310 || chgOCP.DelayMS != 1000 {
		t.Fatalf("ChargeOCP = %+v", chgOCP)
	}

	dsgOCP1, err := ParseCurrentProtectionParameters([]byte(frame([]byte{0x01, 0xFE, 0xCF, 0xFE, 0xCA, 0x0A})), true)
	if err != nil {
		t.Fatal(err)
	}
	if !dsgOCP1.Enabled || dsgOCP1.WarningA != 305 || dsgOCP1.ProtectA != 310 || dsgOCP1.DelayMS != 1000 {
		t.Fatalf("DischargeOCP1 = %+v", dsgOCP1)
	}

	dsgOCP2, err := ParseDischargeOCP2Parameters([]byte(frame([]byte{0x01, 0x90, 0x14, 0x01, 0x90, 0x14})))
	if err != nil {
		t.Fatal(err)
	}
	if dsgOCP2.ProtectA != 400 || dsgOCP2.DelayMS != 500 {
		t.Fatalf("DischargeOCP2 = %+v", dsgOCP2)
	}

	scpUS, err := ParseShortCircuitDelay([]byte(frame([]byte{0x0C})))
	if err != nil {
		t.Fatal(err)
	}
	if scpUS != 300 {
		t.Fatalf("ShortCircuitDelay = %d, want 300", scpUS)
	}

	chgHT, dsgHT, err := ParseDualTemperatureProtectionParameters([]byte(frame([]byte{0x01, 0x0D, 0x02, 0x0D, 0x34, 0x0C, 0xD0, 0x0D, 0x34, 0x0D, 0x66, 0x0D, 0x02})))
	if err != nil {
		t.Fatal(err)
	}
	if chgHT.WarningC != 60 || chgHT.ProtectC != 65 || chgHT.RecoveryC != 55 || dsgHT.WarningC != 65 || dsgHT.ProtectC != 70 || dsgHT.RecoveryC != 60 {
		t.Fatalf("HighTemp = (%+v, %+v)", chgHT, dsgHT)
	}

	mosHT, err := ParseSingleTemperatureProtectionParameters([]byte(frame([]byte{0x01, 0x0E, 0x2E, 0x0F, 0x28, 0x0D, 0xFC})))
	if err != nil {
		t.Fatal(err)
	}
	if mosHT.WarningC != 90 || mosHT.ProtectC != 115 || mosHT.RecoveryC != 85 {
		t.Fatalf("MOSHighTemp = %+v", mosHT)
	}
}

func analogFrame(info []byte) string {
	return frame(info)
}

func frame(info []byte) string {
	body := "~25014600" + lengthChecksum(hexLen(info)) + hexLen(info) + fmt.Sprintf("%X", info)
	return body + payloadChecksum([]byte(body)) + "\r"
}

func hexLen(info []byte) string {
	return fmt.Sprintf("%03X", len(info)*2)
}
