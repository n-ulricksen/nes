package nes

import (
	"fmt"
	"image"
	"image/color"
	"io/ioutil"
	"log"
	"os"
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

	// Relative nametable address
	nameTbl0 uint16 = 0x0000
	nameTbl1 uint16 = 0x0400
	nameTbl2 uint16 = 0x0800
	nameTbl3 uint16 = 0x0C00

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
	ppuCtrl   *PpuReg
	ppuMask   *PpuReg
	ppuStatus *PpuReg

	nmi bool // Set true to signal a non-maskable interrupt

	// Internal PPU variables
	scanline      int  // Scanline count in the current frame
	cycle         int  // Cycle count in the current scanline
	frameComplete bool // Whether the current frame is finished rendering

	frames int // Total number of rendered frames

	dataBuffer byte // PPU reads are delayed 1 cycle, so we buffer the byte being read.

	// Background Rendering ~~~~~~
	// "Loopy" internal registers
	vRam        *PpuLoopyReg
	tRam        *PpuLoopyReg // Temporary ram address
	scrollFineX byte         // internal fine X scroll (3 bits)
	addrLatch   byte         // Address latch to signal high or low byte - used by PPUSCROLL and PPUADDR.

	// Tile/attribute fetching
	nextBgTileId byte
	nextBgAttr   byte
	nextBgTileLo byte
	nextBgTileHi byte

	// Shifters used for fine x scrolling
	bgPatternShifterLo uint16
	bgPatternShifterHi uint16
	bgAttribShifterLo  uint16
	bgAttribShifterHi  uint16
	// ~~~~~

	// Foreground Rendering ~~~~~~
	// Primary OAM
	// Secondary OAM

	// ~~~~~

	display *Display

	paletteRGBA [paletteSize]color.RGBA

	logger *log.Logger
}

func NewPpu() *Ppu {
	return &Ppu{
		nameTable:    [2][1024]byte{},
		paletteTable: [32]byte{},
		patternTable: [2][4096]byte{},

		ppuCtrl:   new(PpuReg),
		ppuMask:   new(PpuReg),
		ppuStatus: new(PpuReg),

		scanline:      0,
		cycle:         0,
		frameComplete: true,

		frames: 0,

		vRam: new(PpuLoopyReg),
		tRam: new(PpuLoopyReg),

		paletteRGBA: loadPalette("./palettes/ntscpalette.pal"),
	}
}

func (p *Ppu) ConnectCartridge(c *Cartridge) {
	p.Cart = c
}

func (p *Ppu) ConnectDisplay(d *Display) {
	p.display = d
}

// For future use if PPU logging is needed.
func newPpuLogger() *log.Logger {
	now := time.Now()
	logFile := fmt.Sprintf("./logs/ppu%s.log", now.Format("20060102-150405"))
	f, err := os.OpenFile(logFile, os.O_WRONLY|os.O_CREATE, 0664)
	if err != nil {
		log.Fatal("Unable to create PPU log file...\n", err)
	}

	return log.New(f, "", 0)
}

// PPU clock cycle.
// 1 frame = 262 scanlines (-1 - 260)
// 1 scanline = 341 PPU clock cycles (0 - 340)
func (p *Ppu) Clock() {
	p.renderBackground()
	p.renderForeground()

	p.cycle++
	if p.cycle >= 341 {
		p.cycle = 0
		p.scanline++

		// Scanline 261 is referred to scanline -1
		if p.scanline >= 261 {
			p.scanline = -1
			p.frameComplete = true
			p.frames++

			p.display.UpdateScreen()
		}
	}
}

