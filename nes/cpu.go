package nes

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"time"
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
	Cycles        byte   // Remaining cycles for current insturction
	Opcode        byte   // Opcode representing next instruction to be executed
	AddrAbs       uint16 // Set by addressing mode functions, used by instructions
	AddrRel       uint16 // Relative displacement address used for branching
	Fetched       byte   // Byte of memory used by CPU instructions
	CycleCount    uint32 // Total # of cycles executed by the CPU
	isImpliedAddr bool   // Whether the current instruction's address mode is implied

	disassembly map[uint16]string

	InstLookup [16 * 16]Instruction // Instruction operation lookup

	AddrModeFns map[AddressingMode]func() byte // Addressing mode name -> function map

	// TODO: remove this completely
	OpDiss string // Dissasembly for the current instruction, used for debug

	Logger *log.Logger // CPU logging
}

const (
	stackBase uint16 = 0x0100
)

func NewCpu6502() *Cpu6502 {
	cpu := &Cpu6502{
		Pc:     0x0000,
		Sp:     0xFD,
		A:      0x00,
		X:      0x00,
		Y:      0x00,
		Status: 0x00,

		Cycles:        0,
		Opcode:        0x00,
		AddrAbs:       0x0000,
		AddrRel:       0x0000,
		Fetched:       0x00,
		isImpliedAddr: false,
		CycleCount:    0,
	}

	// Create log file.
	now := time.Now()
	logFile := fmt.Sprintf("./logs/cpu%s.log", now.Format("20060102-150405"))
	f, err := os.OpenFile(logFile, os.O_WRONLY|os.O_CREATE, 0664)
	if err != nil {
		log.Fatal("Unable to create CPU log file...\n", err)
	}

	cpu.Logger = log.New(f, "", 0)

	// Create the lookup table containing all the CPU instructions.
	// Reference: http://archive.6502.org/datasheets/rockwell_r650x_r651x.pdf
	//            http://www.oxyron.de/html/opcodes02.html
	cpu.InstLookup = [16 * 16]Instruction{
		{"BRK", cpu.opBRK, IMP, 7}, {"ORA", cpu.opORA, IZX, 6}, {"XXX", cpu.opXXX, IMP, 2}, {"XXX", cpu.opXXX, IMP, 8}, {"NOP", cpu.opNOP, ZP0, 3}, {"ORA", cpu.opORA, ZP0, 3}, {"ASL", cpu.opASL, ZP0, 5}, {"XXX", cpu.opXXX, IMP, 5}, {"PHP", cpu.opPHP, IMP, 3}, {"ORA", cpu.opORA, IMM, 2}, {"ASL", cpu.opASL, IMP, 2}, {"XXX", cpu.opXXX, IMP, 2}, {"NOP", cpu.opNOP, ABS, 4}, {"ORA", cpu.opORA, ABS, 4}, {"ASL", cpu.opASL, ABS, 6}, {"XXX", cpu.opXXX, IMP, 6},

		{"BPL", cpu.opBPL, REL, 2}, {"ORA", cpu.opORA, IZY, 5}, {"XXX", cpu.opXXX, IMP, 2}, {"XXX", cpu.opXXX, IMP, 8}, {"NOP", cpu.opNOP, ZPX, 4}, {"ORA", cpu.opORA, ZPX, 4}, {"ASL", cpu.opASL, ZPX, 6}, {"XXX", cpu.opXXX, IMP, 6}, {"CLC", cpu.opCLC, IMP, 2}, {"ORA", cpu.opORA, ABY, 4}, {"NOP", cpu.opNOP, IMP, 2}, {"XXX", cpu.opXXX, IMP, 7}, {"NOP", cpu.opNOP, ABX, 4}, {"ORA", cpu.opORA, ABX, 4}, {"ASL", cpu.opASL, ABX, 7}, {"XXX", cpu.opXXX, IMP, 7},

		{"JSR", cpu.opJSR, ABS, 6}, {"AND", cpu.opAND, IZX, 6}, {"XXX", cpu.opXXX, IMP, 2}, {"XXX", cpu.opXXX, IMP, 8}, {"BIT", cpu.opBIT, ZP0, 3}, {"AND", cpu.opAND, ZP0, 3}, {"ROL", cpu.opROL, ZP0, 5}, {"XXX", cpu.opXXX, IMP, 5}, {"PLP", cpu.opPLP, IMP, 4}, {"AND", cpu.opAND, IMM, 2}, {"ROL", cpu.opROL, IMP, 2}, {"XXX", cpu.opXXX, IMP, 2}, {"BIT", cpu.opBIT, ABS, 4}, {"AND", cpu.opAND, ABS, 4}, {"ROL", cpu.opROL, ABS, 6}, {"XXX", cpu.opXXX, IMP, 6},

		{"BMI", cpu.opBMI, REL, 2}, {"AND", cpu.opAND, IZY, 5}, {"XXX", cpu.opXXX, IMP, 2}, {"XXX", cpu.opXXX, IMP, 8}, {"NOP", cpu.opNOP, ZPX, 4}, {"AND", cpu.opAND, ZPX, 4}, {"ROL", cpu.opROL, ZPX, 6}, {"XXX", cpu.opXXX, IMP, 6}, {"SEC", cpu.opSEC, IMP, 2}, {"AND", cpu.opAND, ABY, 4}, {"NOP", cpu.opNOP, IMP, 2}, {"XXX", cpu.opXXX, IMP, 7}, {"NOP", cpu.opNOP, ABX, 4}, {"AND", cpu.opAND, ABX, 4}, {"ROL", cpu.opROL, ABX, 7}, {"XXX", cpu.opXXX, IMP, 7},

		{"RTI", cpu.opRTI, IMP, 6}, {"EOR", cpu.opEOR, IZX, 6}, {"XXX", cpu.opXXX, IMP, 2}, {"XXX", cpu.opXXX, IMP, 8}, {"NOP", cpu.opNOP, ZP0, 3}, {"EOR", cpu.opEOR, ZP0, 3}, {"LSR", cpu.opLSR, ZP0, 5}, {"XXX", cpu.opXXX, IMP, 5}, {"PHA", cpu.opPHA, IMP, 3}, {"EOR", cpu.opEOR, IMM, 2}, {"LSR", cpu.opLSR, IMP, 2}, {"XXX", cpu.opXXX, IMP, 2}, {"JMP", cpu.opJMP, ABS, 3}, {"EOR", cpu.opEOR, ABS, 4}, {"LSR", cpu.opLSR, ABS, 6}, {"XXX", cpu.opXXX, IMP, 6},

		{"BVC", cpu.opBVC, REL, 2}, {"EOR", cpu.opEOR, IZY, 5}, {"XXX", cpu.opXXX, IMP, 2}, {"XXX", cpu.opXXX, IMP, 8}, {"NOP", cpu.opNOP, ZPX, 4}, {"EOR", cpu.opEOR, ZPX, 4}, {"LSR", cpu.opLSR, ZPX, 6}, {"XXX", cpu.opXXX, IMP, 6}, {"CLI", cpu.opCLI, IMP, 2}, {"EOR", cpu.opEOR, ABY, 4}, {"NOP", cpu.opNOP, IMP, 2}, {"XXX", cpu.opXXX, IMP, 7}, {"NOP", cpu.opNOP, ABX, 4}, {"EOR", cpu.opEOR, ABX, 4}, {"LSR", cpu.opLSR, ABX, 7}, {"XXX", cpu.opXXX, IMP, 7},

		{"RTS", cpu.opRTS, IMP, 6}, {"ADC", cpu.opADC, IZX, 6}, {"XXX", cpu.opXXX, IMP, 2}, {"XXX", cpu.opXXX, IMP, 8}, {"NOP", cpu.opNOP, ZP0, 3}, {"ADC", cpu.opADC, ZP0, 3}, {"ROR", cpu.opROR, ZP0, 5}, {"XXX", cpu.opXXX, IMP, 5}, {"PLA", cpu.opPLA, IMP, 4}, {"ADC", cpu.opADC, IMM, 2}, {"ROR", cpu.opROR, IMP, 2}, {"XXX", cpu.opXXX, IMP, 2}, {"JMP", cpu.opJMP, IND, 5}, {"ADC", cpu.opADC, ABS, 4}, {"ROR", cpu.opROR, ABS, 6}, {"XXX", cpu.opXXX, IMP, 6},

		{"BVS", cpu.opBVS, REL, 2}, {"ADC", cpu.opADC, IZY, 5}, {"XXX", cpu.opXXX, IMP, 2}, {"XXX", cpu.opXXX, IMP, 8}, {"NOP", cpu.opNOP, ZPX, 4}, {"ADC", cpu.opADC, ZPX, 4}, {"ROR", cpu.opROR, ZPX, 6}, {"XXX", cpu.opXXX, IMP, 6}, {"SEI", cpu.opSEI, IMP, 2}, {"ADC", cpu.opADC, ABY, 4}, {"NOP", cpu.opNOP, IMP, 2}, {"XXX", cpu.opXXX, IMP, 7}, {"NOP", cpu.opNOP, ABX, 4}, {"ADC", cpu.opADC, ABX, 4}, {"ROR", cpu.opROR, ABX, 7}, {"XXX", cpu.opXXX, IMP, 7},

		{"NOP", cpu.opNOP, IMM, 2}, {"STA", cpu.opSTA, IZX, 6}, {"NOP", cpu.opNOP, IMM, 2}, {"XXX", cpu.opXXX, IMP, 6}, {"STY", cpu.opSTY, ZP0, 3}, {"STA", cpu.opSTA, ZP0, 3}, {"STX", cpu.opSTX, ZP0, 3}, {"XXX", cpu.opXXX, IMP, 3}, {"DEY", cpu.opDEY, IMP, 2}, {"NOP", cpu.opNOP, IMM, 2}, {"TXA", cpu.opTXA, IMP, 2}, {"XXX", cpu.opXXX, IMP, 2}, {"STY", cpu.opSTY, ABS, 4}, {"STA", cpu.opSTA, ABS, 4}, {"STX", cpu.opSTX, ABS, 4}, {"XXX", cpu.opXXX, IMP, 6},

		{"BCC", cpu.opBCC, REL, 2}, {"STA", cpu.opSTA, IZY, 6}, {"XXX", cpu.opXXX, IMP, 2}, {"XXX", cpu.opXXX, IMP, 6}, {"STY", cpu.opSTY, ZPX, 4}, {"STA", cpu.opSTA, ZPX, 4}, {"STX", cpu.opSTX, ZPY, 4}, {"XXX", cpu.opXXX, IMP, 4}, {"TYA", cpu.opTYA, IMP, 2}, {"STA", cpu.opSTA, ABY, 5}, {"TXS", cpu.opTXS, IMP, 2}, {"XXX", cpu.opXXX, IMP, 5}, {"XXX", cpu.opXXX, IMP, 5}, {"STA", cpu.opSTA, ABX, 5}, {"XXX", cpu.opXXX, IMP, 5}, {"XXX", cpu.opXXX, IMP, 7},

		{"LDY", cpu.opLDY, IMM, 2}, {"LDA", cpu.opLDA, IZX, 6}, {"LDX", cpu.opLDX, IMM, 2}, {"XXX", cpu.opXXX, IMP, 6}, {"LDY", cpu.opLDY, ZP0, 3}, {"LDA", cpu.opLDA, ZP0, 3}, {"LDX", cpu.opLDX, ZP0, 3}, {"XXX", cpu.opXXX, IMP, 3}, {"TAY", cpu.opTAY, IMP, 2}, {"LDA", cpu.opLDA, IMM, 2}, {"TAX", cpu.opTAX, IMP, 2}, {"XXX", cpu.opXXX, IMP, 2}, {"LDY", cpu.opLDY, ABS, 4}, {"LDA", cpu.opLDA, ABS, 4}, {"LDX", cpu.opLDX, ABS, 4}, {"XXX", cpu.opXXX, IMP, 6},

		{"BCS", cpu.opBCS, REL, 2}, {"LDA", cpu.opLDA, IZY, 5}, {"XXX", cpu.opXXX, IMP, 2}, {"XXX", cpu.opXXX, IMP, 5}, {"LDY", cpu.opLDY, ZPX, 4}, {"LDA", cpu.opLDA, ZPX, 4}, {"LDX", cpu.opLDX, ZPY, 4}, {"XXX", cpu.opXXX, IMP, 4}, {"CLV", cpu.opCLV, IMP, 2}, {"LDA", cpu.opLDA, ABY, 4}, {"TSX", cpu.opTSX, IMP, 2}, {"XXX", cpu.opXXX, IMP, 4}, {"LDY", cpu.opLDY, ABX, 4}, {"LDA", cpu.opLDA, ABX, 4}, {"LDX", cpu.opLDX, ABY, 4}, {"XXX", cpu.opXXX, IMP, 7},

		{"CPY", cpu.opCPY, IMM, 2}, {"CMP", cpu.opCMP, IZX, 6}, {"NOP", cpu.opNOP, IMM, 2}, {"XXX", cpu.opXXX, IMP, 8}, {"CPY", cpu.opCPY, ZP0, 3}, {"CMP", cpu.opCMP, ZP0, 3}, {"DEC", cpu.opDEC, ZP0, 5}, {"XXX", cpu.opXXX, IMP, 5}, {"INY", cpu.opINY, IMP, 2}, {"CMP", cpu.opCMP, IMM, 2}, {"DEX", cpu.opDEX, IMP, 2}, {"XXX", cpu.opXXX, IMP, 2}, {"CPY", cpu.opCPY, ABS, 4}, {"CMP", cpu.opCMP, ABS, 4}, {"DEC", cpu.opDEC, ABS, 6}, {"XXX", cpu.opXXX, IMP, 6},

		{"BNE", cpu.opBNE, REL, 2}, {"CMP", cpu.opCMP, IZY, 5}, {"XXX", cpu.opXXX, IMP, 2}, {"XXX", cpu.opXXX, IMP, 8}, {"NOP", cpu.opNOP, ZPX, 4}, {"CMP", cpu.opCMP, ZPX, 4}, {"DEC", cpu.opDEC, ZPX, 6}, {"XXX", cpu.opXXX, IMP, 6}, {"CLD", cpu.opCLD, IMP, 2}, {"CMP", cpu.opCMP, ABY, 4}, {"NOP", cpu.opNOP, IMP, 2}, {"XXX", cpu.opXXX, IMP, 7}, {"NOP", cpu.opNOP, ABX, 4}, {"CMP", cpu.opCMP, ABX, 4}, {"DEC", cpu.opDEC, ABX, 7}, {"XXX", cpu.opXXX, IMP, 7},

		{"CPX", cpu.opCPX, IMM, 2}, {"SBC", cpu.opSBC, IZX, 6}, {"NOP", cpu.opNOP, IMM, 2}, {"XXX", cpu.opXXX, IMP, 8}, {"CPX", cpu.opCPX, ZP0, 3}, {"SBC", cpu.opSBC, ZP0, 3}, {"INC", cpu.opINC, ZP0, 5}, {"XXX", cpu.opXXX, IMP, 5}, {"INX", cpu.opINX, IMP, 2}, {"SBC", cpu.opSBC, IMM, 2}, {"NOP", cpu.opNOP, IMP, 2}, {"XXX", cpu.opXXX, IMP, 2}, {"CPX", cpu.opCPX, ABS, 4}, {"SBC", cpu.opSBC, ABS, 4}, {"INC", cpu.opINC, ABS, 6}, {"XXX", cpu.opXXX, IMP, 6},

		{"BEQ", cpu.opBEQ, REL, 2}, {"SBC", cpu.opSBC, IZY, 5}, {"XXX", cpu.opXXX, IMP, 2}, {"XXX", cpu.opXXX, IMP, 8}, {"NOP", cpu.opNOP, ZPX, 4}, {"SBC", cpu.opSBC, ZPX, 4}, {"INC", cpu.opINC, ZPX, 6}, {"XXX", cpu.opXXX, IMP, 6}, {"SED", cpu.opSED, IMP, 2}, {"SBC", cpu.opSBC, ABY, 4}, {"NOP", cpu.opNOP, IMP, 2}, {"XXX", cpu.opXXX, IMP, 7}, {"NOP", cpu.opNOP, ABX, 4}, {"SBC", cpu.opSBC, ABX, 4}, {"INC", cpu.opINC, ABX, 7}, {"XXX", cpu.opXXX, IMP, 7},
	}

	// Create an address mode map, used to determine addressing mode function
	// by name.
	cpu.AddrModeFns = map[AddressingMode]func() byte{
		IMP: cpu.amIMP,
		IMM: cpu.amIMM,
		REL: cpu.amREL,
		ZP0: cpu.amZP0,
		ZPX: cpu.amZPX,
		ZPY: cpu.amZPY,
		ABS: cpu.amABS,
		ABX: cpu.amABX,
		ABY: cpu.amABY,
		IND: cpu.amIND,
		IZX: cpu.amIZX,
		IZY: cpu.amIZY,
	}

	return cpu
}

