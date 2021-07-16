package nes

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"log"
)

// NES Cartridge. Connected to both main bus and PPU bus.
// The cartridges consist of program (PRG) memory and character (CHR) memory.
type Cartridge struct {
	prgMem []byte // Program memory (PRG)
	chrMem []byte // Character memory (CHR)

	mapper Mapper // Cartridge mapper used to configure CPU/PPU read/write addresses.
}

// iNES file header
// reference: https://wiki.nesdev.com/w/index.php/INES
type CartridgeHeader struct {
	Name         [4]byte // Constant "NES" followed by MS-DOS end of file
	PrgRomChunks byte    // Program memory size in 16KB chunks
	ChrRomChunks byte    // Character memory size in 8KB chunks
	Mapper1      byte    // Flags 6
	Mapper2      byte    // Flags 7
	PrgRamSize   byte    // Flags 8
	TvSystem1    byte    // Flags 9
	TvSystem2    byte    // Flags 10
	Unused       [5]byte // Unused padding
}

// Creates a new NES Cartridge using the file at the given path.
func NewCartridge(filepath string) *Cartridge {
	// Load the NES file.
	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		log.Fatalf("Unable to open %v\n%v\n", filepath, err)
	}
	fmt.Printf("Header data: % x\n", data[:16])
	buf := bytes.NewBuffer(data)

	// Read/decode the NES header.
	header := new(CartridgeHeader)
	err = binary.Read(buf, binary.BigEndian, header)
	if err != nil {
		log.Fatalf("Unable to parse header\n%v\n", err)
	}
	fmt.Printf("Parsed cartridge header: %+v\n\n", header)

	// Check if trainer is used (bit 3 of mapper1 flags).
	if (header.Mapper1 & (0x1 << 3)) > 0 {
		// 512-byte trainer
		// XXX: ignoring trainer data for now
		err = binary.Read(buf, binary.BigEndian, make([]byte, 512))
		if err != nil {
			log.Fatalf("Unable to read trainer data\n%v\n", err)
		}
	}

	// TODO: determine iNES version (0/1/2)

	cartridge := new(Cartridge)

	// Determine mapper ID from high 4 bits of mapper flags.
	mapperLo := header.Mapper1 >> 4
	mapperHi := header.Mapper2 >> 4
	mapperId := (mapperHi << 4) | mapperLo

	// Set Mapper
	var mapper Mapper
	switch mapperId {
	case 0:
		mapper = NewMapper000(header.PrgRomChunks, header.ChrRomChunks)
	}
	cartridge.mapper = mapper
	fmt.Println("Mapper ID:", mapperId)
	fmt.Println("Mapper:", mapper)

	// Read/load PRG memory (16KB chunks).
	cartridge.prgMem = make([]byte, 16*1024*int(header.PrgRomChunks))
	fmt.Printf("PRG ROM size: %v\n", len(cartridge.prgMem))
	err = binary.Read(buf, binary.BigEndian, cartridge.prgMem)
	if err != nil {
		log.Fatalf("Unable to read PRG memory\n%v\n", err)
	}

	// Read/load CHR memory (8KB chunks).
	cartridge.chrMem = make([]byte, 8*1024*int(header.ChrRomChunks))
	fmt.Printf("CHR ROM size: %v\n", len(cartridge.chrMem))
	err = binary.Read(buf, binary.BigEndian, cartridge.chrMem)
	if err != nil {
		log.Fatalf("Unable to read CHR memory\n%v\n", err)
	}

	// Determine if PlayChoice INST-ROM (bit 2 of mapper2 flags).
	if (header.Mapper2 & (0x1 << 2)) > 0 {
		// 8192-bytes
		// XXX: ignoring INST-ROM data for now
		err = binary.Read(buf, binary.BigEndian, make([]byte, 8192))
		if err != nil {
			log.Fatalf("Unable to read PlayChoice INST-ROM data\n%v\n", err)
		}
	}

	return nil
}

// Communicate with main (CPU) bus.
func (c *Cartridge) cpuRead(addr uint16, data *byte) bool {
	var mappedAddr uint16
	if c.mapper.cpuMapRead(addr, &mappedAddr) {
		*data = c.prgMem[mappedAddr]
		return true
	}

	return false
}

func (c *Cartridge) cpuWrite(addr uint16, data byte) bool {
	var mappedAddr uint16
	if c.mapper.cpuMapWrite(addr, &mappedAddr) {
		c.prgMem[mappedAddr] = data
		return true
	}

	return false
}

// Communicate with PPU bus.
func (c *Cartridge) ppuRead(addr uint16, data *byte) bool {
	var mappedAddr uint16
	if c.mapper.ppuMapRead(addr, &mappedAddr) {
		*data = c.chrMem[mappedAddr]
		return true
	}

	return false
}

func (c *Cartridge) ppuWrite(addr uint16, data byte) bool {
	var mappedAddr uint16
	if c.mapper.ppuMapRead(addr, &mappedAddr) {
		c.chrMem[mappedAddr] = data
		return true
	}

	return false
}
