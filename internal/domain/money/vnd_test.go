package money

import "testing"

func TestVNDArithmetic(t *testing.T) {
	a := VND(100000)
	b := VND(50000)

	if got := a.Add(b); got != VND(150000) {
		t.Fatalf("expected 150000, got %d", got)
	}

	if got := a.Sub(b); got != VND(50000) {
		t.Fatalf("expected 50000, got %d", got)
	}

	if got := b.Mul(3); got != VND(150000) {
		t.Fatalf("expected 150000, got %d", got)
	}

	if a.Int64() != 100000 {
		t.Fatalf("expected 100000, got %d", a.Int64())
	}
}

func TestVNDFormat(t *testing.T) {
	tests := []struct {
		val  VND
		want string
	}{
		{VND(0), "0 ₫"},
		{VND(500), "500 ₫"},
		{VND(1000), "1.000 ₫"},
		{VND(100000), "100.000 ₫"},
		{VND(1500000), "1.500.000 ₫"},
		{VND(-30000), "-30.000 ₫"},
	}

	for _, tt := range tests {
		if got := tt.val.Format(); got != tt.want {
			t.Errorf("Format(%d) = %q, want %q", tt.val, got, tt.want)
		}
	}
}
