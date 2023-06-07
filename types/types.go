package types

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"strconv"
	"syscall"
	"time"
	"unsafe"

	"github.com/stealthrocket/wazergo/wasm"
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

// UnsafeLoadObject is a helper which may be used to implement the LoadObject
// method for object types which do not contain any inner pointers.
//
//	func (v T) LoadObject(_ api.Memory, object []byte) T {
//		return types.UnsafeLoadObject[T](object)
//	}
func UnsafeLoadObject[T Object[T]](mem []byte) T {
	return *unsafeCastObject[T](mem)
}

// UnsafeStoreObject is a helper which may be used to implement the StoreObject
// method for object types which do not contain any inner pointers.
//
//	func (v T) StoreObject(_ api.Memory, object []byte) {
//		types.UnsafeStoreObject(object, v)
//	}
func UnsafeStoreObject[T Object[T]](mem []byte, obj T) {
	*unsafeCastObject[T](mem) = obj
}

func unsafeCastObject[T Object[T]](mem []byte) *T {
	return (*T)(unsafe.Pointer(unsafe.SliceData(mem)))
}

func objectSize[T Object[T]]() int {
	var typ T
	return typ.ObjectSize()
}

type Int8 int8

func (arg Int8) Format(w io.Writer) {
	fmt.Fprintf(w, "%d", arg)
}

func (arg Int8) FormatValue(w io.Writer, memory api.Memory, stack []uint64) {
	arg.LoadValue(memory, stack).Format(w)
}

func (arg Int8) FormatObject(w io.Writer, memory api.Memory, object []byte) {
	arg.LoadObject(memory, object).Format(w)
}

func (arg Int8) LoadValue(memory api.Memory, stack []uint64) Int8 {
	return Int8(api.DecodeI32(stack[0]))
}

func (arg Int8) LoadObject(memory api.Memory, object []byte) Int8 {
	return Int8(object[0])
}

func (arg Int8) StoreValue(memory api.Memory, stack []uint64) {
	stack[0] = api.EncodeI32(int32(arg))
}

func (arg Int8) StoreObject(memory api.Memory, object []byte) {
	object[0] = byte(arg)
}

func (arg Int8) ValueTypes() []api.ValueType {
	return []api.ValueType{api.ValueTypeI32}
}

func (arg Int8) ObjectSize() int {
	return 1
}

var (
	_ Object[Int8] = Int8(0)
	_ Param[Int8]  = Int8(0)
	_ Result       = Int8(0)
	_ Formatter    = Int8(0)
)

type Int16 int16

func (arg Int16) Format(w io.Writer) {
	fmt.Fprintf(w, "%d", arg)
}

func (arg Int16) FormatValue(w io.Writer, memory api.Memory, stack []uint64) {
	arg.LoadValue(memory, stack).Format(w)
}

func (arg Int16) FormatObject(w io.Writer, memory api.Memory, object []byte) {
	arg.LoadObject(memory, object).Format(w)
}

func (arg Int16) LoadValue(memory api.Memory, stack []uint64) Int16 {
	return Int16(api.DecodeI32(stack[0]))
}

func (arg Int16) LoadObject(memory api.Memory, object []byte) Int16 {
	return Int16(binary.LittleEndian.Uint16(object))
}

func (arg Int16) StoreValue(memory api.Memory, stack []uint64) {
	stack[0] = api.EncodeI32(int32(arg))
}

func (arg Int16) StoreObject(memory api.Memory, object []byte) {
	binary.LittleEndian.PutUint16(object, uint16(arg))
}

func (arg Int16) ValueTypes() []api.ValueType {
	return []api.ValueType{api.ValueTypeI32}
}

func (arg Int16) ObjectSize() int {
	return 2
}

var (
	_ Object[Int16] = Int16(0)
	_ Param[Int16]  = Int16(0)
	_ Result        = Int16(0)
	_ Formatter     = Int16(0)
)

type Int32 int32

func (arg Int32) Format(w io.Writer) {
	fmt.Fprintf(w, "%d", arg)
}

func (arg Int32) FormatValue(w io.Writer, memory api.Memory, stack []uint64) {
	arg.LoadValue(memory, stack).Format(w)
}

func (arg Int32) FormatObject(w io.Writer, memory api.Memory, object []byte) {
	arg.LoadObject(memory, object).Format(w)
}

