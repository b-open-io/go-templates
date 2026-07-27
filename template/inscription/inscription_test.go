package inscription

import (
	"testing"

	"github.com/bsv-blockchain/go-sdk/script"
)

func TestDecode_InvalidScript(t *testing.T) {
	s := script.NewFromBytes([]byte{0x00, 0x51, 0x52}) // random bytes, not a valid inscription
	insc := Decode(s)
	if insc != nil {
		t.Errorf("expected nil for invalid script, got %+v", insc)
	}
}

func TestLock_Basic(t *testing.T) {
	insc := &Inscription{
		File: File{
			Type:    "text/plain",
			Content: []byte("hello world"),
		},
	}
	script, err := insc.Lock()
	if err != nil {
		t.Fatalf("Lock error: %v", err)
	}
	if script == nil || len(*script) == 0 {
		t.Error("expected non-empty script")
	}
}

func TestRoundTrip_LockDecode(t *testing.T) {
	insc := &Inscription{
		File: File{
			Type:    "text/plain",
			Content: []byte("round trip test content"),
		},
	}
	script, err := insc.Lock()
	if err != nil {
		t.Fatalf("Lock error: %v", err)
	}
	decoded := Decode(script)
	if decoded == nil {
		t.Fatalf("Decode failed, got nil")
	}
	if decoded.File.Type != insc.File.Type {
		t.Errorf("File.Type mismatch: got %q, want %q", decoded.File.Type, insc.File.Type)
	}
	if string(decoded.File.Content) != string(insc.File.Content) {
		t.Errorf("File.Content mismatch: got %q, want %q", string(decoded.File.Content), string(insc.File.Content))
	}
}

// TestDecode_OrdPushNearScriptStart is the regression test for the bounds bug
// fixed in be91c94: the guard checked `pos >= 2` while the body indexed
// `(*scr)[startI-2]`. ReadOp advances pos past startI, so a 3-byte "ord" push at
// or near offset 0 satisfied the guard with startI < 2 and panicked with
// "index out of range [-1]".
//
// Decode is fed arbitrary on-chain scripts from untrusted senders, so a
// malformed envelope must return nil, never panic. Revert the guard to
// `pos >= 2` and every subtest here panics.
func TestDecode_OrdPushNearScriptStart(t *testing.T) {
	ordPush := []byte{script.OpDATA3, 'o', 'r', 'd'}

	tests := []struct {
		name   string
		script []byte
	}{
		{
			// startI == 0: the push is the very first op.
			name:   "ord push at offset 0",
			script: ordPush,
		},
		{
			// startI == 1: one byte of lead-in, still under the 2 the body needs.
			name:   "ord push at offset 1",
			script: append([]byte{script.Op0}, ordPush...),
		},
		{
			// startI == 1 with OP_IF immediately before — the shape closest to a
			// real inscription while still short of the OP_0 OP_IF prefix.
			name:   "ord push after a bare OP_IF",
			script: append([]byte{script.OpIF}, ordPush...),
		},
		{
			// Trailing data, so the decoder would carry on past the push.
			name:   "ord push at offset 0 with trailing data",
			script: append(append([]byte{}, ordPush...), script.Op1, script.Op2),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := script.NewFromBytes(tc.script)
			// Decode must return, not panic. nil is the correct answer: none of
			// these carry the OP_0 OP_IF prefix a real inscription needs.
			if insc := Decode(s); insc != nil {
				t.Errorf("expected nil for a script with no OP_0 OP_IF prefix, got %+v", insc)
			}
		})
	}
}

// A well-formed inscription must still decode — the guard fix tightened a bounds
// check and must not have narrowed what Decode accepts.
func TestDecode_ValidInscriptionStillDecodes(t *testing.T) {
	insc := &Inscription{
		File: File{
			Type:    "text/plain",
			Content: []byte("still decodes"),
		},
	}
	locked, err := insc.Lock()
	if err != nil {
		t.Fatalf("Lock error: %v", err)
	}
	got := Decode(locked)
	if got == nil {
		t.Fatal("expected a valid inscription to decode, got nil")
	}
	if string(got.File.Content) != "still decodes" {
		t.Errorf("content = %q, want %q", got.File.Content, "still decodes")
	}
}
