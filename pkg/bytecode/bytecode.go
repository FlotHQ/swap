package bytecode

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"sync"
)

const (
	MagicNumber uint32 = 0x53574150
	Version     uint32 = 1
)

type Header struct {
	Magic            uint32
	Version          uint32
	ConstantCount    uint32
	InstructionCount uint32
}

func (i Instruction) Serialize(buf []byte) {
	binary.LittleEndian.PutUint32(buf, uint32(i))
}

func SerializeInstruction(instruction Instruction) [4]byte {
	var buf [4]byte
	binary.LittleEndian.PutUint32(buf[:], uint32(instruction))
	return buf
}

var bufferPool = sync.Pool{
	New: func() interface{} {
		return make([]byte, 4096)
	},
}

func SerializeBytecode(w io.Writer, instructions []Instruction, constants []Constant) error {
	buf := bufferPool.Get().([]byte)
	defer bufferPool.Put(buf)

	header := Header{
		Magic:            MagicNumber,
		Version:          Version,
		ConstantCount:    uint32(len(constants)),
		InstructionCount: uint32(len(instructions)),
	}

	binary.LittleEndian.PutUint32(buf[0:], header.Magic)
	binary.LittleEndian.PutUint32(buf[4:], header.Version)
	binary.LittleEndian.PutUint32(buf[8:], header.ConstantCount)
	binary.LittleEndian.PutUint32(buf[12:], header.InstructionCount)

	if _, err := w.Write(buf[:16]); err != nil {
		return err
	}

	for _, constant := range constants {
		if err := writeConstant(w, constant, buf); err != nil {
			return err
		}
	}

	for i := 0; i < len(instructions); i += 1024 {
		end := i + 1024
		if end > len(instructions) {
			end = len(instructions)
		}
		for j, instr := range instructions[i:end] {
			binary.LittleEndian.PutUint32(buf[j*4:], uint32(instr))
		}
		if _, err := w.Write(buf[:(end-i)*4]); err != nil {
			return err
		}
	}

	return nil
}

func writeConstant(w io.Writer, constant Constant, buf []byte) error {
	buf[0] = uint8(constant.Type)

	switch constant.Type {
	case ConstString:
		str := constant.Value.(string)
		binary.LittleEndian.PutUint32(buf[1:], uint32(len(str)))
		if _, err := w.Write(buf[:5]); err != nil {
			return err
		}
		_, err := w.Write([]byte(str))
		return err
	case ConstInteger:
		binary.LittleEndian.PutUint64(buf[1:], uint64(constant.Value.(int64)))
		_, err := w.Write(buf[:9])
		return err
	case ConstFloat:
		binary.LittleEndian.PutUint64(buf[1:], math.Float64bits(constant.Value.(float64)))
		_, err := w.Write(buf[:9])
		return err
	case ConstBoolean:
		if constant.Value.(bool) {
			buf[1] = 1
		} else {
			buf[1] = 0
		}
		_, err := w.Write(buf[:2])
		return err
	default:
		return fmt.Errorf("unknown constant type: %v", constant.Type)
	}
}

func DeserializeBytecode(r io.Reader) ([]Instruction, []Constant, error) {
	buf := bufferPool.Get().([]byte)
	defer bufferPool.Put(buf)

	if _, err := io.ReadFull(r, buf[:16]); err != nil {
		return nil, nil, fmt.Errorf("failed to read header: %w", err)
	}

	header := Header{
		Magic:            binary.LittleEndian.Uint32(buf[0:]),
		Version:          binary.LittleEndian.Uint32(buf[4:]),
		ConstantCount:    binary.LittleEndian.Uint32(buf[8:]),
		InstructionCount: binary.LittleEndian.Uint32(buf[12:]),
	}

	if header.Magic != MagicNumber {
		return nil, nil, fmt.Errorf("invalid magic number")
	}
	if header.Version != Version {
		return nil, nil, fmt.Errorf("unsupported version: %d", header.Version)
	}

	constants := make([]Constant, header.ConstantCount)
	if err := readConstants(r, constants, buf); err != nil {
		return nil, nil, err
	}

	instructions := make([]Instruction, header.InstructionCount)
	instructionBuf := buf[:4]
	for i := uint32(0); i < header.InstructionCount; i++ {
		if _, err := io.ReadFull(r, instructionBuf); err != nil {
			return nil, nil, fmt.Errorf("failed to read instruction: %w", err)
		}
		instructions[i] = Instruction(binary.LittleEndian.Uint32(instructionBuf))
	}

	return instructions, constants, nil
}

func readConstants(r io.Reader, constants []Constant, buf []byte) error {
	for i := range constants {
		if _, err := io.ReadFull(r, buf[:1]); err != nil {
			return fmt.Errorf("failed to read constant type: %w", err)
		}
		constType := ConstantType(buf[0])

		switch constType {
		case ConstString:
			if _, err := io.ReadFull(r, buf[1:5]); err != nil {
				return fmt.Errorf("failed to read string length: %w", err)
			}
			strLen := binary.LittleEndian.Uint32(buf[1:5])
			if uint32(cap(buf)) < strLen {
				buf = make([]byte, strLen)
			} else {
				buf = buf[:strLen]
			}
			if _, err := io.ReadFull(r, buf); err != nil {
				return fmt.Errorf("failed to read string: %w", err)
			}
			constants[i] = Constant{Type: constType, Value: string(buf)}
		case ConstInteger:
			if _, err := io.ReadFull(r, buf[1:9]); err != nil {
				return fmt.Errorf("failed to read integer: %w", err)
			}
			constants[i] = Constant{Type: constType, Value: int64(binary.LittleEndian.Uint64(buf[1:9]))}
		case ConstFloat:
			if _, err := io.ReadFull(r, buf[1:9]); err != nil {
				return fmt.Errorf("failed to read float: %w", err)
			}
			constants[i] = Constant{Type: constType, Value: math.Float64frombits(binary.LittleEndian.Uint64(buf[1:9]))}
		case ConstBoolean:
			if _, err := io.ReadFull(r, buf[1:2]); err != nil {
				return fmt.Errorf("failed to read boolean: %w", err)
			}
			constants[i] = Constant{Type: constType, Value: buf[1] != 0}
		default:
			return fmt.Errorf("unknown constant type: %v", constType)
		}
	}
	return nil
}
