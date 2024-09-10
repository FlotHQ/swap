package vm

import (
	"strings"
	"testing"

	"github.com/flothq/swap/pkg/bytecode"
)

func TestVMHandleLargeInput(t *testing.T) {
	largeSlice := make([]interface{}, 1000000)
	for i := range largeSlice {
		largeSlice[i] = strings.Repeat("a", 1000)
	}

	instructions := []bytecode.Instruction{
		bytecode.PackInstruction(bytecode.OpLoopStart, 0, 0, 0),
		bytecode.PackInstruction(bytecode.OpResolvePrint, 1, 0, 0),
		bytecode.PackInstruction(bytecode.OpLoopEnd, 0, 0, 0),
		bytecode.PackInstruction(bytecode.OpHalt, 0, 0, 0),
	}
	constants := []bytecode.Constant{
		{bytecode.ConstString, "largeSlice"},
		{bytecode.ConstString, "."},
	}
	context := map[string]interface{}{
		"largeSlice": largeSlice,
	}

	vm := NewVM(instructions, context, constants)
	result, err := vm.Run()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	expectedLength := 1000000 * 1000
	if len(result) != expectedLength {
		t.Errorf("Expected result length %d, got %d", expectedLength, len(result))
	}

	if !strings.HasPrefix(string(result), strings.Repeat("a", 1000)) {
		t.Errorf("Result doesn't start with the expected content")
	}
	if !strings.HasSuffix(string(result), strings.Repeat("a", 1000)) {
		t.Errorf("Result doesn't end with the expected content")
	}
}

func TestVM_LoopSize(t *testing.T) {
	instructions := []bytecode.Instruction{
		bytecode.PackInstruction(bytecode.OpLoopStart, 0, 0, 0),
		bytecode.PackInstruction(bytecode.OpPrintConst, 1, 0, 0),
		bytecode.PackInstruction(bytecode.OpLoopEnd, 0, 0, 0),
		bytecode.PackInstruction(bytecode.OpHalt, 0, 0, 0),
	}
	constants := []bytecode.Constant{
		{bytecode.ConstString, "items"},
		{bytecode.ConstString, "A"},
	}
	context := map[string]interface{}{
		"items": make([]interface{}, 1000000),
	}
	vm := NewVM(instructions, context, constants)
	result, err := vm.Run()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(result) != 1000000 {
		t.Errorf("Expected result length 1000000, got %d", len(result))
	}
	vm.Release()
}

func TestVMLargeInput(t *testing.T) {

	largeInput := strings.Repeat("a", 10*1024*1024)

	instructions := []bytecode.Instruction{
		bytecode.PackInstruction(bytecode.OpResolvePrint, 0, 0, 0),
		bytecode.PackInstruction(bytecode.OpHalt, 0, 0, 0),
	}
	constants := []bytecode.Constant{
		{bytecode.ConstString, "largeInput"},
	}
	context := map[string]interface{}{
		"largeInput": largeInput,
	}

	vm := NewVM(instructions, context, constants)
	result, err := vm.Run()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(result) != len(largeInput) {
		t.Errorf("Expected result length %d, got %d", len(largeInput), len(result))
	}
}

func TestVMSafety(t *testing.T) {

	instructions := []bytecode.Instruction{
		bytecode.PackInstruction(bytecode.OpLoopStart, 0, 0, 0),
		bytecode.PackInstruction(bytecode.OpResolvePrint, 1, 0, 0),
		bytecode.PackInstruction(bytecode.OpLoopEnd, 0, 0, 0),
		bytecode.PackInstruction(bytecode.OpHalt, 0, 0, 0),
	}
	constants := []bytecode.Constant{
		{bytecode.ConstString, "items"},
		{bytecode.ConstString, "."},
	}
	context := map[string]interface{}{
		"items": []interface{}{"item1", "item2"},
	}

	vm := NewVM(instructions, context, constants)
	result, err := vm.Run()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if string(result) != "item1item2" {
		t.Errorf("Expected 'item1item2', got %s", result)
	}

	context["items"] = "not a slice"
	vm = NewVM(instructions, context, constants)
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not handle invalid loop variable properly")
		}
	}()
	vm.Run()

	instructions = []bytecode.Instruction{
		bytecode.PackInstruction(bytecode.OpResolvePrint, 0, 0, 0),
		bytecode.PackInstruction(bytecode.OpHalt, 0, 0, 0),
	}
	constants = []bytecode.Constant{
		{bytecode.ConstString, "../../../some/path"},
	}
	vm = NewVM(instructions, context, constants)
	result, err = vm.Run()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if string(result) != "" {
		t.Errorf("Expected empty string for invalid path, got %s", result)
	}

	instructions = []bytecode.Instruction{
		bytecode.PackInstruction(bytecode.OpResolvePrint, 0, 0, 0),
		bytecode.PackInstruction(bytecode.OpHalt, 0, 0, 0),
	}
	constants = []bytecode.Constant{
		{bytecode.ConstString, "{{.newKey = \"hacked\"}}"},
	}
	context = map[string]interface{}{}
	vm = NewVM(instructions, context, constants)
	vm.Run()
	if _, exists := context["newKey"]; exists {
		t.Errorf("Context was manipulated through template")
	}
}

func TestVM_Run(t *testing.T) {
	tests := []struct {
		name         string
		instructions []bytecode.Instruction
		context      map[string]interface{}
		expected     string
		constants    []bytecode.Constant
	}{
		{
			name: "print constant text",
			instructions: []bytecode.Instruction{
				bytecode.PackInstruction(bytecode.OpPrintConst, 0, 0, 0),
				bytecode.PackInstruction(bytecode.OpHalt, 0, 0, 0),
			},
			constants: []bytecode.Constant{
				{bytecode.ConstString, "Hello, World!"},
			},
			context:  map[string]interface{}{},
			expected: "Hello, World!",
		},
		{
			name: "resolve variable and print",
			instructions: []bytecode.Instruction{
				bytecode.PackInstruction(bytecode.OpPrintConst, 0, 0, 0),
				bytecode.PackInstruction(bytecode.OpResolvePrint, 1, 0, 0),
				bytecode.PackInstruction(bytecode.OpHalt, 0, 0, 0),
			},
			context: map[string]interface{}{"name": "World!"},
			constants: []bytecode.Constant{
				{bytecode.ConstString, "Hello, "},
				{bytecode.ConstString, ".name"},
			},
			expected: "Hello, World!",
		},

		{
			name: "simple loop",
			instructions: []bytecode.Instruction{
				bytecode.PackInstruction(bytecode.OpLoopStart, 0, 0, 0),
				bytecode.PackInstruction(bytecode.OpResolvePrint, 1, 0, 0),
				bytecode.PackInstruction(bytecode.OpLoopEnd, 0, 0, 0),
				bytecode.PackInstruction(bytecode.OpHalt, 0, 0, 0),
			},
			context: map[string]interface{}{"items": []interface{}{"0", "1", "2"}},
			constants: []bytecode.Constant{
				{bytecode.ConstString, ".items"},
				{bytecode.ConstString, "."},
			},
			expected: "012",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm := NewVM(tt.instructions, tt.context, tt.constants)
			result, err := vm.Run()
			if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
			if string(result) != tt.expected {
				t.Errorf("Expected %q, but got %q", tt.expected, result)
			}
		})
	}
}
