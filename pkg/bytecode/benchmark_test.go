package bytecode

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"testing"
)

func BenchmarkSerializeBytecode(b *testing.B) {
	sizes := []int{10, 100, 1000, 10000}
	for _, size := range sizes {
		b.Run(fmt.Sprintf("Size-%d", size), func(b *testing.B) {
			instructions := generateLargeInstructionSet(size)
			constants := generateConstants(size)
			buf := &bytes.Buffer{}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				buf.Reset()
				err := SerializeBytecode(buf, instructions, constants)
				if err != nil {
					b.Fatalf("Serialization failed: %v", err)
				}
			}
		})
	}
}

func BenchmarkDeserializeBytecode(b *testing.B) {
	sizes := []int{10, 100, 1000, 10000}
	for _, size := range sizes {
		b.Run(fmt.Sprintf("Size-%d", size), func(b *testing.B) {
			instructions := generateLargeInstructionSet(size)
			constants := generateConstants(size)
			buf := &bytes.Buffer{}
			err := SerializeBytecode(buf, instructions, constants)
			if err != nil {
				b.Fatalf("Serialization failed: %v", err)
			}
			serialized := buf.Bytes()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				reader := bytes.NewReader(serialized)
				_, _, err := DeserializeBytecode(reader)
				if err != nil {
					b.Fatalf("Deserialization failed: %v", err)
				}
			}
		})
	}
}

func BenchmarkSerializeInstruction(b *testing.B) {
	instruction := PackInstruction(OpPrintConst, 1, 2, 3)
	buf := make([]byte, 4)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		instruction.Serialize(buf)
	}
}

func BenchmarkDeserializeInstruction(b *testing.B) {
	instruction := PackInstruction(OpPrintConst, 1, 2, 3)
	buf := make([]byte, 4)
	instruction.Serialize(buf)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		unpacked := UnpackedInstruction{}
		unpacked.Unpack(Instruction(binary.LittleEndian.Uint32(buf)))
		if unpacked.Op != OpPrintConst || unpacked.A != 1 || unpacked.B != 2 || unpacked.C != 3 {
			b.Fatalf("Deserialization failed: unexpected result")
		}
	}
}

func generateLargeInstructionSet(size int) []Instruction {
	instructions := make([]Instruction, size)
	for i := 0; i < size; i++ {
		opcode := OpCode(i % 8)
		instructions[i] = PackInstruction(opcode, uint8(i), uint8(i+1), uint8(i+2))
	}
	return instructions
}

func generateConstants(size int) []Constant {
	constants := make([]Constant, size)
	for i := 0; i < size; i++ {
		switch i % 4 {
		case 0:
			constants[i] = Constant{Type: ConstString, Value: fmt.Sprintf("String%d", i)}
		case 1:
			constants[i] = Constant{Type: ConstInteger, Value: int64(i)}
		case 2:
			constants[i] = Constant{Type: ConstFloat, Value: float64(i)}
		case 3:
			constants[i] = Constant{Type: ConstBoolean, Value: i%2 == 0}
		}
	}
	return constants
}
