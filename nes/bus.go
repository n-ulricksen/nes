package nes

import (
	"fmt"
	"io/ioutil"
	"log"
	"time"
)

// Main bus used by the CPU.
type Bus struct {
	Cpu  *Cpu6502        // NES CPU.
	Ppu  *Ppu            // Picture processing unit.
	Ram  [64 * 1024]byte // 64kb RAM used for initial development.
	Cart *Cartridge      // NES Cartridge.
	Disp *Display

	ClockCount int
}

const (
	// RAM
	ramMinAddr uint16 = 0x0000
	ramMaxAddr uint16 = 0x1FFF
	ramMirror  uint16 = 0x07FF // mirror every 2KB.

	// PPU
	ppuMinAddr uint16 = 0x2000
	ppuMaxAddr uint16 = 0x3FFF
	ppuMirror  uint16 = 0x0007 // mirror every 8 bytes.

	// Cartridge
	cartMinAddr uint16 = 0x4020
	cartMaxAddr uint16 = 0xFFFF

	// Frames per second
	fps float64 = 30.0
)

func NewBus() *Bus {
	// Create a new CPU. Here we use a 6502.
	cpu := NewCpu6502()

	// Attach devices to the bus.
	bus := &Bus{
		Cpu: cpu,
		Ppu: NewPpu(),
		Ram: [64 * 1024]byte{}, // fake RAM for now...
	}

	// Connect this bus to the cpu.
	cpu.ConnectBus(bus)

	return bus
}

func (b *Bus) Run() {
	// Create a PixelGL display for the PPU to render to.
	display := NewDisplay()
	b.Disp = display

	// PPU needs access to the display.
	b.Ppu.ConnectDisplay(display)

	intervalInMilli := (1 / fps) * 1000
	interval := time.Duration(intervalInMilli) * time.Millisecond
	fmt.Println("Frame refresh time:", interval)

	ticker := time.NewTicker(interval)

	// Use a time ticker to keep frames rendered steadily at a set FPS.
	for {
		for !b.Ppu.frameComplete {
			b.Clock()
		}

		<-ticker.C
		ticker.Reset(interval)

		// Prepare for new frame
		b.Ppu.frameComplete = false
	}
}

// Used by the CPU to read data from the main bus at a specified address.
func (b *Bus) CpuRead(addr uint16) byte {
	var data byte

	if addr >= ramMinAddr && addr <= ramMaxAddr {
		data = b.Ram[addr&ramMirror]
	} else if addr >= ppuMinAddr && addr <= ppuMaxAddr {
		data = b.Ppu.cpuRead(addr & ppuMirror)
	} else if addr >= cartMinAddr && addr <= cartMaxAddr {
		data = b.Cart.cpuRead(addr)
	}

	return data
}

// Used by the CPU to write data to the main bus at a specified address.
func (b *Bus) CpuWrite(addr uint16, data byte) {
	if addr >= ramMinAddr && addr <= ramMaxAddr {
		b.Ram[addr&ramMirror] = data
	} else if addr >= ppuMinAddr && addr <= ppuMaxAddr {
		b.Ppu.cpuWrite(addr&ppuMirror, data)
	} else if addr >= cartMinAddr && addr <= cartMaxAddr {
		b.Cart.cpuWrite(addr, data)
	}

}

// Load a cartridge to the NES. The cartridge is connected to both the CPU and PPU.
func (b *Bus) InsertCartridge(cart *Cartridge) {
	b.Cart = cart
	b.Ppu.ConnectCartridge(cart)
}

// Reset the NES.
func (b *Bus) Reset() {
	b.Cpu.Reset()

	b.ClockCount = 0
}

// 1 NES clock cycle.
func (b *Bus) Clock() {
	b.Ppu.Clock()

	// CPU runs 3 times slower than PPU.
	if b.ClockCount%3 == 0 {
		b.Cpu.Clock()
	}

	b.ClockCount++
}

// Load a ROM to the NES.
func (b *Bus) Load(filepath string) {
	data, err := ioutil.ReadFile(filepath)

	if err != nil {
		log.Fatalf("Unable to open %v\n%v\n", filepath, err)
	}

	romOffset := 0x8000

	for i, bte := range data {
		b.Ram[romOffset+i] = bte
	}
}

// Load a slice of bytes to the NES.
func (b *Bus) LoadBytes(rom []byte) {
	romOffset := 0x8000

	for i, bte := range rom {
		b.Ram[romOffset+i] = bte
	}
}

func (b *Bus) LoadNestest() {
	filepath := "./external_tests/nestest/nestest.nes"

	data, err := ioutil.ReadFile(filepath)

	if err != nil {
		log.Fatalf("Unable to open %v\n%v\n", filepath, err)
	}

	// Load 0x4000 bytes starting from 0x0010 (NES headers) from the nestest ROM
	// into addresses 0x8000 & 0xC000.
	for i := 0; i < 0x4000; i++ {
		b.Ram[i+0x8000] = data[i+0x10]
		b.Ram[i+0xC000] = data[i+0x10]
	}

	// Nestest program entry
	b.Cpu.Pc = 0xC000
}

// Used for testing the emulator with nestest.
func (b *Bus) CheckForNestestErrors() {
	errAddr1 := 0x02
	errAddr2 := 0x03

	if b.Ram[errAddr1] != 0x00 {
		log.Fatalf("nestest error %#X\n", b.Ram[errAddr1])
	}
	if b.Ram[errAddr2] != 0x00 {
		log.Fatalf("nestest error %#X\n", b.Ram[errAddr2])
	}
}
