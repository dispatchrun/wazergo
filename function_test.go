package wasm_test

import (
	"context"
	"errors"
	"reflect"
	"strconv"
	"testing"

	"github.com/stealthrocket/wasm-go"
	"github.com/stealthrocket/wasm-go/wasmtest"
	"github.com/tetratelabs/wazero/api"
)

type instance struct{}

func (*instance) Close(context.Context) error { return nil }

type plugin struct{}

func (plugin) Name() string                                    { return "test" }
func (plugin) Functions() wasm.Functions[*instance]            { return nil }
func (plugin) Instantiate(...wasm.Option[*instance]) *instance { return nil }

func TestFunc0(t *testing.T) {
	oops := errors.New("oops")
	testFunc0(t, 1, func(*instance, context.Context) wasm.Int32 { return 1 })
	testFunc0(t, 2, func(*instance, context.Context) wasm.Int64 { return 2 })
	testFunc0(t, 3, func(*instance, context.Context) wasm.Uint32 { return 3 })
	testFunc0(t, 4, func(*instance, context.Context) wasm.Uint64 { return 4 })
	testFunc0(t, 0.1, func(*instance, context.Context) wasm.Float32 { return 0.1 })
	testFunc0(t, 0.5, func(*instance, context.Context) wasm.Float64 { return 0.5 })
	testFunc0(t, wasm.OK, func(*instance, context.Context) wasm.Error { return wasm.OK })
	testFunc0(t, wasm.Fail(^wasm.Errno(0)), func(*instance, context.Context) wasm.Error { return wasm.Fail(oops) })
}

func TestFunc1(t *testing.T) {
	testFunc1(t, 42, 42, func(this *instance, ctx context.Context, v wasm.Int32) wasm.Int32 {
		return v
	})
	testFunc1(t, wasm.Res(wasm.Int32(42)), wasmtest.Bytes("42"),
		func(this *instance, ctx context.Context, v wasmtest.Bytes) wasm.Optional[wasm.Int32] {
			i, err := strconv.Atoi(string(v))
			return wasm.Opt(wasm.Int32(i), err)
		},
	)
}

func TestFunc2(t *testing.T) {
	testFunc2(t, wasm.Res(wasm.Int32(41)), wasmtest.Bytes("42"), wasmtest.Bytes("-1"),
		func(this *instance, ctx context.Context, v1, v2 wasmtest.Bytes) wasm.Optional[wasm.Int32] {
			i1, _ := strconv.Atoi(string(v1))
			i2, _ := strconv.Atoi(string(v2))
			return wasm.Res(wasm.Int32(i1 + i2))
		},
	)
}

func testFunc(t *testing.T, opts []wasm.Option[*instance], test func(*instance, context.Context, api.Module)) {
	t.Helper()
	memory := wasm.NewFixedSizeMemory(wasm.PageSize)
	module := wasmtest.NewModule("test", wasmtest.Memory(memory))
	test(new(instance), context.Background(), module)
}

func testFunc0[R value[R]](t *testing.T, want R, f func(*instance, context.Context) R, opts ...wasm.Option[*instance]) {
	t.Helper()
	testFunc(t, opts, func(this *instance, ctx context.Context, module api.Module) {
		t.Helper()
		assertEqual(t, want, wasmtest.Call[R](wasm.F0(f), ctx, module, this))
	})
}

func testFunc1[R value[R], T value[T]](t *testing.T, want R, arg T, f func(*instance, context.Context, T) R, opts ...wasm.Option[*instance]) {
	t.Helper()
	testFunc(t, opts, func(this *instance, ctx context.Context, module api.Module) {
		t.Helper()
		assertEqual(t, want, wasmtest.Call[R](wasm.F1(f), ctx, module, this, arg))
	})
}

func testFunc2[R value[R], T1 value[T1], T2 value[T2]](t *testing.T, want R, arg1 T1, arg2 T2, f func(*instance, context.Context, T1, T2) R, opts ...wasm.Option[*instance]) {
	t.Helper()
	testFunc(t, opts, func(this *instance, ctx context.Context, module api.Module) {
		t.Helper()
		assertEqual(t, want, wasmtest.Call[R](wasm.F2(f), ctx, module, this, arg1, arg2))
	})
}

func assertEqual(t *testing.T, want, got any) {
	t.Helper()
	if !reflect.DeepEqual(want, got) {
		t.Errorf("result mismatch: want=%+v got=%+v", want, got)
	}
}