// Connect the CPU to a 16-bit address bus.
func (cpu *Cpu6502) ConnectBus(b *Bus) { cpu.bus = b }

// Read from the attached bus.
func (cpu *Cpu6502) read(addr uint16) byte {
	return cpu.bus.CpuRead(addr)
}

// Write to the attached bus.
func (cpu *Cpu6502) write(addr uint16, data byte) {
	cpu.bus.CpuWrite(addr, data)
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
		cpu.Fetched = cpu.read(cpu.AddrAbs)
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
const (
	resetVectAddr uint16 = 0xFFFC
	irqVectAddr          = 0xFFFE
	nmiVectAddr          = 0xFFFA
)

func (cpu *Cpu6502) Reset() {
	// Clear registers, reset stack pointer
	cpu.A = 0x00
	cpu.X = 0x00
	cpu.Y = 0x00
	cpu.Status = 0x00 | byte(StatusFlagX) | byte(StatusFlagI)
	cpu.Sp = 0xFD

	// Set the program counter to the value at the reset vector address.
	cpu.Pc = cpu.readWord(resetVectAddr)

	// Clear internal variables.
	cpu.Opcode = 0x00
	cpu.AddrAbs = 0x0000
	cpu.AddrRel = 0x0000
	cpu.Fetched = 0x00
	cpu.isImpliedAddr = false
	cpu.CycleCount = 0

	// Spend time on reset
	cpu.Cycles = 7
}

// Interrupt Request
func (cpu *Cpu6502) IRQ() {
	// Push program counter to the stack
	pcHi := byte((cpu.Pc >> 8) & 0x00FF)
	pcLo := byte(cpu.Pc & 0x00FF)
	cpu.stackPush(pcHi)
	cpu.stackPush(pcLo)

	// Set flags: interrupt, break, and unused
	cpu.setFlag(StatusFlagI, true)
	cpu.setFlag(StatusFlagB, true)
	cpu.setFlag(StatusFlagX, true)

	// Push status flag to stack
	cpu.stackPush(cpu.Status)

	// Set program counter to value stored at IRQ vector address
	cpu.Pc = cpu.readWord(irqVectAddr)

	// Spend time on IRQ
	cpu.Cycles = 7
}

// Non-maskable Interrupt Request
func (cpu *Cpu6502) NMI() {
	// Push program counter to the stack
	pcHi := byte((cpu.Pc >> 8) & 0x00FF)
	pcLo := byte(cpu.Pc & 0x00FF)
	cpu.stackPush(pcHi)
	cpu.stackPush(pcLo)

	// Set flags: interrupt, break, and unused
	cpu.setFlag(StatusFlagI, true)
	cpu.setFlag(StatusFlagB, true)
	cpu.setFlag(StatusFlagX, true)

	// Push status flag to stack
	cpu.stackPush(cpu.Status)

	// Set program counter to value stored at NMI vector address
	cpu.Pc = cpu.readWord(nmiVectAddr)

	// Spend time on NMI
	cpu.Cycles = 8
}

// Cycle represents one CPU clock cycle.
func (cpu *Cpu6502) Clock() {
	if cpu.Cycles == 0 {
		// Get the next opcode by reading from the bus at the location of the
		// current program counter.
		cpu.Opcode = cpu.read(cpu.Pc)

		// Store CPU state for logging.
		cpuState := fmt.Sprintf("\t\tA:%02X X:%02X Y:%02X P:%02X SP:%02X\tCYC:%d",
			cpu.A, cpu.X, cpu.Y, cpu.Status, cpu.Sp, cpu.CycleCount)
		oldpc := cpu.Pc

		// Lookup by opcode the instruction to be executed.
		inst := cpu.InstLookup[cpu.Opcode]

		//fmt.Printf("from %#x fetched %#x: %v\n", cpu.Pc, cpu.Opcode, inst)

		// Increment program counter.
		cpu.Pc++

		// Set required cycles for instruction execution.
		cpu.Cycles = inst.Cycles

		// Add any additional cycles needed by either the addressing mode or
		// instruction.
		extraCycles1 := cpu.AddrModeFns[inst.AddrMode]()

		// Execute the instruction.
		extraCycles2 := inst.Execute()

		// Log CPU instructions.
		var buf bytes.Buffer
		buf.WriteString(fmt.Sprintf("%04X\t%02X - %s ", oldpc, cpu.Opcode, inst.Name))
		buf.WriteString(cpuState)
		cpu.Logger.Print(buf.String())
		cpu.OpDiss = buf.String()

		cpu.Cycles += (extraCycles1 & extraCycles2)
	}

	// Turn implied address mode off, just in case the last instruction turned it on.
	cpu.isImpliedAddr = false

	cpu.CycleCount++

	cpu.Cycles--
}

////////////////////////////////////////////////////////////////
// Addressing Modes
// These functions return any extra cycles needed for execution.

// Implied:
func (cpu *Cpu6502) amIMP() byte {
	cpu.isImpliedAddr = true

	cpu.Fetched = cpu.A
	return 0x00
}

// Immediate:
func (cpu *Cpu6502) amIMM() byte {
	// The second byte of the instruction contains the operand.
	cpu.AddrAbs = cpu.Pc
	cpu.Pc++

	return 0x00
}

// Relative:
func (cpu *Cpu6502) amREL() byte {
	addr := cpu.read(cpu.Pc)
	cpu.Pc++

	cpu.AddrRel = uint16(addr)

	// Pad left 8 bits if value is negative.
	if cpu.AddrRel > (1 << 7) {
		cpu.AddrRel |= 0xFF00
	}

	return 0x00
}

// Zero Page:
func (cpu *Cpu6502) amZP0() byte {
	// Use the second byte of the instruction to index into page zero.
	lo := cpu.read(cpu.Pc)
	cpu.Pc++

	cpu.AddrAbs = uint16(lo)

	return 0x00
}

// Zero Page, X
func (cpu *Cpu6502) amZPX() byte {
	cpu.AddrAbs = uint16(cpu.read(cpu.Pc)+cpu.X) & 0x00FF
	cpu.Pc++

	return 0x00
}

// Zero Page, Y
func (cpu *Cpu6502) amZPY() byte {
	cpu.AddrAbs = uint16(cpu.read(cpu.Pc)+cpu.Y) & 0x00FF
	cpu.Pc++

	return 0x00
}

// Absolute:
func (cpu *Cpu6502) amABS() byte {
	// The second byte of the instruction contains the low order byte of the
	// address. The third byte of the instruction contains the high order byte.
	cpu.AddrAbs = cpu.readWord(cpu.Pc)
	cpu.Pc += 2

	return 0x00
}

// Absolute, X:
func (cpu *Cpu6502) amABX() byte {
	// This is the same as absolute addressing, but offsetting by the value in
	// register X.
	addr := cpu.readWord(cpu.Pc)
	cpu.Pc += 2

	cpu.AddrAbs = addr + uint16(cpu.X)

	// Add a cycle if page cross occurred.
	if cpu.AddrAbs&0xFF00 != addr&0xFF00 {
		return 1
	}

	return 0x00
}

// Absolute, Y:
func (cpu *Cpu6502) amABY() byte {
	// This is the same as absolute addressing, but offsetting by the value in
	// register Y.
	addr := cpu.readWord(cpu.Pc)
	cpu.Pc += 2

	cpu.AddrAbs = addr + uint16(cpu.Y)

	// Add a cycle if page cross occurred.
	if cpu.AddrAbs&0xFF00 != addr&0xFF00 {
		return 1
	}

	return 0x00
}

// Indirect:
// This addressing mode contains a hardware bug. If the low byte of the address
// read from PC is 0xFF, a page cross should occur. This bug causes the CPU to
// always wraparound the address, resulting in an invalid effective address.
func (cpu *Cpu6502) amIND() byte {
	// The next 16 bits contain a memory address pointing to the effective address.
	addr := cpu.readWord(cpu.Pc)
	cpu.Pc += 2

	if byte(addr) == 0xFF {
		// Hardware bug.
		lo := cpu.read(addr)
		hi := cpu.read(addr & 0xFF00) // Same page, 0th byte.
		cpu.AddrAbs = uint16(hi)<<8 | uint16(lo)
	} else {
		// Expected behavior.
		cpu.AddrAbs = cpu.readWord(addr)
	}

	return 0x00
}

// Indexed Indirect:
func (cpu *Cpu6502) amIZX() byte {
	// Add the second byte of the instruction with the contents of register X.
	// This result is a zero page memory location pointing to the low order byte
	// of the effective address. The next memory location contains the high
	// order byte. Both memory locations must be in page zero.

	// Get the low order byte of the address.
	addr := (cpu.read(cpu.Pc) + cpu.X) & 0x00FF
	cpu.Pc++

	// Read effective address from page zero.
	lo := cpu.read(uint16(addr))
	hi := cpu.read((uint16(addr) + 1) & 0x00FF) // Zero page wraparound
	cpu.AddrAbs = uint16(hi)<<8 | uint16(lo)

	return 0x00
}

// Indirect Indexed:
func (cpu *Cpu6502) amIZY() byte {
	// The second byte of the instruction points to a zero page memory location.
	// The contents of this memory location are added to the contents of
	// register Y to form the low order byte of the effective address. The carry
	// from this addition is added to the contents of the next page zero memory
	// location to form the high order byte of the effective address.
	addr := uint16(cpu.read(cpu.Pc)) & 0x00FF
	cpu.Pc++

	lo := cpu.read(addr)
	hi := cpu.read((addr + 1) & 0x00FF) // Zero page wraparound

	cpu.AddrAbs = (uint16(hi)<<8 | uint16(lo)) + uint16(cpu.Y)

	// Add a cycle if page cross occurred.
	if cpu.AddrAbs&0xFF00 != (uint16(hi) << 8) {
		return 1
	}

	return 0x00
}

////////////////////////////////////////////////////////////////
// Instructions
type Instruction struct {
	Name     string
	Execute  func() byte
	AddrMode AddressingMode
	Cycles   byte
}

// CPU insturctions. Each instruction method returns the number of any extra
// cycles necessary for execution.

// ADC - Add with Carry
func (cpu *Cpu6502) opADC() byte {
	cpu.fetch()

	// 16-bit to keep any carry.
	result := uint16(cpu.A) + uint16(cpu.Fetched) + uint16(cpu.getFlag(StatusFlagC))

	cpu.setFlag(StatusFlagC, result > 0xFF)
	cpu.setFlag(StatusFlagZ, byte(result) == 0)

	// Set negative flag if bit 7 of result is set.
	cpu.setFlag(StatusFlagN, (result&(1<<7) > 0))

	// Determine if overflow using MSB from accumulator, memory, and result:
	// v = (a == m && a != r)
	a := (cpu.A & (1 << 7))
	m := (cpu.Fetched & (1 << 7))
	r := (byte(result) & (1 << 7))

	cpu.setFlag(StatusFlagV, (a == m) && (a != r))

	cpu.A = byte(result)

	return 0x01 // Potential for extra cycle
}

// AND - Logical AND
func (cpu *Cpu6502) opAND() byte {
	cpu.fetch()

	cpu.A &= cpu.Fetched

	cpu.setFlag(StatusFlagZ, cpu.A == 0)

	// Set if bit 7 of result is set.
	cpu.setFlag(StatusFlagN, cpu.A&(1<<7) > 0)

	return 0x01
}

// ASL - Arithmetic Shift Left
func (cpu *Cpu6502) opASL() byte {
	cpu.fetch()

	// Set carry flag to old bit 7.
	cpu.setFlag(StatusFlagC, cpu.Fetched&(1<<7) > 0)

	result := cpu.Fetched << 1

	// Write result to accumulator register if in implied addressing mode, else
	// write to addrAbs location in memory.
	if cpu.isImpliedAddr {
		cpu.A = result
	} else {
		cpu.write(cpu.AddrAbs, result)
	}

	cpu.setFlag(StatusFlagZ, cpu.A == 0)

	// Set if bit 7 of result is set.
	cpu.setFlag(StatusFlagN, result&(1<<7) > 0)

	return 0x00
}

// BCC - Branch if Carry Clear
func (cpu *Cpu6502) opBCC() byte {
	if cpu.getFlag(StatusFlagC) == 0 {
		// Extra cycle when branch succeeds
		cpu.Cycles++

		cpu.AddrAbs = cpu.Pc + cpu.AddrRel

		if cpu.AddrAbs&0xFF00 != cpu.Pc&0xFF00 {
			// Extra cycle if cross pages
			cpu.Cycles++
		}

		cpu.Pc = cpu.AddrAbs
	}

	return 0x00
}

// BCS - Branch if Carry Set
func (cpu *Cpu6502) opBCS() byte {
	if cpu.getFlag(StatusFlagC) != 0 {
		// Extra cycle when branch succeeds
		cpu.Cycles++

		cpu.AddrAbs = cpu.Pc + cpu.AddrRel

		if cpu.AddrAbs&0xFF00 != cpu.Pc&0xFF00 {
			// Extra cycle if cross pages
			cpu.Cycles++
		}

		cpu.Pc = cpu.AddrAbs
	}

	return 0x00
}

// BEQ - Branch if Equal
func (cpu *Cpu6502) opBEQ() byte {
	if cpu.getFlag(StatusFlagZ) != 0 {
		// Extra cycle if branch succeeds
		cpu.Cycles++

		cpu.AddrAbs = cpu.Pc + cpu.AddrRel

		if cpu.AddrAbs&0xFF00 != cpu.Pc&0xFF00 {
			// Extra cycle if cross pages
			cpu.Cycles++
		}

		cpu.Pc = cpu.AddrAbs
	}

	return 0x00
}

// BIT - Bit Test
func (cpu *Cpu6502) opBIT() byte {
	cpu.fetch()

	result := cpu.Fetched & cpu.A

	cpu.setFlag(StatusFlagZ, result == 0)

	// Set if bit 6 of result is set.
	cpu.setFlag(StatusFlagV, cpu.Fetched&(1<<6) > 0)

	// Set if bit 7 of result is set.
	cpu.setFlag(StatusFlagN, cpu.Fetched&(1<<7) > 0)

	return 0x00
}

// BMI - Branch if Minus
func (cpu *Cpu6502) opBMI() byte {
	if cpu.getFlag(StatusFlagN) != 0 {
		// Extra cycle when branch succeeds
		cpu.Cycles++

		cpu.AddrAbs = cpu.Pc + cpu.AddrRel

		if cpu.AddrAbs&0xFF00 != cpu.Pc&0xFF00 {
			// Extra cycle if cross pages
			cpu.Cycles++
		}

		cpu.Pc = cpu.AddrAbs
	}

	return 0x00
}

// BNE - Branch if Not Equal
func (cpu *Cpu6502) opBNE() byte {
	if cpu.getFlag(StatusFlagZ) == 0 {
		// Extra cycle if branch succeeds
		cpu.Cycles++

		cpu.AddrAbs = cpu.Pc + cpu.AddrRel

		if cpu.AddrAbs&0xFF00 != cpu.Pc&0xFF00 {
			// Extra cycle if cross pages
			cpu.Cycles++
		}

		cpu.Pc = cpu.AddrAbs
	}

	return 0x00
}

// BPL - Branch if Positive
func (cpu *Cpu6502) opBPL() byte {
	if cpu.getFlag(StatusFlagN) == 0 {
		// Extra cycle if branch succeeds
		cpu.Cycles++

		cpu.AddrAbs = cpu.Pc + cpu.AddrRel

		if cpu.AddrAbs&0xFF00 != cpu.Pc&0xFF00 {
			// Extra cycle if cross pages
			cpu.Cycles++
		}

		cpu.Pc = cpu.AddrAbs
	}

	return 0x00
}

// BRK - Force Interrupt
func (cpu *Cpu6502) opBRK() byte {
	// Push the high byte of the program counter to the stack.
	cpu.stackPush(byte((cpu.Pc >> 8) & 0xFF))

	// Push the low byte of the program counter to the stack.
	cpu.stackPush(byte(cpu.Pc))

	// Push the CPU status to the stack.
	// Set B flag according to: http://visual6502.org/wiki/index.php?title=6502_BRK_and_B_bit
	cpu.stackPush(cpu.Status | byte(StatusFlagB))

	// Load the IRQ interrupt vector at $FFFE/F to the PC.
	cpu.Pc = cpu.readWord(irqVectAddr)

	// Set break flag to 1.
	cpu.setFlag(StatusFlagB, true)

	return 0x00
}

// BVC - Branch if Overflow Clear
func (cpu *Cpu6502) opBVC() byte {
	if cpu.getFlag(StatusFlagV) == 0 {
		// Add cycle if branch succeeds
		cpu.Cycles++

		cpu.AddrAbs = cpu.Pc + cpu.AddrRel

		if cpu.AddrAbs&0xFF00 != cpu.Pc&0xFF00 {
			// Extra cycle if cross pages
			cpu.Cycles++
		}

		cpu.Pc = cpu.AddrAbs
	}

	return 0x00
}

// BVS - Branch if Overflow Set
func (cpu *Cpu6502) opBVS() byte {
	if cpu.getFlag(StatusFlagV) > 0 {
		// Add cycle if branch succeeds
		cpu.Cycles++

		cpu.AddrAbs = cpu.Pc + cpu.AddrRel

		if cpu.AddrAbs&0xFF00 != cpu.Pc&0xFF00 {
			// Extra cycle if cross pages
			cpu.Cycles++
		}

		cpu.Pc = cpu.AddrAbs
	}

	return 0x00
}

// CLC - Clear Carry Flag
func (cpu *Cpu6502) opCLC() byte {
	cpu.setFlag(StatusFlagC, false)

	return 0x00
}

// CLD - Clear Decimal Mode
func (cpu *Cpu6502) opCLD() byte {
	cpu.setFlag(StatusFlagD, false)

	return 0x00
}

// CLI - Clear Interrupt Disable
func (cpu *Cpu6502) opCLI() byte {
	cpu.setFlag(StatusFlagI, false)

	return 0x00
}

// CLV - Clear Overflow Flag
func (cpu *Cpu6502) opCLV() byte {
	cpu.setFlag(StatusFlagV, false)

	return 0x00
}

// CMP - Compare (Accumulator)
func (cpu *Cpu6502) opCMP() byte {
	cpu.fetch()

	cpu.setFlag(StatusFlagC, cpu.A >= cpu.Fetched)
	cpu.setFlag(StatusFlagZ, cpu.A == cpu.Fetched)
	cpu.setFlag(StatusFlagN, ((cpu.A-cpu.Fetched)&(1<<7) > 0)) // if bit 7 set

	return 0x01
}

// CPX - Compare X Register
func (cpu *Cpu6502) opCPX() byte {
	cpu.fetch()

	cpu.setFlag(StatusFlagC, cpu.X >= cpu.Fetched)
	cpu.setFlag(StatusFlagZ, cpu.X == cpu.Fetched)
	cpu.setFlag(StatusFlagN, ((cpu.X-cpu.Fetched)&(1<<7) > 0)) // if bit 7 set

	return 0x00
}

// CPY - Compare Y Register
func (cpu *Cpu6502) opCPY() byte {
	cpu.fetch()

	cpu.setFlag(StatusFlagC, cpu.Y >= cpu.Fetched)
	cpu.setFlag(StatusFlagZ, cpu.Y == cpu.Fetched)
	cpu.setFlag(StatusFlagN, ((cpu.Y-cpu.Fetched)&(1<<7) > 0)) // if bit 7 set

	return 0x00
}

// DEC - Decrement Memory
func (cpu *Cpu6502) opDEC() byte {
	cpu.fetch()

	cpu.Fetched--

	cpu.write(cpu.AddrAbs, cpu.Fetched)

	cpu.setFlag(StatusFlagZ, cpu.Fetched == 0)         // if A == 0
	cpu.setFlag(StatusFlagN, (cpu.Fetched&(1<<7) > 0)) // if bit 7 set

	return 0x00
}

// DEX - Decrement X Register
func (cpu *Cpu6502) opDEX() byte {
	cpu.X--

	cpu.setFlag(StatusFlagZ, cpu.X == 0)

	// Set negative flag if bit 7 of X register is set.
	cpu.setFlag(StatusFlagN, cpu.X&(1<<7) > 0)

	return 0x00
}

// DEY - Decrement Y Register
func (cpu *Cpu6502) opDEY() byte {
	cpu.Y--

	cpu.setFlag(StatusFlagZ, cpu.Y == 0)

	// Set negative flag if bit 7 of Y register is set.
	cpu.setFlag(StatusFlagN, cpu.Y&(1<<7) > 0)

	return 0x00
}

// EOR - Exclusive OR
func (cpu *Cpu6502) opEOR() byte {
	cpu.fetch()

	cpu.A ^= cpu.Fetched

	cpu.setFlag(StatusFlagZ, cpu.A == 0)

	// Set negative flag if bit 7 is set.
	cpu.setFlag(StatusFlagN, cpu.A&(1<<7) > 0)

	return 0x01
}

// INC - Increment Memory
func (cpu *Cpu6502) opINC() byte {
	cpu.fetch()

	cpu.Fetched++

	cpu.write(cpu.AddrAbs, cpu.Fetched)

	cpu.setFlag(StatusFlagZ, cpu.Fetched == 0)         // if A == 0
	cpu.setFlag(StatusFlagN, (cpu.Fetched&(1<<7) > 0)) // if bit 7 set

	return 0x00
}

// INX - Increment X Register
func (cpu *Cpu6502) opINX() byte {
	cpu.X++

	cpu.setFlag(StatusFlagZ, cpu.X == 0)         // if X == 0
	cpu.setFlag(StatusFlagN, (cpu.X&(1<<7) > 0)) // if bit 7 set

	return 0x00
}

// INY - Increment Y Register
func (cpu *Cpu6502) opINY() byte {
	cpu.Y++

	cpu.setFlag(StatusFlagZ, cpu.Y == 0)         // if Y == 0
	cpu.setFlag(StatusFlagN, (cpu.Y&(1<<7) > 0)) // if bit 7 set

	return 0x00
}

// JMP - Jump
func (cpu *Cpu6502) opJMP() byte {
	cpu.Pc = cpu.AddrAbs

	return 0x00
}

// JSR - Jump to Subroutine
func (cpu *Cpu6502) opJSR() byte {
	// Push the address (minus 1) of the return point to the stack, and sets
	// the program counter to the target memory address.
	cpu.Pc--

	// Push the high byte of the program counter to the stack.
	cpu.stackPush(byte((cpu.Pc >> 8) & 0xFF))

	// Push the low byte of the program counter to the stack.
	cpu.stackPush(byte(cpu.Pc))

	// Set program counter to the given address.
	cpu.Pc = cpu.AddrAbs

	return 0x00
}

// LDA - Load Accumulator
func (cpu *Cpu6502) opLDA() byte {
	cpu.fetch()

	cpu.A = cpu.Fetched

	cpu.setFlag(StatusFlagZ, cpu.A == 0)         // if A == 0
	cpu.setFlag(StatusFlagN, (cpu.A&(1<<7) > 0)) // if bit 7 set

	return 0x01
}

// LDX - Load X Register
func (cpu *Cpu6502) opLDX() byte {
	cpu.fetch()

	cpu.X = cpu.Fetched

	cpu.setFlag(StatusFlagZ, cpu.X == 0)         // if X == 0
	cpu.setFlag(StatusFlagN, (cpu.X&(1<<7) > 0)) // if bit 7 set

	return 0x01
}

// LDY - Load Y Register
func (cpu *Cpu6502) opLDY() byte {
	cpu.fetch()

	cpu.Y = cpu.Fetched

	cpu.setFlag(StatusFlagZ, cpu.Y == 0)         // if Y == 0
	cpu.setFlag(StatusFlagN, (cpu.Y&(1<<7) > 0)) // if bit 7 set

	return 0x01
}

// LSR - Logical Shift Right
func (cpu *Cpu6502) opLSR() byte {
	cpu.fetch()

	// Set carry flag to old bit 0.
	cpu.setFlag(StatusFlagC, cpu.Fetched&0x1 > 0)

	cpu.Fetched = cpu.Fetched >> 1

	cpu.setFlag(StatusFlagZ, cpu.Fetched == 0)
	cpu.setFlag(StatusFlagN, (cpu.Fetched&(1<<7) > 0)) // if bit 7 set

	if cpu.isImpliedAddr {
		cpu.A = cpu.Fetched
	} else {
		cpu.write(cpu.AddrAbs, cpu.Fetched)
	}

	return 0x00
}

// NOP - No Operation
func (cpu *Cpu6502) opNOP() byte { return 0x01 }

// ORA - Logical Inclusive OR
func (cpu *Cpu6502) opORA() byte {
	cpu.fetch()

	cpu.A |= cpu.Fetched

	cpu.setFlag(StatusFlagZ, cpu.A == 0)         // if A == 0
	cpu.setFlag(StatusFlagN, (cpu.A&(1<<7) > 0)) // if bit 7 set

	return 0x01
}

// PHA - Push Accumulator
func (cpu *Cpu6502) opPHA() byte {
	cpu.stackPush(cpu.A)
	return 0x00
}

// PHP - Push Processor Status
func (cpu *Cpu6502) opPHP() byte {
	// Set B flag according to: http://visual6502.org/wiki/index.php?title=6502_BRK_and_B_bit
	cpu.stackPush(cpu.Status | byte(StatusFlagB))

	return 0x00
}

// PLA - Pull Accumulator
func (cpu *Cpu6502) opPLA() byte {
	// Pull value from stack to accumulator.
	cpu.A = cpu.stackPop()

	cpu.setFlag(StatusFlagZ, cpu.A == 0)         // if A == 0
	cpu.setFlag(StatusFlagN, (cpu.A&(1<<7) > 0)) // if bit 7 set

	return 0x00
}

// PLP - Pull Processor Status
func (cpu *Cpu6502) opPLP() byte {
	// Load processor status flags from the stack. B flag should remain unchanged.
	bFlag := cpu.getFlag(StatusFlagB) > 0
	cpu.Status = cpu.stackPop()
	cpu.setFlag(StatusFlagB, bFlag)

	// Always set unused flag.
	cpu.setFlag(StatusFlagX, true)

	return 0x00
}

// ROL - Rotate Left
func (cpu *Cpu6502) opROL() byte {
	cpu.fetch()

	carry := cpu.getFlag(StatusFlagC)

	// Set carry flag to bit 7 of old value.
	cpu.setFlag(StatusFlagC, cpu.Fetched&(1<<7) > 0)

	// Shift left one, set bit 1 to old carry.
	cpu.Fetched = (cpu.Fetched << 1) | carry

	cpu.setFlag(StatusFlagZ, cpu.Fetched == 0)

	// Set negative flag to bit 7 of new value.
	cpu.setFlag(StatusFlagN, cpu.Fetched&(1<<7) > 0)

	if cpu.isImpliedAddr {
		cpu.A = cpu.Fetched
	} else {
		cpu.write(cpu.AddrAbs, cpu.Fetched)
	}

	return 0x00
}

// ROR - Rotate Right
func (cpu *Cpu6502) opROR() byte {
	cpu.fetch()

	carry := cpu.getFlag(StatusFlagC)

	// Set carry flag to bit 1 of old value.
	cpu.setFlag(StatusFlagC, cpu.Fetched&1 > 0)

	// Shift right one, set bit 7 to old carry.
	cpu.Fetched = (cpu.Fetched >> 1) | (carry << 7)

	cpu.setFlag(StatusFlagZ, cpu.Fetched == 0)

	// Set negative flag to bit 7 of new value.
	cpu.setFlag(StatusFlagN, cpu.Fetched&(1<<7) > 0)

	if cpu.isImpliedAddr {
		cpu.A = cpu.Fetched
	} else {
		cpu.write(cpu.AddrAbs, cpu.Fetched)
	}

	return 0x00
}

// RTI - Return from Interrupt
func (cpu *Cpu6502) opRTI() byte {
	// Pull the status flags then the program counter form the stack. B flag should
	// remain unchanged.
	bFlag := cpu.getFlag(StatusFlagB) > 0
	cpu.Status = cpu.stackPop()
	cpu.setFlag(StatusFlagB, bFlag)

	// Always set unused flag.
	cpu.setFlag(StatusFlagX, true)

	lo := cpu.stackPop()
	hi := cpu.stackPop()

	cpu.Pc = uint16(hi)<<8 | uint16(lo)

	return 0x00
}

// RTS - Return from Subroutine
func (cpu *Cpu6502) opRTS() byte {
	// Pull the program counter from the stack.
	lo := cpu.stackPop()
	hi := cpu.stackPop()

	cpu.Pc = uint16(hi)<<8 | uint16(lo)

	// Increment PC (JSR decrements it)
	cpu.Pc++

	return 0x00
}

// SBC - Subtract with Carry
func (cpu *Cpu6502) opSBC() byte {
	cpu.fetch()

	// Invert to subtract
	sub := uint16(cpu.Fetched) ^ 0x00FF

	// 16-bit to keep any carry.
	result := uint16(cpu.A) + sub + uint16(cpu.getFlag(StatusFlagC))

	cpu.setFlag(StatusFlagC, result > 0xFF)
	cpu.setFlag(StatusFlagZ, byte(result) == 0)

	// Set negative flag if bit 7 of result is set.
	cpu.setFlag(StatusFlagN, (result&(1<<7) > 0))

	// Determine if overflow using MSB from accumulator, memory, and result:
	// v = (a != m && m == r)
	a := (cpu.A & (1 << 7))
	m := (cpu.Fetched & (1 << 7))
	r := (byte(result) & (1 << 7))

	cpu.setFlag(StatusFlagV, (a != m) && (m == r))

	cpu.A = byte(result)

	return 0x01
}

// SEC - Set Carry Flag
func (cpu *Cpu6502) opSEC() byte {
	cpu.setFlag(StatusFlagC, true)

	return 0x00
}

// SED - Set Decimal Flag
func (cpu *Cpu6502) opSED() byte {
	cpu.setFlag(StatusFlagD, true)

	return 0x00
}

// SEI - Set Interrupt Disable
func (cpu *Cpu6502) opSEI() byte {
	cpu.setFlag(StatusFlagI, true)

	return 0x00
}

// STA - Store Accumulator
func (cpu *Cpu6502) opSTA() byte {
	cpu.write(cpu.AddrAbs, cpu.A)

	return 0x00
}

// STX - Store X Register
func (cpu *Cpu6502) opSTX() byte {
	cpu.write(cpu.AddrAbs, cpu.X)

	return 0x00
}

// STY - Store Y Register
func (cpu *Cpu6502) opSTY() byte {
	cpu.write(cpu.AddrAbs, cpu.Y)

	return 0x00
}

// TAX - Transfer Accumulator to X
func (cpu *Cpu6502) opTAX() byte {
	cpu.X = cpu.A

	cpu.setFlag(StatusFlagZ, cpu.X == 0)         // if X == 0
	cpu.setFlag(StatusFlagN, (cpu.X&(1<<7) > 0)) // if bit 7 set

	return 0x00
}

// TAY - Transfer Accumulator to Y
func (cpu *Cpu6502) opTAY() byte {
	cpu.Y = cpu.A

	cpu.setFlag(StatusFlagZ, cpu.Y == 0)         // if Y == 0
	cpu.setFlag(StatusFlagN, (cpu.Y&(1<<7) > 0)) // if bit 7 set

	return 0x00
}

// TSX - Transfer Stack Pointer to X
func (cpu *Cpu6502) opTSX() byte {
	cpu.X = cpu.Sp

	cpu.setFlag(StatusFlagZ, cpu.X == 0)         // if X == 0
	cpu.setFlag(StatusFlagN, (cpu.X&(1<<7) > 0)) // if bit 7 set

	return 0x00
}

// TXA - Transfer X to Accumulator
func (cpu *Cpu6502) opTXA() byte {
	cpu.A = cpu.X

	cpu.setFlag(StatusFlagZ, cpu.A == 0)         // if A == 0
	cpu.setFlag(StatusFlagN, (cpu.A&(1<<7) > 0)) // if bit 7 set

	return 0x00
}

// TXS - Transfer X to Stack Pointer
func (cpu *Cpu6502) opTXS() byte {
	cpu.Sp = cpu.X

	return 0x00
}

// TYA - Transfer Y to Accumulator
func (cpu *Cpu6502) opTYA() byte {
	cpu.A = cpu.Y

	cpu.setFlag(StatusFlagZ, cpu.A == 0)         // if A == 0
	cpu.setFlag(StatusFlagN, (cpu.A&(1<<7) > 0)) // if bit 7 set

	return 0x00
}

// Catch-all instruction for illegal opcodes.
func (cpu *Cpu6502) opXXX() byte { return 0x00 }
