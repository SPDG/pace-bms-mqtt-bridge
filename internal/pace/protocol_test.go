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

func analogFrame(info []byte) string {
	body := "~25014600" + lengthChecksum(hexLen(info)) + hexLen(info) + fmt.Sprintf("%X", info)
	return body + payloadChecksum([]byte(body)) + "\r"
}

func hexLen(info []byte) string {
	return fmt.Sprintf("%03X", len(info)*2)
}
