package nes

type Cartridge struct{}

// Communicate with main (CPU) bus.
func (c *Cartridge) cpuRead(addr uint16, data *byte) bool { return false }
func (c *Cartridge) cpuWrite(addr uint16, data byte) bool { return false }

// Communicate with PPU bus.
func (c *Cartridge) ppuRead(addr uint16, data *byte) bool { return false }
func (c *Cartridge) ppuWrite(addr uint16, data byte) bool { return false }
