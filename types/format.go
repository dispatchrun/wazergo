package types

import (
	"fmt"
	"io"
	"reflect"
)

// Formatter is an interface used to customize the behavior of Format.
type Formatter interface {
	Format(io.Writer)
}

// Format is a helper function which can be used in the implementation of the
// FormatValue and FormatObject methods of Value and Object[T].
//
// The code below represents a common usage pattern:
//
//	func (v T) FormatObject(w io.Writer, memory api.Memory, object []byte) {
//		types.Format(w, v.LoadObject(memory, object))
//	}
//
// If T is a struct type, the output is wrapped in "{...}" and the struct fields
// are iterated and printed as comma-separated "name:value" pairs. The name may
// be customized by defining a "name" struct field tag such as:
//
//	type T struct {
//		Field int32 `name:"field"`
//	}
//
// If any of the values impelement the Formatter interface, formatting is
// delegated to the Format method.
//
// The implementation of Format has to use reflection, so it may not be best
// suited to use in contexts where performance is critical, in which cases the
// program is better off providing a custom implementation of the method.
func Format(w io.Writer, v any) { format(w, reflect.ValueOf(v)) }

var formatterInterface = reflect.TypeOf((*Formatter)(nil)).Elem()

func format(w io.Writer, v reflect.Value) {
	// TODO: to improve performance we could generate the formatters once and
	// keep track of them in a cache (e.g. similar to what encoding/json does).
	t := v.Type()
	if t.Implements(formatterInterface) {
		v.Interface().(Formatter).Format(w)
		return
	}
	switch t.Kind() {
	case reflect.Bool:
		formatBool(w, v.Bool())
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
		formatInt(w, v.Int())
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint, reflect.Uintptr:
		formatUint(w, v.Uint())
	case reflect.Float32, reflect.Float64:
		formatFloat(w, v.Float())
	case reflect.String:
		formatString(w, v.String())
	case reflect.Array:
		formatArray(w, v)
	case reflect.Slice:
		if v.Type().Elem().Kind() == reflect.Uint8 {
			formatBytes(w, v.Bytes())
		} else {
			formatArray(w, v)
		}
	case reflect.Struct:
		formatStruct(w, v)
	case reflect.Pointer:
		formatPointer(w, v)
	default:
		formatUnsupported(w, v)
	}
}

func formatBool(w io.Writer, v bool) {
	fmt.Fprintf(w, "%t", v)
}

func formatInt(w io.Writer, v int64) {
	fmt.Fprintf(w, "%d", v)
}

func formatUint(w io.Writer, v uint64) {
	fmt.Fprintf(w, "%d", v)
}

func formatFloat(w io.Writer, v float64) {
	fmt.Fprintf(w, "%g", v)
}

func formatString(w io.Writer, v string) {
	fmt.Fprintf(w, "%q", v)
}

func formatBytes(w io.Writer, v []byte) {
	Bytes(v).Format(w)
}

func formatArray(w io.Writer, v reflect.Value) {
	io.WriteString(w, "[")
	for i, n := 0, v.Len(); i < n; i++ {
		if i != 0 {
			io.WriteString(w, ",")
		}
		format(w, v.Index(i))
	}
	io.WriteString(w, "]")
}

func formatStruct(w io.Writer, v reflect.Value) {
	io.WriteString(w, "{")
	t := v.Type()
	for i, f := range reflect.VisibleFields(t) {
		if i != 0 {
			io.WriteString(w, ",")
		}
		name := f.Tag.Get("name")
		if name == "" {
			name = f.Name
		}
		io.WriteString(w, name)
		io.WriteString(w, ":")
		format(w, v.FieldByIndex(f.Index))
	}
	io.WriteString(w, "}")
}

func formatPointer(w io.Writer, v reflect.Value) {
	if v.IsNil() {
		io.WriteString(w, "<nil>")
	} else {
		format(w, v.Elem())
	}
}

func formatUnsupported(w io.Writer, v reflect.Value) {
	fmt.Fprintf(w, "<%s>", v.Type().Name())
}
