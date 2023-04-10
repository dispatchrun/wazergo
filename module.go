package wazergo

import (
	"context"

	. "github.com/stealthrocket/wazergo/types"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
)

// Module is a type constraint used to validate that all module instances
// created from wazero host modules abide to the same set of requirements.
type Module interface{ api.Closer }

// HostModule is an interface representing type-safe wazero host modules.
// The interface is parametrized on the module type that it instantiates.
//
// HostModule instances are expected to be immutable and therfore safe to use
// concurrently from multiple goroutines.
type HostModule[T Module] interface {
	// Returns the name of the host module (e.g. "wasi_snapshot_preview1").
	Name() string
	// Returns the collection of functions exported by the host module.
	// The method may return the same value across multiple calls to this
	// method, the program is expected to treat it as a read-only value.
	Functions() Functions[T]
	// Creates a new instance of the host module type, using the list of options
	// passed as arguments to configure it. This method is intended to be called
	// automatically when instantiating a module via an instantiation context.
	Instantiate(...Option[T]) T
}

// Build builds the host module p in the wazero runtime r, returning the
// instance of HostModuleBuilder that was created. This is a low level function
// which is only exposed for certain advanced use cases where a program might
// not be able to leverage Compile/Instantiate, most application should not need
// to use this function.
func Build[T Module](runtime wazero.Runtime, mod HostModule[T], decorators ...Decorator[T]) wazero.HostModuleBuilder {
	moduleName := mod.Name()
	builder := runtime.NewHostModuleBuilder(moduleName)

	for export, fn := range mod.Functions() {
		if fn.Name == "" {
			fn.Name = export
		}

		for _, d := range decorators {
			fn = d.Decorate(moduleName, fn)
		}

		paramTypes := concatValueTypes(fn.Params)
		resultTypes := concatValueTypes(fn.Results)

		builder.NewFunctionBuilder().
			WithGoModuleFunction(bind(fn.Func), paramTypes, resultTypes).
			WithName(fn.Name).
			Export(export)
	}

	return builder
}

func concatValueTypes(values []Value) []api.ValueType {
	numValueTypes := 0
	for _, v := range values {
		numValueTypes += len(v.ValueTypes())
	}
	valueTypes := make([]api.ValueType, 0, numValueTypes)
	for _, v := range values {
		valueTypes = append(valueTypes, v.ValueTypes()...)
	}
	return valueTypes
}

func bind[T Module](f func(T, context.Context, api.Module, []uint64)) api.GoModuleFunction {
	return contextualizedGoModuleFunction[T](f)
}

type contextualizedGoModuleFunction[T Module] func(T, context.Context, api.Module, []uint64)

func (f contextualizedGoModuleFunction[T]) Call(ctx context.Context, module api.Module, stack []uint64) {
	modules := ctx.Value(modulesKey{}).(modules)
	this := modules[contextKey[T]{}].(T)
	f(this, ctx, module, stack)
}

// CompiledModule represents a compiled version of a wazero host module.
type CompiledModule[T Module] struct {
	HostModule HostModule[T]
	wazero.CompiledModule
}

// Compile compiles a wazero host module within the given context.
func Compile[T Module](ctx context.Context, runtime wazero.Runtime, mod HostModule[T], decorators ...Decorator[T]) (*CompiledModule[T], error) {
	compiledModule, err := Build(runtime, mod, decorators...).Compile(ctx)
	if err != nil {
		return nil, err
	}
	return &CompiledModule[T]{mod, compiledModule}, nil
}

// InstantiationContext is a type carrying the state of instantiated wazero
// host modules. This context must be used to create call contexts to invoke
// exported functions of WebAssembly modules (see NewCallContext).
type InstantiationContext struct {
	context context.Context
	runtime wazero.Runtime
	modules modules
}

// NewInstantiationContext creates a new wazero host module instantiation
// context.
func NewInstantiationContext(ctx context.Context, rt wazero.Runtime) *InstantiationContext {
	return &InstantiationContext{
		context: ctx,
		runtime: rt,
		modules: make(modules),
	}
}

