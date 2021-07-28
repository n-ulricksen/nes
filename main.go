package main

import (
	"fmt"
	"log"

	"github.com/n-ulricksen/nes-emulator/nes"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"
	"golang.org/x/image/colornames"
	"golang.org/x/image/font/basicfont"
)

func main() {
	fmt.Println("Starting NES...")
	nesEmulator := nes.NewBus()

	// Load a test cartridge
	//cart := nes.NewCartridge("./roms/DK.nes")
	cart := nes.NewCartridge("./roms/SMB.nes")
	//cart := nes.NewCartridge("./external_tests/nestest/nestest.nes")
	nesEmulator.InsertCartridge(cart)

	nesEmulator.Cpu.Disassemble(0x0000, 0xFFFF)

	fmt.Println("Resetting NES...")
	nesEmulator.Cpu.Reset()

	//startDebugMode(nesEmulator)
	pixelgl.Run(nesEmulator.Run)
}

func startDebugMode(nesEmu *nes.Bus) {
	// Channel used to signal CPU to cycle.
	cycleChan := make(chan bool)

	// Function to run the NES emulator. This must be run on a separate goroutine
	// due to PixelGL requiring main.
	go func(nesEmu *nes.Bus, cycleChan <-chan bool) {
		// There may be cycles left to complete for current instruction/interrupt.
		cycles := int(nesEmu.Cpu.Cycles)

		for {
			// While debug mode is running, we wait for signals from the PixelGL
			// subroutine to trigger CPU cycles. Each signal on the channel will
			// cause 1 CPU instruction worth of cycles.
			select {
			case <-cycleChan:
				if cycles == 0 {
					cycles = int(nesEmu.Cpu.InstLookup[nesEmu.CpuRead(nesEmu.Cpu.Pc)].Cycles)
				}

				for i := 0; i < cycles; i++ {
					nesEmu.Clock()
				}
				cycles = 0
				nesEmu.CheckForNestestErrors()
			}
		}
	}(nesEmu, cycleChan)

	// PixelGL must run on main thread.
	pixelgl.Run(RunDebugWindow(nesEmu, cycleChan))
}

func RunDebugWindow(nesEmu *nes.Bus, cycleChan chan<- bool) func() {
	return func() {
		// Config
		cfg := pixelgl.WindowConfig{
			Title:  "NES Emulator Debug",
			Bounds: pixel.R(0, 0, 1024, 768),
			VSync:  true,
		}

		// Create new window.
		win, err := pixelgl.NewWindow(cfg)
		if err != nil {
			log.Fatal("Unable to create new PixelGL window...\n", err)
		}

		// Create atlas and text boxes.
		basicAtlas := text.NewAtlas(basicfont.Face7x13, text.ASCII)

		ramText := text.New(pixel.V(10, 768-20), basicAtlas)
		regText := text.New(pixel.V(400, 768-20), basicAtlas)

		// Main display (the running game)
		// XXX: Might need to use a different graphics library

		isAutoRun := false

		// Draw to the screen
		for !win.Closed() {
			win.Clear(colornames.Black)

			ramText.Clear()
			printDebugMem(ramText, nesEmu)
			ramText.Draw(win, pixel.IM.Scaled(ramText.Orig, 1))

			regText.Clear()
			printDebugCpu(regText, nesEmu)
			regText.Draw(win, pixel.IM.Scaled(regText.Orig, 1))

			win.Update()

			// Run a CPU instruction everytime the SPACE key is pressed or if autorun
			// is toggled (triggered by ENTER key).
			if isAutoRun || win.JustPressed(pixelgl.KeySpace) {
				cycleChan <- true
			}

			// Toggle CPU autorun with ENTER key.
			if win.JustPressed(pixelgl.KeyEnter) {
				isAutoRun = !isAutoRun
			}
		}
	}
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
