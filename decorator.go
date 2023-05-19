package wazergo

import (
	"context"
	"fmt"
	"io"
	"log"
	"strings"

	. "github.com/stealthrocket/wazergo/types"
	"github.com/tetratelabs/wazero/api"
)

// Decorator is an interface type which applies a transformation to a function.
type Decorator[T Module] interface {
	Decorate(module string, fn Function[T]) Function[T]
}

// DecoratorFunc is a helper used to create decorators from functions using type
// inference to keep the syntax simple.
func DecoratorFunc[T Module](d func(string, Function[T]) Function[T]) Decorator[T] {
	return decoratorFunc[T](d)
}

type decoratorFunc[T Module] func(string, Function[T]) Function[T]

func (d decoratorFunc[T]) Decorate(module string, fn Function[T]) Function[T] { return d(module, fn) }

// Log constructs a function decorator which adds logging to function calls.
func Log[T Module](logger *log.Logger) Decorator[T] {
	return DecoratorFunc(func(module string, fn Function[T]) Function[T] {
		if logger == nil {
			return fn
		}
		n := fn.StackParamCount()
		return fn.WithFunc(func(this T, ctx context.Context, module api.Module, stack []uint64) {
			params := make([]uint64, n)
			copy(params, stack)

			panicked := true
			defer func() {
				memory := module.Memory()
				buffer := new(strings.Builder)
				defer logger.Printf("%s", buffer)

				fmt.Fprintf(buffer, "%s::%s(", module, fn.Name)
				formatValues(buffer, memory, params, fn.Params)
				fmt.Fprintf(buffer, ")")

				if panicked {
					fmt.Fprintf(buffer, " PANIC!")
				} else {
					fmt.Fprintf(buffer, " → ")
					formatValues(buffer, memory, stack, fn.Results)
				}
			}()

			fn.Func(this, ctx, module, stack)
			panicked = false
		})
	})
}

func formatValues(w io.Writer, memory api.Memory, stack []uint64, values []Value) {
	for i, v := range values {
		if i > 0 {
			fmt.Fprintf(w, ", ")
		}
		v.FormatValue(w, memory, stack)
		stack = stack[len(v.ValueTypes()):]
	}
}

// Decorate returns a version of the given host module where the decorators were
// applied to all its functions.
func Decorate[T Module](mod HostModule[T], decorators ...Decorator[T]) HostModule[T] {
	functions := mod.Functions()
	decorated := &decoratedHostModule[T]{
		hostModule: mod,
		functions:  make(Functions[T], len(functions)),
	}
	moduleName := mod.Name()
	for name, function := range functions {
		for _, decorator := range decorators {
			function = decorator.Decorate(moduleName, function)
		}
		decorated.functions[name] = function
	}
	return decorated
}

type decoratedHostModule[T Module] struct {
	hostModule HostModule[T]
	functions  Functions[T]
}

func (m *decoratedHostModule[T]) Name() string {
	return m.hostModule.Name()
}

func (m *decoratedHostModule[T]) Functions() Functions[T] {
	return m.functions
}

func (m *decoratedHostModule[T]) Instantiate(ctx context.Context, options ...Option[T]) (T, error) {
	return m.hostModule.Instantiate(ctx, options...)
}