// PPU background rendering, 1 scanline/pixel at a time.
// https://wiki.nesdev.com/w/index.php/PPU_rendering
func (p *Ppu) renderBackground() {
	// Rendering visible scanlines. We must include scanline -1 here because
	// that is when the data used in scanline 0 is fetched.
	if p.scanline >= -1 && p.scanline < 240 {
		if p.scanline == -1 && p.cycle == 1 {
			p.ppuStatus.clearFlag(statusVBlank)
		}

		// Last cycle of the scanline -1 is skipped every odd rendered
		// frame. We skip this 0 cycle every other frame to emulate this
		// behavior.
		if p.scanline == 0 && p.cycle == 0 {
			if p.frames%2 == 1 {
				p.cycle++
			}
		}

		// Repeated cycles - these memory accesses take 2 cycles on a real NES
		// PPU, but we will perform them in one for emulation.
		// Reference:
		//   https://wiki.nesdev.com/w/index.php/PPU_scrolling#Tile_and_attribute_fetching
		if (p.cycle >= 2 && p.cycle <= 257) || (p.cycle >= 321 && p.cycle <= 337) {
			p.updateShifters()

			var fetchAddr uint16
			switch (p.cycle - 1) % 8 {
			case 0:
				p.loadBackgroundShifters()

				// Nametable byte
				fetchAddr = nameTblAddr | (p.vRam.value() & 0x0FFF)
				p.nextBgTileId = p.ppuRead(fetchAddr)
			case 2:
				// Attribute table byte
				fetchAddr = 0x23C0 | (p.vRam.value() & 0x0C00) |
					((p.vRam.value() >> 4) & 0x38) | ((p.vRam.value() >> 2) & 0x07)
				p.nextBgAttr = p.ppuRead(fetchAddr)

				// TODO: figure this out and document it
				if (p.vRam.getCoarseY() & 0x2) > 0 {
					p.nextBgAttr >>= 4
				}
				if (p.vRam.getCoarseX() & 0x2) > 0 {
					p.nextBgAttr >>= 2
				}
				p.nextBgAttr &= 0x3
			case 4:
				// Pattern table tile low
				fetchAddr = uint16(p.ppuCtrl.getFlag(ctrlBgPatternTbl))<<12 |
					uint16(p.nextBgTileId)<<4 | uint16(p.vRam.getFineY()) + 0x0
				p.nextBgTileLo = p.ppuRead(fetchAddr)
			case 6:
				// Pattern table tile high
				fetchAddr = uint16(p.ppuCtrl.getFlag(ctrlBgPatternTbl))<<12 |
					uint16(p.nextBgTileId)<<4 | uint16(p.vRam.getFineY()) + 0x8
				p.nextBgTileHi = p.ppuRead(fetchAddr)
			case 7:
				// Increment horizontal scroll
				if p.shouldRender() {
					if p.vRam.getCoarseX() == 31 {
						// Wrap around (nametable is 32 tiles wide)
						p.vRam.setCoarseX(0)
						p.vRam.toggleNametableH()
					} else {
						// Course X is last bits of vRam address
						*p.vRam += 1
					}

				}
			}
		}

		if p.cycle == 256 {
			if p.shouldRender() {
				p.incrementVerticalScroll()
			}
		}

		// End of visible scanline, transfer x position from tRam to vRam
		if p.cycle == 257 {
			p.loadBackgroundShifters()
			if p.shouldRender() {
				p.vRam.setNametable(p.tRam.getNametable() & 0b01)
				p.vRam.setCoarseX(p.tRam.getCoarseX())
			}
		}

		// Unused nametable fetches at the end of each scnaline
		if p.cycle == 337 || p.cycle == 339 {
			fetchAddr := nameTblAddr | (p.vRam.value() & 0x0FFF)
			p.nextBgTileId = p.ppuRead(fetchAddr)
		}

		// End of visible frame, transfer y position from tRam to vRam
		if p.scanline == -1 && p.cycle >= 280 && p.cycle <= 304 {
			if p.shouldRender() {
				p.vRam.setNametable(p.tRam.getNametable() & 0b10)
				p.vRam.setCoarseY(p.tRam.getCoarseY())
				p.vRam.setFineY(p.tRam.getFineY())
			}
		}
	}

	// Post-render scanline - PPU idle
	if p.scanline == 240 {
	}

	// Enter vertical blank
	if p.scanline == 241 && p.cycle == 1 {
		p.ppuStatus.setFlag(statusVBlank)

		if p.ppuCtrl.getFlag(ctrlNmi) == 1 {
			p.nmi = true
		}
	}

	// Scanlines 241-260 don't do much of anything.

	// Get the palette and pixel vlues used to lookup the color to render at
	// this scanline/pixel.
	var bgPixel, bgPalette byte

	if p.ppuMask.getFlag(maskBgShow) > 0 {
		bitMux := uint16(0x8000 >> p.scrollFineX)

		var pixelLo, pixelHi byte
		if p.bgPatternShifterLo&bitMux > 0 {
			pixelLo = 1
		}
		if p.bgPatternShifterHi&bitMux > 0 {
			pixelHi = 1
		}
		bgPixel = (pixelHi << 1) | pixelLo

		var paletteLo, paletteHi byte
		if p.bgAttribShifterLo&bitMux > 0 {
			paletteLo = 1
		}
		if p.bgAttribShifterHi&bitMux > 0 {
			paletteHi = 1
		}
		bgPalette = (paletteHi << 1) | paletteLo
	}

	// Finally draw the correct color to the current pixel.
	p.display.DrawPixel(p.cycle-1, p.scanline,
		p.getColorFromPalette(bgPalette, bgPixel))
}

