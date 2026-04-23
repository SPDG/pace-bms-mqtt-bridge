package pace

import "testing"

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
