package app

// trayIcon holds the PNG rendered in the system tray. main supplies it before
// Startup runs so the icon travels with the binary rather than depending on an
// installed icon theme.
var trayIcon []byte

// SetTrayIcon provides the tray icon image. Must be called before Startup.
func SetTrayIcon(icon []byte) {
	trayIcon = icon
}
