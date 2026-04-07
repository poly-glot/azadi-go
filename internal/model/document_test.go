package model

import "testing"

func TestDocument_FileType(t *testing.T) {
	tests := []struct {
		name     string
		fileName string
		want     string
	}{
		{"pdf extension", "statement.pdf", "pdf"},
		{"uppercase extension", "REPORT.PDF", "pdf"},
		{"mixed case", "Invoice.Pdf", "pdf"},
		{"csv extension", "data.csv", "csv"},
		{"xlsx extension", "report.XLSX", "xlsx"},
		{"multiple dots", "my.report.v2.pdf", "pdf"},
		{"no extension", "README", "pdf"},
		{"empty filename", "", "pdf"},
		{"dot only", ".", ""},
		{"ends with dot", "file.", ""},
		{"png extension", "photo.PNG", "png"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &Document{FileName: tt.fileName}
			if got := d.FileType(); got != tt.want {
				t.Errorf("FileType() = %q, want %q", got, tt.want)
			}
		})
	}
}
