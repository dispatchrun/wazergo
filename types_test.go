package wasm_test

import (
	"reflect"
	"testing"

	"github.com/stealthrocket/wasm-go"
)

type value[T any] interface {
	wasm.Param[T]
	wasm.Result
}

func TestLoadAndStoreValue(t *testing.T) {
	testLoadAndStoreValue(t, wasm.None{})
	testLoadAndStoreValue(t, wasm.OK)

	testLoadAndStoreValue(t, wasm.Int8(-1))
	testLoadAndStoreValue(t, wasm.Int16(-2))
	testLoadAndStoreValue(t, wasm.Int32(-3))
	testLoadAndStoreValue(t, wasm.Int64(-4))

	testLoadAndStoreValue(t, wasm.Uint8(1))
	testLoadAndStoreValue(t, wasm.Uint16(2))
	testLoadAndStoreValue(t, wasm.Uint32(3))
	testLoadAndStoreValue(t, wasm.Uint64(4))

	testLoadAndStoreValue(t, wasm.Float32(0.1))
	testLoadAndStoreValue(t, wasm.Float64(0.5))
}

func testLoadAndStoreValue[T value[T]](t *testing.T, value T) {
	var loaded T
	var stack = make([]uint64, len(value.ValueTypes()))

	value.StoreValue(nil, stack)
	loaded = loaded.LoadValue(nil, stack)

	if !reflect.DeepEqual(value, loaded) {
		t.Errorf("values mismatch: want=%#v got=%#v", value, loaded)
	}

	for i := range stack {
		stack[i] = 0
	}

	var optionalValue wasm.Optional[T]
	var optionalLoaded wasm.Optional[T]

	stack = make([]uint64, len(optionalValue.ValueTypes()))
	optionalValue = wasm.Res(value)
	optionalValue.StoreValue(nil, stack)
	optionalLoaded = optionalLoaded.LoadValue(nil, stack)

	if !reflect.DeepEqual(optionalValue, optionalLoaded) {
		t.Errorf("optional values mismatch: want=%#v got=%#v", optionalValue, optionalLoaded)
	}
}

func TestLoadAndStoreObject(t *testing.T) {
	testLoadAndStoreObject(t, wasm.None{})

	testLoadAndStoreObject(t, wasm.Int8(-1))
	testLoadAndStoreObject(t, wasm.Int16(-2))
	testLoadAndStoreObject(t, wasm.Int32(-3))
	testLoadAndStoreObject(t, wasm.Int64(-4))

	testLoadAndStoreObject(t, wasm.Uint8(1))
	testLoadAndStoreObject(t, wasm.Uint16(2))
	testLoadAndStoreObject(t, wasm.Uint32(3))
	testLoadAndStoreObject(t, wasm.Uint64(4))

	testLoadAndStoreObject(t, wasm.Float32(0.1))
	testLoadAndStoreObject(t, wasm.Float64(0.5))
}

func testLoadAndStoreObject[T wasm.Object[T]](t *testing.T, value T) {
	var loaded T
	var object = make([]byte, value.ObjectSize())

	value.StoreObject(nil, object)
	loaded = loaded.LoadObject(nil, object)

	if !reflect.DeepEqual(value, loaded) {
		t.Errorf("objects mismatch: want=%#v got=%#v", value, loaded)
	}
}
