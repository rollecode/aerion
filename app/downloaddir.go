package app

import (
	"os"
	"path/filepath"
	"strings"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// downloadDirName is the directory created under $HOME when nothing else is
// configured or discoverable.
const downloadDirName = "Downloads"

// xdgDownloadDir reads XDG_DOWNLOAD_DIR so localised setups (Lataukset,
// Téléchargements) resolve to the directory the desktop actually uses.
// Returns "" when the entry is missing or unparseable.
func xdgDownloadDir(home string) string {
	data, err := os.ReadFile(filepath.Join(home, ".config", "user-dirs.dirs"))
	if err != nil {
		return ""
	}

	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "XDG_DOWNLOAD_DIR=") {
			continue
		}

		value := strings.Trim(strings.TrimPrefix(line, "XDG_DOWNLOAD_DIR="), `"`)
		if value == "" {
			return ""
		}
		if strings.HasPrefix(value, "$HOME/") {
			return filepath.Join(home, strings.TrimPrefix(value, "$HOME/"))
		}
		if filepath.IsAbs(value) {
			return value
		}
		return ""
	}
	return ""
}

// DownloadDir resolves the directory save dialogs should open in: the user's
// setting first, then the desktop's download directory, then ~/Downloads. The
// directory is created when missing, and an unusable one falls back to $HOME
// rather than failing the save.
func (a *App) DownloadDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = ""
	}

	dir := a.configuredDownloadDir()
	if dir == "" && home != "" {
		dir = xdgDownloadDir(home)
	}
	if dir == "" && home != "" {
		dir = filepath.Join(home, downloadDirName)
	}
	if dir == "" {
		return ""
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return home
	}
	return dir
}

// configuredDownloadDir returns the stored setting, expanding a leading ~.
func (a *App) configuredDownloadDir() string {
	if a.settingsStore == nil {
		return ""
	}

	dir, err := a.settingsStore.GetDownloadDirectory()
	if err != nil || strings.TrimSpace(dir) == "" {
		return ""
	}

	dir = strings.TrimSpace(dir)
	if !strings.HasPrefix(dir, "~") {
		return dir
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, strings.TrimPrefix(dir, "~"))
}

// GetDownloadDirectory reports the directory save dialogs will open in.
func (a *App) GetDownloadDirectory() (string, error) {
	return a.DownloadDir(), nil
}

// SetDownloadDirectory stores a new save directory. An empty value restores
// the platform default.
func (a *App) SetDownloadDirectory(dir string) error {
	return a.settingsStore.SetDownloadDirectory(strings.TrimSpace(dir))
}

// PickDownloadDirectory opens a directory chooser and stores the result.
// Returns the directory in use, unchanged when the dialog is cancelled.
func (a *App) PickDownloadDirectory() (string, error) {
	chosen, err := wailsRuntime.OpenDirectoryDialog(a.ctx, wailsRuntime.OpenDialogOptions{
		Title:            "Choose download folder",
		DefaultDirectory: a.DownloadDir(),
	})
	if err != nil {
		return "", err
	}
	if chosen == "" {
		return a.DownloadDir(), nil
	}
	if err := a.SetDownloadDirectory(chosen); err != nil {
		return "", err
	}
	return a.DownloadDir(), nil
}
