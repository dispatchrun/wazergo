package wazergo_test

import (
	"context"
	"os"
	"testing"

	"github.com/stealthrocket/wazergo"
	. "github.com/stealthrocket/wazergo/types"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
)

var hostModule wazergo.HostModule[*hostInstance] = hostFunctions{
	"answer": wazergo.F0((*hostInstance).Answer),
}

type hostFunctions wazergo.Functions[*hostInstance]

func (m hostFunctions) Name() string {
	return "test"
}

func (m hostFunctions) Functions() wazergo.Functions[*hostInstance] {
	return (wazergo.Functions[*hostInstance](m))
}

func (m hostFunctions) Instantiate(ctx context.Context, opts ...wazergo.Option[*hostInstance]) (*hostInstance, error) {
	ins := new(hostInstance)
	wazergo.Configure(ins, opts...)
	return ins, nil
}

type hostInstance struct {
	answer int
}

func (m *hostInstance) Close(ctx context.Context) error {
	return nil
}

func (m *hostInstance) Answer(ctx context.Context) Int32 {
	return Int32(m.answer)
}

func answer(a int) wazergo.Option[*hostInstance] {
	return wazergo.OptionFunc(func(m *hostInstance) { m.answer = a })
}

func TestMultipleHostModuleInstances(t *testing.T) {
	ctx := context.Background()

	runtime := wazero.NewRuntime(ctx)
	defer runtime.Close(ctx)

	// three copies, all share the same host module name but different state
	instance0 := wazergo.MustInstantiate(ctx, runtime, hostModule, answer(0))
	instance1 := wazergo.MustInstantiate(ctx, runtime, hostModule, answer(21))
	instance2 := wazergo.MustInstantiate(ctx, runtime, hostModule, answer(42))

	defer instance0.Close(ctx)
	defer instance1.Close(ctx)
	defer instance2.Close(ctx)

	guest, err := loadModule(ctx, runtime, "testdata/answer.wasm")
	if err != nil {
		t.Fatal(err)
	}

	answer := guest.ExportedFunction("answer")
	r0, _ := answer.Call(wazergo.WithModuleInstance(ctx, instance0))
	r1, _ := answer.Call(wazergo.WithModuleInstance(ctx, instance1))
	r2, _ := answer.Call(wazergo.WithModuleInstance(ctx, instance2))

	for i, test := range [...]struct{ want, got int }{
		{want: 0, got: int(r0[0])},
		{want: 21, got: int(r1[0])},
		{want: 42, got: int(r2[0])},
	} {
		if test.want != test.got {
			t.Errorf("result %d is wrong: want=%d got=%d", i, test.want, test.got)
		}
	}
}

func loadModule(ctx context.Context, runtime wazero.Runtime, filePath string) (api.Module, error) {
	b, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	return runtime.Instantiate(ctx, b)
}
