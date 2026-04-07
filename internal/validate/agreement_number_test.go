package validate

import "testing"

func TestAgreementNumber(t *testing.T) {
	tests := []struct {
		input string
		valid bool
	}{
		{"AGR-100001", true},
		{"AGR-1", true},
		{"AGR-999999", true},
		{"agr-100001", false},
		{"100001", false},
		{"AGR100001", false},
		{"AGR-", false},
		{"", false},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := AgreementNumber(tt.input); got != tt.valid {
				t.Errorf("AgreementNumber(%q) = %v, want %v", tt.input, got, tt.valid)
			}
		})
	}
}
