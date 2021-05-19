package nes

import (
	"testing"
)

////////////////////////////////////////////////////////////////
// Addressing Modes
// TODO:
func TestAmIMP(t *testing.T) {}
func TestAmIMM(t *testing.T) {}
func TestAmREL(t *testing.T) {}
func TestAmZP0(t *testing.T) {}
func TestAmZPX(t *testing.T) {}
func TestAmZPY(t *testing.T) {}
func TestAmABS(t *testing.T) {}
func TestAmABX(t *testing.T) {}
func TestAmABY(t *testing.T) {}
func TestAmIND(t *testing.T) {}
func TestAmIZX(t *testing.T) {}
func TestAmIZY(t *testing.T) {}

////////////////////////////////////////////////////////////////
// Instructions
func TestOpAND(t *testing.T) {
	nes := NewBus()
	cpu := nes.cpu

	// Snapshot
	flags := cpu.Status
	before := cpu.A

	// Operate
	cpu.opAND()

	after := cpu.A

	tests := []struct {
		got  interface{}
		want interface{}
	}{
		{cpu.getFlag(StatusFlagC), flags & byte(StatusFlagC)}, // unchanged
		{cpu.getFlag(StatusFlagZ) > 0, cpu.A == 0},            // set if A == 0
		{cpu.getFlag(StatusFlagI), flags & byte(StatusFlagI)}, // unchanged
		{cpu.getFlag(StatusFlagD), flags & byte(StatusFlagD)}, // unchanged
		{cpu.getFlag(StatusFlagB), flags & byte(StatusFlagB)}, // unchanged
		{cpu.getFlag(StatusFlagV), flags & byte(StatusFlagV)}, // unchanged
		{cpu.getFlag(StatusFlagN) > 0, cpu.A&(1<<7) > 0},      // set if bit 7 of accumulator is set

		{before & cpu.fetched, after}, // compare logical AND results
	}

	// Test
	for _, test := range tests {
		if test.got != test.want {
			t.Errorf("got %v, want %v\n", test.got, test.want)
		}
	}
}

func TestOpASL(t *testing.T) {
	nes := NewBus()
	cpu := nes.cpu

	// Snapshot
	flags := cpu.Status
	before := cpu.fetched

	// Operate
	cpu.opASL()

	var after byte
	if cpu.isImpliedAddr {
		after = cpu.A
	} else {
		after = cpu.read(cpu.addrAbs)
	}

	tests := []struct {
		got  interface{}
		want interface{}
	}{
		{cpu.getFlag(StatusFlagC) > 0, before&(1<<7) > 0},     // set to contents of old bit 7
		{cpu.getFlag(StatusFlagZ) > 0, cpu.A == 0},            // set if A == 0
		{cpu.getFlag(StatusFlagI), flags & byte(StatusFlagI)}, // unchanged
		{cpu.getFlag(StatusFlagD), flags & byte(StatusFlagD)}, // unchanged
		{cpu.getFlag(StatusFlagB), flags & byte(StatusFlagB)}, // unchanged
		{cpu.getFlag(StatusFlagV), flags & byte(StatusFlagV)}, // unchanged
		{cpu.getFlag(StatusFlagN) > 0, after&(1<<7) > 0},      // set if bit 7 of result is set

		{before, after}, // compare arithmetic shift left results
	}

	// Test
	for _, test := range tests {
		if test.got != test.want {
			t.Errorf("got %v, want %v\n", test.got, test.want)
		}
	}
}

func TestOpBPL(t *testing.T) {
	nes := NewBus()
	cpu := nes.cpu

	// Snapshot
	flags := cpu.Status

	// Operate
	cpu.opBPL()

	tests := []struct {
		got  interface{}
		want interface{}
	}{
		{cpu.getFlag(StatusFlagC), flags & byte(StatusFlagC)}, // unchanged
		{cpu.getFlag(StatusFlagZ), flags & byte(StatusFlagZ)}, // unchanged
		{cpu.getFlag(StatusFlagI), flags & byte(StatusFlagI)}, // unchanged
		{cpu.getFlag(StatusFlagD), flags & byte(StatusFlagD)}, // unchanged
		{cpu.getFlag(StatusFlagB), flags & byte(StatusFlagB)}, // unchanged
		{cpu.getFlag(StatusFlagV), flags & byte(StatusFlagV)}, // unchanged
		{cpu.getFlag(StatusFlagN), flags & byte(StatusFlagN)}, // unchanged
	}

	// Test
	for _, test := range tests {
		if test.got != test.want {
			t.Errorf("got %v, want %v\n", test.got, test.want)
		}
	}
}

func TestOpBRK(t *testing.T) {
	nes := NewBus()
	cpu := nes.cpu

	// Snapshot
	flags := cpu.Status

	// Operate
	cpu.opBRK()

	tests := []struct {
		got  interface{}
		want interface{}
	}{
		{cpu.getFlag(StatusFlagC), flags & byte(StatusFlagC)}, // unchanged
		{cpu.getFlag(StatusFlagZ), flags & byte(StatusFlagZ)}, // unchanged
		{cpu.getFlag(StatusFlagI), flags & byte(StatusFlagI)}, // unchanged
		{cpu.getFlag(StatusFlagD), flags & byte(StatusFlagD)}, // unchanged
		{cpu.getFlag(StatusFlagB) > 0, true},                  // set to 1
		{cpu.getFlag(StatusFlagV), flags & byte(StatusFlagV)}, // unchanged
		{cpu.getFlag(StatusFlagN), flags & byte(StatusFlagN)}, // unchanged

		{cpu.Pc, cpu.readWord(irqVectAddr)}, // new PC is from IRQ vector
		// TODO: check that the old program counter and status flags are on the
		// stack at [cpu.Sp+2]
	}

	// Test
	for _, test := range tests {
		if test.got != test.want {
			t.Errorf("got %v, want %v\n", test.got, test.want)
		}
	}
}