func (arg Int32) LoadValue(memory api.Memory, stack []uint64) Int32 {
	return Int32(api.DecodeI32(stack[0]))
}

func (arg Int32) LoadObject(memory api.Memory, object []byte) Int32 {
	return Int32(binary.LittleEndian.Uint32(object))
}

func (arg Int32) StoreValue(memory api.Memory, stack []uint64) {
	stack[0] = api.EncodeI32(int32(arg))
}

func (arg Int32) StoreObject(memory api.Memory, object []byte) {
	binary.LittleEndian.PutUint32(object, uint32(arg))
}

func (arg Int32) ValueTypes() []api.ValueType {
	return []api.ValueType{api.ValueTypeI32}
}

func (arg Int32) ObjectSize() int {
	return 4
}

var (
	_ Object[Int32] = Int32(0)
	_ Param[Int32]  = Int32(0)
	_ Result        = Int32(0)
	_ Formatter     = Int32(0)
)

type Int64 int64

func (arg Int64) Format(w io.Writer) {
	fmt.Fprintf(w, "%d", arg)
}

func (arg Int64) FormatValue(w io.Writer, memory api.Memory, stack []uint64) {
	arg.LoadValue(memory, stack).Format(w)
}

func (arg Int64) FormatObject(w io.Writer, memory api.Memory, object []byte) {
	arg.LoadObject(memory, object).Format(w)
}

func (arg Int64) LoadValue(memory api.Memory, stack []uint64) Int64 {
	return Int64(stack[0])
}

func (arg Int64) LoadObject(memory api.Memory, object []byte) Int64 {
	return Int64(binary.LittleEndian.Uint64(object))
}

func (arg Int64) StoreValue(memory api.Memory, stack []uint64) {
	stack[0] = uint64(arg)
}

func (arg Int64) StoreObject(memory api.Memory, object []byte) {
	binary.LittleEndian.PutUint64(object, uint64(arg))
}

func (arg Int64) ValueTypes() []api.ValueType {
	return []api.ValueType{api.ValueTypeI64}
}

func (arg Int64) ObjectSize() int {
	return 8
}

var (
	_ Object[Int64] = Int64(0)
	_ Param[Int64]  = Int64(0)
	_ Result        = Int64(0)
	_ Formatter     = Int64(0)
)

type Bool bool

func (arg Bool) Format(w io.Writer) {
	fmt.Fprintf(w, "%t", arg)
}

func (arg Bool) FormatValue(w io.Writer, memory api.Memory, stack []uint64) {
	arg.LoadValue(memory, stack).Format(w)
}

func (arg Bool) FormatObject(w io.Writer, memory api.Memory, object []byte) {
	arg.LoadObject(memory, object).Format(w)
}

func (arg Bool) LoadValue(memory api.Memory, stack []uint64) Bool {
	return Bool(api.DecodeU32(stack[0]) != 0)
}

func (arg Bool) LoadObject(memory api.Memory, object []byte) Bool {
	return Bool(object[0] != 0)
}

func (arg Bool) StoreValue(memory api.Memory, stack []uint64) {
	stack[0] = api.EncodeU32(uint32(arg.byte()))
}

func (arg Bool) StoreObject(memory api.Memory, object []byte) {
	object[0] = arg.byte()
}

func (arg Bool) ValueTypes() []api.ValueType {
	return []api.ValueType{api.ValueTypeI32}
}

func (arg Bool) ObjectSize() int {
	return 1
}

func (arg Bool) byte() byte {
	if arg {
		return 1
	}
	return 0
}

var (
	_ Object[Bool] = Bool(false)
	_ Param[Bool]  = Bool(false)
	_ Result       = Bool(false)
	_ Formatter    = Bool(false)
)

type Uint8 uint8

func (arg Uint8) Format(w io.Writer) {
	fmt.Fprintf(w, "%d", arg)
}

func (arg Uint8) FormatValue(w io.Writer, memory api.Memory, stack []uint64) {
	arg.LoadValue(memory, stack).Format(w)
}

func (arg Uint8) FormatObject(w io.Writer, memory api.Memory, object []byte) {
	arg.LoadObject(memory, object).Format(w)
}

func (arg Uint8) LoadValue(memory api.Memory, stack []uint64) Uint8 {
	return Uint8(api.DecodeU32(stack[0]))
}

