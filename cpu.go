package nes

import (
	"fmt"
)

type Cpu6502 struct {
	Pc     uint16 // Program Counter
	Sp     byte   // Stack Pointer: low 8 bits of next free location on stack.
	A      byte   // Accumulator Register
	X      byte   // X Register
	Y      byte   // Y Register
	Status byte   // Processor Status Flags

	bus *Bus // Communication Bus

	// Internal variables
	cycles        byte   // Remaining cycles for current insturction
	opcode        byte   // Opcode representing next instruction to be executed
	addrAbs       uint16 // Set by addressing mode functions, used by instructions
	fetched       byte   // Byte of memory used by CPU instructions
	isImpliedAddr bool   // Whether the current instruction's address mode is implied
	cycleCount    uint32 // Total # of cycles executed by the CPU

	instLookup [16 * 16]Instruction // Instruction operation lookup
}

const stackBase uint16 = 0x0100

func NewCpu6502() *Cpu6502 {
	cpu := &Cpu6502{
		Pc:     0x0000,
		Sp:     0xFD,
		A:      0x00,
		X:      0x00,
		Y:      0x00,
		Status: 0x00,

		cycles:        0,
		opcode:        0x00,
		addrAbs:       0x0000,
		fetched:       0x00,
		isImpliedAddr: false,
		cycleCount:    0,
	}

	// Create the lookup table containing all the CPU instructions.
	// Reference: http://archive.6502.org/datasheets/rockwell_r650x_r651x.pdf
	cpu.instLookup = [16 * 16]Instruction{
		{"BRK", cpu.opBRK, cpu.amIMP, 7}, {"ORA", cpu.opORA, cpu.amIZX, 6}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"ORA", cpu.opORA, cpu.amZP0, 3}, {"ASL", cpu.opASL, cpu.amZP0, 5}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"PHP", cpu.opPHP, cpu.amIMP, 3}, {"ORA", cpu.opORA, cpu.amIMM, 2}, {"ASL", cpu.opASL, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"ORA", cpu.opORA, cpu.amABS, 4}, {"ASL", cpu.opASL, cpu.amABS, 6}, {"XXX", cpu.opXXX, cpu.amIMP, 2},

		{"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2},

		{"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2},

		{"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2},

		{"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2},

		{"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2},

		{"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2},

		{"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2},

		{"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2},

		{"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2},

		{"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2},

		{"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2},

		{"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2},

		{"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2},

		{"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2},

		{"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2}, {"XXX", cpu.opXXX, cpu.amIMP, 2},
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

// Read a word from memory (little endian order).
func (cpu *Cpu6502) readWord(addr uint16) uint16 {
	lo := cpu.read(addr)
	hi := cpu.read(addr + 1)

	return (uint16(hi) << 8) | uint16(lo)
}

// Read a byte from memory at the address previously set by the appropriate
// addressing mode function. Avoid if current instruction's address mode is implied.
func (cpu *Cpu6502) fetch() {
	if !cpu.isImpliedAddr {
		cpu.fetched = cpu.read(cpu.addrAbs)
	}
}

// Functions to push and pop from the stack.
func (cpu *Cpu6502) stackPush(data byte) {
	cpu.write((stackBase | uint16(cpu.Sp)), data)
	cpu.Sp--
}

func (cpu *Cpu6502) stackPop() byte {
	cpu.Sp++
	return cpu.read(stackBase | uint16(cpu.Sp))
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
// Interrupts
const resetVectAddr = 0xFFFC
const irqVectAddr = 0xFFFE

func (cpu *Cpu6502) Reset() {
	// TODO: set the program counter to the absolute address read from location
	// 0xFFFC.

	// Clear registers, reset stack pointer
	cpu.A = 0x00
	cpu.X = 0x00
	cpu.Y = 0x00
	cpu.Status = 0x00 | byte(StatusFlagX)
	cpu.Sp = 0xFD

	// Get the program counter from the reset vector location in RAM.
	cpu.Pc = cpu.readWord(resetVectAddr)

	// TODO: clear internal variables (absolute/relative addresses, fetched)

	// Spend time on reset
	cpu.cycles = 8
}

// Interrupt Request
func (cpu *Cpu6502) IRQ() {}
func (cpu *Cpu6502) NMI() {}

// Cycle represents one CPU clock cycle.
func (cpu *Cpu6502) Cycle() {
	if cpu.cycles == 0 {
		// Get the next opcode by reading from the bus at the location of the
		// current program counter.
		cpu.opcode = cpu.read(cpu.Pc)

		// Lookup by opcode the instruction to be executed.
		inst := cpu.instLookup[cpu.opcode]

		fmt.Printf("from %#x fetched %#x: %v\n", cpu.Pc, cpu.opcode, inst)

		cpu.Pc++

		// Set required cycles for instruction execution.
		cpu.cycles = inst.Cycles

		// Add any additional cycles needed by either the addressing mode or
		// instruction.
		extraCycles1 := inst.AddrMode()
		extraCycles2 := inst.Execute()

		cpu.cycles += (extraCycles1 & extraCycles2)
	}

	// Turn implied address mode off, just in case the last instruction turned it on.
	cpu.isImpliedAddr = false

	cpu.cycleCount++

	cpu.cycles--
}

////////////////////////////////////////////////////////////////
// Addressing Modes
// These functions return any extra cycles needed for execution.

// Implied:
func (cpu *Cpu6502) amIMP() byte {
	cpu.isImpliedAddr = true

	cpu.fetched = cpu.A
	return 0x00
}

// Immediate:
func (cpu *Cpu6502) amIMM() byte {
	// The second byte of the instruction contains the operand.
	cpu.addrAbs = cpu.Pc
	cpu.Pc++

	return 0x00
}

func (cpu *Cpu6502) amREL() byte { return 0x00 }

// Zero Page:
func (cpu *Cpu6502) amZP0() byte {
	// Use the second byte of the instruction to index into page zero.
	lo := cpu.read(cpu.Pc)
	cpu.Pc++

	cpu.addrAbs = uint16(lo)

	return 0x00
}

func (cpu *Cpu6502) amZPX() byte { return 0x00 }
func (cpu *Cpu6502) amZPY() byte { return 0x00 }

// Absolute:
func (cpu *Cpu6502) amABS() byte {
	// The second byte of the instruction contains the low order byte of the
	// address. The third byte of the instruction contains the high order byte.
	lo := cpu.read(cpu.Pc)
	cpu.Pc++

	hi := cpu.read(cpu.Pc)
	cpu.Pc++

	cpu.addrAbs = uint16(hi)<<8 | uint16(lo)

	return 0x00
}

func (cpu *Cpu6502) amABX() byte { return 0x00 }
func (cpu *Cpu6502) amABY() byte { return 0x00 }
func (cpu *Cpu6502) amIND() byte { return 0x00 }

// Indexed Indirect:
func (cpu *Cpu6502) amIZX() byte {
	// Add the second byte of the instruction with the contents of register X.
	// This result is a zero page memory location pointing to the low order byte
	// of the effective address. The next memory location contains the high
	// order byte. Both memory locations must be in page zero.

	// Get the low order byte of the address.
	addr := cpu.read(cpu.Pc) + cpu.X
	cpu.Pc++

	// Read effective address from page zero.
	lo := cpu.read(uint16(addr))
	hi := cpu.read((uint16(addr) + 1) % 0x0100) // Zero page wraparound
	cpu.addrAbs = uint16(hi)<<8 | uint16(lo)

	return 0x00
}

func (cpu *Cpu6502) amIZY() byte { return 0x00 }

////////////////////////////////////////////////////////////////
// Instructions
type Instruction struct {
	Name     string
	Execute  func() byte
	AddrMode func() byte
	Cycles   byte
}

// CPU insturctions. Each instruction method returns the number of any extra
// cycles necessary for execution.
func (cpu *Cpu6502) opADC() byte { return 0x00 }
func (cpu *Cpu6502) opAND() byte { return 0x00 }

// ASL - Arithmetic Shift Left
func (cpu *Cpu6502) opASL() byte {
	cpu.fetch()

	// Set carry flag to old bit 7.
	cpu.setFlag(StatusFlagC, cpu.fetched&(1<<7) > 0)

	result := cpu.fetched << 1

	// Write result to accumulator register if in implied addressing mode, else
	// write to addrAbs location in memory.
	if cpu.isImpliedAddr {
		cpu.A = result
	} else {
		cpu.write(cpu.addrAbs, result)
	}

	cpu.setFlag(StatusFlagZ, cpu.A == 0)

	// Set negative flag if bit 7 of result is set.
	if result&(1<<7) > 0 {
		cpu.setFlag(StatusFlagN, true)
	}

	return 0x00
}

func (cpu *Cpu6502) opBCC() byte { return 0x00 }
func (cpu *Cpu6502) opBCS() byte { return 0x00 }
func (cpu *Cpu6502) opBEQ() byte { return 0x00 }
func (cpu *Cpu6502) opBIT() byte { return 0x00 }
func (cpu *Cpu6502) opBMI() byte { return 0x00 }
func (cpu *Cpu6502) opBNE() byte { return 0x00 }
func (cpu *Cpu6502) opBPL() byte { return 0x00 }

// BRK - Force Interrupt
func (cpu *Cpu6502) opBRK() byte {
	// Push the high byte of the program counter to the stack.
	cpu.stackPush(byte((cpu.Pc >> 8) & 0xFF))

	// Push the low byte of the program counter to the stack.
	cpu.stackPush(byte(cpu.Pc))

	// Push the CPU status to the stack.
	cpu.stackPush(cpu.Status)

	// Load the IRQ interrupt vector at $FFFE/F to the PC.
	cpu.Pc = cpu.readWord(irqVectAddr)

	// Set break flag to 1.
	cpu.setFlag(StatusFlagB, true)

	return 0x00
}

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

// ORA - Logical Inclusive OR
func (cpu *Cpu6502) opORA() byte {
	cpu.fetch()

	cpu.A |= cpu.fetched

	cpu.setFlag(StatusFlagZ, cpu.A == 0x00)      // if A == 0
	cpu.setFlag(StatusFlagN, (cpu.A&(1<<7) > 0)) // if bit 7 set

	return 0x00
}

func (cpu *Cpu6502) opPHA() byte { return 0x00 }

// PHP - Push Processor Status
func (cpu *Cpu6502) opPHP() byte {
	cpu.stackPush(cpu.Status)

	return 0x00
}

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

// Catch-all instruction for illegal opcodes.
func (cpu *Cpu6502) opXXX() byte { return 0x00 }
