package nes

type Mapper000 struct {
	PrgBanks byte
	ChrBanks byte
}

func NewMapper000(prgRomChunks, chrRomChunks byte) Mapper000 {
	return Mapper000{
		PrgBanks: prgRomChunks,
		ChrBanks: chrRomChunks,
	}
}

// Address Mapping
//
// if 16KB ROM size:
// 	 0x8000-0xBFFF -> 0x0000-0x3FFF
//   0xC000-0xFFFF -> 0x0000-0x3FFF (mirror)
//
// if 32KB ROM size:
//   0x8000-0xFFFF -> 0x0000-0x7FFF

func (m Mapper000) cpuMapRead(addr uint16) uint16 {
	if addr >= 0x8000 && addr <= 0xFFFF {
		if m.PrgBanks > 1 {
			addr &= 0x7FFF // 32KB ROM
		} else {
			addr &= 0x3FFF // 16KB ROM, need to mirror
		}
	}

	return addr
}

func (m Mapper000) cpuMapWrite(addr uint16) uint16 {
	if addr >= 0x8000 && addr <= 0xFFFF {
		if m.PrgBanks > 1 {
			addr &= 0x3FFF // 16KB ROM, need to mirror
		} else {
			addr &= 0x7FFF // 32KB ROM
		}
	}

	return addr
}

// No PPU mapping
func (m Mapper000) ppuMapRead(addr uint16) uint16 {
	if addr >= 0x0000 && addr <= 0x1FFF {
	}

	return addr
}

func (m Mapper000) ppuMapWrite(addr uint16) uint16 {
	if addr >= 0x0000 && addr <= 0x1FFF {
	}

	return addr
}
