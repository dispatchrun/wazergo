package wazergo

import (
	"context"

	. "github.com/stealthrocket/wazergo/types"
	"github.com/tetratelabs/wazero/api"
)

// Functions is a map type representing the collection of functions exported
// by a plugin. The map keys are the names of that each function gets exported
// as. The function value is the description of the wazero host function to
// be added when building a plugin. The type parameter T is used to ensure
// consistency between the plugin definition and the functions that compose it.
type Functions[T any] map[string]Function[T]

// Function represents a single function exported by a plugin. Programs may
// configure the fields individually but it is often preferrable to use one of
// the Func* constructors instead to let the Go compiler ensure type and memory
// safety when generating the code to bridge between WebAssembly and Go.
type Function[T any] struct {
	Name    string
	Params  []Value
	Results []Value
	Func    func(T, context.Context, api.Module, []uint64)
}

// NumParams is the number of parameters this function reads from the stack.
//
// Note that this is not necessarily the same as len(f.Params), since Params
// holds higher level values that may correspond to more than one stack param.
func (f *Function[T]) NumParams() int {
	return countStackValues(f.Params)
}

// NumResults is the number of return values this function writes to the stack.
//
// Note that this is not necessarily the same as len(f.Results), since Results
// holds higher level values that may correspond to more than one stack result.
func (f *Function[T]) NumResults() int {
	return countStackValues(f.Results)
}

// WithFunc returns a copy of the Function with the internal Func field
// replaced.
func (f Function[T]) WithFunc(fn func(T, context.Context, api.Module, []uint64)) Function[T] {
	f.Func = fn
	return f
}

func countStackValues(values []Value) (count int) {
	for _, v := range values {
		count += len(v.ValueTypes())
	}
	return
}

// F0 is the Function constructor for functions accepting no parameters.
func F0[T any, R Result](fn func(T, context.Context) R) Function[T] {
	var ret R
	return Function[T]{
		Results: []Value{ret},
		Func: func(this T, ctx context.Context, module api.Module, stack []uint64) {
			fn(this, ctx).StoreValue(module.Memory(), stack)
		},
	}
}

// F1 is the Function constructor for functions accepting one parameter.
func F1[T any, P Param[P], R Result](fn func(T, context.Context, P) R) Function[T] {
	var ret R
	var arg P
	return Function[T]{
		Params:  []Value{arg},
		Results: []Value{ret},
		Func: func(this T, ctx context.Context, module api.Module, stack []uint64) {
			var arg P
			var memory = module.Memory()
			fn(this, ctx, arg.LoadValue(memory, stack)).StoreValue(memory, stack)
		},
	}
}

// F2 is the Function constructor for functions accepting two parameters.
func F2[
	T any,
	P1 Param[P1],
	P2 Param[P2],
	R Result,
](fn func(T, context.Context, P1, P2) R) Function[T] {
	var ret R
	var arg1 P1
	var arg2 P2
	params1 := arg1.ValueTypes()
	params2 := arg2.ValueTypes()
	a := len(params1)
	b := len(params2) + a
	return Function[T]{
		Params:  []Value{arg1, arg2},
		Results: []Value{ret},
		Func: func(this T, ctx context.Context, module api.Module, stack []uint64) {
			var arg1 P1
			var arg2 P2
			var memory = module.Memory()
			fn(this, ctx,
				arg1.LoadValue(memory, stack[0:a:a]),
				arg2.LoadValue(memory, stack[a:b:b]),
			).StoreValue(memory, stack)
		},
	}
}

// F3 is the Function constructor for functions accepting three parameters.
func F3[
	T any,
	P1 Param[P1],
	P2 Param[P2],
	P3 Param[P3],
	R Result,
](fn func(T, context.Context, P1, P2, P3) R) Function[T] {
	var ret R
	var arg1 P1
	var arg2 P2
	var arg3 P3
	params1 := arg1.ValueTypes()
	params2 := arg2.ValueTypes()
	params3 := arg3.ValueTypes()
	a := len(params1)
	b := len(params2) + a
	c := len(params3) + b
	return Function[T]{
		Params:  []Value{arg1, arg2, arg3},
		Results: []Value{ret},
		Func: func(this T, ctx context.Context, module api.Module, stack []uint64) {
			var arg1 P1
			var arg2 P2
			var arg3 P3
			var memory = module.Memory()
			fn(this, ctx,
				arg1.LoadValue(memory, stack[0:a:a]),
				arg2.LoadValue(memory, stack[a:b:b]),
				arg3.LoadValue(memory, stack[b:c:c]),
			).StoreValue(memory, stack)
		},
	}
}