func (p *Ppu) renderForeground() {

}

// Communicate with main (CPU) bus - used for PPU register access.
func (p *Ppu) cpuRead(addr uint16) byte {
	var data byte

	switch addr {
	case 0x0000: // Controller
	case 0x0001: // Mask
	case 0x0002: // Status
		data = byte(*p.ppuStatus) & 0xE0

		// Reading the status register clears the VBlank flag and the PPU address latch.
		p.ppuStatus.clearFlag(statusVBlank)
		p.addrLatch = 0
	case 0x0003: // OAM Address
	case 0x0004: // OAM Data
	case 0x0005: // Scroll
	case 0x0006: // Address
	case 0x0007: // Data
		// CPU reads from VRAM are delayed by one cycle. The data to be read is
		// stored in a buffer on the PPU. Reading from VRAM returns the current
		// value stored on the buffer.
		data = p.dataBuffer
		p.dataBuffer = p.ppuRead(p.vRam.value())

		// The buffer is not used when reading palette data. The data is instead
		// placed directly onto the bus, bypassing the PPU data buffer.
		if p.vRam.value() >= paletteAddr {
			data = p.dataBuffer
		}

		// Accessing this port increments the VRAM address.
		// Bit 2 of PPUCTRL determines the amount to increment by:
		// 	0: increment by 1 (across)
		// 	1: increment by 32 (down)
		inc := p.ppuCtrl.getFlag(ctrlVramInc)
		if inc == 0 {
			*p.vRam += 1
		} else {
			*p.vRam += 32
		}
	}

	return data
}

func (p *Ppu) cpuWrite(addr uint16, data byte) {
	//fmt.Printf("CPU writing %x to address %x.\n", data, addr)
	switch addr {
	case 0x0000: // Controller
		*p.ppuCtrl = PpuReg(data)

		// 2 LSB used to set TRAM nametable bits.
		p.tRam.setNametable(data & 0b11)
	case 0x0001: // Mask
		*p.ppuMask = PpuReg(data)
	case 0x0002: // Status
	case 0x0003: // OAM Address
	case 0x0004: // OAM Data
	case 0x0005: // Scroll
		if p.addrLatch == 0 {
			// First write (coarse/fine X scroll values)
			coarseX := (data & (0b11111 << 3)) >> 3
			fineX := data & 0b111
			p.tRam.setCoarseX(coarseX)
			p.scrollFineX = fineX

			p.addrLatch = 1
		} else {
			// Second write (coarse/fine Y scroll values)
			coarseY := (data & (0b11111 << 3)) >> 3
			fineY := data & 0b111
			p.tRam.setCoarseY(coarseY)
			p.tRam.setFineY(fineY)

			p.addrLatch = 0
		}
	case 0x0006: // Address
		if p.addrLatch == 0 {
			// First write (high byte)
			setBits := uint16(data&0b111111) << 8
			*p.tRam = PpuLoopyReg(setBits) | *p.tRam&0xFF

			// First read also clears bit 14 of tRam
			*p.tRam &^= PpuLoopyReg(0b1 << 14)

			p.addrLatch = 1
		} else {
			// Second write (low byte)
			setBits := uint16(data)
			*p.tRam = (*p.tRam & 0xFF00) | PpuLoopyReg(setBits)

			// Second read transfers tRam to vRam
			*p.vRam = *p.tRam

			p.addrLatch = 0
		}
	case 0x0007: // Data
		p.ppuWrite(p.vRam.value(), data)

		// Accessing this port increments the VRAM address.
		// Bit 2 of PPUCTRL determines the amount to increment by:
		// 	0: increment by 1 (across)
		// 	1: increment by 32 (down)
		inc := p.ppuCtrl.getFlag(ctrlVramInc)
		if inc == 0 {
			*p.vRam += 1
		} else {
			*p.vRam += 32
		}
	}
}

