package nes

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"time"
)

// Main bus used by the CPU.
type Bus struct {
	Cpu        *Cpu6502        // NES CPU.
	Ppu        *Ppu            // Picture processing unit.
	Ram        [64 * 1024]byte // 64kb RAM used for initial development.
	Cart       *Cartridge      // NES Cartridge.
	Controller *Controller     // NES Controller.
	Disp       *Display

	ClockCount int

	isDebug   bool // Enable debug panel
	isLogging bool // Enable logging
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
	//cartMinAddr uint16 = 0x4020
	cartMinAddr uint16 = 0x8000 // XXX: changing this for now to get disassembler to work
	cartMaxAddr uint16 = 0xFFFF

	// Frames per second
	fps float64 = 30.0
)

func NewBus(isDebug, isLogging bool) *Bus {
	// Create a new CPU. Here we use a 6502.
	cpu := NewCpu6502()

	// Attach devices to the bus.
	bus := &Bus{
		Cpu:        cpu,
		Ppu:        NewPpu(),
		Ram:        [64 * 1024]byte{},
		Controller: NewController(),
		isDebug:    isDebug,
		isLogging:  isLogging,
	}

	// Connect this bus to the cpu.
	cpu.ConnectBus(bus)

	return bus
}

// Run the NES.
func (b *Bus) Run() {
	// Create a PixelGL display for the PPU to render to.
	display := NewDisplay(b.isDebug)
	b.Disp = display

	// PPU needs access to the display.
	b.Ppu.ConnectDisplay(display)

	intervalInMilli := (1 / fps) * 1000
	interval := time.Duration(intervalInMilli) * time.Millisecond
	fmt.Println("Frame refresh time:", interval)

	// Use a timer to keep frames rendered steadily at a set FPS.
	var t time.Time
	for !display.window.Closed() {
		// Run 1 whole frame.
		t = time.Now()
		for !b.Ppu.frameComplete {
			b.Clock()
		}

		b.Controller.updateControllerInput(b.Disp.window)

		if b.isDebug {
			b.DrawDebugPanel()
		}

		time.Sleep(interval - time.Since(t))

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

	if b.Ppu.nmi {
		b.Ppu.nmi = false
		b.Cpu.NMI()
	}

	b.ClockCount++
}

// TODO: move this out of Bus, and into main or something. Also, rewrite this.
func (b *Bus) DrawDebugPanel() {
	// Pattern tables
	patternTable0 := b.Ppu.GetPatternTable(0)
	patternTable1 := b.Ppu.GetPatternTable(1)

	b.Disp.DrawDebugRGBA(8, int(gameH)-128-8, patternTable0)
	b.Disp.DrawDebugRGBA(128+16, int(gameH)-128-8, patternTable1)

	b.Disp.debugRegText.Clear()
	debugStr := b.getCpuDebugString()
	b.Disp.WriteRegDebugString(debugStr)

	// Disassembly
	diss := b.getDisassemblyLines()
	b.Disp.WriteInstDebugString(diss)
}

func (b *Bus) getDisassemblyLines() string {
	var buf bytes.Buffer

	pc := b.Cpu.Pc

	idx := pc
	for i := 0; i < 10; i++ {
		idx, err := getNextIdx(&b.Cpu.disassembly, idx)
		if err != nil {
			// End of the map
			break
		}
		idx++
		buf.WriteString(b.Cpu.disassembly[idx])
		buf.WriteByte('\n')
	}

	return buf.String()
}

// Items are stored by memory address, not all memory address are filled. This
// function returns the next item at or after the given memory address.
func getNextIdx(m *map[uint16]string, addr uint16) (uint16, error) {
	for _, ok := (*m)[addr]; !ok; addr++ {
		if addr >= 0xFFFF {
			return 0, errors.New("End of map")
		}
	}

	return addr, nil
}

func (b *Bus) getCpuDebugString() string {
	var buf bytes.Buffer

	buf.WriteString(fmt.Sprintf("Flags: %08b\n", b.Cpu.Status))
	buf.WriteString(fmt.Sprintf("PC: %#04X\n", b.Cpu.Pc))
	buf.WriteString(fmt.Sprintf("A: %#02X\n", b.Cpu.A))
	buf.WriteString(fmt.Sprintf("X: %#02X\n", b.Cpu.X))
	buf.WriteString(fmt.Sprintf("Y: %#02X\n", b.Cpu.Y))
	buf.WriteString(fmt.Sprintf("SP: %#02X\n\n", b.Cpu.Sp))

	// Cycles
	buf.WriteString(fmt.Sprintf("Cycle Count: %d\n\n", b.Cpu.CycleCount))

	// Instructions
	//buf.WriteString(fmt.Sprintf(t, "%#02X: %s\n\n", b.Cpu.Opcode, nesEmu.Cpu.InstLookup[nesEmu.Cpu.Opcode].Name)
	buf.WriteString(fmt.Sprintf("Previous Instruction:\n%s\n", b.Cpu.OpDiss))

	return buf.String()
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

// Used for testing the emulator with nestest.
func (b *Bus) CheckForNestestErrors() {
	errAddr1 := 0x02
	errAddr2 := 0x03

	if b.Ram[errAddr1] != 0x00 {
		log.Printf("nestest error %#X\n", b.Ram[errAddr1])
	}
	if b.Ram[errAddr2] != 0x00 {
		log.Printf("nestest error %#X\n", b.Ram[errAddr2])
	}
}