// F4 is the Function constructor for functions accepting four parameters.
func F4[
	T any,
	P1 Param[P1],
	P2 Param[P2],
	P3 Param[P3],
	P4 Param[P4],
	R Result,
](fn func(T, context.Context, P1, P2, P3, P4) R) Function[T] {
	var ret R
	var arg1 P1
	var arg2 P2
	var arg3 P3
	var arg4 P4
	params1 := arg1.ValueTypes()
	params2 := arg2.ValueTypes()
	params3 := arg3.ValueTypes()
	params4 := arg4.ValueTypes()
	a := len(params1)
	b := len(params2) + a
	c := len(params3) + b
	d := len(params4) + c
	return Function[T]{
		Params:  []Value{arg1, arg2, arg3, arg4},
		Results: []Value{ret},
		Func: func(this T, ctx context.Context, module api.Module, stack []uint64) {
			var arg1 P1
			var arg2 P2
			var arg3 P3
			var arg4 P4
			var memory = module.Memory()
			fn(this, ctx,
				arg1.LoadValue(memory, stack[0:a:a]),
				arg2.LoadValue(memory, stack[a:b:b]),
				arg3.LoadValue(memory, stack[b:c:c]),
				arg4.LoadValue(memory, stack[c:d:d]),
			).StoreValue(memory, stack)
		},
	}
}

// F5 is the Function constructor for functions accepting five parameters.
func F5[
	T any,
	P1 Param[P1],
	P2 Param[P2],
	P3 Param[P3],
	P4 Param[P4],
	P5 Param[P5],
	R Result,
](fn func(T, context.Context, P1, P2, P3, P4, P5) R) Function[T] {
	var ret R
	var arg1 P1
	var arg2 P2
	var arg3 P3
	var arg4 P4
	var arg5 P5
	params1 := arg1.ValueTypes()
	params2 := arg2.ValueTypes()
	params3 := arg3.ValueTypes()
	params4 := arg4.ValueTypes()
	params5 := arg5.ValueTypes()
	a := len(params1)
	b := len(params2) + a
	c := len(params3) + b
	d := len(params4) + c
	e := len(params5) + d
	return Function[T]{
		Params:  []Value{arg1, arg2, arg3, arg4, arg5},
		Results: []Value{ret},
		Func: func(this T, ctx context.Context, module api.Module, stack []uint64) {
			var arg1 P1
			var arg2 P2
			var arg3 P3
			var arg4 P4
			var arg5 P5
			var memory = module.Memory()
			fn(this, ctx,
				arg1.LoadValue(memory, stack[0:a:a]),
				arg2.LoadValue(memory, stack[a:b:b]),
				arg3.LoadValue(memory, stack[b:c:c]),
				arg4.LoadValue(memory, stack[c:d:d]),
				arg5.LoadValue(memory, stack[d:e:e]),
			).StoreValue(memory, stack)
		},
	}
}

