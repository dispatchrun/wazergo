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
		f := fn.Func
		n := 0

		for _, v := range fn.Params {
			n += len(v.ValueTypes())
		}

		fn.Func = func(this T, ctx context.Context, module api.Module, stack []uint64) {
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

			f(this, ctx, module, stack)
			panicked = false
		}
		return fn
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
