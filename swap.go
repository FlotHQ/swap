package swap

import (
	"bytes"
	"fmt"
	"io"
	"sync"
	"unsafe"

	"github.com/flothq/swap/internal/compiler"
	"github.com/flothq/swap/internal/lexer"
	"github.com/flothq/swap/internal/lru"
	"github.com/flothq/swap/internal/vm"
	"github.com/flothq/swap/pkg/bytecode"
)

type Engine struct {
	cache      *lru.Cache[string, *vm.Program]
	engineOpts EngineOpts
}

type EngineOpts struct {
	CacheSize    int
	CacheEnabled bool
}

type EngineOption func(*EngineOpts)

func WithCacheEnabled(enabled bool) EngineOption {
	return func(opts *EngineOpts) {
		opts.CacheEnabled = enabled
	}
}

func WithCacheSize(size int) EngineOption {
	return func(opts *EngineOpts) {
		opts.CacheSize = size
	}
}

func NewEngine(opts ...EngineOption) *Engine {
	e := &Engine{
		engineOpts: EngineOpts{},
	}
	for _, opt := range opts {
		opt(&e.engineOpts)
	}
	if e.engineOpts.CacheEnabled {
		if e.engineOpts.CacheSize <= 0 {
			e.engineOpts.CacheSize = 1024 * 1024
		}
		e.cache = lru.New[string, *vm.Program](e.engineOpts.CacheSize, func(k string, v *vm.Program) int {
			return len(k) + int(unsafe.Sizeof(*v))
		})
	}
	return e
}

func (e *Engine) Execute(template string, context map[string]interface{}) (string, error) {
	program, err := e.Compile(template)
	if err != nil {
		return "", fmt.Errorf("compilation error: %w", err)
	}

	if err != nil {
		return "", fmt.Errorf("deserialization error: %w", err)
	}

	result, err := e.Run(program, context)
	if err != nil {
		return "", fmt.Errorf("execution error: %w", err)
	}

	programPool.Put(program)
	return string(result), nil
}

func (e *Engine) Compile(template string) (*vm.Program, error) {
	if e.cache != nil {
		if cached, ok := e.cache.Get(template); ok {
			return cached, nil
		}
	}

	lex := lexer.NewLexer(template)
	defer lex.Release()
	tokens := lex.Lex()

	comp := compiler.NewCompiler(tokens)
	defer comp.Release()
	instructions, constants, err := comp.Compile(tokens)
	if err != nil {
		return nil, fmt.Errorf("compilation error: %w", err)
	}

	var buf bytes.Buffer
	err = bytecode.SerializeBytecode(&buf, instructions, constants)
	if err != nil {
		return nil, fmt.Errorf("serialization error: %w", err)
	}

	program, err := e.deserializeBytecode(&buf)
	if err != nil {
		return nil, fmt.Errorf("deserialization error: %w", err)
	}

	if e.cache != nil {
		e.cache.Set(template, program)
	}

	return program, nil
}

var programPool = sync.Pool{
	New: func() interface{} {
		return &vm.Program{}
	},
}

func (e *Engine) deserializeBytecode(r io.Reader) (*vm.Program, error) {
	instructions, constants, err := bytecode.DeserializeBytecode(r)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize bytecode: %w", err)
	}

	program := programPool.Get().(*vm.Program)
	program.Instructions = instructions
	program.Constants = constants

	return program, nil
}

func (e *Engine) Run(program *vm.Program, context map[string]interface{}) ([]byte, error) {
	vm := vm.NewVM(program.Instructions, context, program.Constants)
	defer vm.Release()

	result, err := vm.Run()
	if err != nil {
		return nil, fmt.Errorf("VM execution failed: %w", err)
	}

	return result, nil
}
