package nes

import (
	"fmt"
	"image"
	"image/color"
	"io/ioutil"
	"log"
	"math/rand"
	"time"
)

const (
	paletteSize byte = 0x40

	// PPU addresses
	patternTblAddr    uint16 = 0x0000
	patternTblAddrEnd uint16 = 0x1FFF
	patternTblSize    uint16 = 0x1000 // Single pattern table - size in bytes

	nameTblAddr    uint16 = 0x2000
	nameTblAddrEnd uint16 = 0x3EFF

	paletteAddr    uint16 = 0x3F00
	paletteAddrEnd uint16 = 0x3FFF
)

// References:
// http://wiki.nesdev.com/w/index.php/PPU_registers
// https://www.youtube.com/watch?v=xdzOvpYPmGE (javidx9)
type Ppu struct {
	Cart *Cartridge

	nameTable    [2][1024]byte // NES allows storage for 2 nametables
	paletteTable [32]byte
	patternTable [2][4096]byte

	// PPU Registers
	ppuController PpuReg
	ppuMask       PpuReg
	ppuStatus     PpuReg

	// Intertal PPU variables
	scanline      int  // Scanline count in the current frame
	cycle         int  // Cycle count in the current scanline
	frameComplete bool // Whether or not the current frame is finished rendering

	addrLatch  byte   // Address latch to signal high or low byte - used by PPUSCROLL and PPUADDR.
	dataBuffer byte   // PPU reads are delayed 1 cycle, so we buffer the byte being read.
	addressTmp uint16 // Used to store the compiled address used for PPU data reads/writes.

	display *Display

	paletteRGBA [paletteSize]color.RGBA
}

func NewPpu() *Ppu {
	return &Ppu{
		nameTable:    [2][1024]byte{},
		paletteTable: [32]byte{},
		patternTable: [2][4096]byte{},

		ppuMask:   0x0,
		ppuStatus: 0x0,

		scanline:      0,
		cycle:         0,
		frameComplete: true,

		paletteRGBA: loadPalette("./palettes/ntscpalette.pal"),
	}
}

func init() {
	// XXX: only needed for generating screen noise before implementing PPU.
	rand.Seed(time.Now().UnixNano())
}

func (p *Ppu) ConnectCartridge(c *Cartridge) {
	p.Cart = c
}

func (p *Ppu) ConnectDisplay(d *Display) {
	p.display = d
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
		// Reading the status register clears bit 7 and the PPU address latch.
		p.ppuStatus.clearFlag(statusVBlank)
		p.addrLatch = 0x0
		data = byte(p.ppuStatus)
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
		p.ppuController = PpuReg(data)
	case 0x0001: // Mask
		p.ppuMask = PpuReg(data)
	case 0x0002: // Status
	case 0x0003: // OAM Address
	case 0x0004: // OAM Data
	case 0x0005: // Scroll
	case 0x0006: // Address
		if p.addrLatch == 0x0 {
			// First write
			p.addrLatch = data
		} else {
			// Second write
			p.addressTmp = uint16(p.addrLatch)<<8 | uint16(data)

			p.addrLatch = 0x0
		}

	case 0x0007: // Data
	}
}

// Communicate with PPU bus.
func (p *Ppu) ppuRead(addr uint16) byte {
	addr &= ppuMaxAddr

	var data byte

	if addr >= patternTblAddr && addr <= patternTblAddrEnd {
		tbl := (addr >> 12) & 0x1
		idx := addr & 0x0FFF
		data = p.patternTable[tbl][idx]
	} else if addr >= nameTblAddr && addr <= nameTblAddrEnd {
	} else if addr >= paletteAddr && addr <= paletteAddrEnd {
		// Mirrored addresses
		if addr == 0x3F10 || addr == 0x3F14 || addr == 0x3F18 || addr == 0x3F1C {
			addr -= 0x10
		}
		data = p.paletteTable[addr&0xFF]
	}

	return data
}

func (p *Ppu) ppuWrite(addr uint16, data byte) {
	addr &= ppuMaxAddr // Max addressable range.

	if addr >= patternTblAddr && addr <= patternTblAddrEnd {
		tbl := (addr >> 12) & 0x1
		idx := addr & 0x0FFF
		p.patternTable[tbl][idx] = data
	} else if addr >= nameTblAddr && addr <= nameTblAddrEnd {
	} else if addr >= paletteAddr && addr <= paletteAddrEnd {
		// Mirrored addresses
		if addr == 0x3F10 || addr == 0x3F14 || addr == 0x3F18 || addr == 0x3F1C {
			addr -= 0x10
		}
		p.paletteTable[addr&0xFF] = data
	}
}

// LoadPalette loads an NES palette from the specified file path, and returns
// and array of RGBA colors.
func loadPalette(filepath string) [paletteSize]color.RGBA {
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

// Convenience functions for development.

// Pattern tables are 16x16 grids of tiles or sprites. Each tile is 8x8 pixels
// and 16 bytes of memory.
func (p *Ppu) GetPatternTable(i int) *image.RGBA {
	rgba := image.NewRGBA(image.Rect(0, 0, 128, 128))

	for tileY := 0; tileY < 16; tileY++ {
		for tileX := 0; tileX < 16; tileX++ {
			// Tile
			memOffset := uint16(tileY*(16*16) + tileX*16)

			for row := 0; row < 8; row++ {
				// 2 bytes represent an 8 pixel row.
				tileLo := p.ppuRead(patternTblSize*uint16(i) + memOffset + uint16(row))
				tileHi := p.ppuRead(patternTblAddr*uint16(i) + memOffset + uint16(row) + 8)

				for col := 0; col < 8; col++ {
					// Calculate each pixel's value (0-3). The LSB represents
					// the last pixel in the row of 8. Use bit shifts to place the
					// required bit in the correct position each iteration.
					pixel := (tileLo & 0x01) + ((tileHi & 0x01) << 1)
					tileLo >>= 1
					tileHi >>= 1

					// Pixel position
					x := tileX*8 + (7 - col) // Inverted x-axis
					y := tileY*8 + row

					// Pixel color
					c := p.getColorFromPalette(0, pixel)

					// Draw the pixel
					rgba.Set(x, y, c)
				}
			}
		}
	}

	return rgba
}

func (p *Ppu) getColorFromPalette(palette, pixel byte) color.RGBA {
	idx := p.ppuRead(paletteAddr + uint16((palette<<2)+pixel))
	return p.paletteRGBA[idx]
}
