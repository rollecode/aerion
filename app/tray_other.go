//go:build !linux

package app

// The tray is currently implemented for Linux only.
func (a *App) startTray() {}

func (a *App) stopTray() {}

func (a *App) applyTrayIcon() {}
