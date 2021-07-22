package nes

// PPU Registers
type PpuReg byte
type PpuRegFlag byte

// PPUCTRL flags
const (
	ctrlNameTblLo PpuRegFlag = 1 << iota
	ctrlNameTblHi
	ctrlVramInc
	ctrlSpritePatternTbl
	ctrlBgPatternTbl
	ctrlSpriteSize
	ctrlExtMode
	ctrlNmi
)

// PPUMASK flags
const (
	maskGreyscale PpuRegFlag = 1 << iota
	maskBgLeft
	maskSpriteLeft
	maskBgShow
	maskSpriteShow
	maskEmphasizeRed
	maskEmphasizeGreen
	maskEmphasizeBlue
)

// PPUSTATUS flags
const (
	statusSpriteOverflow PpuRegFlag = 1 << (iota + 5)
	statusSprite0Hit
	statusVBlank
)

func (r PpuReg) setFlag(flag PpuRegFlag) {
	r |= PpuReg(flag)
}

func (r PpuReg) clearFlag(flag PpuRegFlag) {
	r &^= PpuReg(flag)
}

func (r PpuReg) toggleFlag(flag PpuRegFlag) {
	r ^= PpuReg(flag)
}

func (r PpuReg) isFlagSet(flag PpuRegFlag) bool {
	return (r & PpuReg(flag)) != 0
}
