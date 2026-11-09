package monad

import "fmt"

// Result represents the outcome of a process: a success, optionally holding a
// value of type T, or a failure holding an error.
//
// A Result is a failure whenever it holds an error, and any value it was
// created with is then hidden. A Result with neither a value nor an error is a
// valid, empty success.
type Result[T any] struct {
	value *T
	err   error
}

// ResultOf returns a Result holding res and err. If err is not nil the Result
// is a failure and res is hidden.
func ResultOf[T any](res *T, err error) (result Result[T]) {
	return Result[T]{value: res, err: err}
}

// String returns the error message if the Result holds an error. Otherwise, it
// returns the value formatted with fmt.Sprint, or "<nil>" if there is no value.
func (r Result[T]) String() (str string) {
	if r.err != nil {
		return r.err.Error()
	}
	if r.value == nil {
		return "<nil>"
	}
	return fmt.Sprint(*r.value)
}

// Map applies f to the held value and returns a Result with the result.
// If the Result holds an error, f is not called and a Result with that error is
// returned. If the Result has no value, f is not called and an empty success is
// returned.
func (r Result[T]) Map[U any](f func(T) U) (result Result[U]) {
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
// If the Result holds an error, f is not called and a Result with that error is
// returned. If the Result has no value, f is not called and an empty success is
// returned.
func (r Result[T]) FlatMap[U any](f func(T) Result[U]) (result Result[U]) {
	if r.err != nil {
		return Result[U]{err: r.err}
	}
	if r.value == nil {
		return Result[U]{value: nil, err: nil}
	}
	return f(*r.value)
}

// IsFailure returns true if there is an error
func (r Result[T]) IsFailure() (isFailure bool) {
	return r.err != nil
}

// IsSuccess returns true if there is no error
func (r Result[T]) IsSuccess() (isSuccess bool) {
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

// GetOrDefault returns the value if it exists and there is no error,
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

// ToOption returns an Option holding the same pointer as the Result, or an
// empty Option if the Result is a failure or has no value.
func (r Result[T]) ToOption() (option Option[T]) {
	if r.err != nil || r.value == nil {
		return OptionOf[T](nil)
	}
	return OptionOf(r.value)
}
