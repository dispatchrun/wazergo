# wazergo

This package is a library of generic types intended to help create WebAssembly
host modules for [wazero](https://github.com/tetratelabs/wazero).

## Motivation

WebAssembly imports provide powerful features to express dependencies between
modules. A module can invoke functions of another module by declaring imports
which are mapped to exports of another module. Programs using wazero can create
such modules entirely in Go to provide extensions built into the host: those are
called *host modules*.

When defining host modules, the Go program declares the list of exported
functions using one of these two APIs of the
[`wazero.HostFunctionBuilder`](https://pkg.go.dev/github.com/tetratelabs/wazero#HostFunctionBuilder):

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
capabilities of the WebAssembly programs. We wanted to improve safety and
maintainability of our host modules, while maintaining the performance overhead
to a minimum. We wanted to test the hypothesis that Go generics could be used to
achieve these goals, and this repository is the outcome of that experiment.

## Usage

This package is intended to be used as a library to create host modules for
wazero. The code is separated in two packages: the top level [`wazergo`][wazergo]
package contains the type and functions used to build host modules, including
the declaration of functions they export. The [`types`][types] subpackage
contains the declaration of generic types representing integers, floats,
pointers, arrays, etc...

Programs using the [`types`][types] package often import its symbols directly
into their package name namespace(s), which helps declare the host module
functions. For example:

```go
import (
    . "github.com/stealthrocket/wazergo/types"
)

...

// Answer returns a Int32 declared in the types package.
func (m *Module) Answer(ctx context.Context) Int32 {
    return 42
}
```

### Building Host Modules

To construct a host module, the program must declare a type satisfying the
[`Module`][Module] interface, and construct a [`HostModule[T]`][HostModule]
of that type, along with the list of its exported functions. The following model
is often useful:

```go
package my_host_module

import (
    "github.com/stealthrocket/wazergo"
)

// Declare the host module from a set of exported functions.
var HostModule waszero.HostModule[*Module] = functions{
    ...
}

// The `functions` type impements `HostModule[*Module]`, providing the
// module name, map of exported functions, and the ability to create instances
// of the module type.
type functions waszergo.Functions[*Module]

func (f functions) Name() string {
    return "my_host_module"
}

func (f functions) Functions() waszergo.Functions[*Module] {
    return (wasm.Functions[*Module])(f)
}

func (f functions) Instantiate(opts ...Option) *Module {
    return NewModule(opts...)
}

type Option = wazergo.Option[*Module]

// Module will be the Go type we use to maintain the state of our module
// instances.
type Module struct {
    ...
}

func NewModule(opts ...Option) *Module {
    ...
}
```

There are a few concepts of the library that we are getting exposed to in this
example:

- [`HostModule[T]`][HostModule] is an interface parametrized on the type
  of our module instances. This interface is the bridge between the library and
  the wazero APIs.

- [`Functions[T]`][Functions] is a map type parametrized on the module
  type, it associates the exported function names to the method of the module
  type that will be invoked when WebAssembly programs invoke them as imported
  symbols.

- [`Optional[T]`][Optional] is an interface type parameterized on the
  module type and representing the configuration options available on the
  module. It is common for the package to declare options using function
  constructors, for example:

  ```go
  func CustomValue(value int) Option {
    return wazergo.OptionFunc(func(m *Module) { ... })
  }
  ```

These types are helpers to glue the Go type where the host module is implemented
(`Module` in our example) to the generic abstractions provided by the library to
drive configuration and instantiation of the modules in wazero.

### Declaring Host Functions

The declaration of host functions is done by constructing a map of exported
names to methods of the module type, and is where the `types` subpackage can be
employed to define parameters and return values.

```go
package my_host_module

import (
    . "github.com/stealthrocket/wazergo"
    . "github.com/stealthrocket/wazergo/types"
)

var HostModule HostModule[*Module] = functions{
    "answer": F0((*Module).Answer),
    "double": F1((*Module).Double),
}

...

func (m *Module) Answer(ctx context.Context) Int32 {
    return 42
}

func (m *Module) Double(ctx context.Context, f Float32) Float32 {
    return f + f
}
```

- Exported methods of a host module must always start with a
  [`context.Context`][context] parameter.

- The parameters and return values must satisfy [`Param[T]`][Param] and
  [`Result`][Result] interfaces. The [`types`][types] subpackage contains types
  that do, but the application can construct its own for more advanced use
  cases (e.g. struct types).

- When constructing the [`Functions[T]`][Functions] map, the program
  must use one of the [`F{n}`][F] generics constructors to create a
  [`Function[T]`][Function] value from methods of the module.
  The program must use a function constructor matching the number of parameter
  to the method (e.g. [`F2`][F2] if there are two parameters, not including the
  context). The function constructors handle the conversion of Go function
  signatures to WebAssembly function types using information about their generic
  type parameters.

- Methods of the module must have a single return value. For the common case of
  having to return either a value or an error (in which case the WebAssembly
  function has two results), the generic [`Optional[T]`][Optional]
  type can be used, or the application may declare its own result types.

### Composite Parameter Types

[`Array[T]`][Array] type is base generic type used to represent
contiguous sequences of fixed-length primitive values such as integers and
floats. Array values map to a pair of `i32` values for the memory offset and
number of elements in the array. For example, the [`Bytes`][Bytes] type
(equivalent to a Go `[]byte`) is expressed as `Array[byte]`.

[`Param[T]`][Param] and [`Result`][Result] are the interfaces used
as type constraints in generic type paramaeters

To express sequences of non-primitive types, the generic [`List[T]`][List]
type can represent lists of types implementing the [`Object[T]`][Object]
interface. [`Object[T]`][Object] is used by types that can be loaded from,
or stored to the module memory.

### Memory Safety

Memory safety is guaranteed both by the use of wazero's `Memory` type, and
triggering a panic with a value of type [`SEGFAULT`][SEGFAULT] if the program
attempts to access a memory address outside of its own linear memory.

The panic effectively interrupts the program flow at the call site of the host
function, and is turned into an error by wazero so the host application can
safely handle the module termination.

### Type Safety

Type safety is guaranteed by the package at multiple levels.

Due to the use of generics, the compiler is able to verify that the host module
constructed by the program is semantically correct. For example, the compiler
will refuse to create a host function where one of the return value is a type
which does not implement the [`Result`][Result] interface.

Runtime validation is then added by wazero when mapping module imports to ensure
that the low level WebAssembly signatures of the imports match with those of the
host module.

### Context Propagation

Calls to the host functions of a module require injecting the context in which
the host module was instantiated into the context in which the exported functions
of a module instante that depend on it are called (e.g. binding of the method
receiver to the calls to carry state across invocations).

This is currently done using an [`InstantiationContext`][InstantiationContext],
which acts as a container for instantiated host modules. The following example
shows how to leverage it:

```go
instantiation := wazergo.NewInstantiationContext(ctx, runtime)
...
// When invoking exported functions of a module; this may also be done
// automatically via calls to wazero.Runtime.InstantiateModule which
// invoke the start function(s).
ctx = wasm.NewCallContext(ctx, instantiation)

start := module.ExportedFunction("_start")
r, err := start.Call(ctx)
if err != nil {
	...
}
```

_Note: this part of the wazergo package is the most experiemental and might be
changed based on user feedback and experience using the library. Any changes will
aim at simplifying use the package._

## Contributing

No software is ever complete, and while there will be porbably be additions and
fixes brought to the library, it is usable in its current state, and while we
aim to maintain backward compatibility, breaking changes might be introduced if
necessary to improve usability as we learn more from using the library.

Pull requests are welcome! Anything that is not a simple fix would probably
benefit from being discussed in an issue first.

Remember to be respectful and open minded!

[context]: https://pkg.go.dev/context#Context
[F]: https://pkg.go.dev/github.com/stealthrocket/wazergo#F0
[F2]: https://pkg.go.dev/github.com/stealthrocket/wazergo#F2
[Function]: https://pkg.go.dev/github.com/stealthrocket/wazergo#Function
[Functions]: https://pkg.go.dev/github.com/stealthrocket/wazergo#Functions
[HostModule]: https://pkg.go.dev/github.com/stealthrocket/wazergo#HostModule
[Module]: https://pkg.go.dev/github.com/stealthrocket/wazergo#Module
[Optional]: https://pkg.go.dev/github.com/stealthrocket/wazergo/types#Optional
[Array]: https://pkg.go.dev/github.com/stealthrocket/wazergo/types#Array
[Bytes]: https://pkg.go.dev/github.com/stealthrocket/wazergo/types#Bytes
[Object]: https://pkg.go.dev/github.com/stealthrocket/wazergo/types#Object
[Param]: https://pkg.go.dev/github.com/stealthrocket/wazergo/types#Param
[types]: https://pkg.go.dev/github.com/stealthrocket/wazergo/types
[wazergo]: https://pkg.go.dev/github.com/stealthrocket/wazergo
[SEGFAULT]: https://pkg.go.dev/github.com/stealthrocket/wazergo#SEGFAULT
[InstantiationContext]: https://pkg.go.dev/github.com/stealthrocket/wazergo#InstantiationContext
