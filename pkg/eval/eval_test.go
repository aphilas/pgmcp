package eval

import (
	"testing"
)

func TestEval(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    int
		wantErr bool
	}{
		{"integer literal", "42", 42, false},
		{"addition", "1 + 2", 3, false},
		{"subtraction", "10 - 3", 7, false},
		{"multiplication", "4 * 5", 20, false},
		{"division", "10 / 2", 5, false},
		{"nested expression", "(1 + 2) * 3", 9, false},
		{"operator precedence", "2 + 3 * 4", 14, false},
		{"complex expression", "(10 - 2) / (1 + 3)", 2, false},
		{"zero", "0", 0, false},
		{"parse error", "1 +", 0, true},
		{"identifier rejected", "abc", 0, true},
		{"division by zero", "1 / 0", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Eval(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got == nil {
				t.Fatal("expected non-nil result")
			}
			if *got != tt.want {
				t.Errorf("Eval(%q) = %d, want %d", tt.input, *got, tt.want)
			}
		})
	}
}
