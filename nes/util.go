package nes

import (
	"fmt"
	"log"
	"regexp"
	"runtime"
	"time"
)

// Function time tracking thanks to:
// https://stackoverflow.com/questions/45766572/is-there-an-efficient-way-to-calculate-execution-time-in-golang
func TimeTrack(start time.Time) {
	elapsed := time.Since(start)

	// Skip this function, and fetch the PC and file for its parent.
	pc, _, _, _ := runtime.Caller(1)

	// Retrieve a function object this functions parent.
	funcObj := runtime.FuncForPC(pc)

	// Regex to extract just the function name (and not the module path).
	runtimeFunc := regexp.MustCompile(`^.*\.(.*)$`)
	name := runtimeFunc.ReplaceAllString(funcObj.Name(), "$1")

	log.Println(fmt.Sprintf("%s took %s", name, elapsed))
}

// Flip a byte's bits.
func flipByte(b byte) byte {
	for i := 0; i < 4; i++ {
		bitLo := i
		bitHi := 7 - i

		newLo := (b & (1 << bitHi)) >> bitHi
		newHi := b & (1 << bitLo)

		setBit(&b, bitLo, newLo)
		setBit(&b, bitHi, newHi)
	}

	return b
}

// Set a bit in b at the given bit index.
func setBit(b *byte, bitIdx int, newBit byte) {
	if newBit == 0 {
		*b &^= (1 << bitIdx)
	} else {
		*b |= (1 << bitIdx)
	}
}
