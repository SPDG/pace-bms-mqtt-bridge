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
