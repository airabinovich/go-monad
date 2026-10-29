// Package monad provides basic functional programming monads for Go.
package monad

import "fmt"

// Option represents an optional value: it either holds a value of type T
// or is empty. The zero value of Option is empty.
type Option[T comparable] struct {
	value *T
}

// OptionOf returns an Option holding the value pointed to by elem,
// or an empty Option if elem is nil.
func OptionOf[T comparable](elem *T) (option Option[T]) {
	return Option[T]{value: elem}
}

// String returns a string representation of the Option depending on its content
func (o Option[T]) String() string {
	if o.value == nil {
		return "<nil>"
	}
	return fmt.Sprint(*o.value)
}

// Map applies f to the held value and returns an Option with the result.
// If the Option is empty, f is not called and an empty Option is returned.
func (o Option[T]) Map[U comparable](f func(T) U) (mapped Option[U]) {
	if o.value == nil {
		return Option[U]{}
	}
	result := f(*o.value)
	return Option[U]{value: &result}
}

// FlatMap applies f to the held value and returns the Option produced by f.
// If the Option is empty, f is not called and an empty Option is returned.
func (o Option[T]) FlatMap[U comparable](f func(T) Option[U]) (mapped Option[U]) {
	if o.value == nil {
		return Option[U]{}
	}
	return f(*o.value)
}

// Filter returns the Option unchanged if it holds a value that satisfies
// predicate, or an empty Option otherwise. If the Option is empty,
// predicate is not called.
func (o Option[T]) Filter(predicate func(T) bool) (filtered Option[T]) {
	if o.value == nil || !predicate(*o.value) {
		return Option[T]{}
	}
	return o
}

// IsPresent reports whether the Option holds a value.
func (o Option[T]) IsPresent() (present bool) {
	return o.value != nil
}

// GetOrNil returns a pointer to the held value, or nil if the Option is empty.
func (o Option[T]) GetOrNil() (value *T) {
	return o.value
}

// GetOrDefault returns the held value, or other if the Option is empty.
func (o Option[T]) GetOrDefault(other T) (value T) {
	if o.value == nil {
		return other
	}
	return *o.value
}

// ToList returns a List holding a copy of the held value, or an empty List
// if the Option is empty.
func (o Option[T]) ToList() (list List[T]) {
	if o.value == nil {
		return ListOf[T]()
	}
	return ListOf(*o.value)
}

// ToSequence returns a Sequence of a copy of the held value, or an empty
// Sequence if the Option is empty.
func (o Option[T]) ToSequence() (sequence Sequence[T]) {
	if o.value == nil {
		return SequenceOf[T]()
	}
	return SequenceOf(*o.value)
}

// ToResult returns a Result with the held value if it exists,
// or with an error provided by ifEmpty function if the Option is empty.
func (o Option[T]) ToResult(ifEmpty func() error) Result[T] {
	if o.value == nil {
		return ResultOf[T](nil, ifEmpty())
	}
	return ResultOf[T](o.value, nil)
}
