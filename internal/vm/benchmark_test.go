package vm

import (
	"bytes"
	"fmt"
	"math/rand"
	"regexp"
	"testing"
	"text/template"

	"github.com/flothq/swap/pkg/bytecode"
)

func BenchmarkRegexReplace(b *testing.B) {
	template := "Hello, {{.name}}"

	context := map[string]interface{}{"name": "World"}

	re := regexp.MustCompile(`\{\{\.(\w+)\}\}`)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		result := re.ReplaceAllStringFunc(template, func(match string) string {
			key := re.FindStringSubmatch(match)[1]
			if value, ok := context[key]; ok {
				return fmt.Sprint(value)
			}
			return match
		})
		_ = result
	}
}

func BenchmarkVMReplace(b *testing.B) {
	instructions := []bytecode.Instruction{
		bytecode.PackInstruction(bytecode.OpPrintConst, 0, 0, 0),
		bytecode.PackInstruction(bytecode.OpResolvePrint, 1, 0, 0),
		bytecode.PackInstruction(bytecode.OpHalt, 0, 0, 0),
	}
	context := map[string]interface{}{"name": "World"}
	constants := []bytecode.Constant{
		{bytecode.ConstString, "Hello, "},
		{bytecode.ConstString, ".name"},
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		vm := NewVM(instructions, context, constants)
		vm.Run()
		vm.Release()
	}
}

func BenchmarkVM_Basic(b *testing.B) {
	instructions := []bytecode.Instruction{
		bytecode.PackInstruction(bytecode.OpPrintConst, 0, 0, 0),
		bytecode.PackInstruction(bytecode.OpResolvePrint, 1, 0, 0),
		bytecode.PackInstruction(bytecode.OpHalt, 0, 0, 0),
	}
	context := map[string]interface{}{"name": "World"}
	constants := []bytecode.Constant{
		{bytecode.ConstString, "Hello, "},
		{bytecode.ConstString, ".name"},
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		vm := NewVM(instructions, context, constants)
		vm.Run()
		vm.Release()
	}
}

func BenchmarkVM_Medium(b *testing.B) {
	instructions := []bytecode.Instruction{
		bytecode.PackInstruction(bytecode.OpLoopStart, 0, 0, 0),
		bytecode.PackInstruction(bytecode.OpPrintConst, 1, 0, 0),
		bytecode.PackInstruction(bytecode.OpResolvePrint, 2, 0, 0),
		bytecode.PackInstruction(bytecode.OpLoopEnd, 0, 0, 0),
		bytecode.PackInstruction(bytecode.OpHalt, 0, 0, 0),
	}
	context := map[string]interface{}{"items": []interface{}{1, 2, 3, 4, 5}}
	constants := []bytecode.Constant{
		{bytecode.ConstString, ".items"},
		{bytecode.ConstString, "Item "},
		{bytecode.ConstString, "."},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		vm := NewVM(instructions, context, constants)
		vm.Run()
		vm.Release()
	}
}

