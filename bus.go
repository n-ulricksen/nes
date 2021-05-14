package nes

type Bus struct {
	cpu *Cpu6502
	ram [64 * 1024]byte // 64kb RAM used for initial development.
}

func NewBus() *Bus {
	// Create a new CPU. Here we use a 6502.
	cpu := NewCpu6502()

	// Attach devices to the bus.
	bus := &Bus{
		cpu: cpu,
		ram: [64 * 1024]byte{},
	}

	// Connect this bus to the cpu.
	cpu.ConnectBus(bus)

	return bus
}

func (b *Bus) Read(addr uint16) byte {
	if addr >= 0x0000 && addr <= 0xFFFF {
		return b.ram[addr]
	}
	return 0x00
}

func (b *Bus) Write(addr uint16, data byte) {
	if addr >= 0x0000 && addr <= 0xFFFF {
		b.ram[addr] = data
	}
}
