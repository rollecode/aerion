//go:build !linux

package app

import "fmt"

func renderHTMLToPDF(html, path string) error {
	return fmt.Errorf("saving as PDF is only supported on Linux")
}