// Communicate with PPU bus.
func (p *Ppu) ppuRead(addr uint16) byte {
	addr &= ppuMaxAddr

	var data byte

	if addr >= patternTblAddr && addr <= patternTblAddrEnd {
		//tbl := (addr >> 12) & 0x1
		//idx := addr & 0x0FFF
		//data = p.patternTable[tbl][idx]
		data = p.Cart.ppuRead(addr)
	} else if addr >= nameTblAddr && addr <= nameTblAddrEnd {
		// Nametable read with the correct mirroring set by the game cartridge
		data = p.nametableRead(addr)
	} else if addr >= paletteAddr && addr <= paletteAddrEnd {
		// Mirrored addresses
		addr &= 0x1F
		if addr == 0x0010 || addr == 0x0014 || addr == 0x0018 || addr == 0x001C {
			addr -= 0x10
		}
		data = p.paletteTable[addr]
	}

	return data
}

func (p *Ppu) ppuWrite(addr uint16, data byte) {
	addr &= ppuMaxAddr // Max addressable range.

	if addr >= patternTblAddr && addr <= patternTblAddrEnd {
		//tbl := (addr >> 12) & 0x1
		//idx := addr & 0x0FFF
		//p.patternTable[tbl][idx] = data
		p.Cart.ppuWrite(addr, data)
	} else if addr >= nameTblAddr && addr <= nameTblAddrEnd {
		// Nametable write with the correct mirroring set by the game cartridge
		p.nametableWrite(addr, data)
	} else if addr >= paletteAddr && addr <= paletteAddrEnd {
		// Mirrored addresses
		addr &= 0x1F
		if addr == 0x0010 || addr == 0x0014 || addr == 0x0018 || addr == 0x001C {
			addr -= 0x10
		}
		p.paletteTable[addr] = data
	}
}

// Gets a byte of data from the nametable memory using a given memory address.
func (p *Ppu) nametableRead(addr uint16) byte {
	var data byte

	// Get an address relative to the nametable space (0x0000-0x0FFF)
	addr &= 0x0FFF
	nameTblId := getNametableId(addr)

	switch nameTblId {
	case 0:
		data = p.nameTable[0][addr&0x3FF]
	case 1:
		if p.Cart.mirroring == mirrorHorizontal {
			data = p.nameTable[0][addr&0x3FF] // mirror
		} else if p.Cart.mirroring == mirrorVertical {
			data = p.nameTable[1][addr&0x3FF]
		}
	case 2:
		if p.Cart.mirroring == mirrorHorizontal {
			data = p.nameTable[1][addr&0x3FF]
		} else if p.Cart.mirroring == mirrorVertical {
			data = p.nameTable[0][addr&0x3FF] // mirror
		}
	case 3:
		data = p.nameTable[1][addr&0x3FF] // always mirror
	}

	return data
}