// Close closes the instantiation context, making it unusable to the program.
//
// Closing the context alos closes all modules that were instantiated from it
// and implement the io.Closer interface.
func (ins *InstantiationContext) Close(ctx context.Context) error {
	for _, mod := range ins.modules {
		mod.Close(ctx)
	}
	ins.context = nil
	ins.runtime = nil
	ins.modules = nil
	return nil
}

// Instantiate creates an module instance for the given compiled wazero host
// module. The list of options is used to pass configuration to the module
// instance.
//
// The function returns the wazero module instance that was created from the
// underlying compiled module. The returned module is bound to the instantiation
// context. If the module is closed, its state is automatically removed from the
// parent context, as well as removed from the parent wazero runtime like any
// other module instance closed by the application.
func Instantiate[T Module](ctx *InstantiationContext, compiled *CompiledModule[T], opts ...Option[T]) (api.Module, error) {
	instance := compiled.HostModule.Instantiate(opts...)
	ctx.modules[contextKey[T]{}] = instance
	callContext := NewCallContext(ctx.context, ctx)
	module, err := ctx.runtime.InstantiateModule(callContext, compiled.CompiledModule, wazero.NewModuleConfig().
		WithStartFunctions(), // TODO: is it OK not to run _start for library-style modules?
	)
	if err != nil {
		return nil, err
	}
	return &moduleInstance[T]{module, instance, ctx.modules}, nil
}

type contextKey[T any] struct{}

type modules map[any]api.Closer

type modulesKey struct{}

type moduleInstance[T Module] struct {
	api.Module
	instance T
	modules  modules
}

func (m *moduleInstance[T]) close(ctx context.Context) {
	delete(m.modules, contextKey[T]{})
	m.modules = nil
	m.instance.Close(ctx)
}

func (m *moduleInstance[T]) Close(ctx context.Context) error {
	defer m.close(ctx)
	return m.Module.Close(ctx)
}

func (m *moduleInstance[T]) CloseWithExitCode(ctx context.Context, exitCode uint32) error {
	defer m.close(ctx)
	return m.Module.CloseWithExitCode(ctx, exitCode)
}

// NewCallContext returns a Go context inheriting from ctx and containing the
// state needed for module instantiated from wazero host module to properly bind
// their methods to their receiver (e.g. the module instance).
//
// Use this function when calling methods of an instantiated WebAssenbly module
// which may invoke exported functions of a wazero host module, for example:
//
//	// The program first creates the instantiation context and uses it to
//	// instantiate compiled host module (not shown here).
//	instiation := wazergo.NewInstantiationContext(...)
//
//	...
//
//	// In this example the parent is the background context, but it might be any
//	// other Go context relevant to the application.
//	ctx = wazergo.NewCallContext(context.Background(), instantiation)
//
//	start := module.ExportedFunction("_start")
//	r, err := start.Call(ctx)
//	if err != nil {
//		...
//	}
func NewCallContext(ctx context.Context, ins *InstantiationContext) context.Context {
	return context.WithValue(ctx, modulesKey{}, ins.modules)
}

// WithCallContext returns a Go context inheriting from ctx and containig the
// necessary state to be used in calls to exported functions of the given wazero
// host modul. This function is rarely used by applications, it is often more
// useful in tests to setup the test state without constructing the entire
// compilation and instantiation contexts (see NewCallContext instead).
func WithCallContext[T Module](ctx context.Context, mod HostModule[T], opts ...Option[T]) (context.Context, func()) {
	prev, _ := ctx.Value(modulesKey{}).(modules)
	next := make(modules, len(prev)+1)
	for k, v := range prev {
		next[k] = v
	}
	instance := mod.Instantiate(opts...)
	next[contextKey[T]{}] = instance
	return context.WithValue(ctx, modulesKey{}, next), func() { instance.Close(ctx) }
}
