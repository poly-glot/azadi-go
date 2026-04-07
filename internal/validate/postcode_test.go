package validate

import "testing"

func TestUKPostcode(t *testing.T) {
	tests := []struct {
		input string
		valid bool
	}{
		{"SW1A 1AA", true},
		{"M1 1AE", true},
		{"B1 1BB", true},
		{"LS1 1BA", true},
		{"EH1 1YZ", true},
		{"sw1a1aa", true},
		{"123", false},
		{"", false},
		{"INVALID", false},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := UKPostcode(tt.input); got != tt.valid {
				t.Errorf("UKPostcode(%q) = %v, want %v", tt.input, got, tt.valid)
			}
		})
	}
}

func TestNormalizePostcode(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"sw1a1aa", "SW1A 1AA"},
		{"SW1A 1AA", "SW1A 1AA"},
		{"m1 1ae", "M1 1AE"},
		{"  b1 1bb  ", "B1 1BB"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := NormalizePostcode(tt.input); got != tt.want {
				t.Errorf("NormalizePostcode(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestNormalizePostcodeForComparison(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"SW1A 1AA", "SW1A1AA"},
		{"sw1a 1aa", "SW1A1AA"},
		{"  m1 1ae  ", "M11AE"},
		{"B11BB", "B11BB"},
		{"ls1  1ba", "LS11BA"},
		{"", ""},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := NormalizePostcodeForComparison(tt.input); got != tt.want {
				t.Errorf("NormalizePostcodeForComparison(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