// Write data to the appropriate nametable, determined by the address and what
// mirroring mode is being used by the cartridge.
func (p *Ppu) nametableWrite(addr uint16, data byte) {
	// Relative nametable address
	addr &= 0x0FFF
	nameTblId := getNametableId(addr)

	switch nameTblId {
	case 0:
		p.nameTable[0][addr&0x3FF] = data
	case 1:
		if p.Cart.mirroring == mirrorHorizontal {
			p.nameTable[0][addr&0x3FF] = data // mirror
		} else if p.Cart.mirroring == mirrorVertical {
			p.nameTable[1][addr&0x3FF] = data
		}
	case 2:
		if p.Cart.mirroring == mirrorHorizontal {
			p.nameTable[1][addr&0x3FF] = data
		} else if p.Cart.mirroring == mirrorVertical {
			p.nameTable[0][addr&0x3FF] = data // mirror
		}
	case 3:
		p.nameTable[1][addr&0x3FF] = data // always mirror
	}
}

// Returns the nametable ID (0, 1, 2, 3) for the given relative memory address.
func getNametableId(addr uint16) byte {
	var id byte

	if addr >= nameTbl0 && addr < nameTbl1 {
		id = 0
	} else if addr >= nameTbl1 && addr < nameTbl2 {
		id = 1
	} else if addr >= nameTbl2 && addr < nameTbl3 {
		id = 2
	} else {
		id = 3
	}

	return id
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

	return palette
}

// Get a color from the given palette ID, offset by the given pixel value.
func (p *Ppu) getColorFromPalette(palette, pixel byte) color.RGBA {
	idx := p.ppuRead(paletteAddr + uint16((palette<<2)+pixel))

	return p.paletteRGBA[idx&0x3F]
}

// Check whether the PPU is in render mode. This is set by the maskBgShow and
// maskSpriteShow flags.
func (p *Ppu) shouldRender() bool {
	showBg := p.ppuMask.getFlag(maskBgShow) > 0
	showSprites := p.ppuMask.getFlag(maskSpriteShow) > 0
	return showBg || showSprites
}

// Update the shifters used to implement fine x scrolling.
func (p *Ppu) updateShifters() {
	// Pattern table shifters
	p.bgPatternShifterLo <<= 1
	p.bgPatternShifterHi <<= 1

	// Palette attribute shifters
	p.bgAttribShifterLo <<= 1
	p.bgAttribShifterHi <<= 1
}

// Load the background shifters with background tile pattern and attributes.
func (p *Ppu) loadBackgroundShifters() {
	// Tile patterns
	p.bgPatternShifterLo = (p.bgPatternShifterLo & 0xFF00) | uint16(p.nextBgTileLo)
	p.bgPatternShifterHi = (p.bgPatternShifterHi & 0xFF00) | uint16(p.nextBgTileHi)

	// Tile attributes
	var attrLo byte

	if p.nextBgAttr&0b01 > 0 {
		attrLo = 0xFF
	} else {
		attrLo = 0x00
	}
	p.bgAttribShifterLo = (p.bgAttribShifterLo & 0xFF00) | uint16(attrLo)

	if p.nextBgAttr&0b10 > 0 {
		attrLo = 0xFF
	} else {
		attrLo = 0x00
	}
	p.bgAttribShifterHi = (p.bgAttribShifterHi & 0xFF00) | uint16(attrLo)
}

// Increment fine y scroll, overflowing to coarse y. Wrap around nametables
// vertically.
// TODO: make this readable?
func (p *Ppu) incrementVerticalScroll() {
	if p.vRam.getFineY() < 7 {
		// Normal increment
		*p.vRam += 0x1000
	} else {
		p.vRam.setFineY(0x0)

		y := p.vRam.getCoarseY()
		if y == 29 {
			y = 0
			p.vRam.toggleNametableV()
		} else if y == 31 {
			y = 0
		} else {
			y++
		}
		p.vRam.setCoarseY(y)
	}
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
				tileHi := p.ppuRead(patternTblSize*uint16(i) + memOffset + uint16(row) + 8)

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