func (arg Uint8) LoadObject(memory api.Memory, object []byte) Uint8 {
	return Uint8(object[0])
}

func (arg Uint8) StoreValue(memory api.Memory, stack []uint64) {
	stack[0] = api.EncodeU32(uint32(arg))
}

func (arg Uint8) StoreObject(memory api.Memory, object []byte) {
	object[0] = byte(arg)
}

func (arg Uint8) ValueTypes() []api.ValueType {
	return []api.ValueType{api.ValueTypeI32}
}

func (arg Uint8) ObjectSize() int {
	return 1
}

var (
	_ Object[Uint8] = Uint8(0)
	_ Param[Uint8]  = Uint8(0)
	_ Result        = Uint8(0)
	_ Formatter     = Uint8(0)
)

type Uint16 uint16

func (arg Uint16) Format(w io.Writer) {
	fmt.Fprintf(w, "%d", arg)
}

func (arg Uint16) FormatValue(w io.Writer, memory api.Memory, stack []uint64) {
	arg.LoadValue(memory, stack).Format(w)
}

func (arg Uint16) FormatObject(w io.Writer, memory api.Memory, object []byte) {
	arg.LoadObject(memory, object).Format(w)
}

func (arg Uint16) LoadValue(memory api.Memory, stack []uint64) Uint16 {
	return Uint16(api.DecodeU32(stack[0]))
}

func (arg Uint16) LoadObject(memory api.Memory, object []byte) Uint16 {
	return Uint16(binary.LittleEndian.Uint16(object))
}

func (arg Uint16) StoreValue(memory api.Memory, stack []uint64) {
	stack[0] = api.EncodeU32(uint32(arg))
}

func (arg Uint16) StoreObject(memory api.Memory, object []byte) {
	binary.LittleEndian.PutUint16(object, uint16(arg))
}

func (arg Uint16) ValueTypes() []api.ValueType {
	return []api.ValueType{api.ValueTypeI32}
}

func (arg Uint16) ObjectSize() int {
	return 2
}

var (
	_ Object[Uint16] = Uint16(0)
	_ Param[Uint16]  = Uint16(0)
	_ Result         = Uint16(0)
	_ Formatter      = Uint16(0)
)

type Uint32 uint32

func (arg Uint32) Format(w io.Writer) {
	fmt.Fprintf(w, "%d", arg)
}

func (arg Uint32) FormatValue(w io.Writer, memory api.Memory, stack []uint64) {
	arg.LoadValue(memory, stack).Format(w)
}

func (arg Uint32) FormatObject(w io.Writer, memory api.Memory, object []byte) {
	arg.LoadObject(memory, object).Format(w)
}

func (arg Uint32) LoadValue(memory api.Memory, stack []uint64) Uint32 {
	return Uint32(api.DecodeU32(stack[0]))
}

func (arg Uint32) LoadObject(memory api.Memory, object []byte) Uint32 {
	return Uint32(binary.LittleEndian.Uint32(object))
}

func (arg Uint32) StoreValue(memory api.Memory, stack []uint64) {
	stack[0] = api.EncodeU32(uint32(arg))
}

func (arg Uint32) StoreObject(memory api.Memory, object []byte) {
	binary.LittleEndian.PutUint32(object, uint32(arg))
}

func (arg Uint32) ValueTypes() []api.ValueType {
	return []api.ValueType{api.ValueTypeI32}
}

func (arg Uint32) ObjectSize() int {
	return 4
}

var (
	_ Object[Uint32] = Uint32(0)
	_ Param[Uint32]  = Uint32(0)
	_ Result         = Uint32(0)
	_ Formatter      = Uint32(0)
)

type Uint64 uint64

func (arg Uint64) Format(w io.Writer) {
	fmt.Fprintf(w, "%d", arg)
}

func (arg Uint64) FormatValue(w io.Writer, memory api.Memory, stack []uint64) {
	arg.LoadValue(memory, stack).Format(w)
}

func (arg Uint64) FormatObject(w io.Writer, memory api.Memory, object []byte) {
	arg.LoadObject(memory, object).Format(w)
}

func (arg Uint64) LoadValue(memory api.Memory, stack []uint64) Uint64 {
	return Uint64(stack[0])
}

