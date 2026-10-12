package argon2

import (
	"testing"
)

func TestArgon2Hasher(t *testing.T) {
	hasher := New()

	password := "CorrectHorseBatteryStaple123!"
	hash, err := hasher.Hash(password)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	if !hasher.Verify(password, hash) {
		t.Fatalf("expected password verification to succeed")
	}

	if hasher.Verify("WrongPassword!", hash) {
		t.Fatalf("expected wrong password verification to fail")
	}

	if hasher.Verify(password, "invalid-hash-string") {
		t.Fatalf("expected invalid hash verification to fail")
	}
}
