package nes

type Cpu6502 struct {
	Pc     uint16 // Program Counter
	Sp     byte   // Stack Pointer: low 8 bits of next free location on stack.
	A      byte   // Accumulator Register
	X      byte   // X Register
	Y      byte   // Y Register
	Status byte   // Processor Status Flags

	bus *Bus // Communication Bus

	instLookup [16 * 16]Instruction // Instruction operation lookup
}

func NewCpu6502() *Cpu6502 {
	cpu := &Cpu6502{
		Pc:     0x0000,
		Sp:     0x00,
		A:      0x00,
		X:      0x00,
		Y:      0x00,
		Status: 0x00,
	}

	// Create the lookup table containing all the CPU instructions.
	// Reference: http://archive.6502.org/datasheets/rockwell_r650x_r651x.pdf
	cpu.instLookup = [16 * 16]Instruction{
		{"BRK", cpu.amIMP, cpu.opBRK, 7}, {"ORA", cpu.amIZX, cpu.opORA, 6},
	}

	return cpu
}

// Connect the CPU to a 16-bit address bus.
func (cpu *Cpu6502) ConnectBus(b *Bus) { cpu.bus = b }

// Read from the attached bus.
func (cpu *Cpu6502) read(addr uint16) byte {
	return cpu.bus.Read(addr)
}

// Write to the attached bus.
func (cpu *Cpu6502) write(addr uint16, data byte) {
	cpu.bus.Write(addr, data)
}

////////////////////////////////////////////////////////////////
// Status Flags
type SF6502 byte // 6502 Status Flag

const (
	StatusFlagC SF6502 = 1 << iota // Carry
	StatusFlagZ                    // Zero
	StatusFlagI                    // Interrupt Disable
	StatusFlagD                    // Decimal Mode (not used on NES)
	StatusFlagB                    // Break Command
	StatusFlagX                    // UNUSED
	StatusFlagV                    // Overflow
	StatusFlagN                    // Negative
)

// Convenience functions used to get and set CPU status flags.
func (cpu *Cpu6502) getFlag(f SF6502) byte {
	return cpu.Status & byte(f)
}

func (cpu *Cpu6502) setFlag(f SF6502, b bool) {
	if b {
		cpu.Status |= byte(f)
	} else {
		cpu.Status &^= byte(f)
	}
}

////////////////////////////////////////////////////////////////
// Addressing Modes
// These functions return any extra cycles needed for execution.
func (cpu *Cpu6502) amIMP() byte { return 0x00 }
func (cpu *Cpu6502) amIMM() byte { return 0x00 }
func (cpu *Cpu6502) amREL() byte { return 0x00 }
func (cpu *Cpu6502) amZP0() byte { return 0x00 }
func (cpu *Cpu6502) amZPX() byte { return 0x00 }
func (cpu *Cpu6502) amZPY() byte { return 0x00 }
func (cpu *Cpu6502) amABS() byte { return 0x00 }
func (cpu *Cpu6502) amABX() byte { return 0x00 }
func (cpu *Cpu6502) amABY() byte { return 0x00 }
func (cpu *Cpu6502) amIND() byte { return 0x00 }
func (cpu *Cpu6502) amIZX() byte { return 0x00 }
func (cpu *Cpu6502) amIZY() byte { return 0x00 }

////////////////////////////////////////////////////////////////
// Instructions
type Instruction struct {
	Name     string
	AddrMode func() byte
	Execute  func() byte
	Cycles   byte
}

// CPU insturctions. Each instruction method returns the number of any extra
// cycles necessary for execution.
func (cpu *Cpu6502) opADC() byte { return 0x00 }
func (cpu *Cpu6502) opAND() byte { return 0x00 }
func (cpu *Cpu6502) opASL() byte { return 0x00 }
func (cpu *Cpu6502) opBCC() byte { return 0x00 }
func (cpu *Cpu6502) opBCS() byte { return 0x00 }
func (cpu *Cpu6502) opBEQ() byte { return 0x00 }
func (cpu *Cpu6502) opBIT() byte { return 0x00 }
func (cpu *Cpu6502) opBMI() byte { return 0x00 }
func (cpu *Cpu6502) opBNE() byte { return 0x00 }
func (cpu *Cpu6502) opBPL() byte { return 0x00 }
func (cpu *Cpu6502) opBRK() byte { return 0x00 }
func (cpu *Cpu6502) opBVC() byte { return 0x00 }
func (cpu *Cpu6502) opBVS() byte { return 0x00 }
func (cpu *Cpu6502) opCLC() byte { return 0x00 }
func (cpu *Cpu6502) opCLD() byte { return 0x00 }
func (cpu *Cpu6502) opCLI() byte { return 0x00 }
func (cpu *Cpu6502) opCLV() byte { return 0x00 }
func (cpu *Cpu6502) opCMP() byte { return 0x00 }
func (cpu *Cpu6502) opCPX() byte { return 0x00 }
func (cpu *Cpu6502) opCPY() byte { return 0x00 }
func (cpu *Cpu6502) opDEC() byte { return 0x00 }
func (cpu *Cpu6502) opDEX() byte { return 0x00 }
func (cpu *Cpu6502) opDEY() byte { return 0x00 }
func (cpu *Cpu6502) opEOR() byte { return 0x00 }
func (cpu *Cpu6502) opINC() byte { return 0x00 }
func (cpu *Cpu6502) opINX() byte { return 0x00 }
func (cpu *Cpu6502) opINY() byte { return 0x00 }
func (cpu *Cpu6502) opJMP() byte { return 0x00 }
func (cpu *Cpu6502) opJSR() byte { return 0x00 }
func (cpu *Cpu6502) opLDA() byte { return 0x00 }
func (cpu *Cpu6502) opLDX() byte { return 0x00 }
func (cpu *Cpu6502) opLDY() byte { return 0x00 }
func (cpu *Cpu6502) opLSR() byte { return 0x00 }
func (cpu *Cpu6502) opNOP() byte { return 0x00 }
func (cpu *Cpu6502) opORA() byte { return 0x00 }
func (cpu *Cpu6502) opPHA() byte { return 0x00 }
func (cpu *Cpu6502) opPHP() byte { return 0x00 }
func (cpu *Cpu6502) opPLA() byte { return 0x00 }
func (cpu *Cpu6502) opPLP() byte { return 0x00 }
func (cpu *Cpu6502) opROL() byte { return 0x00 }
func (cpu *Cpu6502) opROR() byte { return 0x00 }
func (cpu *Cpu6502) opRTI() byte { return 0x00 }
func (cpu *Cpu6502) opRTS() byte { return 0x00 }
func (cpu *Cpu6502) opSBC() byte { return 0x00 }
func (cpu *Cpu6502) opSEC() byte { return 0x00 }
func (cpu *Cpu6502) opSED() byte { return 0x00 }
func (cpu *Cpu6502) opSEI() byte { return 0x00 }
func (cpu *Cpu6502) opSTA() byte { return 0x00 }
func (cpu *Cpu6502) opSTX() byte { return 0x00 }
func (cpu *Cpu6502) opSTY() byte { return 0x00 }
func (cpu *Cpu6502) opTAX() byte { return 0x00 }
func (cpu *Cpu6502) opTAY() byte { return 0x00 }
func (cpu *Cpu6502) opTSX() byte { return 0x00 }
func (cpu *Cpu6502) opTXA() byte { return 0x00 }
func (cpu *Cpu6502) opTXS() byte { return 0x00 }
func (cpu *Cpu6502) opTYA() byte { return 0x00 }