func (arg Uint64) LoadObject(memory api.Memory, object []byte) Uint64 {
	return Uint64(binary.LittleEndian.Uint64(object))
}

func (arg Uint64) StoreValue(memory api.Memory, stack []uint64) {
	stack[0] = uint64(arg)
}

func (arg Uint64) StoreObject(memory api.Memory, object []byte) {
	binary.LittleEndian.PutUint64(object, uint64(arg))
}

func (arg Uint64) ValueTypes() []api.ValueType {
	return []api.ValueType{api.ValueTypeI64}
}

func (arg Uint64) ObjectSize() int {
	return 8
}

var (
	_ Object[Uint64] = Uint64(0)
	_ Param[Uint64]  = Uint64(0)
	_ Result         = Uint64(0)
	_ Formatter      = Uint64(0)
)

type Float32 float32

func (arg Float32) Format(w io.Writer) {
	fmt.Fprintf(w, "%g", arg)
}

func (arg Float32) FormatValue(w io.Writer, memory api.Memory, stack []uint64) {
	arg.LoadValue(memory, stack).Format(w)
}

func (arg Float32) FormatObject(w io.Writer, memory api.Memory, object []byte) {
	arg.LoadObject(memory, object).Format(w)
}

func (arg Float32) LoadValue(memory api.Memory, stack []uint64) Float32 {
	return Float32(api.DecodeF32(stack[0]))
}

func (arg Float32) LoadObject(memory api.Memory, object []byte) Float32 {
	return Float32(math.Float32frombits(binary.LittleEndian.Uint32(object)))
}

func (arg Float32) StoreValue(memory api.Memory, stack []uint64) {
	stack[0] = api.EncodeF32(float32(arg))
}

func (arg Float32) StoreObject(memory api.Memory, object []byte) {
	binary.LittleEndian.PutUint32(object, math.Float32bits(float32(arg)))
}

func (arg Float32) ValueTypes() []api.ValueType {
	return []api.ValueType{api.ValueTypeF32}
}

func (arg Float32) ObjectSize() int {
	return 4
}

var (
	_ Object[Float32] = Float32(0)
	_ Param[Float32]  = Float32(0)
	_ Result          = Float32(0)
	_ Formatter       = Float32(0)
)

type Float64 float64

func (arg Float64) Format(w io.Writer) {
	fmt.Fprintf(w, "%g", arg)
}

func (arg Float64) FormatValue(w io.Writer, memory api.Memory, stack []uint64) {
	arg.LoadValue(memory, stack).Format(w)
}

func (arg Float64) FormatObject(w io.Writer, memory api.Memory, object []byte) {
	arg.LoadObject(memory, object).Format(w)
}

func (arg Float64) LoadValue(memory api.Memory, stack []uint64) Float64 {
	return Float64(api.DecodeF64(stack[0]))
}

func (arg Float64) LoadObject(memory api.Memory, object []byte) Float64 {
	return Float64(math.Float64frombits(binary.LittleEndian.Uint64(object)))
}

func (arg Float64) StoreValue(memory api.Memory, stack []uint64) {
	stack[0] = api.EncodeF64(float64(arg))
}

func (arg Float64) StoreObject(memory api.Memory, object []byte) {
	binary.LittleEndian.PutUint64(object, math.Float64bits(float64(arg)))
}

func (arg Float64) ValueTypes() []api.ValueType {
	return []api.ValueType{api.ValueTypeF64}
}

func (arg Float64) ObjectSize() int {
	return 8
}

var (
	_ Param[Float64] = Float64(0)
	_ Result         = Float64(0)
	_ Formatter      = Float64(0)
)

type Duration time.Duration

func (arg Duration) Format(w io.Writer) {
	fmt.Fprintf(w, "%s", time.Duration(arg))
}

func (arg Duration) FormatValue(w io.Writer, memory api.Memory, stack []uint64) {
	arg.LoadValue(memory, stack).Format(w)
}

func (arg Duration) FormatObject(w io.Writer, memory api.Memory, object []byte) {
	arg.LoadObject(memory, object).Format(w)
}

func (arg Duration) LoadValue(memory api.Memory, stack []uint64) Duration {
	return Duration(stack[0])
}

func (arg Duration) LoadObject(memory api.Memory, object []byte) Duration {
	return Duration(binary.LittleEndian.Uint64(object))
}

