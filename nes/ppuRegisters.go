package nes

// PPU Registers
type PpuReg byte
type PpuRegFlag byte

// PPUCTRL flags - $2000
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

// PPUMASK flags - $2001
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

// PPUSTATUS flags - $2002
const (
	statusSpriteOverflow PpuRegFlag = 1 << (iota + 5)
	statusSprite0Hit
	statusVBlank
)

func (r *PpuReg) setFlag(flag PpuRegFlag) {
	*r |= PpuReg(flag)
}

func (r *PpuReg) clearFlag(flag PpuRegFlag) {
	*r &^= PpuReg(flag)
}

func (r *PpuReg) toggleFlag(flag PpuRegFlag) {
	*r ^= PpuReg(flag)
}

func (r *PpuReg) getFlag(flag PpuRegFlag) byte {
	if (*r & PpuReg(flag)) == 0 {
		return 0
	}
	return 1
}
