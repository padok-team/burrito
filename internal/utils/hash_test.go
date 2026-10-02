package utils

import "testing"

func TestShortHashIsDeterministicAndBounded(t *testing.T) {
	if ShortHash("a") != ShortHash("a") {
		t.Fatal("expected ShortHash to be deterministic for the same input")
	}
	if ShortHash("a") == ShortHash("b") {
		t.Fatal("expected ShortHash to differ for different inputs")
	}
	if got := len(ShortHash("a")); got != 16 {
		t.Fatalf("expected ShortHash length 16, got %d", got)
	}
}
