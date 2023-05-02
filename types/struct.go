package types

import (
	"io"
	"reflect"
	"strconv"
	"sync/atomic"
	"unsafe"

	"github.com/tetratelabs/wazero/api"
)

// Struct is an implementation of the Object[T] interface intended to
// facilitate the definition of custom struct types in the signature of
// host functions.
type Struct[T any] struct {
	Value T
}

func (s Struct[T]) FormatObject(w io.Writer, memory api.Memory, object []byte) {
	cachedObjectType[T]().formatObject(unsafe.Pointer(&s.Value), w, memory, object)
}

func (s Struct[T]) LoadObject(memory api.Memory, object []byte) Struct[T] {
	cachedObjectType[T]().loadObject(unsafe.Pointer(&s.Value), memory, object)
	return s
}

func (s Struct[T]) StoreObject(memory api.Memory, object []byte) {
	cachedObjectType[T]().storeObject(unsafe.Pointer(&s.Value), memory, object)
}

func (s Struct[T]) ObjectSize() int {
	return cachedObjectType[T]().objectSize()
}

var (
	_ Object[Struct[None]] = Struct[None]{}

	objectTypes atomic.Value // map[unsafe.Pointer]objectType
)

type objectType interface {
	formatObject(unsafe.Pointer, io.Writer, api.Memory, []byte)

	loadObject(unsafe.Pointer, api.Memory, []byte)

	storeObject(unsafe.Pointer, api.Memory, []byte)

	objectSize() int
}

func typeid(t reflect.Type) unsafe.Pointer {
	return (*[2]unsafe.Pointer)(unsafe.Pointer(&t))[1]
}

func cachedObjectType[T any]() objectType {
	return cachedObjectTypeOf(reflect.TypeOf((*T)(nil)).Elem())
}

func cachedObjectTypeOf(t reflect.Type) objectType {
	cache, _ := objectTypes.Load().(map[unsafe.Pointer]objectType)
	typ, ok := cache[typeid(t)]
	if !ok {
		typ = objectTypeOf(t)
		newCache := make(map[unsafe.Pointer]objectType, len(cache)+1)
		for k, v := range cache {
			newCache[k] = v
		}
		newCache[typeid(t)] = typ
		objectTypes.Store(newCache)
	}
	return typ
}

func objectTypeOf(t reflect.Type) objectType {
	switch t.Kind() {
	case reflect.Int8:
		return object[Int8]{}
	case reflect.Int16:
		return object[Int16]{}
	case reflect.Int32:
		return object[Int32]{}
	case reflect.Int64:
		return object[Int64]{}
	case reflect.Uint8:
		return object[Uint8]{}
	case reflect.Uint16:
		return object[Uint16]{}
	case reflect.Uint32:
		return object[Uint32]{}
	case reflect.Uint64:
		return object[Uint64]{}
	case reflect.Float32:
		return object[Float32]{}
	case reflect.Float64:
		return object[Float64]{}
	case reflect.Struct:
		return structTypeOf(t)
	case reflect.Array:
		return arrayTypeOf(t)
	default:
		panic("cannot construct wasm type from Go value of type: " + t.String())
	}
}

type arrayType struct {
	typ  objectType
	len  int
	size int
	elem uintptr
}

func arrayTypeOf(t reflect.Type) *arrayType {
	elemType := t.Elem()
	itemType := objectTypeOf(elemType)
	return &arrayType{
		typ:  itemType,
		len:  t.Len(),
		size: itemType.objectSize(),
		elem: uintptr(elemType.Size()),
	}
}

func (t *arrayType) formatObject(p unsafe.Pointer, w io.Writer, m api.Memory, object []byte) {
	io.WriteString(w, "[")

	for i := 0; i < t.len; i++ {
		if i != 0 {
			io.WriteString(w, ",")
		}
		t.typ.formatObject(unsafe.Add(p, uintptr(i)*t.elem), w, m, object[:t.size])
		object = object[t.size:]
	}

	io.WriteString(w, "]")
}

