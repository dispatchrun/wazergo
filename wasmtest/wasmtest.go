// Package wasmtest provides building blocks useful to write tests for wazero
// host modules.
package wasmtest

import (
	"context"
	"log"
	"math"
	"os"
	"reflect"

	"github.com/stealthrocket/wasm-go"
	"github.com/tetratelabs/wazero"
)

type Context struct {
	context       context.Context
	runtime       wazero.Runtime
	logger        *log.Logger
	compilation   *wasm.CompilationContext
	instantiation *wasm.InstantiationContext
}

func NewContext(ctx context.Context, logger *log.Logger) *Context {
	runtime := wazero.NewRuntime(ctx)
	return &Context{
		context:       ctx,
		runtime:       runtime,
		logger:        logger,
		compilation:   wasm.NewCompilationContext(ctx, runtime),
		instantiation: wasm.NewInstantiationContext(ctx, runtime),
	}
}

func (c *Context) Close() error {
	c.instantiation.Close(c.context)
	c.compilation.Close(c.context)
	c.runtime.Close(c.context)
	return nil
}

func Load[T wasm.Module](ctx *Context, m wasm.HostModule[T], opts ...wasm.Option[T]) {
	c, err := wasm.Compile(ctx.compilation, m, wasm.Log[T](ctx.logger))
	if err != nil {
		panic(err)
	}
	if _, err := wasm.Instantiate(ctx.instantiation, c, opts...); err != nil {
		panic(err)
	}
}

type Cmd struct {
	Entrypoint string
	Params     []uint64
}

type CmdOption = wasm.Option[*Cmd]

func Entrypoint(fn string) CmdOption {
	return wasm.OptionFunc(func(cmd *Cmd) { cmd.Entrypoint = fn })
}

func Params(values ...any) CmdOption {
	params := make([]uint64, len(values))
	push(params, values)
	return wasm.OptionFunc(func(cmd *Cmd) { cmd.Params = params })
}

func push(stack []uint64, values []any) {
	for i, value := range values {
		switch v := reflect.ValueOf(value); v.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			stack[i] = uint64(v.Int())
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			stack[i] = v.Uint()
		case reflect.Float32, reflect.Float64:
			stack[i] = math.Float64bits(v.Float())
		default:
			panic("cannot construct WebAssembly parameter from Go value of type " + v.Type().String())
		}
	}
}

func Exec(ctx *Context, path string, opts ...CmdOption) ([]uint64, error) {
	cmd := &Cmd{
		Entrypoint: "_start",
	}
	wasm.Configure(cmd, opts...)
	binary, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	compiledModule, err := ctx.runtime.CompileModule(ctx.context, binary)
	if err != nil {
		panic(err)
	}
	defer compiledModule.Close(ctx.context)
	moduleInstance, err := ctx.runtime.InstantiateModule(ctx.context, compiledModule,
		wazero.NewModuleConfig().
			WithStartFunctions(),
	)
	if err != nil {
		panic(err)
	}
	defer moduleInstance.Close(ctx.context)
	callContext := wasm.NewCallContext(ctx.context, ctx.instantiation)
	return moduleInstance.ExportedFunction(cmd.Entrypoint).Call(callContext, cmd.Params...)
}
