package wasm

// Option is a generic interface used to represent options that apply
// configuration to a value.
type Option[T any] interface {
	// Configure is called to apply the configuration option to the value passed
	// as argument.
	Configure(T)
}

// OptionFunc is a constructor which creates an option from a function.
// This function is useful to leverage type inference and not have to repeat
// the type T in the type parameter.
func OptionFunc[T any](opt func(T)) Option[T] { return option[T](opt) }

type option[T any] func(T)

func (option option[T]) Configure(value T) { option(value) }

// Configure applies the list of options to the value passed as first argument.
func Configure[T any](value T, options ...Option[T]) {
	for _, opt := range options {
		opt.Configure(value)
	}
}
