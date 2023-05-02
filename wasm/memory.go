package wasm

import (
	"encoding/binary"
	"math"

	"github.com/tetratelabs/wazero/api"
)

// PageSize is the size of memory pages in WebAssembly programs (64 KiB).
const PageSize = 64 * 1024

func ceil(size uint32) uint32 {
	size += PageSize - 1
	size /= PageSize
	size *= PageSize
	return size
}

type memoryDefinition struct{ *Memory }

func (def memoryDefinition) ModuleName() string { return "" }

func (def memoryDefinition) Index() uint32 { return 0 }

func (def memoryDefinition) Import() (moduleName, name string, isImport bool) { return }

func (def memoryDefinition) ExportNames() []string { return nil }

func (def memoryDefinition) Min() uint32 { return 0 }

func (def memoryDefinition) Max() (uint32, bool) { return ceil(uint32(len(def.memory))), true }

// Memory is an implementation of the api.Memory interface of wazero backed by
// a Go byte slice. The memory has a fixed size and cannot grow nor shrink.
//
// This type is mostly useful in tests to construct memory areas where output
// parameters can be stored.
type Memory struct {
	memory []byte
	api.Memory
}

// NewFixedSizeMemory constructs a Memory instance of size bytes aligned on the
// WebAssembly page size.
func NewFixedSizeMemory(size uint32) *Memory {
	return &Memory{
		memory: make([]byte, ceil(size)),
	}
}

func (mem *Memory) Definition() api.MemoryDefinition { return memoryDefinition{Memory: mem} }

func (mem *Memory) Size() uint32 { return uint32(len(mem.memory)) }

func (mem *Memory) Grow(uint32) (uint32, bool) { return ceil(uint32(len(mem.memory))), false }

func (mem *Memory) ReadByte(offset uint32) (byte, bool) {
	if mem.isOutOfRange(offset, 1) {
		return 0, false
	}
	return mem.memory[offset], true
}

func (mem *Memory) ReadUint16Le(offset uint32) (uint16, bool) {
	if mem.isOutOfRange(offset, 2) {
		return 0, false
	}
	return binary.LittleEndian.Uint16(mem.memory[offset:]), true
}

func (mem *Memory) ReadUint32Le(offset uint32) (uint32, bool) {
	if mem.isOutOfRange(offset, 4) {
		return 0, false
	}
	return binary.LittleEndian.Uint32(mem.memory[offset:]), true
}

func (mem *Memory) ReadUint64Le(offset uint32) (uint64, bool) {
	if mem.isOutOfRange(offset, 8) {
		return 0, false
	}
	return binary.LittleEndian.Uint64(mem.memory[offset:]), true
}

func (mem *Memory) ReadFloat32Le(offset uint32) (float32, bool) {
	v, ok := mem.ReadUint32Le(offset)
	return math.Float32frombits(v), ok
}

func (mem *Memory) ReadFloat64Le(offset uint32) (float64, bool) {
	v, ok := mem.ReadUint64Le(offset)
	return math.Float64frombits(v), ok
}

func (mem *Memory) Read(offset, length uint32) ([]byte, bool) {
	if mem.isOutOfRange(offset, length) {
		return nil, false
	}
	return mem.memory[offset : offset+length : offset+length], true
}

func (mem *Memory) WriteByte(offset uint32, value byte) bool {
	if mem.isOutOfRange(offset, 1) {
		return false
	}
	mem.memory[offset] = value
	return true
}

func (mem *Memory) WriteUint16Le(offset uint32, value uint16) bool {
	if mem.isOutOfRange(offset, 2) {
		return false
	}
	binary.LittleEndian.PutUint16(mem.memory[offset:], value)
	return true
}

func (mem *Memory) WriteUint32Le(offset uint32, value uint32) bool {
	if mem.isOutOfRange(offset, 4) {
		return false
	}
	binary.LittleEndian.PutUint32(mem.memory[offset:], value)
	return true
}

func (mem *Memory) WriteUint64Le(offset uint32, value uint64) bool {
	if mem.isOutOfRange(offset, 4) {
		return false
	}
	binary.LittleEndian.PutUint64(mem.memory[offset:], value)
	return true
}

func (mem *Memory) WriteFloat32Le(offset uint32, value float32) bool {
	return mem.WriteUint32Le(offset, math.Float32bits(value))
}

func (mem *Memory) WriteFloat64Le(offset uint32, value float64) bool {
	return mem.WriteUint64Le(offset, math.Float64bits(value))
}

func (mem *Memory) Write(offset uint32, value []byte) bool {
	if mem.isOutOfRange(offset, uint32(len(value))) {
		return false
	}
	copy(mem.memory[offset:], value)
	return true
}

func (mem *Memory) WriteString(offset uint32, value string) bool {
	if mem.isOutOfRange(offset, uint32(len(value))) {
		return false
	}
	copy(mem.memory[offset:], value)
	return true
}

func (mem *Memory) isOutOfRange(offset, length uint32) bool {
	size := mem.Size()
	return offset >= size || length > size || offset > (size-length)
}