func (t *arrayType) loadObject(p unsafe.Pointer, m api.Memory, object []byte) {
	for i := 0; i < t.len; i++ {
		t.typ.loadObject(unsafe.Add(p, uintptr(i)*t.elem), m, object[:t.size])
		object = object[t.size:]
	}
}

func (t *arrayType) storeObject(p unsafe.Pointer, m api.Memory, object []byte) {
	for i := 0; i < t.len; i++ {
		t.typ.storeObject(unsafe.Add(p, uintptr(i)*t.elem), m, object[:t.size])
		object = object[t.size:]
	}
}

func (t *arrayType) objectSize() int {
	return t.len * t.size
}

type structType struct {
	fields []structField
	size   int
}

func structTypeOf(t reflect.Type) *structType {
	st := &structType{
		fields: structFieldsOf(t),
	}
	for _, f := range st.fields {
		st.size += f.size
	}
	return st
}

func (t *structType) formatObject(p unsafe.Pointer, w io.Writer, m api.Memory, object []byte) {
	io.WriteString(w, "{")

	for i, f := range t.fields {
		if i != 0 {
			io.WriteString(w, ",")
		}
		io.WriteString(w, f.name)
		io.WriteString(w, ":")
		n := f.size
		f.typ.formatObject(unsafe.Add(p, f.offset), w, m, object[:n])
		object = object[n:]
	}

	io.WriteString(w, "}")
}

func (t *structType) loadObject(p unsafe.Pointer, m api.Memory, object []byte) {
	for i := range t.fields {
		f := &t.fields[i]
		n := f.size
		f.typ.loadObject(unsafe.Add(p, f.offset), m, object[:n])
		object = object[n:]
	}
}

func (t *structType) storeObject(p unsafe.Pointer, m api.Memory, object []byte) {
	for i := range t.fields {
		f := &t.fields[i]
		n := f.size
		f.typ.storeObject(unsafe.Add(p, f.offset), m, object[:n])
		object = object[n:]
	}
}

func (t *structType) objectSize() int {
	return t.size
}

type structField struct {
	name   string
	typ    objectType
	size   int
	offset uintptr
}

func structFieldsOf(t reflect.Type) []structField {
	fields := make([]structField, 0, t.NumField())
	return appendStructFields(fields, t, 0)
}

func appendStructFields(fields []structField, t reflect.Type, offset uintptr) []structField {
	for _, f := range reflect.VisibleFields(t) {
		fieldOffset := offset + f.Offset
		if f.Anonymous {
			fields = appendStructFields(fields, f.Type, fieldOffset)
			continue
		}
		fieldName := f.Name
		fieldType := objectTypeOf(f.Type)
		fieldSize := fieldType.objectSize()
		if name := f.Tag.Get("name"); name != "" {
			fieldName = name
		}
		if size := f.Tag.Get("size"); size != "" {
			n, err := strconv.Atoi(size)
			if err != nil {
				panic(t.String() + "." + f.Name + ": invalid size tag")
			}
			fieldSize = n
		}
		if fieldName == "-" {
			continue
		}
		fields = append(fields, structField{
			name:   fieldName,
			typ:    fieldType,
			size:   fieldSize,
			offset: fieldOffset,
		})
	}
	return fields
}

type object[T Object[T]] struct{}

func (object[T]) formatObject(p unsafe.Pointer, w io.Writer, m api.Memory, object []byte) {
	x := (*T)(p)
	(*x).FormatObject(w, m, object)
}

func (object[T]) loadObject(p unsafe.Pointer, m api.Memory, object []byte) {
	x := (*T)(p)
	*x = (*x).LoadObject(m, object)
}

func (object[T]) storeObject(p unsafe.Pointer, m api.Memory, object []byte) {
	x := (*T)(p)
	(*x).StoreObject(m, object)
}

func (object[T]) objectSize() int {
	return objectSize[T]()
}

var _ objectType = object[None]{}
