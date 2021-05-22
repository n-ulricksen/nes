package nes

import (
	"io/ioutil"
	"log"
)

type Bus struct {
	Cpu *Cpu6502
	Ram [64 * 1024]byte // 64kb RAM used for initial development.
}

func NewBus() *Bus {
	// Create a new CPU. Here we use a 6502.
	cpu := NewCpu6502()

	// Attach devices to the bus.
	bus := &Bus{
		Cpu: cpu,
		Ram: [64 * 1024]byte{},
	}

	// Connect this bus to the cpu.
	cpu.ConnectBus(bus)

	return bus
}

func (b *Bus) Read(addr uint16) byte {
	if addr >= 0x0000 && addr <= 0xFFFF {
		return b.Ram[addr]
	}
	return 0x00
}

func (b *Bus) Write(addr uint16, data byte) {
	if addr >= 0x0000 && addr <= 0xFFFF {
		b.Ram[addr] = data
	}
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
