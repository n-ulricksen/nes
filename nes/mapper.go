package nes

// Mapper functions return whether or not the given address was successfully mapped.
type Mapper interface {
	cpuMapRead(uint16) uint16
	cpuMapWrite(uint16) uint16
	ppuMapRead(uint16) uint16
	ppuMapWrite(uint16) uint16
}
