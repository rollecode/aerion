package app

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPDFFilename(t *testing.T) {
	tests := []struct {
		name    string
		subject string
		want    string
	}{
		{"plain subject", "Quarterly report", "Quarterly report.pdf"},
		{"path separators replaced", "Invoice 10/2026", "Invoice 10-2026.pdf"},
		{"reserved characters replaced", `a:b*c?d"e<f>g|h`, "a-b-c-d-e-f-g-h.pdf"},
		{"newlines replaced", "line one\nline two", "line one-line two.pdf"},
		{"empty falls back", "   ", pdfFilenameFallback + ".pdf"},
		{"dot-only falls back", "...", pdfFilenameFallback + ".pdf"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := pdfFilename(tt.subject); got != tt.want {
				t.Errorf("pdfFilename(%q) = %q, want %q", tt.subject, got, tt.want)
			}
		})
	}
}

func TestPDFFilenameLength(t *testing.T) {
	long := ""
	for len(long) < 300 {
		long += "subject "
	}
	got := pdfFilename(long)
	if len(got) > pdfFilenameMaxLen+len(".pdf") {
		t.Errorf("pdfFilename produced %d chars, want <= %d", len(got), pdfFilenameMaxLen+len(".pdf"))
	}
}

func TestPDFDefaultDirCreatesMissing(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	dir := pdfDefaultDir()

	want := filepath.Join(home, "Downloads")
	if dir != want {
		t.Fatalf("pdfDefaultDir() = %q, want %q", dir, want)
	}
	st, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("directory was not created: %v", err)
	}
	if !st.IsDir() {
		t.Fatal("expected a directory")
	}
}

func TestPDFDefaultDirFallsBackWhenUncreatable(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	// A regular file where Downloads should go makes MkdirAll fail.
	if err := os.WriteFile(filepath.Join(home, "Downloads"), []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}

	if dir := pdfDefaultDir(); dir != home {
		t.Fatalf("pdfDefaultDir() = %q, want fallback %q", dir, home)
	}
}
