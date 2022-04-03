package nes

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"time"
)

// Main bus used by the CPU.
type Bus struct {
	Cpu             *Cpu6502       // NES CPU.
	Ppu             *Ppu           // Picture processing unit.
	Ram             [8 * 1024]byte // 8KiB RAM.
	Cart            *Cartridge     // NES Cartridge.
	Controller      [2]*Controller // NES Controller.
	ControllerState [2]byte        // 8 bit shifter representing each button's state
	Disp            *Display

	ClockCount int

	// Direct memory access
	dmaPage byte
	dmaAddr byte
	dmaData byte // Memory to be sent from the CPU to the OAM

	dmaTransfer bool // Set to enable DMA transfer
	dmaNeedSync bool // Set when CPU should wait 1 cycle for DMA

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
	cartMinAddr uint16 = 0x8000 // XXX: changing this for now to get disassembler to work
	cartMaxAddr uint16 = 0xFFFF

	// Direct memory access
	dmaAddr uint16 = 0x4014

	// Controller
	ctrlMinAddr uint16 = 0x4016
	ctrlMaxAddr uint16 = 0x4017

	// Frames per second
	fps float64 = 60
)

func NewBus(isDebug, isLogging bool) *Bus {
	// Create a new CPU. Here we use a 6502.
	cpu := NewCpu6502(isLogging)

	controllers := [2]*Controller{}
	for i := range controllers {
		controllers[i] = NewController()
	}

	// Attach devices to the bus.
	bus := &Bus{
		Cpu:         cpu,
		Ppu:         NewPpu(),
		Controller:  controllers,
		dmaTransfer: false,
		dmaNeedSync: true,

		isDebug:   isDebug,
		isLogging: isLogging,
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

		for i := range b.Controller {
			b.Controller[i].updateControllerInput(b.Disp.window)
		}

		if b.isDebug {
			b.DrawDebugPanel()
		}

		since := time.Since(t)
		toSleep := interval - since
		time.Sleep(toSleep)

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
	} else if addr >= ctrlMinAddr && addr <= ctrlMaxAddr {
		data = (b.ControllerState[addr&1] & (1 << 7)) >> 7
		b.ControllerState[addr&1] <<= 1 // shift
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
	} else if addr == dmaAddr {
		b.dmaPage = data
		b.dmaAddr = 0x00
		b.dmaTransfer = true
	} else if addr >= ctrlMinAddr && addr <= ctrlMaxAddr {
		b.ControllerState[addr&1] = b.Controller[addr&1].GetState()
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
		if b.dmaTransfer {
			// A DMA transfer suspends the CPU until complete
			b.initDmaTransfer()
		} else {
			b.Cpu.Clock()
		}
	}

	if b.Ppu.nmi {
		b.Ppu.nmi = false
		b.Cpu.NMI()
	}

	b.ClockCount++
}

func (b *Bus) initDmaTransfer() {
	if b.dmaNeedSync {
		if b.ClockCount%2 == 1 {
			b.dmaNeedSync = false
		}
	} else {
		if b.ClockCount%2 == 0 {
			// read from CPU memory
			addr := uint16(b.dmaPage)<<8 | uint16(b.dmaAddr)
			b.dmaData = b.CpuRead(addr)
		} else {
			// write to OAM memory
			b.Ppu.oam.write(b.dmaAddr, b.dmaData)
			b.dmaAddr++

			if b.dmaAddr == 0x00 {
				// DMA transfer has finishied
				b.dmaTransfer = false
				b.dmaNeedSync = true
			}
		}
	}
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

	// Keyboard input

	// Disassembly
	diss := b.getDisassemblyLines()
	b.Disp.WriteInstDebugString(diss)
}

// getDisassemblyLines returns the last 15 lines of disassembly, separated by
// new lines, as a string.
func (b *Bus) getDisassemblyLines() string {
	var buf bytes.Buffer

	idx := b.Cpu.PrevInstIdx
	length := len(b.Cpu.PrevInstructions)

	for i := 1; i <= length; i++ {
		inst := b.Cpu.PrevInstructions[(idx+i)%length]
		if inst != "" {
			buf.WriteString(inst)
			buf.WriteByte('\n')
		}
	}

	return buf.String()
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