// F6 is the Function constructor for functions accepting six parameters.
func F6[
	T any,
	P1 Param[P1],
	P2 Param[P2],
	P3 Param[P3],
	P4 Param[P4],
	P5 Param[P5],
	P6 Param[P6],
	R Result,
](fn func(T, context.Context, P1, P2, P3, P4, P5, P6) R) Function[T] {
	var ret R
	var arg1 P1
	var arg2 P2
	var arg3 P3
	var arg4 P4
	var arg5 P5
	var arg6 P6
	params1 := arg1.ValueTypes()
	params2 := arg2.ValueTypes()
	params3 := arg3.ValueTypes()
	params4 := arg4.ValueTypes()
	params5 := arg5.ValueTypes()
	params6 := arg6.ValueTypes()
	a := len(params1)
	b := len(params2) + a
	c := len(params3) + b
	d := len(params4) + c
	e := len(params5) + d
	f := len(params6) + e
	return Function[T]{
		Params:  []Value{arg1, arg2, arg3, arg4, arg5, arg6},
		Results: []Value{ret},
		Func: func(this T, ctx context.Context, module api.Module, stack []uint64) {
			var arg1 P1
			var arg2 P2
			var arg3 P3
			var arg4 P4
			var arg5 P5
			var arg6 P6
			var memory = module.Memory()
			fn(this, ctx,
				arg1.LoadValue(memory, stack[0:a:a]),
				arg2.LoadValue(memory, stack[a:b:b]),
				arg3.LoadValue(memory, stack[b:c:c]),
				arg4.LoadValue(memory, stack[c:d:d]),
				arg5.LoadValue(memory, stack[d:e:e]),
				arg6.LoadValue(memory, stack[e:f:f]),
			).StoreValue(memory, stack)
		},
	}
}

// F7 is the Function constructor for functions accepting seven parameters.
func F7[
	T any,
	P1 Param[P1],
	P2 Param[P2],
	P3 Param[P3],
	P4 Param[P4],
	P5 Param[P5],
	P6 Param[P6],
	P7 Param[P7],
	R Result,
](fn func(T, context.Context, P1, P2, P3, P4, P5, P6, P7) R) Function[T] {
	var ret R
	var arg1 P1
	var arg2 P2
	var arg3 P3
	var arg4 P4
	var arg5 P5
	var arg6 P6
	var arg7 P7
	params1 := arg1.ValueTypes()
	params2 := arg2.ValueTypes()
	params3 := arg3.ValueTypes()
	params4 := arg4.ValueTypes()
	params5 := arg5.ValueTypes()
	params6 := arg6.ValueTypes()
	params7 := arg7.ValueTypes()
	a := len(params1)
	b := len(params2) + a
	c := len(params3) + b
	d := len(params4) + c
	e := len(params5) + d
	f := len(params6) + e
	g := len(params7) + f
	return Function[T]{
		Params:  []Value{arg1, arg2, arg3, arg4, arg5, arg6, arg7},
		Results: []Value{ret},
		Func: func(this T, ctx context.Context, module api.Module, stack []uint64) {
			var arg1 P1
			var arg2 P2
			var arg3 P3
			var arg4 P4
			var arg5 P5
			var arg6 P6
			var arg7 P7
			var memory = module.Memory()
			fn(this, ctx,
				arg1.LoadValue(memory, stack[0:a:a]),
				arg2.LoadValue(memory, stack[a:b:b]),
				arg3.LoadValue(memory, stack[b:c:c]),
				arg4.LoadValue(memory, stack[c:d:d]),
				arg5.LoadValue(memory, stack[d:e:e]),
				arg6.LoadValue(memory, stack[e:f:f]),
				arg7.LoadValue(memory, stack[f:g:g]),
			).StoreValue(memory, stack)
		},
	}
}

