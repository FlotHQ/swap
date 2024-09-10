package bytecode

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func TestSerializeDeserializeBytecode(t *testing.T) {
	testCases := []struct {
		name         string
		instructions []Instruction
		constants    []Constant
	}{
		{
			name: "Simple instructions and constants",
			instructions: []Instruction{
				PackInstruction(OpPrintConst, 0, 0, 0),
				PackInstruction(OpResolvePrint, 1, 0, 0),
				PackInstruction(OpHalt, 0, 0, 0),
			},
			constants: []Constant{
				{Type: ConstString, Value: "Hello"},
				{Type: ConstInteger, Value: int64(42)},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := SerializeBytecode(&buf, tc.instructions, tc.constants)
			if err != nil {
				t.Fatalf("Serialization failed: %v", err)
			}

			data := buf.Bytes()
			var header Header
			err = binary.Read(bytes.NewReader(data), binary.LittleEndian, &header)
			if err != nil {
				t.Fatalf("Failed to read header: %v", err)
			}
			if header.Magic != MagicNumber {
				t.Errorf("Invalid magic number. Expected: %d, Got: %d", MagicNumber, header.Magic)
			}
			if header.Version != Version {
				t.Errorf("Invalid version. Expected: %d, Got: %d", Version, header.Version)
			}
			if header.ConstantCount != uint32(len(tc.constants)) {
				t.Errorf("Invalid constant count. Expected: %d, Got: %d", len(tc.constants), header.ConstantCount)
			}
			if header.InstructionCount != uint32(len(tc.instructions)) {
				t.Errorf("Invalid instruction count. Expected: %d, Got: %d", len(tc.instructions), header.InstructionCount)
			}

			deserializedInstructions, deserializedConstants, err := DeserializeBytecode(bytes.NewReader(data))
			if err != nil {
				t.Fatalf("Deserialization failed: %v", err)
			}

			if len(tc.instructions) != len(deserializedInstructions) {
				t.Fatalf("Instruction count mismatch. Expected: %d, Got: %d", len(tc.instructions), len(deserializedInstructions))
			}

			if len(tc.constants) != len(deserializedConstants) {
				t.Fatalf("Constant count mismatch. Expected: %d, Got: %d", len(tc.constants), len(deserializedConstants))
			}

			for i, original := range tc.instructions {
				if original != deserializedInstructions[i] {
					t.Errorf("Instruction %d mismatch. Expected: %v, Got: %v", i, original, deserializedInstructions[i])
				}
			}

			for i, original := range tc.constants {
				if original.Type != deserializedConstants[i].Type {
					t.Errorf("Constant %d type mismatch. Expected: %v, Got: %v", i, original.Type, deserializedConstants[i].Type)
				}
				switch original.Type {
				case ConstString:
					if original.Value.(string) != deserializedConstants[i].Value.(string) {
						t.Errorf("Constant %d value mismatch. Expected: %v, Got: %v", i, original.Value, deserializedConstants[i].Value)
					}
				case ConstInteger:
					if original.Value != deserializedConstants[i].Value {
						t.Errorf("Constant %d mismatch. Expected: %v, Got: %v", i, original, deserializedConstants[i])
					}
				}
			}
		})
	}
}
