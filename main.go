package main

import (
	"flag"
	"fmt"

	"github.com/n-ulricksen/nes-emulator/nes"

	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"
)

// Command line flags
var (
	flagDebug   bool
	flagLogging bool
)

func main() {
	parseFlags()

	fmt.Println("Starting NES...")
	nesEmulator := nes.NewBus(flagDebug, flagLogging)

	// Load a test cartridge
	cart := nes.NewCartridge("./roms/DK.nes")
	//cart := nes.NewCartridge("./roms/SMB.nes")
	//cart := nes.NewCartridge("./external_tests/nestest/nestest.nes")
	nesEmulator.InsertCartridge(cart)

	nesEmulator.Cpu.Disassemble(0x0000, 0xFFFF)

	fmt.Println("Resetting NES...")
	nesEmulator.Cpu.Reset()

	pixelgl.Run(nesEmulator.Run)
}

func parseFlags() {
	flag.BoolVar(&flagDebug, "d", false, "enable debug panel")
	flag.BoolVar(&flagLogging, "l", false, "enable logging")

	flag.Parse()
}

func printDebugMem(t *text.Text, nesEmu *nes.Bus) {
	// Print 16 bytes per line.
	ramRowLimit := 0x0010

	// RAM: 0x0000-0x00FF
	ram1Base := 0x0000
	ram1Limit := ram1Base + 0x100
	for i := ram1Base; i < ram1Limit; i += ramRowLimit {
		fmt.Fprintf(t, "$%04X: % x\n", i,
			nesEmu.Ram[i:i+ramRowLimit])
	}

	fmt.Fprintf(t, "\n\n")

	// RAM: 0x8000-0x80FF
	ram2Base := 0xC000
	ram2Limit := ram2Base + 0x100
	for i := ram2Base; i < ram2Limit; i += ramRowLimit {
		fmt.Fprintf(t, "$%04X: % x\n", i,
			nesEmu.Ram[i:i+ramRowLimit])
	}
}

// XXX: remove
func printDebugCpu(t *text.Text, nesEmu *nes.Bus) {
	fmt.Fprintf(t, "Flags: %08b\n", nesEmu.Cpu.Status)
	fmt.Fprintf(t, "PC: %#04X\n", nesEmu.Cpu.Pc)
	fmt.Fprintf(t, "A: %#02X\n", nesEmu.Cpu.A)
	fmt.Fprintf(t, "X: %#02X\n", nesEmu.Cpu.X)
	fmt.Fprintf(t, "Y: %#02X\n", nesEmu.Cpu.Y)
	fmt.Fprintf(t, "SP: %#02X\n\n", nesEmu.Cpu.Sp)

	// Cycles
	fmt.Fprintf(t, "Cycle Count: %d\n\n", nesEmu.Cpu.CycleCount)

	// Instructions
	//fmt.Fprintf(t, "%#02X: %s\n\n", nesEmu.Cpu.Opcode, nesEmu.Cpu.InstLookup[nesEmu.Cpu.Opcode].Name)
	//fmt.Fprintf(t, "Previous Instruction:\n%s\n", nesEmu.Cpu.OpDiss)
}

func trimWhitespace(s string) string {
	newString := ""

	for _, c := range s {
		if !(c == ' ' || c == '\t' || c == '\n') {
			newString += string(c)
		}
	}

	return newString
}