func TestOpCLC(t *testing.T) {
	nes := NewBus()
	cpu := nes.cpu

	// Snapshot
	flags := cpu.Status

	// Operate
	cpu.opCLC()

	tests := []struct {
		got  interface{}
		want interface{}
	}{
		{cpu.getFlag(StatusFlagC), byte(0)},                   // set to 0
		{cpu.getFlag(StatusFlagZ), flags & byte(StatusFlagZ)}, // unchanged
		{cpu.getFlag(StatusFlagI), flags & byte(StatusFlagI)}, // unchanged
		{cpu.getFlag(StatusFlagD), flags & byte(StatusFlagD)}, // unchanged
		{cpu.getFlag(StatusFlagB), flags & byte(StatusFlagB)}, // unchanged
		{cpu.getFlag(StatusFlagV), flags & byte(StatusFlagV)}, // unchanged
		{cpu.getFlag(StatusFlagN), flags & byte(StatusFlagN)}, // unchanged
	}

	// Test
	for _, test := range tests {
		if test.got != test.want {
			t.Errorf("got %v, want %v\n", test.got, test.want)
		}
	}
}

func TestOpJSR(t *testing.T) {
	nes := NewBus()
	cpu := nes.cpu

	// Snapshot
	flags := cpu.Status

	// Operate
	cpu.opJSR()

	tests := []struct {
		got  interface{}
		want interface{}
	}{
		{cpu.getFlag(StatusFlagC), flags & byte(StatusFlagC)}, // unchanged
		{cpu.getFlag(StatusFlagZ), flags & byte(StatusFlagZ)}, // unchanged
		{cpu.getFlag(StatusFlagI), flags & byte(StatusFlagI)}, // unchanged
		{cpu.getFlag(StatusFlagD), flags & byte(StatusFlagD)}, // unchanged
		{cpu.getFlag(StatusFlagB), flags & byte(StatusFlagB)}, // unchanged
		{cpu.getFlag(StatusFlagV), flags & byte(StatusFlagV)}, // unchanged
		{cpu.getFlag(StatusFlagN), flags & byte(StatusFlagN)}, // unchanged
	}

	// Test
	for _, test := range tests {
		if test.got != test.want {
			t.Errorf("got %v, want %v\n", test.got, test.want)
		}
	}
}

func TestOpORA(t *testing.T) {
	nes := NewBus()
	cpu := nes.cpu

	// Snapshot
	flags := cpu.Status
	regA := cpu.A

	// Operate
	cpu.opORA()

	tests := []struct {
		got  interface{}
		want interface{}
	}{
		{cpu.getFlag(StatusFlagC), flags & byte(StatusFlagC)}, // unchanged
		{cpu.getFlag(StatusFlagZ) > 0, cpu.A == 0},            // set if A == 0
		{cpu.getFlag(StatusFlagI), flags & byte(StatusFlagI)}, // unchanged
		{cpu.getFlag(StatusFlagD), flags & byte(StatusFlagD)}, // unchanged
		{cpu.getFlag(StatusFlagB), flags & byte(StatusFlagB)}, // unchanged
		{cpu.getFlag(StatusFlagV), flags & byte(StatusFlagV)}, // unchanged
		{cpu.getFlag(StatusFlagN) > 0, cpu.A&(1<<7) > 0},      // set if bit 7 set

		{cpu.A, regA | cpu.read(cpu.addrAbs)}, // check the inclusive OR
	}

	// Test
	for _, test := range tests {
		if test.got != test.want {
			t.Errorf("got %v, want %v\n", test.got, test.want)
		}
	}
}

func TestOpPHP(t *testing.T) {
	nes := NewBus()
	cpu := nes.cpu

	// Snapshot
	flags := cpu.Status

	// Operate
	cpu.opPHP()

	tests := []struct {
		got  interface{}
		want interface{}
	}{
		{cpu.getFlag(StatusFlagC), flags & byte(StatusFlagC)}, // unchanged
		{cpu.getFlag(StatusFlagZ), flags & byte(StatusFlagZ)}, // unchanged
		{cpu.getFlag(StatusFlagI), flags & byte(StatusFlagI)}, // unchanged
		{cpu.getFlag(StatusFlagD), flags & byte(StatusFlagD)}, // unchanged
		{cpu.getFlag(StatusFlagB), flags & byte(StatusFlagB)}, // unchanged
		{cpu.getFlag(StatusFlagV), flags & byte(StatusFlagV)}, // unchanged
		{cpu.getFlag(StatusFlagN), flags & byte(StatusFlagN)}, // unchanged

		{cpu.stackPop(), cpu.Status}, // check flags were pushed to stack
	}

	// Test
	for _, test := range tests {
		if test.got != test.want {
			t.Errorf("got %v, want %v\n", test.got, test.want)
		}
	}
}