func BenchmarkVM_Heavy(b *testing.B) {
	instructions := []bytecode.Instruction{
		bytecode.PackInstruction(bytecode.OpLoopStart, 0, 0, 0),
		bytecode.PackInstruction(bytecode.OpPrintConst, 1, 0, 0),
		bytecode.PackInstruction(bytecode.OpResolvePrint, 2, 0, 0),
		bytecode.PackInstruction(bytecode.OpPrintConst, 3, 0, 0),
		bytecode.PackInstruction(bytecode.OpResolvePrint, 4, 0, 0),
		bytecode.PackInstruction(bytecode.OpLoopEnd, 0, 0, 0),
		bytecode.PackInstruction(bytecode.OpHalt, 0, 0, 0),
	}

	context := map[string]interface{}{
		"users": []interface{}{
			map[string]interface{}{"name": "Alice", "age": 30},
			map[string]interface{}{"name": "Bob", "age": 25},
			map[string]interface{}{"name": "Charlie", "age": 35},
			map[string]interface{}{"name": "David", "age": 28},
			map[string]interface{}{"name": "Eve", "age": 32},
		},
	}
	constants := []bytecode.Constant{
		{bytecode.ConstString, ".users"},
		{bytecode.ConstString, "User "},
		{bytecode.ConstString, ".name"},
		{bytecode.ConstString, ": "},
		{bytecode.ConstString, ".age"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		vm := NewVM(instructions, context, constants)
		vm.Run()
		vm.Release()
	}
}

func BenchmarkNativeTemplate_Basic(b *testing.B) {
	tmpl, err := template.New("basic").Parse("Hello, {{.name}}")
	if err != nil {
		b.Fatal(err)
	}
	data := map[string]interface{}{"name": "World"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		err := tmpl.Execute(&buf, data)
		if err != nil {
			b.Fatal(err)
		}
		_ = buf.String()
	}
}

func BenchmarkVM_JSON(b *testing.B) {
	instructions := []bytecode.Instruction{
		bytecode.PackInstruction(bytecode.OpPrintConst, 0, 0, 0),
		bytecode.PackInstruction(bytecode.OpResolvePrint, 1, 0, 0),
		bytecode.PackInstruction(bytecode.OpPrintConst, 2, 0, 0),
		bytecode.PackInstruction(bytecode.OpResolvePrint, 3, 0, 0),
		bytecode.PackInstruction(bytecode.OpPrintConst, 4, 0, 0),
		bytecode.PackInstruction(bytecode.OpLoopStart, 5, 0, 0),
		bytecode.PackInstruction(bytecode.OpPrintConst, 6, 0, 0),
		bytecode.PackInstruction(bytecode.OpResolvePrint, 7, 0, 0),
		bytecode.PackInstruction(bytecode.OpPrintConst, 8, 0, 0),
		bytecode.PackInstruction(bytecode.OpResolvePrint, 9, 0, 0),
		bytecode.PackInstruction(bytecode.OpPrintConst, 10, 0, 0),
		bytecode.PackInstruction(bytecode.OpResolvePrint, 11, 0, 0),
		bytecode.PackInstruction(bytecode.OpPrintConst, 12, 0, 0),
		bytecode.PackInstruction(bytecode.OpLoopEnd, 0, 0, 0),
		bytecode.PackInstruction(bytecode.OpPrintConst, 13, 0, 0),
		bytecode.PackInstruction(bytecode.OpResolvePrint, 14, 0, 0),
		bytecode.PackInstruction(bytecode.OpPrintConst, 15, 0, 0),
		bytecode.PackInstruction(bytecode.OpHalt, 0, 0, 0),
	}

	constants := []bytecode.Constant{
		{bytecode.ConstString, ".customer.name"},
		{bytecode.ConstString, ", \"customerAddress\": "},
		{bytecode.ConstString, "{\"customerName\": "},
		{bytecode.ConstString, ".customer.address"},
		{bytecode.ConstString, ", \"items\": ["},
		{bytecode.ConstString, ".items"},
		{bytecode.ConstString, "{\"name\": "},
		{bytecode.ConstString, ".name"},
		{bytecode.ConstString, ", \"quantity\": "},
		{bytecode.ConstString, ".quantity"},
		{bytecode.ConstString, ", \"price\": "},
		{bytecode.ConstString, ".price"},
		{bytecode.ConstString, "}"},
		{bytecode.ConstString, "], \"total\": "},
		{bytecode.ConstString, "total"},
		{bytecode.ConstString, "}"},
	}

	context := generateLongInvoiceContext()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		vm := NewVM(instructions, context, constants)
		vm.Run()
		vm.Release()
	}
}

func BenchmarkNativeTemplate_Medium(b *testing.B) {
	tmpl, err := template.New("medium").Parse("{{range $index, $element := .items}}Item {{$index}}{{end}}")
	if err != nil {
		b.Fatal(err)
	}
	data := map[string]interface{}{"items": []interface{}{1, 2, 3, 4, 5}}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		err := tmpl.Execute(&buf, data)
		if err != nil {
			b.Fatal(err)
		}
		_ = buf.String()
	}
}