// F8 is the Function constructor for functions accepting eight parameters.
func F8[
	T any,
	P1 Param[P1],
	P2 Param[P2],
	P3 Param[P3],
	P4 Param[P4],
	P5 Param[P5],
	P6 Param[P6],
	P7 Param[P7],
	P8 Param[P8],
	R Result,
](fn func(T, context.Context, P1, P2, P3, P4, P5, P6, P7, P8) R) Function[T] {
	var ret R
	var arg1 P1
	var arg2 P2
	var arg3 P3
	var arg4 P4
	var arg5 P5
	var arg6 P6
	var arg7 P7
	var arg8 P8
	params1 := arg1.ValueTypes()
	params2 := arg2.ValueTypes()
	params3 := arg3.ValueTypes()
	params4 := arg4.ValueTypes()
	params5 := arg5.ValueTypes()
	params6 := arg6.ValueTypes()
	params7 := arg7.ValueTypes()
	params8 := arg8.ValueTypes()
	a := len(params1)
	b := len(params2) + a
	c := len(params3) + b
	d := len(params4) + c
	e := len(params5) + d
	f := len(params6) + e
	g := len(params7) + f
	h := len(params8) + g
	return Function[T]{
		Params:  []Value{arg1, arg2, arg3, arg4, arg5, arg6, arg7, arg8},
		Results: []Value{ret},
		Func: func(this T, ctx context.Context, module api.Module, stack []uint64) {
			var arg1 P1
			var arg2 P2
			var arg3 P3
			var arg4 P4
			var arg5 P5
			var arg6 P6
			var arg7 P7
			var arg8 P8
			var memory = module.Memory()
			fn(this, ctx,
				arg1.LoadValue(memory, stack[0:a:a]),
				arg2.LoadValue(memory, stack[a:b:b]),
				arg3.LoadValue(memory, stack[b:c:c]),
				arg4.LoadValue(memory, stack[c:d:d]),
				arg5.LoadValue(memory, stack[d:e:e]),
				arg6.LoadValue(memory, stack[e:f:f]),
				arg7.LoadValue(memory, stack[f:g:g]),
				arg8.LoadValue(memory, stack[g:h:h]),
			).StoreValue(memory, stack)
		},
	}
}

// F9 is the Function constructor for functions accepting nine parameters.
func F9[
	T any,
	P1 Param[P1],
	P2 Param[P2],
	P3 Param[P3],
	P4 Param[P4],
	P5 Param[P5],
	P6 Param[P6],
	P7 Param[P7],
	P8 Param[P8],
	P9 Param[P9],
	R Result,
](fn func(T, context.Context, P1, P2, P3, P4, P5, P6, P7, P8, P9) R) Function[T] {
	var ret R
	var arg1 P1
	var arg2 P2
	var arg3 P3
	var arg4 P4
	var arg5 P5
	var arg6 P6
	var arg7 P7
	var arg8 P8
	var arg9 P9
	params1 := arg1.ValueTypes()
	params2 := arg2.ValueTypes()
	params3 := arg3.ValueTypes()
	params4 := arg4.ValueTypes()
	params5 := arg5.ValueTypes()
	params6 := arg6.ValueTypes()
	params7 := arg7.ValueTypes()
	params8 := arg8.ValueTypes()
	params9 := arg9.ValueTypes()
	a := len(params1)
	b := len(params2) + a
	c := len(params3) + b
	d := len(params4) + c
	e := len(params5) + d
	f := len(params6) + e
	g := len(params7) + f
	h := len(params8) + g
	i := len(params9) + h
	return Function[T]{
		Params:  []Value{arg1, arg2, arg3, arg4, arg5, arg6, arg7, arg8, arg9},
		Results: []Value{ret},
		Func: func(this T, ctx context.Context, module api.Module, stack []uint64) {
			var arg1 P1
			var arg2 P2
			var arg3 P3
			var arg4 P4
			var arg5 P5
			var arg6 P6
			var arg7 P7
			var arg8 P8
			var arg9 P9
			var memory = module.Memory()
			fn(this, ctx,
				arg1.LoadValue(memory, stack[0:a:a]),
				arg2.LoadValue(memory, stack[a:b:b]),
				arg3.LoadValue(memory, stack[b:c:c]),
				arg4.LoadValue(memory, stack[c:d:d]),
				arg5.LoadValue(memory, stack[d:e:e]),
				arg6.LoadValue(memory, stack[e:f:f]),
				arg7.LoadValue(memory, stack[f:g:g]),
				arg8.LoadValue(memory, stack[g:h:h]),
				arg9.LoadValue(memory, stack[h:i:i]),
			).StoreValue(memory, stack)
		},
	}
}

