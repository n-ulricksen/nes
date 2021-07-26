package nes

import (
	"bytes"
	"fmt"
)

// Disassemble the loaded 6502 program into human-readable CPU instructions
// mapped to their respective memory address.
//
// Much help from https://github.com/OneLoneCoder/olcNES
func (cpu *Cpu6502) Disassemble(startAddr, endAddr uint16) map[uint16]string {
	// Current CPU instruction, disassembled
	var lineDiss bytes.Buffer
	var value, lo, hi byte

	// this needs to be bigger than uint16, to determine when larger than endAddr
	var addr uint32 = uint32(startAddr)

	disassembly := make(map[uint16]string)

	for addr <= uint32(endAddr) {
		// Instruction memory address
		lineAddr := uint16(addr)
		lineDiss.WriteString(fmt.Sprintf("$%04X: ", lineAddr))

		// Readable instruction name
		opcode := cpu.read(uint16(addr))
		addr++
		instName := cpu.InstLookup[opcode].Name
		lineDiss.WriteString(fmt.Sprintf("%s ", instName))

		// Get addressing mode.
		addrMode := cpu.InstLookup[opcode].AddrMode

		switch addrMode {
		case IMP:
			lineDiss.WriteString("{IMP}")
		case IMM:
			value = cpu.read(uint16(addr))
			addr++
			lineDiss.WriteString(fmt.Sprintf("#$%04X {IMM}", value))
		case REL:
			value = cpu.read(uint16(addr))
			addr++
			lineDiss.WriteString(fmt.Sprintf("$%04X [%04X] {REL}", value, uint16(value)+uint16(addr)))
		case ZP0:
			lo = cpu.read(uint16(addr))
			addr++
			hi = 0x00
			lineDiss.WriteString(fmt.Sprintf("$%04X {ZP0}", lo))
		case ZPX:
			lo = cpu.read(uint16(addr))
			addr++
			hi = 0x00
			lineDiss.WriteString(fmt.Sprintf("$%04X, X {ZPX}", lo))
		case ZPY:
			lo = cpu.read(uint16(addr))
			addr++
			hi = 0x00
			lineDiss.WriteString(fmt.Sprintf("$%04X, Y {ZPY}", lo))
		case ABS:
			lo = cpu.read(uint16(addr))
			addr++
			hi = cpu.read(uint16(addr))
			addr++
			lineDiss.WriteString(fmt.Sprintf("$%04X {ABS}", uint16(hi)<<8|uint16(lo)))
		case ABX:
			lo = cpu.read(uint16(addr))
			addr++
			hi = cpu.read(uint16(addr))
			addr++
			lineDiss.WriteString(fmt.Sprintf("$%04X, X {ABX}", uint16(hi)<<8|uint16(lo)))
		case ABY:
			lo = cpu.read(uint16(addr))
			addr++
			hi = cpu.read(uint16(addr))
			addr++
			lineDiss.WriteString(fmt.Sprintf("$%04X, Y {ABY}", uint16(hi)<<8|uint16(lo)))
		case IND:
			lo = cpu.read(uint16(addr))
			addr++
			hi = cpu.read(uint16(addr))
			addr++
			lineDiss.WriteString(fmt.Sprintf("($%04X) {IND}", uint16(hi)<<8|uint16(lo)))
		case IZX:
			lo = cpu.read(uint16(addr))
			addr++
			hi = 0x00
			lineDiss.WriteString(fmt.Sprintf("($%04X, X) {IZX}", lo))
		case IZY:
			lo = cpu.read(uint16(addr))
			addr++
			hi = 0x00
			lineDiss.WriteString(fmt.Sprintf("($%04X, Y) {IZY}", lo))
		}

		// Add to map
		disassembly[lineAddr] = lineDiss.String()
		lineDiss.Reset()
	}

	return disassembly
}
