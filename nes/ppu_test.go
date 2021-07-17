package nes

import "testing"

func TestLoadPalette(t *testing.T) {
	ppu := new(Ppu)

	ppu.LoadPalette("../palettes/ntscpalette.pal")
}