func (arg Duration) StoreValue(memory api.Memory, stack []uint64) {
	stack[0] = uint64(arg)
}

func (arg Duration) StoreObject(memory api.Memory, object []byte) {
	binary.LittleEndian.PutUint64(object, uint64(arg))
}

func (arg Duration) ValueTypes() []api.ValueType {
	return []api.ValueType{api.ValueTypeI64}
}

func (arg Duration) ObjectSize() int {
	return 8
}

var (
	_ Object[Duration] = Duration(0)
	_ Param[Duration]  = Duration(0)
	_ Result           = Duration(0)
	_ Formatter        = Duration(0)
)

type primitive interface {
	~int8 | ~int16 | ~int32 | ~int64 | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr | ~float32 | ~float64
}

// Array is a type representing a sequence of contiguous items in memory. Array
// values are composed of a pair of pointer and number of items. The item size
// is determined by the size of the type T.
//
// Arrays only satisfy the Param interface, they cannot be used as return value
// of functions.
//
// At this time, arrays may only be composed of primitive Go types, but this
// restriction may be relaxed in a future version. Use List to laod sequences of
// complex types.
type Array[T primitive] []T

func (arg Array[T]) Format(w io.Writer) {
	fmt.Fprintf(w, "[")
	for i, v := range arg {
		if i > 0 {
			fmt.Fprintf(w, ", ")
		}
		fmt.Fprintf(w, "%v", v)
	}
	fmt.Fprintf(w, "]")
}

func (arg Array[T]) FormatObject(w io.Writer, memory api.Memory, object []byte) {
	arg.LoadObject(memory, object).Format(w)
}

func (arg Array[T]) FormatValue(w io.Writer, memory api.Memory, stack []uint64) {
	arg.LoadValue(memory, stack).Format(w)
}

func (arg Array[T]) LoadObject(memory api.Memory, object []byte) Array[T] {
	offset := binary.LittleEndian.Uint32(object[:4])
	length := binary.LittleEndian.Uint32(object[4:])
	return arg.load(memory, offset, length)
}

func (arg Array[T]) LoadValue(memory api.Memory, stack []uint64) Array[T] {
	offset := api.DecodeU32(stack[0])
	length := api.DecodeU32(stack[1])
	return arg.load(memory, offset, length)
}

func (arg Array[T]) load(memory api.Memory, offset, length uint32) Array[T] {
	size := unsafe.Sizeof(T(0))
	data := wasm.Read(memory, offset, length*uint32(size))
	return unsafe.Slice(*(**T)(unsafe.Pointer(&data)), length)
}

func (arg Array[T]) StoreObject(memory api.Memory, object []byte) {
	// TODO: move this to a compile time check by separating the load/store
	// capabilities of the Object interface.
	panic("NOT IMPLEMENTED")
}

func (arg Array[T]) ObjectSize() int {
	return 8
}

func (arg Array[T]) ValueTypes() []api.ValueType {
	return []api.ValueType{api.ValueTypeI32, api.ValueTypeI32}
}

var (
	_ Param[Array[byte]] = Array[byte](nil)
	_ Formatter          = Array[byte](nil)
)

// Bytes is a type alias for arrays of bytes, which is a common use case
// (e.g. I/O functions working on a byte buffer).
type Bytes Array[byte]

func (arg Bytes) Format(w io.Writer) {
	if len(arg) <= 32 {
		fmt.Fprintf(w, "%q", arg)
		return
	}
	b := arg[:20:20]
	b = append(b, "... ("...)
	b = strconv.AppendUint(b, uint64(len(arg)), 10)
	b = append(b, " bytes)"...)
	fmt.Fprintf(w, "%q", b)
}

func (arg Bytes) FormatObject(w io.Writer, memory api.Memory, object []byte) {
	arg.LoadObject(memory, object).Format(w)
}

func (arg Bytes) FormatValue(w io.Writer, memory api.Memory, stack []uint64) {
	arg.LoadValue(memory, stack).Format(w)
}

func (arg Bytes) LoadObject(memory api.Memory, object []byte) Bytes {
	return Bytes(arg.array().LoadObject(memory, object))
}

func (arg Bytes) LoadValue(memory api.Memory, stack []uint64) Bytes {
	return Bytes(arg.array().LoadValue(memory, stack))
}

