//go:build linux

package app

import (
	"fyne.io/systray"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// startTray publishes a StatusNotifierItem so the app stays reachable while
// running with its window hidden. Desktops without a StatusNotifierItem host
// ignore it, so no capability check is needed.
func (a *App) startTray() {
	if len(trayIcon) == 0 || a.stopTrayFn != nil {
		return
	}
	if a.settingsStore != nil {
		if enabled, err := a.settingsStore.GetShowTrayIcon(); err == nil && !enabled {
			return
		}
	}

	start, stop := systray.RunWithExternalLoop(a.buildTrayMenu, nil)
	a.stopTrayFn = stop
	start()
}

// stopTray removes the tray item.
func (a *App) stopTray() {
	if a.stopTrayFn == nil {
		return
	}
	a.stopTrayFn()
	a.stopTrayFn = nil
}

// applyTrayIcon swaps the icon variant on the running tray item. The tray
// itself cannot be restarted once stopped, so enabling or disabling it is
// applied at launch instead.
func (a *App) applyTrayIcon() {
	if a.stopTrayFn == nil {
		return
	}
	systray.SetIcon(a.currentTrayIcon())
}

func (a *App) buildTrayMenu() {
	systray.SetIcon(a.currentTrayIcon())
	systray.SetTitle("Aerion")
	systray.SetTooltip("Aerion")

	open := systray.AddMenuItem("Open Aerion", "")
	settings := systray.AddMenuItem("Settings", "")
	quit := systray.AddMenuItem("Quit Aerion", "")

	go func() {
		defer recoverPanic("app", "tray")
		for {
			select {
			case _, ok := <-open.ClickedCh:
				if !ok {
					return
				}
				a.ShowWindow()
			case _, ok := <-settings.ClickedCh:
				if !ok {
					return
				}
				a.ShowWindow()
				wailsRuntime.EventsEmit(a.ctx, "app:open-settings")
			case _, ok := <-quit.ClickedCh:
				if !ok {
					return
				}
				a.QuitApp()
			}
		}
	}()
}