// F10 is the Function constructor for functions accepting ten parameters.
func F10[
	T any,
	P1 Param[P1],
	P2 Param[P2],
	P3 Param[P3],
	P4 Param[P4],
	P5 Param[P5],
	P6 Param[P6],
	P7 Param[P7],
	P8 Param[P8],
	P9 Param[P9],
	P10 Param[P10],
	R Result,
](fn func(T, context.Context, P1, P2, P3, P4, P5, P6, P7, P8, P9, P10) R) Function[T] {
	var ret R
	var arg1 P1
	var arg2 P2
	var arg3 P3
	var arg4 P4
	var arg5 P5
	var arg6 P6
	var arg7 P7
	var arg8 P8
	var arg9 P9
	var arg10 P10
	params1 := arg1.ValueTypes()
	params2 := arg2.ValueTypes()
	params3 := arg3.ValueTypes()
	params4 := arg4.ValueTypes()
	params5 := arg5.ValueTypes()
	params6 := arg6.ValueTypes()
	params7 := arg7.ValueTypes()
	params8 := arg8.ValueTypes()
	params9 := arg9.ValueTypes()
	params10 := arg10.ValueTypes()
	a := len(params1)
	b := len(params2) + a
	c := len(params3) + b
	d := len(params4) + c
	e := len(params5) + d
	f := len(params6) + e
	g := len(params7) + f
	h := len(params8) + g
	i := len(params9) + h
	j := len(params10) + i
	return Function[T]{
		Params:  []Value{arg1, arg2, arg3, arg4, arg5, arg6, arg7, arg8, arg9, arg10},
		Results: []Value{ret},
		Func: func(this T, ctx context.Context, module api.Module, stack []uint64) {
			var arg1 P1
			var arg2 P2
			var arg3 P3
			var arg4 P4
			var arg5 P5
			var arg6 P6
			var arg7 P7
			var arg8 P8
			var arg9 P9
			var arg10 P10
			var memory = module.Memory()
			fn(this, ctx,
				arg1.LoadValue(memory, stack[0:a:a]),
				arg2.LoadValue(memory, stack[a:b:b]),
				arg3.LoadValue(memory, stack[b:c:c]),
				arg4.LoadValue(memory, stack[c:d:d]),
				arg5.LoadValue(memory, stack[d:e:e]),
				arg6.LoadValue(memory, stack[e:f:f]),
				arg7.LoadValue(memory, stack[f:g:g]),
				arg8.LoadValue(memory, stack[g:h:h]),
				arg9.LoadValue(memory, stack[h:i:i]),
				arg10.LoadValue(memory, stack[i:j:j]),
			).StoreValue(memory, stack)
		},
	}
}

// F11 is the Function constructor for functions accepting eleven parameters.
func F11[
	T any,
	P1 Param[P1],
	P2 Param[P2],
	P3 Param[P3],
	P4 Param[P4],
	P5 Param[P5],
	P6 Param[P6],
	P7 Param[P7],
	P8 Param[P8],
	P9 Param[P9],
	P10 Param[P10],
	P11 Param[P11],
	R Result,
](fn func(T, context.Context, P1, P2, P3, P4, P5, P6, P7, P8, P9, P10, P11) R) Function[T] {
	var ret R
	var arg1 P1
	var arg2 P2
	var arg3 P3
	var arg4 P4
	var arg5 P5
	var arg6 P6
	var arg7 P7
	var arg8 P8
	var arg9 P9
	var arg10 P10
	var arg11 P11
	params1 := arg1.ValueTypes()
	params2 := arg2.ValueTypes()
	params3 := arg3.ValueTypes()
	params4 := arg4.ValueTypes()
	params5 := arg5.ValueTypes()
	params6 := arg6.ValueTypes()
	params7 := arg7.ValueTypes()
	params8 := arg8.ValueTypes()
	params9 := arg9.ValueTypes()
	params10 := arg10.ValueTypes()
	params11 := arg11.ValueTypes()
	a := len(params1)
	b := len(params2) + a
	c := len(params3) + b
	d := len(params4) + c
	e := len(params5) + d
	f := len(params6) + e
	g := len(params7) + f
	h := len(params8) + g
	i := len(params9) + h
	j := len(params10) + i
	k := len(params11) + j
	return Function[T]{
		Params:  []Value{arg1, arg2, arg3, arg4, arg5, arg6, arg7, arg8, arg9, arg10, arg11},
		Results: []Value{ret},
		Func: func(this T, ctx context.Context, module api.Module, stack []uint64) {
			var arg1 P1
			var arg2 P2
			var arg3 P3
			var arg4 P4
			var arg5 P5
			var arg6 P6
			var arg7 P7
			var arg8 P8
			var arg9 P9
			var arg10 P10
			var arg11 P11
			var memory = module.Memory()
			fn(this, ctx,
				arg1.LoadValue(memory, stack[0:a:a]),
				arg2.LoadValue(memory, stack[a:b:b]),
				arg3.LoadValue(memory, stack[b:c:c]),
				arg4.LoadValue(memory, stack[c:d:d]),
				arg5.LoadValue(memory, stack[d:e:e]),
				arg6.LoadValue(memory, stack[e:f:f]),
				arg7.LoadValue(memory, stack[f:g:g]),
				arg8.LoadValue(memory, stack[g:h:h]),
				arg9.LoadValue(memory, stack[h:i:i]),
				arg10.LoadValue(memory, stack[i:j:j]),
				arg11.LoadValue(memory, stack[j:k:k]),
			).StoreValue(memory, stack)
		},
	}
}

