package app

import (
	"os"
	"path/filepath"
	"testing"
)

func TestXDGDownloadDir(t *testing.T) {
	tests := []struct {
		name     string
		contents string
		want     string
	}{
		{"home relative", `XDG_DOWNLOAD_DIR="$HOME/Lataukset"`, "Lataukset"},
		{"unquoted", `XDG_DOWNLOAD_DIR=$HOME/Hentos`, "Hentos"},
		{"missing entry", `XDG_MUSIC_DIR="$HOME/Music"`, ""},
		{"empty value", `XDG_DOWNLOAD_DIR=""`, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			home := t.TempDir()
			if err := os.MkdirAll(filepath.Join(home, ".config"), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(home, ".config", "user-dirs.dirs"), []byte(tt.contents), 0o600); err != nil {
				t.Fatal(err)
			}

			want := ""
			if tt.want != "" {
				want = filepath.Join(home, tt.want)
			}
			if got := xdgDownloadDir(home); got != want {
				t.Errorf("xdgDownloadDir() = %q, want %q", got, want)
			}
		})
	}
}

func TestXDGDownloadDirNoFile(t *testing.T) {
	if got := xdgDownloadDir(t.TempDir()); got != "" {
		t.Errorf("xdgDownloadDir() = %q, want empty", got)
	}
}

func TestDownloadDirCreatesMissing(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	app := &App{}
	dir := app.DownloadDir()

	want := filepath.Join(home, downloadDirName)
	if dir != want {
		t.Fatalf("DownloadDir() = %q, want %q", dir, want)
	}
	if st, err := os.Stat(dir); err != nil || !st.IsDir() {
		t.Fatalf("directory was not created: %v", err)
	}
}

func TestDownloadDirPrefersXDG(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	if err := os.MkdirAll(filepath.Join(home, ".config"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, ".config", "user-dirs.dirs"),
		[]byte(`XDG_DOWNLOAD_DIR="$HOME/Lataukset"`), 0o600); err != nil {
		t.Fatal(err)
	}

	app := &App{}
	if got, want := app.DownloadDir(), filepath.Join(home, "Lataukset"); got != want {
		t.Fatalf("DownloadDir() = %q, want %q", got, want)
	}
}

func TestDownloadDirFallsBackWhenUncreatable(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	if err := os.WriteFile(filepath.Join(home, downloadDirName), []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}

	app := &App{}
	if got := app.DownloadDir(); got != home {
		t.Fatalf("DownloadDir() = %q, want fallback %q", got, home)
	}
}
