package model

import "testing"

func TestFormatPence(t *testing.T) {
	tests := []struct {
		name  string
		pence int64
		want  string
	}{
		{"zero", 0, "£0.00"},
		{"one pound", 100, "£1.00"},
		{"typical amount", 13265200, "£132,652.00"},
		{"small amount", 15, "£0.15"},
		{"thousand", 100000, "£1,000.00"},
		{"million", 100000000, "£1,000,000.00"},
		{"with pence", 12345, "£123.45"},
		{"negative", -500, "-£5.00"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatPence(tt.pence)
			if got != tt.want {
				t.Errorf("FormatPence(%d) = %q, want %q", tt.pence, got, tt.want)
			}
		})
	}
}

func TestDayWithSuffix(t *testing.T) {
	tests := []struct {
		day  int
		want string
	}{
		{1, "1st"}, {2, "2nd"}, {3, "3rd"}, {4, "4th"},
		{11, "11th"}, {12, "12th"}, {13, "13th"},
		{21, "21st"}, {22, "22nd"}, {23, "23rd"},
		{28, "28th"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := DayWithSuffix(tt.day)
			if got != tt.want {
				t.Errorf("DayWithSuffix(%d) = %q, want %q", tt.day, got, tt.want)
			}
		})
	}
}
