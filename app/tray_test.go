package app

import (
	"bytes"
	"image/color"
	"image/png"
	"os"
	"testing"
)

func TestWhiteTrayIcon(t *testing.T) {
	src, err := os.ReadFile("../build/appicon.png")
	if err != nil {
		t.Fatal(err)
	}
	out := whiteTrayIcon(src)
	img, err := png.Decode(bytes.NewReader(out))
	if err != nil {
		t.Fatalf("output is not a valid png: %v", err)
	}
	b := img.Bounds()
	opaque, coloured := 0, 0
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			// NRGBA keeps the colour un-premultiplied, so a half-transparent
			// white pixel still reads 255,255,255.
			c := color.NRGBAModel.Convert(img.At(x, y)).(color.NRGBA)
			if c.A == 0 {
				continue
			}
			opaque++
			if c.R != 0xff || c.G != 0xff || c.B != 0xff {
				coloured++
			}
		}
	}
	if opaque == 0 {
		t.Fatal("no opaque pixels survived, alpha was lost")
	}
	if coloured != 0 {
		t.Fatalf("%d of %d opaque pixels are not white", coloured, opaque)
	}
	t.Logf("%d opaque pixels, all white", opaque)
}

func TestWhiteTrayIconRejectsGarbage(t *testing.T) {
	in := []byte("not a png")
	if got := whiteTrayIcon(in); !bytes.Equal(got, in) {
		t.Fatal("undecodable input should be returned unchanged")
	}
}
