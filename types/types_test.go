package types_test

import (
	"reflect"
	"testing"

	. "github.com/stealthrocket/wazergo/types"
)

func TestLoadAndStoreValue(t *testing.T) {
	testLoadAndStoreValue(t, None{})
	testLoadAndStoreValue(t, OK)

	testLoadAndStoreValue(t, Bool(false))
	testLoadAndStoreValue(t, Bool(true))

	testLoadAndStoreValue(t, Int8(-1))
	testLoadAndStoreValue(t, Int16(-2))
	testLoadAndStoreValue(t, Int32(-3))
	testLoadAndStoreValue(t, Int64(-4))

	testLoadAndStoreValue(t, Uint8(1))
	testLoadAndStoreValue(t, Uint16(2))
	testLoadAndStoreValue(t, Uint32(3))
	testLoadAndStoreValue(t, Uint64(4))

	testLoadAndStoreValue(t, Float32(0.1))
	testLoadAndStoreValue(t, Float64(0.5))

	testLoadAndStoreValue(t, Duration(0))
	testLoadAndStoreValue(t, Duration(1e9))
}

func testLoadAndStoreValue[T ParamResult[T]](t *testing.T, value T) {
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

	var optionalValue Optional[T]
	var optionalLoaded Optional[T]

	stack = make([]uint64, len(optionalValue.ValueTypes()))
	optionalValue = Res(value)
	optionalValue.StoreValue(nil, stack)
	optionalLoaded = optionalLoaded.LoadValue(nil, stack)

	if !reflect.DeepEqual(optionalValue, optionalLoaded) {
		t.Errorf("optional values mismatch: want=%#v got=%#v", optionalValue, optionalLoaded)
	}
}

func TestLoadAndStoreObject(t *testing.T) {
	testLoadAndStoreObject(t, None{})

	testLoadAndStoreValue(t, Bool(false))
	testLoadAndStoreValue(t, Bool(true))

	testLoadAndStoreObject(t, Int8(-1))
	testLoadAndStoreObject(t, Int16(-2))
	testLoadAndStoreObject(t, Int32(-3))
	testLoadAndStoreObject(t, Int64(-4))

	testLoadAndStoreObject(t, Uint8(1))
	testLoadAndStoreObject(t, Uint16(2))
	testLoadAndStoreObject(t, Uint32(3))
	testLoadAndStoreObject(t, Uint64(4))

	testLoadAndStoreObject(t, Float32(0.1))
	testLoadAndStoreObject(t, Float64(0.5))

	testLoadAndStoreObject(t, Duration(0))
	testLoadAndStoreObject(t, Duration(1e9))
}

func testLoadAndStoreObject[T Object[T]](t *testing.T, value T) {
	var loaded T
	var object = make([]byte, value.ObjectSize())

	value.StoreObject(nil, object)
	loaded = loaded.LoadObject(nil, object)

	if !reflect.DeepEqual(value, loaded) {
		t.Errorf("objects mismatch: want=%#v got=%#v", value, loaded)
	}
}
