package types_test

import (
	"fmt"
	"io"
	"reflect"
	"strings"
	"testing"

	. "github.com/stealthrocket/wazergo/types"
	"github.com/tetratelabs/wazero/api"
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

type Vec3d struct {
	X, Y, Z float32
}

func (v Vec3d) FormatObject(w io.Writer, m api.Memory, object []byte) {
	v = v.LoadObject(m, object)
	fmt.Fprintf(w, "{x:%v,y:%v,z:%v}", v.X, v.Y, v.Z)
}

func (v Vec3d) LoadObject(_ api.Memory, object []byte) Vec3d {
	return UnsafeLoadObject[Vec3d](object)
}

func (v Vec3d) StoreObject(_ api.Memory, object []byte) {
	UnsafeStoreObject[Vec3d](object, v)
}

func (v Vec3d) ObjectSize() int {
	return 12
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

	testLoadAndStoreObject(t, Vec3d{1, 2, 3})
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

func TestFormatObject(t *testing.T) {
	testFormatObject(t, None{}, `(none)`)

	testFormatObject(t, Bool(false), `false`)
	testFormatObject(t, Bool(true), `true`)

	testFormatObject(t, Int8(-1), `-1`)
	testFormatObject(t, Int16(-2), `-2`)
	testFormatObject(t, Int32(-3), `-3`)
	testFormatObject(t, Int64(-4), `-4`)

	testFormatObject(t, Uint8(1), `1`)
	testFormatObject(t, Uint16(2), `2`)
	testFormatObject(t, Uint32(3), `3`)
	testFormatObject(t, Uint64(4), `4`)

	testFormatObject(t, Float32(0.1), `0.1`)
	testFormatObject(t, Float64(0.5), `0.5`)

	testFormatObject(t, Duration(0), `0s`)
	testFormatObject(t, Duration(1e9), `1s`)

	testFormatObject(t, Vec3d{1, 2, 3}, `{x:1,y:2,z:3}`)
}

func testFormatObject[T Object[T]](t *testing.T, value T, format string) {
	buffer := new(strings.Builder)
	object := make([]byte, value.ObjectSize())

	value.StoreObject(nil, object)
	value.FormatObject(buffer, nil, object)

	if s := buffer.String(); s != format {
		t.Errorf("object format mismatch: want=%q got=%q", format, s)
	}
}
