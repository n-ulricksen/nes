package nes

import (
	"io/ioutil"
	"log"
)

// Main bus used by the CPU.
type Bus struct {
	Cpu  *Cpu6502        // NES CPU.
	Ppu  *Ppu            // Picture processing unit.
	Ram  [64 * 1024]byte // 64kb RAM used for initial development.
	Cart *Cartridge      // NES Cartridge.

	ClockCount int
}

const (
	// RAM
	minRamAddr uint16 = 0x0000
	maxRamAddr uint16 = 0x1FFF
	ramMirror  uint16 = 0x07FF // mirror every 2KB.

	// PPU registers
	minPpuAddr uint16 = 0x2000
	maxPpuAddr uint16 = 0x3FFF
	ppuMirror  uint16 = 0x0008 // mirror every 8 bytes.
)

func NewBus() *Bus {
	// Create a new CPU. Here we use a 6502.
	cpu := NewCpu6502()

	// Attach devices to the bus.
	bus := &Bus{
		Cpu: cpu,
		Ram: [64 * 1024]byte{}, // fake RAM for now...
	}

	// Connect this bus to the cpu.
	cpu.ConnectBus(bus)

	return bus
}

// Used by the CPU to read data from the main bus at a specified address.
func (b *Bus) CpuRead(addr uint16) byte {
	var data byte

	if addr >= minRamAddr && addr <= maxRamAddr {
		data = b.Ram[addr&ramMirror]
	} else if addr >= minPpuAddr && addr <= maxPpuAddr {
		data = b.Ppu.cpuRead(addr & ppuMirror)
	}

	return data
}

// Used by the CPU to write data to the main bus at a specified address.
func (b *Bus) CpuWrite(addr uint16, data byte) {
	if addr >= minRamAddr && addr <= maxRamAddr {
		b.Ram[addr&ramMirror] = data
	} else if addr >= minPpuAddr && addr <= maxPpuAddr {
		b.Ppu.cpuWrite(addr&ppuMirror, data)
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
