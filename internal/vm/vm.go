package vm

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
	"unsafe"

	"github.com/flothq/swap/pkg/bytecode"
)

type loopInfo struct {
	items   unsafe.Pointer
	itemLen int
	itemCap int
	index   int
	startPC int
}

type Program struct {
	Instructions []bytecode.Instruction
	Constants    []bytecode.Constant
}

func NewProgram(instructions []bytecode.Instruction, constants []bytecode.Constant) *Program {
	return &Program{Instructions: instructions, Constants: constants}
}

func (p *Program) Serialize() ([]byte, error) {
	var buf bytes.Buffer
	err := bytecode.SerializeBytecode(&buf, p.Instructions, p.Constants)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize program: %w", err)
	}
	return buf.Bytes(), nil
}

type VM struct {
	instructions []bytecode.Instruction
	registers    []unsafe.Pointer
	context      map[string]interface{}
	constants    []bytecode.Constant
	buffer       []byte
	loopStack    []loopInfo
	pc           int
	unpacked     bytecode.UnpackedInstruction
}

var vmPool = sync.Pool{
	New: func() interface{} {
		return &VM{
			loopStack: make([]loopInfo, 0, 4),
			unpacked:  bytecode.UnpackedInstruction{},
			registers: make([]unsafe.Pointer, 8),
		}
	},
}

func NewVM(instructions []bytecode.Instruction, context map[string]interface{}, constants []bytecode.Constant) *VM {
	vm := vmPool.Get().(*VM)
	vm.instructions = instructions
	vm.context = context
	if cap(vm.buffer) < 64 {
		vm.buffer = make([]byte, 0, 64)
	} else {
		vm.buffer = vm.buffer[:0]
	}
	vm.loopStack = vm.loopStack[:0]
	vm.constants = constants
	vm.unpacked.Reset()
	vm.pc = 0
	for i := range vm.registers {
		vm.registers[i] = nil
	}

	return vm
}

func (vm *VM) Release() {
	vm.instructions = nil
	vm.context = nil
	vm.buffer = vm.buffer[:0]
	vm.loopStack = vm.loopStack[:0]
	vm.registers = vm.registers[:0]
	vm.constants = vm.constants[:0]
	vm.pc = 0
	vm.unpacked.Reset()
	vm.constants = vm.constants[:0]
	vmPool.Put(vm)
}

func (vm *VM) handleLoopStart(a, b, c uint8) {
	key := vm.constants[a].Value.(string)
	res := vm.resolveVar(key)

	if res == nil {
		panic(fmt.Sprintf("loop key not found: %s", key))
	}

	var info loopInfo
	info.startPC = vm.pc + 1
	info.index = 0

	switch v := res.(type) {
	case []interface{}:
		info.items = unsafe.Pointer(&v)
		info.itemLen = len(v)
		info.itemCap = cap(v)
	case []string:
		info.items = unsafe.Pointer(&v)
		info.itemLen = len(v)
		info.itemCap = cap(v)
	default:
		panic(fmt.Sprintf("unsupported type: %T", v))
	}

	if len(vm.loopStack) < cap(vm.loopStack) {
		vm.loopStack = vm.loopStack[:len(vm.loopStack)+1]
	} else {
		newCap := cap(vm.loopStack) * 2
		if newCap == 0 {
			newCap = 4
		}
		newStack := make([]loopInfo, len(vm.loopStack)+1, newCap)
		copy(newStack, vm.loopStack)
		vm.loopStack = newStack
	}
	vm.loopStack[len(vm.loopStack)-1] = info
}

func (vm *VM) handleLoopEnd() {
	if len(vm.loopStack) > 0 {
		last := &vm.loopStack[len(vm.loopStack)-1]
		last.index++
		if last.index < last.itemLen {
			vm.pc = last.startPC - 1
			return
		} else {
			vm.loopStack = vm.loopStack[:len(vm.loopStack)-1]
		}
	}
}

