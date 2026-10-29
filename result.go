package monad

import "fmt"

// Result represents the outcome of a process that can either return a
// value of type T or an error
type Result[T comparable] struct {
	value *T
	err   error
}

// ResultOf returns a Result holding the provided value and error
func ResultOf[T comparable](res *T, err error) (result Result[T]) {
	return Result[T]{value: res, err: err}
}

// String returns a string representation of the Result.
// The value's string representation if present
// or the error in case there is one
func (r Result[T]) String() string {
	if r.err != nil {
		return r.err.Error()
	}
	return fmt.Sprintf("%v", *r.value)
}

// Map applies f to the held value and returns a Result with the result.
// If the Result hold an error, f is not called and a Result with an error is returned.
func (r Result[T]) Map[U comparable](f func(T) U) (res Result[U]) {
	if r.err != nil {
		return Result[U]{err: r.err}
	}
	if r.value == nil {
		return Result[U]{value: nil, err: nil}
	}
	u := f(*r.value)
	return Result[U]{
		value: &u,
		err:   nil,
	}
}

// FlatMap applies f to the held value and returns the Result produced by f.
// If the Result holds an error, f is not called and a Result with an error is returned.
func (r Result[T]) FlatMap[U comparable](f func(T) Result[U]) (res Result[U]) {
	if r.err != nil {
		return Result[U]{err: r.err}
	}
	if r.value == nil {
		return Result[U]{value: nil, err: nil}
	}
	return f(*r.value)
}

// IsFailure returns true if there is an error
func (r Result[T]) IsFailure() bool {
	return r.err != nil
}

// IsSuccess returns true if there is no error
func (r Result[T]) IsSuccess() bool {
	return r.err == nil
}

// GetOrNil returns a pointer to the value if it exists and there is no error,
// or nil otherwise
func (r Result[T]) GetOrNil() (val *T) {
	if r.value == nil || r.err != nil {
		return nil
	}
	return r.value
}

// GetOrDefault returns the value if it exists and there ir no error,
// or the default value otherwise
func (r Result[T]) GetOrDefault(def T) (val T) {
	if r.value == nil || r.err != nil {
		return def
	}
	return *r.value
}

// Error returns the held error if it exists
func (r Result[T]) Error() (err error) {
	return r.err
}

// ToOption returns an Option holding the held value if it exists
func (r Result[T]) ToOption() (option Option[T]) {
	if r.err != nil || r.value == nil {
		return OptionOf[T](nil)
	}
	return OptionOf(r.value)
}
