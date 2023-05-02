package types_test

import (
	"strings"
	"testing"

	. "github.com/stealthrocket/wazergo/types"
	"github.com/stealthrocket/wazergo/wasm"
)

type T0 struct{}
type T1 struct{ F int8 }
type T2 struct{ F int16 }
type T3 struct{ F int32 }
type T4 struct{ F int64 }
type T5 struct{ F uint8 }
type T6 struct{ F uint16 }
type T7 struct{ F uint32 }
type T8 struct{ F uint64 }
type T9 struct{ F float32 }
type T10 struct{ F float64 }
type T11 struct{ F [3]uint32 }

type Vec3d struct {
	X float32 `name:"x"`
	Y float32 `name:"y"`
	Z float32 `name:"z"`
}

func TestLoadAndStoreStruct(t *testing.T) {
	testLoadAndStoreObject(t, Struct[T0]{})
	testLoadAndStoreObject(t, Struct[T1]{Value: T1{F: 1}})
	testLoadAndStoreObject(t, Struct[T2]{Value: T2{F: 2}})
	testLoadAndStoreObject(t, Struct[T3]{Value: T3{F: 3}})
	testLoadAndStoreObject(t, Struct[T4]{Value: T4{F: 4}})
	testLoadAndStoreObject(t, Struct[T5]{Value: T5{F: 5}})
	testLoadAndStoreObject(t, Struct[T6]{Value: T6{F: 6}})
	testLoadAndStoreObject(t, Struct[T7]{Value: T7{F: 7}})
	testLoadAndStoreObject(t, Struct[T8]{Value: T8{F: 8}})
	testLoadAndStoreObject(t, Struct[T9]{Value: T9{F: 9}})
	testLoadAndStoreObject(t, Struct[T10]{Value: T10{F: 10}})
	testLoadAndStoreObject(t, Struct[T11]{Value: T11{F: [3]uint32{1, 2, 3}}})
	testLoadAndStoreObject(t, Struct[Vec3d]{Value: Vec3d{1, 2, 3}})
}

func TestFormatArray(t *testing.T) {
	value := Struct[T11]{
		Value: T11{
			F: [3]uint32{1, 2, 3},
		},
	}
	offset := uint32(0)
	length := uint32(value.ObjectSize())
	memory := wasm.NewFixedSizeMemory(length)
	object := wasm.Read(memory, offset, length)
	value.StoreObject(memory, object)

	output := new(strings.Builder)
	value.FormatObject(output, memory, object)

	if s := output.String(); s != `{F:[1,2,3]}` {
		t.Errorf("wrong format: %s", s)
	}
}

func TestFormatStruct(t *testing.T) {
	value := Struct[Vec3d]{
		Value: Vec3d{
			X: 1,
			Y: 2,
			Z: 3,
		},
	}
	offset := uint32(0)
	length := uint32(value.ObjectSize())
	memory := wasm.NewFixedSizeMemory(length)
	object := wasm.Read(memory, offset, length)
	value.StoreObject(memory, object)

	output := new(strings.Builder)
	value.FormatObject(output, memory, object)

	if s := output.String(); s != `{x:1,y:2,z:3}` {
		t.Errorf("wrong format: %s", s)
	}
}

func BenchmarkStructObjectSize(b *testing.B) {
	v := Struct[T0]{}

	for i := 0; i < b.N; i++ {
		_ = v.ObjectSize()
	}
}

func BenchmarkStructLoadObject(b *testing.B) {
	v := Struct[Vec3d]{}
	m := make([]byte, v.ObjectSize())

	for i := 0; i < b.N; i++ {
		v = v.LoadObject(nil, m)
	}
}

func BenchmarkStructStoreObject(b *testing.B) {
	v := Struct[Vec3d]{}
	m := make([]byte, v.ObjectSize())

	for i := 0; i < b.N; i++ {
		v.StoreObject(nil, m)
	}
}
