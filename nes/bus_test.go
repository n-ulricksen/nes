package nes

import (
	"io/ioutil"
	"log"
	"testing"
)

func TestInstructions(t *testing.T) {
	nes := NewBus()

	// Load the test file
	romData := readFile("../external_tests/nestest/nestest.nes")
	//romData := []byte{}

	// Location in memory to load ROM
	romOffset := 0x8000

	// Load ROM into memory
	for i, b := range romData {
		nes.Ram[romOffset+i] = b
		//if b != 0 {
		//fmt.Printf("%#x - %#x\n", i, b)
		//}
	}

	//fmt.Printf("%v\n", nes.ram)

	//for i := 0xBF00; i <= 0xC000; i++ {
	//fmt.Printf("%#x: %#x\n", i, nes.ram[i])
	//}
	//fmt.Printf("%#x\n", resetVectAddr)
	//fmt.Printf("%#v\n", nes.ram[resetVectAddr:resetVectAddr+2])

	nes.Cpu.Reset()

	nes.Cpu.Pc = 0x8000

	cyclesToRun := 100
	for i := 0; i < cyclesToRun; i++ {
		nes.Cpu.Cycle()
	}
}

func readFile(filepath string) []byte {
	data, err := ioutil.ReadFile(filepath)

	if err != nil {
		log.Fatalf("Unable to open %v\n%v\n", filepath, err)
	}

	return data
}