func (arg Bytes) StoreObject(memory api.Memory, object []byte) {
	arg.array().StoreObject(memory, object)
}

func (arg Bytes) ObjectSize() int {
	return arg.array().ObjectSize()
}

func (arg Bytes) ValueTypes() []api.ValueType {
	return arg.array().ValueTypes()
}

func (arg Bytes) array() Array[byte] {
	return (Array[byte])(arg)
}

var (
	_ Param[Bytes] = Bytes(nil)
	_ Formatter    = Bytes(nil)
)

// String is similar to Bytes but holds the value as a Go string which is not
// sharing memory with the WebAssembly program memory anymore.
type String string

func (arg String) Format(w io.Writer) {
	fmt.Fprintf(w, "%q", arg)
}

func (arg String) FormatValue(w io.Writer, memory api.Memory, stack []uint64) {
	arg.LoadValue(memory, stack).Format(w)
}

func (arg String) LoadValue(memory api.Memory, stack []uint64) String {
	offset := api.DecodeU32(stack[0])
	length := api.DecodeU32(stack[1])
	return String(wasm.Read(memory, offset, length))
}

func (arg String) ValueTypes() []api.ValueType {
	return []api.ValueType{api.ValueTypeI32, api.ValueTypeI32}
}

var (
	_ Param[String] = String("")
	_ Formatter     = String("")
)

// Pointer is a parameter type used to represent a pointer to an object held in
// program memory.
type Pointer[T Object[T]] struct {
	memory api.Memory
	offset uint32
}

// New constructs a pointer to an object of type T backed by Go memory.
//
// This function is mostly useful to construct pointers to pass to module
// methods in tests, its usage in actual production code should be rare.
func New[T Object[T]]() Pointer[T] {
	return Ptr[T](wasm.NewFixedSizeMemory(uint32(objectSize[T]())), 0)
}

// Ptr constructs a pointer of objects T backed by a memory area at a specified
// offset.
//
// This function is mostly useful to construct pointers to pass to module
// methods in tests, its usage in actual production code should be rare.
func Ptr[T Object[T]](memory api.Memory, offset uint32) Pointer[T] {
	return Pointer[T]{memory, offset}
}

func (arg Pointer[T]) FormatValue(w io.Writer, memory api.Memory, stack []uint64) {
	arg = arg.LoadValue(memory, stack)
	fmt.Fprintf(w, "&")
	arg.Load().FormatObject(w, memory, arg.Object())
}

func (arg Pointer[T]) LoadValue(memory api.Memory, stack []uint64) Pointer[T] {
	return Pointer[T]{memory, api.DecodeU32(stack[0])}
}

func (arg Pointer[T]) ValueTypes() []api.ValueType {
	return []api.ValueType{api.ValueTypeI32}
}

func (arg Pointer[T]) FormatObject(w io.Writer, memory api.Memory, object []byte) {
	// TODO: recursively follow pointer
	fmt.Fprintf(w, "Pointer(%#x)", arg.offset)
}

func (arg Pointer[T]) LoadObject(memory api.Memory, object []byte) Pointer[T] {
	offset := uint32(binary.LittleEndian.Uint32(object))
	return Pointer[T]{memory, offset}
}

func (arg Pointer[T]) StoreObject(memory api.Memory, object []byte) {
	binary.LittleEndian.PutUint32(object, arg.offset)
}

func (arg Pointer[T]) ObjectSize() int {
	return 4
}

func (arg Pointer[T]) Memory() api.Memory {
	return arg.memory
}

func (arg Pointer[T]) Offset() uint32 {
	return arg.offset
}

func (arg Pointer[T]) Object() []byte {
	return wasm.Read(arg.memory, arg.offset, uint32(objectSize[T]()))
}

func (arg Pointer[T]) Load() (value T) {
	return value.LoadObject(arg.memory, arg.Object())
}

func (arg Pointer[T]) Store(value T) {
	value.StoreObject(arg.memory, arg.Object())
}

func (arg Pointer[T]) Index(index int) Pointer[T] {
	return Pointer[T]{arg.memory, arg.offset + uint32(index*objectSize[T]())}
}

func (arg Pointer[T]) Append(buffer []T, count int) []T {
	for i := 0; i < count; i++ {
		buffer = append(buffer, arg.Index(i).Load())
	}
	return buffer
}

