package blurb

import "testing"

func TestNormalizeTrimsAndValidates(t *testing.T) {
	got, err := Normalize("  speedy  ")
	if err != nil {
		t.Fatalf("Normalize returned error: %v", err)
	}
	if got != "speedy" {
		t.Fatalf("Normalize = %q, want %q", got, "speedy")
	}
}

func TestNormalizeRejectsLong(t *testing.T) {
	long := make([]byte, 129)
	for i := range long {
		long[i] = 'a'
	}

	if _, err := Normalize(string(long)); err == nil {
		t.Fatal("expected error for long blurb")
	}
}

func TestRandomReturnsDefault(t *testing.T) {
	seen := map[string]bool{}
	for i := 0; i < 10; i++ {
		seen[Random()] = true
	}

	if len(seen) == 0 {
		t.Fatal("Random returned empty values")
	}
}
