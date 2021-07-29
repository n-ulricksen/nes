package nes

// Loopy registers are 15 bit internal PPU registers used for implementing
// scrolling.
// Loopy register layout:
//   yyy NN YYYYY XXXXX
//
//   yyy   - fine Y scroll
//   NN    - nametable select
//   YYYYY - coarse Y scroll
//   XXXXX - coarse X scroll
type PpuLoopyReg uint16
type PpuLoopyFlagset uint16

const (
	loopyCoarseX   PpuLoopyFlagset = 0b11111
	loopyCoarseY                   = 0b11111 << 5
	loopyNametable                 = 0b11 << 10
	loopyFineY                     = 0b111 << 12
)

// Returns the value fo the loopy register as a unsigned 16-bit integer.
func (r *PpuLoopyReg) value() uint16 {
	return uint16(*r)
}

// Sets coarse X (bits 0-4) of the loopy register with the low 5 bits of the
// given value.
func (r *PpuLoopyReg) setCoarseX(val byte) {
	// Get relevant 5 bits
	setBits := uint16(val) & 0b11111

	// Clear bits about to be set
	*r &^= PpuLoopyReg(loopyCoarseX)

	// Set new bits
	*r |= PpuLoopyReg(setBits)
}

// Sets coarse Y (bits 5-9) of the loopy register with the low 5 bits of the
// given value.
func (r *PpuLoopyReg) setCoarseY(val byte) {
	// Get relevant 5 bits
	setBits := uint16(val) & 0b11111

	// Clear bits about to be set
	*r &^= PpuLoopyReg(loopyCoarseY)

	// Set new bits
	*r |= PpuLoopyReg(setBits << 5)
}

// Sets nametable (bits 10-11) of the loopy register with the low 2 bits of the
// given value.
func (r *PpuLoopyReg) setNametable(val byte) {
	// Get relevant 2 bits
	setBits := uint16(val) & 0b11

	// Clear bits about to be set
	*r &^= PpuLoopyReg(loopyNametable)

	// Set new bits
	*r |= PpuLoopyReg(setBits << 10)
}

// Sets fine Y (bits 12-14) of the loopy register with the low 3 bits of the
// given value.
func (r *PpuLoopyReg) setFineY(val byte) {
	// Get relevant 3 bits
	setBits := uint16(val) & 0b111

	// Clear bits about to be set
	*r &^= PpuLoopyReg(loopyFineY)

	// Set new bits
	*r |= PpuLoopyReg(setBits << 12)
}