func (arg Pointer[T]) Slice(count int) []T {
	return arg.Append(make([]T, 0, count), count)
}

func (arg Pointer[T]) UnsafeSlice(count int) []T {
	var typ T
	size := typ.ObjectSize()
	if count == 0 || size == 0 {
		return nil
	}
	data := wasm.Read(arg.memory, arg.offset, uint32(count*size))
	return unsafe.Slice((*T)(unsafe.Pointer(&data[0])), count)
}

var (
	_ Param[Pointer[None]] = Pointer[None]{}
)

// List represents a sequence of objects held in module memory.
type List[T Object[T]] struct {
	ptr Pointer[T]
	len uint32
}

// MakeList constructs a list from a pointer to an object of type T
// and a length.
func MakeList[T Object[T]](ptr Pointer[T], length int) List[T] {
	return List[T]{ptr, uint32(length)}
}

func (arg List[T]) FormatValue(w io.Writer, memory api.Memory, stack []uint64) {
	fmt.Fprintf(w, "[")
	arg = arg.LoadValue(memory, stack)
	for i := 0; i < arg.Len(); i++ {
		if i > 0 {
			fmt.Fprintf(w, ",")
		}
		p := arg.ptr.Index(i)
		v := p.Load()
		v.FormatObject(w, memory, p.Object())
	}
	fmt.Fprintf(w, "]")
}

func (arg List[T]) LoadValue(memory api.Memory, stack []uint64) List[T] {
	return List[T]{
		ptr: arg.ptr.LoadValue(memory, stack),
		len: api.DecodeU32(stack[1]),
	}
}

func (arg List[T]) ValueTypes() []api.ValueType {
	return []api.ValueType{api.ValueTypeI32, api.ValueTypeI32}
}

func (arg List[T]) Len() int {
	return int(arg.len)
}

func (arg List[T]) Index(index int) Pointer[T] {
	if index < 0 || index >= arg.Len() {
		panic(fmt.Errorf("%T: index out of bounds (%d/%d)", arg, index, arg.Len()))
	}
	return arg.ptr.Index(index)
}

func (arg List[T]) Range(fn func(int, T) bool) {
	for i := 0; i < arg.Len(); i++ {
		if !fn(i, arg.ptr.Index(i).Load()) {
			break
		}
	}
}

func (arg List[T]) Append(buffer []T) []T {
	return arg.ptr.Append(buffer, arg.Len())
}

func (arg List[T]) Slice() []T {
	return arg.ptr.Slice(arg.Len())
}

func (arg List[T]) UnsafeSlice() []T {
	return arg.ptr.UnsafeSlice(arg.Len())
}

var (
	_ Param[List[None]] = List[None]{}
)

// Optional represents a function result which may be missing due to the program
// encountering an error. The type contains either a value of type T or an error.
type Optional[T ParamResult[T]] struct {
	res T
	err error
}

// Value returns the underlying value of the optional. The method panics if opt
// contained an error.
func (opt Optional[T]) Result() T {
	if opt.err != nil {
		panic(opt.err)
	}
	return opt.res
}

// Error returns the error embedded in opt, or nil if opt contains a value.
func (opt Optional[T]) Error() error {
	return opt.err
}

func (opt Optional[T]) FormatValue(w io.Writer, memory api.Memory, stack []uint64) {
	if opt.err != nil {
		fmt.Fprintf(w, "ERROR: %v", opt.err)
	} else {
		opt.res.FormatValue(w, memory, stack)
	}
}

func (opt Optional[T]) LoadValue(memory api.Memory, stack []uint64) Optional[T] {
	n := len(opt.res.ValueTypes())
	opt.res = opt.res.LoadValue(memory, stack[:n:n])
	opt.err = makeErrno(api.DecodeI32(stack[n]))
	return opt
}

func (opt Optional[T]) StoreValue(memory api.Memory, stack []uint64) {
	if n := len(opt.res.ValueTypes()); opt.err != nil {
		for i := range stack[:n] {
			stack[i] = 0
		}
		stack[n] = api.EncodeI32(int32(AsErrno(opt.err)))
	} else {
		opt.res.StoreValue(memory, stack[:n:n])
		stack[n] = 0
	}
}

func (opt Optional[T]) ValueTypes() []api.ValueType {
	return append(opt.res.ValueTypes(), api.ValueTypeI32)
}

