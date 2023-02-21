package wasmtest

import (
	"context"

	"github.com/stealthrocket/wasm-go"
	"github.com/tetratelabs/wazero/api"
)

func Call[R wasm.Param[R], T any](fn wasm.Function[T], ctx context.Context, module api.Module, this T, args ...wasm.Result) (ret R) {
	malloc = 0

	numParams := countValueTypes(fn.Params)
	numResults := countValueTypes(fn.Results)

	stackSize := numParams
	if numResults > stackSize {
		stackSize = numResults
	}

	stack := make([]uint64, stackSize)
	memory := module.Memory()
	offset := 0

	for _, arg := range args {
		arg.StoreValue(memory, stack[offset:])
		offset += len(arg.ValueTypes())
	}

	fn.Func(this, ctx, module, stack)
	return ret.LoadValue(memory, stack)
}

func countValueTypes(values []wasm.Value) (n int) {
	for _, v := range values {
		n += len(v.ValueTypes())
	}
	return n
}