func (vm *VM) Run() ([]byte, error) {
	for vm.pc < len(vm.instructions) {
		instruction := vm.instructions[vm.pc]
		vm.unpacked.Unpack(instruction)

		switch vm.unpacked.Op {
		case bytecode.OpPrintConst:
			vm.appendConstantToBuffer(vm.unpacked.A)
		case bytecode.OpResolvePrint:
			vm.resolveAndWriteVar(vm.getConstantString(vm.unpacked.A))
		case bytecode.OpLoadConst:
			vm.loadConstantToRegister(vm.unpacked.A, vm.unpacked.B)
		case bytecode.OpResolveLoad:
			vm.resolveAndLoadToRegister(vm.unpacked.A, vm.unpacked.B)
		case bytecode.OpLoopStart:
			vm.handleLoopStart(vm.unpacked.A, vm.unpacked.B, vm.unpacked.C)
		case bytecode.OpLoopEnd:
			vm.handleLoopEnd()
		case bytecode.OpCall:
			vm.handleFunctionCall(vm.unpacked.A)
		case bytecode.OpHalt:
			return vm.buffer, nil
		default:
			return nil, fmt.Errorf("unknown opcode: %s", vm.unpacked.Op)
		}

		vm.pc++
	}

	return nil, fmt.Errorf("halt instruction not found")
}

func (vm *VM) appendConstantToBuffer(index uint8) {
	vm.buffer = append(vm.buffer, vm.constants[index].Value.(string)...)
}

func (vm *VM) getConstantString(index uint8) string {
	return vm.constants[index].Value.(string)
}

func (vm *VM) loadConstantToRegister(registerIndex, constantIndex uint8) {
	vm.registers[registerIndex] = unsafe.Pointer(&vm.constants[constantIndex].Value)
}

func (vm *VM) resolveAndLoadToRegister(registerIndex, keyIndex uint8) {
	key := vm.getConstantString(keyIndex)
	value := vm.resolveVar(key[1:])
	vm.registers[registerIndex] = unsafe.Pointer(&value)
}

func (vm *VM) handleFunctionCall(fnKeyIndex uint8) {
	fnKey := vm.getConstantString(fnKeyIndex)
	result := vm.callFunction(fnKey)
	vm.buffer = append(vm.buffer, result...)
}

func (vm *VM) resolveVar(path string) interface{} {
	if path == "." {
		if len(vm.loopStack) > 0 {
			last := &vm.loopStack[len(vm.loopStack)-1]
			slice := *(*[]interface{})(last.items)
			return slice[last.index]
		}
		return vm.context
	}
	if path[0] == '.' {
		key := path[1:]
		if len(vm.loopStack) > 0 {
			last := &vm.loopStack[len(vm.loopStack)-1]
			slice := *(*[]interface{})(last.items)
			item := slice[last.index]
			if mapItem, ok := item.(map[string]interface{}); ok {
				return mapItem[key]
			}
		}
		return vm.context[key]
	}
	return vm.context[path]
}

func (vm *VM) resolveAndWriteVar(path string) {
	value := vm.resolveVar(path)
	switch v := value.(type) {
	case string:
		vm.buffer = append(vm.buffer, v...)
	case int:
		vm.buffer = strconv.AppendInt(vm.buffer, int64(v), 10)
	case float64:
		vm.buffer = strconv.AppendFloat(vm.buffer, v, 'f', -1, 64)
	default:
		vm.buffer = append(vm.buffer, fmt.Sprintf("%v", v)...)
	}
}

func (vm *VM) callFunction(fnKey string) []byte {

	switch fnKey {
	case "upper":
		arg1 := *(*interface{})(vm.registers[0])
		return []byte(strings.ToUpper(arg1.(string)))
	case "lower":
		arg1 := *(*interface{})(vm.registers[0])
		return []byte(strings.ToLower(arg1.(string)))
	case "formatDate":
		arg1 := *(*interface{})(vm.registers[0])
		arg2 := *(*interface{})(vm.registers[1])

		date, err := time.Parse(time.RFC3339, arg1.(string))
		if err != nil {
			panic(err)
		}
		return []byte(date.Format(arg2.(string)))
	default:
		panic(fmt.Sprintf("unknown function: %s", fnKey))
	}
}
