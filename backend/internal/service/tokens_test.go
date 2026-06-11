package service

import (
	"strings"
	"testing"
)

func TestNewRoomCode(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 1000; i++ {
		code, err := newRoomCode()
		if err != nil {
			t.Fatalf("newRoomCode() error: %v", err)
		}
		if len(code) != roomCodeLength {
			t.Fatalf("len(code) = %d, want %d", len(code), roomCodeLength)
		}
		for _, r := range code {
			if !strings.ContainsRune(roomCodeAlphabet, r) {
				t.Fatalf("code %q contains char %q outside the alphabet", code, r)
			}
		}
		seen[code] = true
	}
	// Not a strict guarantee, but 1000 6-char codes from a 31-char alphabet
	// colliding heavily would signal a broken RNG.
	if len(seen) < 990 {
		t.Errorf("only %d unique codes out of 1000 — RNG looks weak", len(seen))
	}
}

func TestNewToken(t *testing.T) {
	tok, err := newToken("student")
	if err != nil {
		t.Fatalf("newToken() error: %v", err)
	}
	if !strings.HasPrefix(tok, "student_") {
		t.Errorf("token %q missing prefix", tok)
	}
	// "student_" + 24 bytes hex-encoded = 8 + 48.
	if want := len("student_") + 48; len(tok) != want {
		t.Errorf("len(token) = %d, want %d", len(tok), want)
	}

	other, _ := newToken("student")
	if tok == other {
		t.Error("two tokens collided — RNG looks broken")
	}
}
