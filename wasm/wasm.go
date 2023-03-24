// Package wasm provides the generic components used to build wazero plugins.
package wasm

import (
	"fmt"

	"github.com/tetratelabs/wazero/api"
)

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
