// Package wasm provides the generic components used to build wazero plugins.
package wasm

import (
	"fmt"
	"io"

	"github.com/tetratelabs/wazero/api"
)

// Value represents a field in a function signature, which may be a parameter
// or a result.
type Value interface {
	// Writes a human-readable representation of the field to w, using the
	// memory and stack of a program to locate the field value.
	FormatValue(w io.Writer, memory api.Memory, stack []uint64)
	// Returns the sequence of primitive types that the result value is
	// composed of. Values of primitive types will have a single type,
	// but more complex types will have more (e.g. a buffer may hold two
	// 32 bits integers for the pointer and length). The number of types
	// indicates the number of words that are consumed from the stack when
	// loading the parameter.
	ValueTypes() []api.ValueType
}

// Param is an interface representing parameters of WebAssembly functions which
// are read form the stack.
//
// The interface is mainly used as a constraint for generic type parameters.
type Param[T any] interface {
	Value
	// Loads and returns the parameter value from the stack and optionally
	// reading its content from memory.
	LoadValue(memory api.Memory, stack []uint64) T
}

// Result is an interface reprenting results of WebAssembly functions which
// are written to the stack.
//
// The interface is mainly used as a constraint for generic type parameters.
type Result interface {
	Value
	// Stores the result value onto the stack and optionally writing its
	// content to memory.
	StoreValue(memory api.Memory, stack []uint64)
}

// ParamResult is an interface implemented by types which can be used as both a
// parameter and a result.
type ParamResult[T any] interface {
	Param[T]
	Result
}

// Object is an interface which represents values that can be loaded or stored
// in memory.
//
// The interface is mainly used as a constraint for generic type parameters.
type Object[T any] interface {
	// Writes a human-readable representation of the object to w.
	FormatObject(w io.Writer, memory api.Memory, object []byte)
	// Loads and returns the object from the given byte slice. If the object
	// contains pointers it migh read them from the module memory passed as
	// first argument.
	LoadObject(memory api.Memory, object []byte) T
	// Stores the object to the given tye slice. If the object contains pointers
	// it might write them to the module memory passed as first argument (doing
	// so without writing to arbitrary location is difficult so this use case is
	// rather rare).
	StoreObject(memory api.Memory, object []byte)
	// Returns the size of the the object in bytes. The byte slices passed to
	// LoadObject and StoreObjects are guaranteed to be at least of the length
	// returned by this method.
	ObjectSize() int
}

// SEGFAULT is an error type used as value in panics triggered by reading
// outside of the addressable memory of a program.
type SEGFAULT struct{ Offset, Length uint32 }

func (err SEGFAULT) Error() string {
	return fmt.Sprintf("segmentation fault: @%08x/%d", err.Offset, err.Length)
}

// Read returns a byte slice from a module memory. The function calls Read on
// the given memory and panics if offset/length are beyond the range of memory.
func Read(memory api.Memory, offset, length uint32) []byte {
	b, ok := memory.Read(offset, length)
	if !ok {
		panic(SEGFAULT{offset, length})
	}
	return b
}