func BenchmarkNativeTemplate_Heavy(b *testing.B) {
	tmpl, err := template.New("heavy").Parse(`
		{{range $index, $user := .users}}
			User {{$index}}: {{$user.name}} ({{$user.age}})
		{{end}}
	`)
	if err != nil {
		b.Fatal(err)
	}
	data := map[string]interface{}{
		"users": []interface{}{
			map[string]interface{}{"name": "Alice", "age": 30},
			map[string]interface{}{"name": "Bob", "age": 25},
			map[string]interface{}{"name": "Charlie", "age": 35},
			map[string]interface{}{"name": "David", "age": 28},
			map[string]interface{}{"name": "Eve", "age": 32},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		err := tmpl.Execute(&buf, data)
		if err != nil {
			b.Fatal(err)
		}
		_ = buf.String()
	}
}

func BenchmarkVM_LongInvoice(b *testing.B) {
	instructions := []bytecode.Instruction{
		bytecode.PackInstruction(bytecode.OpPrintConst, 0, 0, 0),
		bytecode.PackInstruction(bytecode.OpPrintConst, 1, 0, 0),
		bytecode.PackInstruction(bytecode.OpResolvePrint, 2, 0, 0),
		bytecode.PackInstruction(bytecode.OpPrintConst, 3, 0, 0),
		bytecode.PackInstruction(bytecode.OpResolvePrint, 4, 0, 0),
		bytecode.PackInstruction(bytecode.OpPrintConst, 5, 0, 0),
		bytecode.PackInstruction(bytecode.OpLoopStart, 6, 0, 0),
		bytecode.PackInstruction(bytecode.OpPrintConst, 7, 0, 0),
		bytecode.PackInstruction(bytecode.OpResolvePrint, 8, 0, 0),
		bytecode.PackInstruction(bytecode.OpPrintConst, 9, 0, 0),
		bytecode.PackInstruction(bytecode.OpResolvePrint, 10, 0, 0),
		bytecode.PackInstruction(bytecode.OpPrintConst, 11, 0, 0),
		bytecode.PackInstruction(bytecode.OpResolvePrint, 12, 0, 0),
		bytecode.PackInstruction(bytecode.OpPrintConst, 13, 0, 0),
		bytecode.PackInstruction(bytecode.OpLoopEnd, 0, 0, 0),
		bytecode.PackInstruction(bytecode.OpPrintConst, 14, 0, 0),
		bytecode.PackInstruction(bytecode.OpResolvePrint, 15, 0, 0),
		bytecode.PackInstruction(bytecode.OpHalt, 0, 0, 0),
	}

	constants := []bytecode.Constant{
		{bytecode.ConstString, "Invoice\n\n"},
		{bytecode.ConstString, "Bill To: "},
		{bytecode.ConstString, ".customer.name"},
		{bytecode.ConstString, "\n"},
		{bytecode.ConstString, ".customer.address"},
		{bytecode.ConstString, "\n\nItems:\n"},
		{bytecode.ConstString, "items"},
		{bytecode.ConstString, "- "},
		{bytecode.ConstString, ".name"},
		{bytecode.ConstString, " (Quantity: "},
		{bytecode.ConstString, ".quantity"},
		{bytecode.ConstString, ", Price: $"},
		{bytecode.ConstString, ".price"},
		{bytecode.ConstString, ")\n"},
		{bytecode.ConstString, "\nTotal: $"},
		{bytecode.ConstString, "total"},
	}

	context := generateLongInvoiceContext()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		vm := NewVM(instructions, context, constants)
		vm.Run()
		vm.Release()
	}
}

func BenchmarkNativeTemplate_LongInvoice(b *testing.B) {
	tmpl, err := template.New("invoice").Parse(`
Invoice

Bill To: {{.customer.name}}
{{.customer.address}}

Items:
{{range .items}}- {{.name}} (Quantity: {{.quantity}}, Price: ${{.price}})
{{end}}

Total: ${{.total}}
`)
	if err != nil {
		b.Fatal(err)
	}

	data := generateLongInvoiceContext()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		err := tmpl.Execute(&buf, data)
		if err != nil {
			b.Fatal(err)
		}
		_ = buf.String()
	}
}

func generateLongInvoiceContext() map[string]interface{} {
	items := make([]interface{}, 500)
	for i := 0; i < 500; i++ {
		items[i] = map[string]interface{}{
			"name":     fmt.Sprintf("Item %d", i+1),
			"quantity": rand.Intn(10) + 1,
			"price":    20.00,
		}
	}

	return map[string]interface{}{
		"customer.name":    "John Doe",
		"customer.address": "123 Main St, Anytown, AN 12345",
		"items":            items,
		"total":            1234.56,
	}
}
