package app

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"image/png"
)

// trayIcon holds the PNG rendered in the system tray. main supplies it before
// Startup runs so the icon travels with the binary rather than depending on an
// installed icon theme.
var trayIcon []byte

// SetTrayIcon provides the tray icon image. Must be called before Startup.
func SetTrayIcon(icon []byte) {
	trayIcon = icon
}

// whiteTrayIcon repaints every pixel white while keeping the original alpha,
// producing a silhouette that stays legible on dark panels. Returns the input
// unchanged when it cannot be decoded.
func whiteTrayIcon(icon []byte) []byte {
	src, err := png.Decode(bytes.NewReader(icon))
	if err != nil {
		return icon
	}

	bounds := src.Bounds()
	out := image.NewNRGBA(bounds)
	draw.Draw(out, bounds, src, bounds.Min, draw.Src)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			_, _, _, a := out.At(x, y).RGBA()
			out.Set(x, y, color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: uint8(a >> 8)})
		}
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, out); err != nil {
		return icon
	}
	return buf.Bytes()
}

// GetShowTrayIcon reports whether the tray item is enabled.
func (a *App) GetShowTrayIcon() (bool, error) {
	return a.settingsStore.GetShowTrayIcon()
}

// SetShowTrayIcon enables or disables the tray item. Applied at next launch,
// because the tray cannot be republished once it has been torn down.
func (a *App) SetShowTrayIcon(enabled bool) error {
	return a.settingsStore.SetShowTrayIcon(enabled)
}

// GetTrayIconWhite reports whether the tray icon is the white silhouette.
func (a *App) GetTrayIconWhite() (bool, error) {
	return a.settingsStore.GetTrayIconWhite()
}

// SetTrayIconWhite switches the tray icon variant and applies it immediately.
func (a *App) SetTrayIconWhite(white bool) error {
	if err := a.settingsStore.SetTrayIconWhite(white); err != nil {
		return err
	}
	a.applyTrayIcon()
	return nil
}

// currentTrayIcon returns the icon variant the settings ask for.
func (a *App) currentTrayIcon() []byte {
	if a.settingsStore == nil {
		return trayIcon
	}
	white, err := a.settingsStore.GetTrayIconWhite()
	if err != nil || !white {
		return trayIcon
	}
	return whiteTrayIcon(trayIcon)
}
