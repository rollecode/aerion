package app

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hkdb/aerion/internal/logging"
	"github.com/hkdb/aerion/internal/platform"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// pdfFilenameFallback is used when a conversation has no usable subject.
const pdfFilenameFallback = "conversation"

// pdfFilenameMaxLen keeps generated names well inside filesystem limits.
const pdfFilenameMaxLen = 80

// pdfFilename turns a mail subject into a safe .pdf filename.
func pdfFilename(subject string) string {
	cleaned := strings.Map(func(r rune) rune {
		switch r {
		case '/', '\\', ':', '*', '?', '"', '<', '>', '|', '\n', '\r', '\t':
			return '-'
		}
		return r
	}, subject)

	cleaned = strings.TrimSpace(cleaned)
	cleaned = strings.Trim(cleaned, ".")
	if cleaned == "" {
		cleaned = pdfFilenameFallback
	}
	if len(cleaned) > pdfFilenameMaxLen {
		cleaned = strings.TrimSpace(cleaned[:pdfFilenameMaxLen])
	}
	return cleaned + ".pdf"
}

// pdfDefaultDir returns the directory the save dialog opens in, creating it
// when absent so a machine without ~/Downloads still gets a sensible default.
// Falls back to the home directory, then to the dialog's own default.
func pdfDefaultDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	dir := filepath.Join(home, "Downloads")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return home
	}
	return dir
}

// SaveConversationPDF writes print-ready HTML to a PDF file chosen by the user.
// Reports false when the save dialog is cancelled.
func (a *App) SaveConversationPDF(html, subject string) (bool, error) {
	if strings.TrimSpace(html) == "" {
		return false, fmt.Errorf("nothing to save")
	}

	log := logging.WithComponent("app")

	name := pdfFilename(subject)
	dir := pdfDefaultDir()

	path, err := a.pickPDFPath(name, dir)
	if err != nil {
		return false, err
	}
	if path == "" {
		return false, nil
	}

	if !strings.EqualFold(filepath.Ext(path), ".pdf") {
		path += ".pdf"
	}

	if err := renderHTMLToPDF(html, path); err != nil {
		log.Error().Err(err).Msg("Failed to render PDF")
		return false, err
	}

	log.Info().Str("path", path).Msg("Saved conversation as PDF")
	return true, nil
}

// pickPDFPath asks the user where to write the PDF. Flatpak has to go through
// the portal because the Wails GTK dialog is not routed through it.
func (a *App) pickPDFPath(filename, dir string) (string, error) {
	if platform.IsFlatpak() {
		path, err := platform.PortalSaveFile("Save as PDF", filename, dir)
		if err != nil {
			return "", fmt.Errorf("failed to show save dialog: %w", err)
		}
		return path, nil
	}

	path, err := wailsRuntime.SaveFileDialog(a.ctx, wailsRuntime.SaveDialogOptions{
		DefaultDirectory: dir,
		DefaultFilename:  filename,
		Title:            "Save as PDF",
		Filters: []wailsRuntime.FileFilter{
			{DisplayName: "PDF documents (*.pdf)", Pattern: "*.pdf"},
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to show save dialog: %w", err)
	}
	return path, nil
}
