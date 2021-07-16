package nes

// Mapper functions return whether or not the given address was successfully mapped.
type Mapper interface {
	cpuMapRead(uint16, *uint16) bool
	cpuMapWrite(uint16, *uint16) bool
	ppuMapRead(uint16, *uint16) bool
	ppuMapWrite(uint16, *uint16) bool
}