// Opt constructs an optional from a pair of a result and error.
func Opt[T ParamResult[T]](res T, err error) Optional[T] {
	return Optional[T]{res: res, err: err}
}

// Res constructs an optional from a value.
func Res[T ParamResult[T]](res T) Optional[T] {
	return Optional[T]{res: res}
}

// Err constructs an optional from an error. The function panics if the error is
// nil since a nil error indicates that the optional should contain a value.
func Err[T ParamResult[T]](err error) Optional[T] {
	if err == nil {
		panic("cannot create an optional error value from a nil error")
	}
	return Optional[T]{err: err}
}

var (
	_ Param[Optional[None]] = Optional[None]{}
	_ Result                = Optional[None]{}
)

// None is a special type of size zero bytes.
type None struct{}

func formatNone(w io.Writer) { io.WriteString(w, "(none)") }

func (None) Format(w io.Writer) { formatNone(w) }

func (None) FormatValue(w io.Writer, _ api.Memory, _ []uint64) { formatNone(w) }

func (None) FormatObject(w io.Writer, _ api.Memory, _ []byte) { formatNone(w) }

func (None) LoadValue(api.Memory, []uint64) (none None) { return }

func (None) LoadObject(api.Memory, []byte) (none None) { return }

func (None) StoreValue(api.Memory, []uint64) {}

func (None) StoreObject(api.Memory, []byte) {}

func (None) ValueTypes() []api.ValueType { return nil }

func (None) ObjectSize() int { return 0 }

var (
	_ Object[None] = None{}
	_ Param[None]  = None{}
	_ Result       = None{}
	_ Formatter    = None{}
)

// Error is a special optional type which either contains an error or no values.
// It is useful to represent return values of functions that may error but do
// not return any other values.
type Error = Optional[None]

// OK is a special value indicating that a function which returns nothing has
// not errored either.
var OK Optional[None]

// Fail constructs an Error value from the given Go error. The function panics
// if err is nil.
func Fail(err error) Error { return Err[None](err) }

// Errno is an error type representing error codes that are often returned by
// WebAssembly module functions.
//
// This type is employed when converting Go errors embedded into optional values
// to WebAssembly error codes. Go errors are converted to Errno values by
// unwrapping and inspecting the values to determine their error code. If a Go
// error has an Errno method, it is called to be converted to an Errno value.
//
// The special value 0 always indicates that there were no errors, similarly to
// how a Go program would treat a nil error.
type Errno int32

// Errno returns err as an int32 value.
func (err Errno) Errno() int32 {
	return int32(err)
}

// Error returns a human readable representation of err.
func (err Errno) Error() string {
	if i := int(err); i >= 0 && i < len(ErrorStrings) {
		if s := ErrorStrings[i]; s != "" {
			return s
		}
	}
	return fmt.Sprintf("errno(%d)", err)
}

func (err Errno) FormatValue(w io.Writer, memory api.Memory, stack []uint64) {
	io.WriteString(w, err.LoadValue(memory, stack).Error())
}

func (err Errno) LoadValue(memory api.Memory, stack []uint64) Errno {
	return Errno(api.DecodeI32(stack[0]))
}

func (err Errno) StoreValue(memory api.Memory, stack []uint64) {
	stack[0] = api.EncodeI32(int32(err))
}

func (err Errno) ValueTypes() []api.ValueType {
	return []api.ValueType{api.ValueTypeI32}
}

var (
	_ Param[Errno] = Errno(0)
	_ Result       = Errno(0)

	// ErrorStrings is a global used in the formatting of Errno values.
	//
	// The table is empty by default, but the program may assign a table indexed
	// by error code to customize the error messages.
	//
	// There is no synchronization so it is recommended to assign this global
	// during program initialization (e.g. in an init function).
	ErrorStrings []string
)

func makeErrno(errno int32) error {
	if errno == 0 {
		return nil
	}
	return Errno(errno)
}

func AsErrno(err error) Errno {
	if err == nil {
		return 0
	}
	for {
		switch e := errors.Unwrap(err).(type) {
		case nil:
			return -1 // unknown, just don't return 0
		case interface{ Errno() int32 }:
			return Errno(e.Errno())
		case syscall.Errno:
			return Errno(int32(e))
		default:
			err = e
		}
	}
}
