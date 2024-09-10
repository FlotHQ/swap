package main

import (
	"fmt"
	"os"

	"github.com/flothq/swap/internal/compiler"
	"github.com/flothq/swap/internal/lexer"
	"github.com/flothq/swap/internal/vm"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: swap <template>")
		os.Exit(1)
	}

	template := os.Args[1]
	context := map[string]interface{}{
		"name":  "World",
		"items": []string{"apple", "banana", "cherry"},
	}

	lex := lexer.NewLexer(template)
	tokens := lex.Lex()

	comp := compiler.NewCompiler(tokens)
	instructions, constants, err := comp.Compile(tokens)
	if err != nil {
		fmt.Printf("Compilation error: %v\n", err)
		os.Exit(1)
	}

	vm := vm.NewVM(instructions, context, constants)
	result, err := vm.Run()
	if err != nil {
		fmt.Printf("Runtime error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(result))
}
