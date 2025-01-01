package grpc

import (
	"testing"
)

func TestWireMessageRoundtrip(t *testing.T) {
	msg := NewWireMessage()
	msg.SetString(1, "kestrel-run-4482")
	msg.SetVarint(2, 42)
	msg.SetBytes(3, []byte{0xDE, 0xAD, 0xBE, 0xEF})

	encoded := msg.Encode()
	if len(encoded) == 0 {
		t.Fatal("encoded message is empty")
	}

	decoded, err := DecodeWireMessage(encoded)
	if err != nil {
		t.Fatalf("failed to decode message: %v", err)
	}

	if decoded.GetString(1) != "kestrel-run-4482" {
		t.Errorf("tag 1 expected kestrel-run-4482, got %s", decoded.GetString(1))
	}

	v2, err := decoded.GetVarint(2)
	if err != nil || v2 != 42 {
		t.Errorf("tag 2 expected 42, got %d (err: %v)", v2, err)
	}

	bytes3 := decoded.GetBytes(3)
	if len(bytes3) != 4 || bytes3[0] != 0xDE || bytes3[3] != 0xEF {
		t.Errorf("tag 3 bytes mismatch: %v", bytes3)
	}
}
