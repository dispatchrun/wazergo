package wasmtest

import (
	"io"

	"github.com/stealthrocket/wasm-go"
	"github.com/tetratelabs/wazero/api"
)

var malloc uint32

func brk(memory api.Memory, offset, length uint32) ([]byte, uint32) {
	b := wasm.Read(memory, offset, length)
	malloc = offset + length
	return b, offset
}

func sbrk(memory api.Memory, size uint32) ([]byte, uint32) {
	return brk(memory, malloc, size)
}

// Bytes is an extension of the wasm.Bytes type which adds the ability to treat
// those values as results, so they can be passed as argument to Call* functions
// in tests. The content of the byte slice is copied to the WebAssembly module
// memory, starting at address zero in each invocation of Call* functions. This
// would be really unsafe to do in a production applicaiton, which is why the
// feature is only made available to unit tests.
type Bytes wasm.Bytes

func (arg Bytes) FormatValue(w io.Writer, memory api.Memory, stack []uint64) {
	wasm.Bytes(arg).FormatValue(w, memory, stack)
}

func (arg Bytes) LoadValue(memory api.Memory, stack []uint64) Bytes {
	return Bytes(wasm.Bytes(arg).LoadValue(memory, stack))
}

func (arg Bytes) StoreValue(memory api.Memory, stack []uint64) {
	b, offset := sbrk(memory, uint32(len(arg)))
	copy(b, arg)
	stack[0] = api.EncodeU32(offset)
	stack[1] = api.EncodeU32(uint32(len(arg)))
}

func (arg Bytes) ValueTypes() []api.ValueType {
	return wasm.Bytes(arg).ValueTypes()
}
