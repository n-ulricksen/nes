package nes

type AddressingMode int

const (
	IMP AddressingMode = iota
	IMM
	REL
	ZP0
	ZPX
	ZPY
	ABS
	ABX
	ABY
	IND
	IZX
	IZY
)
