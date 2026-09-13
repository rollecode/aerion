//go:build linux

package app

import "fyne.io/systray"

// startTray publishes a StatusNotifierItem so the app stays reachable while
// running with its window hidden. Desktops without a StatusNotifierItem host
// simply ignore it, so no capability check is needed.
func (a *App) startTray() {
	if len(trayIcon) == 0 {
		return
	}

	start, stop := systray.RunWithExternalLoop(a.buildTrayMenu, nil)
	a.stopTrayFn = stop
	start()
}

// stopTray removes the tray item during shutdown.
func (a *App) stopTray() {
	if a.stopTrayFn == nil {
		return
	}
	a.stopTrayFn()
	a.stopTrayFn = nil
}

func (a *App) buildTrayMenu() {
	systray.SetIcon(trayIcon)
	systray.SetTitle("Aerion")
	systray.SetTooltip("Aerion")

	open := systray.AddMenuItem("Open Aerion", "")
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
			case _, ok := <-quit.ClickedCh:
				if !ok {
					return
				}
				a.QuitApp()
			}
		}
	}()
}
