# wasm-go

This package is a library of generic types intended to help create WebAssembly
host modules for [wazero](https://github.com/tetratelabs/wazero).

## Motivation

WebAssembly imports provide powerful features to express dependencies between
modules. A module can invoke functions of another module by declaring imports
which are mapped to exports of another module. Programs using wazero can create
such modules entirely in Go to provide extensions built into the host: those are
called *host modules*.

When defining host modules, the Go program declares the list of exported
functions using one of these two APIs of the [`wazero.HostFunctionBuilder`](https://pkg.go.dev/github.com/tetratelabs/wazero#HostFunctionBuilder):

```go
// WithGoModuleFunction is an advanced feature for those who need higher
// performance than WithFunc at the cost of more complexity.
//
// Here's an example addition function that loads operands from memory:
//
//	builder.WithGoModuleFunction(api.GoModuleFunc(func(ctx context.Context, mod api.Module, params []uint64) []uint64 {
//		mem := m.Memory()
//		offset := uint32(params[0])
//
//		x, _ := mem.ReadUint32Le(ctx, offset)
//		y, _ := mem.ReadUint32Le(ctx, offset + 4) // 32 bits == 4 bytes!
//		sum := x + y
//
//		return []uint64{sum}
//	}, []api.ValueType{api.ValueTypeI32, api.ValueTypeI32}, []api.ValueType{api.ValueTypeI32})
//
// As you can see above, defining in this way implies knowledge of which
// WebAssembly api.ValueType is appropriate for each parameter and result.
//
// ...
//
WithGoModuleFunction(fn api.GoModuleFunction, params, results []api.ValueType) HostFunctionBuilder
```

```go
// WithFunc uses reflect.Value to map a go `func` to a WebAssembly
// compatible Signature. An input that isn't a `func` will fail to
// instantiate.
//
// Here's an example of an addition function:
//
//	builder.WithFunc(func(cxt context.Context, x, y uint32) uint32 {
//		return x + y
//	})
//
// ...
//
WithFunc(interface{}) HostFunctionBuilder
```

The first is a low level API which offers the highest performance but also comes
with usability challenges. The user needs to properly map the stack state to
function parameters and return values, as well as declare the correspondingg
function signature, _manually_ doing the mapping between Go and WebAssembly
types.

The second is a higher level API that most developers should probably prefer to
use. However, it comes with limitations, both in terms of performance due to the
use of reflection, but also usability since the parameters can only be primitive
integer or floating point types:

```go
// Except for the context.Context and optional api.Module, all parameters
// or result types must map to WebAssembly numeric value types. This means
// uint32, int32, uint64, int64, float32 or float64.
```

At [Stealth Rocket](https://github.com/stealthrocket), we leverage wazero as
a core WebAssembly runtime, that we extend with host modules to enhance the
capabilities of the WebAssembly programs. We needed to improve safety and
usability of our host modules, while maintaining the performance overhead to
a minimum. We wanted to test the hypothesis that Go generics could be used to
achieve these goals, and this repository is the outcome of that experiment.

## Usage

This package is intended to be used as a library to create host modules for
wazero.

### Building Host Modules

### Declaring Host Functions

### Composite Types in Function Signatures

### Context Propagation

## Contributing

No software is ever complete, and while there will be porbably be additions and
fixes brought to the library, it is dependable in its current state.

Pull requests are welcome! Anything that is not a simple fix would probably
benefit from being discussed in an issue first.

Remember to be respectful and open minded!
