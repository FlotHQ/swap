# Swap Template Engine

> [!CAUTION]
> This project is experimental and not recommended for production use. It prioritizes performance over feature completeness or safety.


Swap is an experimental minimal Go template engine optimized for rapid variable substitution and execution, with a focus on performance and simplicity. 

1. Bytecode compilation for faster rendering
2. Optimized variable substitution
3. Zero dependencies
4. Low memory footprint
5. Simple API

Ideal for high-throughput applications like web servers or report generators, Swap offers a lightweight, high-performance solution for Go template rendering.

## Features
- Fast template rendering using a bytecode VM
- Compilation of templates into bytecode
- Optional in-memory caching of compiled templates for improved performance
- Support for basic control structures (e.g., loops)
- Limited set of built-in functions

## Benchmarks
The project includes benchmarks for:
- Basic execution
- Execution with caching
- Pre-compiled template execution

Template used for benchmarks:
```
Hello, {{ .name }}!
```

Benchmark results:

| Benchmark | Iterations | Time (ns/op) | Bytes/op | Allocs/op |
|-----------|------------|--------------|----------|-----------|
| BenchmarkRun-24 | 47,031,710 | 25.65 | 0 | 0 |
| BenchmarkExecuteWithCache-24 | 16,992,112 | 70.06 | 48 | 1 |
| BenchmarkExecute-24 | 2,100,693 | 568.8 | 442 | 19 |

*goos: windows, goarch: amd64, cpu: 13th Gen Intel(R) Core(TM) i7-13700K*

## Examples

Here are three examples demonstrating how to use the Swap template engine:

### 1. Basic Template Execution

```go
package main

import (
	"fmt"
	"github.com/flothq/swap"
)

func main() {
	template := "Hello, {{ .name }}!"
	context := map[string]interface{}{"name": "John Doe"}

	engine := swap.NewEngine()
	result, err := engine.Execute(template, context)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println(string(result)) // Output: Hello, John Doe!
}
```

### 2. Template Compilation and Execution

```go
package main

import (
	"fmt"
	"github.com/flothq/swap"
)

func main() {
	template := "Hello, {{ .name }}!"
	context := map[string]interface{}{"name": "John Doe"}

	engine := swap.NewEngine()
	program, _ := engine.Compile(template)


    // save program to file or anywhere
    bin, _ := program.Serialize()
    os.WriteFile("program.bin", program, 0o644)

	result, err := engine.Run(program, context)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println(string(result)) // Output: Hello, John Doe!
}

```	

### 3. Loops and Built-in Functions

```go
package main

import (
	"fmt"
	"github.com/flothq/swap"
)

func main() {
	template := "Users: {{ range .users }}{{ .name }} is a {{ upper(.occupation) }}\n{{ end }}"

	context := map[string]interface{}{"users": []map[string]interface{}{
		{"name": "Alice", "occupation": "Engineer"},
		{"name": "Bob", "occupation": "Manager"},
		{"name": "Charlie", "occupation": "Designer"},
		{"name": "David", "occupation": "Developer"},
		{"name": "Eve", "occupation": "Analyst"},
	}}

	engine := swap.NewEngine()
	result, err := engine.Execute(template, context)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println(string(result))  
}
```
