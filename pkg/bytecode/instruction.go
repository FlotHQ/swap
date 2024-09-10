package bytecode

import "fmt"

type Instruction uint64

func (i Instruction) String() string {
	unpacked := UnpackedInstruction{}
	unpacked.Unpack(i)
	return fmt.Sprintf("%s, %d, %d, %d", unpacked.Op, unpacked.A, unpacked.B, unpacked.C)
}

type OpCode byte

const (
	OpPrintConst OpCode = iota
	OpResolvePrint
	OpMove
	OpCall
	OpLoadConst
	OpResolveLoad
	OpLoopStart
	OpLoopEnd
	OpHalt
)

func (op OpCode) String() string {
	switch op {
	case OpPrintConst:
		return "OpPrintConst"
	case OpResolvePrint:
		return "OpResolvePrint"
	case OpMove:
		return "OpMove"
	case OpCall:
		return "OpCall"
	case OpLoopStart:
		return "OpLoopStart"
	case OpLoopEnd:
		return "OpLoopEnd"
	case OpHalt:
		return "OpHalt"
	case OpResolveLoad:
		return "OpResolveLoad"
	case OpLoadConst:
		return "OpLoadConst"
	default:
		return "Unknown"
	}
}

type UnpackedInstruction struct {
	Op OpCode
	A  uint8
	B  uint8
	C  uint8
}

func (u *UnpackedInstruction) Reset() {
	u.Op = 0
	u.A = 0
	u.B = 0
	u.C = 0
}

func (u *UnpackedInstruction) Unpack(i Instruction) {
	u.Op = OpCode(i & 0xFF)
	u.A = uint8((i >> 8) & 0xFF)
	u.B = uint8((i >> 24) & 0xFF)
	u.C = uint8((i >> 40) & 0xFF)
}

func PackInstruction(op OpCode, a, b, c uint8) Instruction {
	return Instruction(uint64(op) | uint64(a)<<8 | uint64(b)<<24 | uint64(c)<<40)
}
