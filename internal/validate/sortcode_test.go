package validate

import "testing"

func TestSortCode(t *testing.T) {
	tests := []struct {
		input string
		valid bool
	}{
		{"20-30-40", true},
		{"00-00-00", true},
		{"99-99-99", true},
		{"20-30-4", false},
		{"20-30-400", false},
		{"203040", false},
		{"ab-cd-ef", false},
		{"", false},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := SortCode(tt.input); got != tt.valid {
				t.Errorf("SortCode(%q) = %v, want %v", tt.input, got, tt.valid)
			}
		})
	}
}
