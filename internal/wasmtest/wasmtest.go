// Package wasmtest provides building blocks useful to write tests for wazero
// host modules.
package wasmtest

import (
	"context"
	"log"
	"math"
	"os"
	"reflect"

	"github.com/stealthrocket/wazergo"
	"github.com/tetratelabs/wazero"
)

type Context struct {
	context       context.Context
	runtime       wazero.Runtime
	logger        *log.Logger
	instantiation *wazergo.InstantiationContext
}

func NewContext(ctx context.Context, logger *log.Logger) *Context {
	runtime := wazero.NewRuntime(ctx)
	return &Context{
		context:       ctx,
		runtime:       runtime,
		logger:        logger,
		instantiation: wazergo.NewInstantiationContext(ctx, runtime),
	}
}

func (c *Context) Close() error {
	c.instantiation.Close(c.context)
	c.runtime.Close(c.context)
	return nil
}

func Load[T wazergo.Module](ctx *Context, m wazergo.HostModule[T], opts ...wazergo.Option[T]) {
	c, err := wazergo.Compile(ctx.context, ctx.runtime, m, wazergo.Log[T](ctx.logger))
	if err != nil {
		panic(err)
	}
	if _, err := wazergo.Instantiate(ctx.instantiation, c, opts...); err != nil {
		panic(err)
	}
}

type Cmd struct {
	Entrypoint string
	Params     []uint64
}

type CmdOption = wazergo.Option[*Cmd]

func Entrypoint(fn string) CmdOption {
	return wazergo.OptionFunc(func(cmd *Cmd) { cmd.Entrypoint = fn })
}

func Params(values ...any) CmdOption {
	params := make([]uint64, len(values))
	push(params, values)
	return wazergo.OptionFunc(func(cmd *Cmd) { cmd.Params = params })
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
	wazergo.Configure(cmd, opts...)
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
	callContext := wazergo.NewCallContext(ctx.context, ctx.instantiation)
	return moduleInstance.ExportedFunction(cmd.Entrypoint).Call(callContext, cmd.Params...)
}
