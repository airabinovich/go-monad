package monad

import (
	"fmt"
	"strings"
)

// List is an ordered sequence of values of type T. It is a plain slice,
// so it adds no overhead and supports len, range and indexing directly.
// T is comparable so that a List can be turned into a Set.
// The zero value of List is empty. Every function and method returns a
// non-nil List, even when it is empty, so calls can always be chained.
// List is not safe for concurrent use.
type List[T comparable] []T

// ListOf returns a List holding elems, in order, or an empty List if no
// elems are given. When called as ListOf(s...), the List shares s's
// backing array instead of copying it.
func ListOf[T comparable](elems ...T) (list List[T]) {
	return ListOfSlice(elems)
}

// ListOfSlice returns a List holding the elements of elems, in order, or an
// empty List if elems is empty or nil. The List shares elems's backing
// array instead of copying it.
func ListOfSlice[T comparable](elems []T) (list List[T]) {
	if elems == nil {
		return List[T]{}
	}
	return List[T](elems)
}

// Len returns the length of this list
func (l List[T]) Len() int {
	return len(l)
}

// String returns a string representation of this List
func (l List[T]) String() string {
	builder := strings.Builder{}
	builder.WriteRune('[')
	length := l.Len()
	for i, elem := range l {
		builder.WriteString(fmt.Sprintf("%v", elem))
		if i < length-1 {
			builder.WriteRune(',')
		}
	}
	builder.WriteRune(']')
	return builder.String()
}

// Map applies f to each element and returns a List with the results, in
// order. If the List is empty, f is not called and an empty List is returned.
func (l List[T]) Map[U comparable](f func(T) U) (mapped List[U]) {
	mapped = make(List[U], len(l))
	for i, elem := range l {
		mapped[i] = f(elem)
	}
	return mapped
}

// FlatMap applies f to each element and returns the concatenation of the
// Lists produced by f, in order. If the List is empty, f is not called and
// an empty List is returned.
func (l List[T]) FlatMap[U comparable](f func(T) List[U]) (mapped List[U]) {
	mapped = List[U]{}
	for _, elem := range l {
		mapped = append(mapped, f(elem)...)
	}
	return mapped
}

// Filter returns a new List with the elements that satisfy predicate, in
// order, or an empty List if none do. If the List is empty, predicate is
// not called.
func (l List[T]) Filter(predicate func(T) bool) (filtered List[T]) {
	filtered = List[T]{}
	for _, elem := range l {
		if predicate(elem) {
			filtered = append(filtered, elem)
		}
	}
	return filtered
}

// IsEmpty reports whether the List has no elements.
func (l List[T]) IsEmpty() (empty bool) {
	return len(l) == 0
}

// First returns an Option holding a pointer to the first element,
// or an empty Option if the List is empty. The pointer holds the
// real element, not a copy.
func (l List[T]) First() (first Option[T]) {
	if len(l) == 0 {
		return OptionOf[T](nil)
	}
	return OptionOf(&l[0])
}

// FirstOrNil returns a pointer to the first element in the List,
// or nil if the List is empty. The pointer holds the
// real element, not a copy
func (l List[T]) FirstOrNil() (first *T) {
	if len(l) == 0 {
		return nil
	}
	return &l[0]
}

// ToSet returns a Set holding the distinct elements of the List, or an
// empty Set if the List is empty. Duplicates are removed by using the
// elements as map keys. It panics if an element is an interface value
// whose dynamic type is not comparable.
func (l List[T]) ToSet() (set Set[T]) {
	return SetOfSlice(l)
}

// ToSequence returns a Sequence of the elements of the List, in order, or an
// empty Sequence if the List is empty. The Sequence reads the List's backing
// array each time it is evaluated instead of copying it.
func (l List[T]) ToSequence() (sequence Sequence[T]) {
	return SequenceOfSlice(l)
}