// F12 is the Function constructor for functions accepting twelve parameters.
func F12[
	T any,
	P1 Param[P1],
	P2 Param[P2],
	P3 Param[P3],
	P4 Param[P4],
	P5 Param[P5],
	P6 Param[P6],
	P7 Param[P7],
	P8 Param[P8],
	P9 Param[P9],
	P10 Param[P10],
	P11 Param[P11],
	P12 Param[P12],
	R Result,
](fn func(T, context.Context, P1, P2, P3, P4, P5, P6, P7, P8, P9, P10, P11, P12) R) Function[T] {
	var ret R
	var arg1 P1
	var arg2 P2
	var arg3 P3
	var arg4 P4
	var arg5 P5
	var arg6 P6
	var arg7 P7
	var arg8 P8
	var arg9 P9
	var arg10 P10
	var arg11 P11
	var arg12 P12
	params1 := arg1.ValueTypes()
	params2 := arg2.ValueTypes()
	params3 := arg3.ValueTypes()
	params4 := arg4.ValueTypes()
	params5 := arg5.ValueTypes()
	params6 := arg6.ValueTypes()
	params7 := arg7.ValueTypes()
	params8 := arg8.ValueTypes()
	params9 := arg9.ValueTypes()
	params10 := arg10.ValueTypes()
	params11 := arg11.ValueTypes()
	params12 := arg12.ValueTypes()
	a := len(params1)
	b := len(params2) + a
	c := len(params3) + b
	d := len(params4) + c
	e := len(params5) + d
	f := len(params6) + e
	g := len(params7) + f
	h := len(params8) + g
	i := len(params9) + h
	j := len(params10) + i
	k := len(params11) + j
	l := len(params12) + k
	return Function[T]{
		Params:  []Value{arg1, arg2, arg3, arg4, arg5, arg6, arg7, arg8, arg9, arg10, arg11, arg12},
		Results: []Value{ret},
		Func: func(this T, ctx context.Context, module api.Module, stack []uint64) {
			var arg1 P1
			var arg2 P2
			var arg3 P3
			var arg4 P4
			var arg5 P5
			var arg6 P6
			var arg7 P7
			var arg8 P8
			var arg9 P9
			var arg10 P10
			var arg11 P11
			var arg12 P12
			var memory = module.Memory()
			fn(this, ctx,
				arg1.LoadValue(memory, stack[0:a:a]),
				arg2.LoadValue(memory, stack[a:b:b]),
				arg3.LoadValue(memory, stack[b:c:c]),
				arg4.LoadValue(memory, stack[c:d:d]),
				arg5.LoadValue(memory, stack[d:e:e]),
				arg6.LoadValue(memory, stack[e:f:f]),
				arg7.LoadValue(memory, stack[f:g:g]),
				arg8.LoadValue(memory, stack[g:h:h]),
				arg9.LoadValue(memory, stack[h:i:i]),
				arg10.LoadValue(memory, stack[i:j:j]),
				arg11.LoadValue(memory, stack[j:k:k]),
				arg12.LoadValue(memory, stack[k:l:l]),
			).StoreValue(memory, stack)
		},
	}
}
