package nes

import (
	"testing"
)

//const testRom = "../roms/LegendOfZelda.nes"
const testRom = "../roms/DK.nes"

func TestNewCartridge(t *testing.T) {
	_ = NewCartridge(testRom)
}
