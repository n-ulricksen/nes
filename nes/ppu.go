package nes

import (
	"fmt"
	"image/color"
	"io/ioutil"
	"log"
	"math/rand"
	"time"
)

const (
	paletteSize = 0x40
)

// References:
// http://wiki.nesdev.com/w/index.php/PPU_registers
// https://www.youtube.com/watch?v=xdzOvpYPmGE (javidx9)
type Ppu struct {
	Cart *Cartridge

	nameTable    [2][1024]byte // NES allows storage for 2 nametables
	paletteTable [32]byte

	// Intertal PPU variables
	scanline      int  // Scanline count in the current frame
	cycle         int  // Cycle count in the current scanline
	frameComplete bool // Whether or not the current frame is finished rendering

	display *Display

	paletteRGBA [paletteSize]color.RGBA
}

func NewPpu() *Ppu {
	return &Ppu{
		nameTable:    [2][1024]byte{},
		paletteTable: [32]byte{},

		scanline:      0,
		cycle:         0,
		frameComplete: true,

		paletteRGBA: LoadPalette("./palettes/ntscpalette.pal"),
	}
}

func init() {
	// XXX: only needed for generating screen noise before implementing PPU.
	rand.Seed(time.Now().UnixNano())
}

func (p *Ppu) ConnectCartridge(c *Cartridge) {
	p.Cart = c
}

// LoadPalette loads an NES palette from the specified file path, and returns
// and array of RGBA colors.
func LoadPalette(filepath string) [paletteSize]color.RGBA {
	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		log.Fatal("Unable to open palette file.\n", err)
	}

	palette := [paletteSize]color.RGBA{}

	for i := 0; i < len(data); i += 3 {
		r := data[i]
		g := data[i+1]
		b := data[i+2]
		palette[i/3] = color.RGBA{r, g, b, 255}
	}
	fmt.Println("Palette data:")
	fmt.Println(palette)

	return palette
}

// PPU clock cycle.
// 1 frame = 262 scanlines
// 1 scanline = 341 PPU clock cycles
func (p *Ppu) Clock() {
	p.cycle++

	// Draw static to the screen for now (random color pixel)
	i := uint8(rand.Intn(0x40))
	p.display.DrawPixel(p.cycle-1, p.scanline, p.paletteRGBA[i])

	if p.cycle >= 341 {
		p.cycle = 0
		p.scanline++

		if p.scanline >= 262 {
			p.scanline = -1
			p.frameComplete = true

			p.display.UpdateScreen()
		}
	}
}

// Communicate with main (CPU) bus - used for PPU register access.
func (p *Ppu) cpuRead(addr uint16) byte {
	var data byte

	switch addr {
	case 0x0000: // Controller
	case 0x0001: // Mask
	case 0x0002: // Status
	case 0x0003: // OAM Address
	case 0x0004: // OAM Data
	case 0x0005: // Scroll
	case 0x0006: // Address
	case 0x0007: // Data
	}

	return data
}

func (p *Ppu) cpuWrite(addr uint16, data byte) {
	switch addr {
	case 0x0000: // Controller
	case 0x0001: // Mask
	case 0x0002: // Status
	case 0x0003: // OAM Address
	case 0x0004: // OAM Data
	case 0x0005: // Scroll
	case 0x0006: // Address
	case 0x0007: // Data
	}
}

// Communicate with PPU bus.
func (p *Ppu) ppuRead(addr uint16) byte {
	addr &= ppuMaxAddr

	// Read through the cartridge/mapper.
	return p.Cart.ppuRead(addr)
}

func (p *Ppu) ppuWrite(addr uint16, data byte) {
	addr &= ppuMaxAddr // Max addressable range.

	// Write through the cartridge/mapper.
	p.Cart.ppuWrite(addr, data)
}
