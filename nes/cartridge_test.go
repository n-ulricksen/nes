package nes

import (
	"fmt"
	"testing"
)

const testRom = "../roms/LegendOfZelda.nes"

func TestNewCartridge(t *testing.T) {
	cart := NewCartridge(testRom)
	fmt.Println(cart)
}
