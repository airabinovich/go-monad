package monad

import (
	"fmt"
	"iter"
	"reflect"
	"strings"
)

// List is an ordered sequence of values of type T.
// The zero value of List is empty. Every function and method returns a
// non-nil List, even when it is empty, so calls can always be chained.
// A nil *List is safe to use with every method and behaves like an empty List.
// Only Add cannot modify a nil receiver, so it returns a new List instead.
// List is not safe for concurrent use.
type List[T any] struct {
	elements []T
}

// ListOf returns a List holding elems, in order, or an empty List if no
// elems are given. When called as ListOf(s...), the List shares s's
// backing array instead of copying it.
func ListOf[T any](elems ...T) (list *List[T]) {
	return ListOfSlice(elems)
}

// ListOfSlice returns a List holding the elements of elems, in order, or an
// empty List if elems is empty or nil. The List shares elems's backing
// array instead of copying it.
func ListOfSlice[T any](elems []T) (list *List[T]) {
	if elems == nil {
		return &List[T]{}
	}
	return &List[T]{elems}
}

// slice returns the elements of the List, or nil for a nil List. It is the
// single place that makes the methods safe for a nil receiver.
func (l *List[T]) slice() []T {
	if l == nil {
		return nil
	}
	return l.elements
}

// Len returns the number of elements in the List, or 0 for a nil List.
func (l *List[T]) Len() int {
	return len(l.slice())
}

// String returns the elements formatted with %v, separated by commas and
// wrapped in square brackets, e.g. [1,2,3]. A nil List is formatted as [].
func (l *List[T]) String() string {
	builder := strings.Builder{}
	builder.WriteRune('[')
	length := l.Len()
	for i, elem := range l.slice() {
		builder.WriteString(fmt.Sprintf("%v", elem))
		if i < length-1 {
			builder.WriteRune(',')
		}
	}
	builder.WriteRune(']')
	return builder.String()
}

// Equal reports whether both Lists hold the same elements in the same order.
// Elements are compared with reflect.DeepEqual. A nil List is equal to an
// empty List.
func (l *List[T]) Equal(other *List[T]) bool {
	elems, otherElems := l.slice(), other.slice()
	if len(elems) != len(otherElems) {
		return false
	}
	for i := range elems {
		if !reflect.DeepEqual(elems[i], otherElems[i]) {
			return false
		}
	}
	return true
}

// Add appends element to the List and returns the List. If the receiver is
// nil, nothing is modified and a new List holding only element is returned.
func (l *List[T]) Add(element T) (list *List[T]) {
	if l == nil {
		return ListOf[T](element)
	}
	if l.elements == nil {
		l.elements = []T{}
	}
	l.elements = append(l.elements, element)
	return l
}

// Contains reports whether element is in the List. Elements are compared with
// reflect.DeepEqual.
func (l *List[T]) Contains(element T) bool {
	for _, elem := range l.slice() {
		if reflect.DeepEqual(elem, element) {
			return true
		}
	}
	return false
}

// Map applies f to each element and returns a List with the results, in
// order. If the List is empty, f is not called and an empty List is returned.
func (l *List[T]) Map[U any](f func(T) U) (mapped *List[U]) {
	mappedElems := make([]U, 0, l.Len())
	for _, elem := range l.slice() {
		mappedElems = append(mappedElems, f(elem))
	}
	return ListOfSlice(mappedElems)
}

// FlatMap applies f to each element and returns the concatenation of the
// Lists produced by f, in order. A nil List returned by f is treated as empty.
// If the List is empty, f is not called and an empty List is returned.
func (l *List[T]) FlatMap[U any](f func(T) *List[U]) (mapped *List[U]) {
	mappedElems := make([]U, 0, l.Len())
	for _, elem := range l.slice() {
		mappedElems = append(mappedElems, f(elem).slice()...)
	}
	return ListOfSlice(mappedElems)
}

// Filter returns a new List with the elements that satisfy predicate, in
// order, or an empty List if none do. If the List is empty, predicate is
// not called.
func (l *List[T]) Filter(predicate func(T) bool) (filtered *List[T]) {
	filteredElems := make([]T, 0, l.Len())
	for _, elem := range l.slice() {
		if predicate(elem) {
			filteredElems = append(filteredElems, elem)
		}
	}
	return ListOfSlice[T](filteredElems)
}

// IsEmpty reports whether the List has no elements.
func (l *List[T]) IsEmpty() (empty bool) {
	return l.Len() == 0
}

// First returns an Option holding a copy of the first element, or an empty
// Option if the List is empty. Changing the held value does not change the
// List. The copy is shallow: if T is a pointer, slice or map, the copy still
// refers to the same underlying data.
func (l *List[T]) First() (first Option[T]) {
	return OptionOf(l.FirstOrNil())
}

// FirstOrNil returns a pointer to a copy of the first element, or nil if the
// List is empty. Changing the pointed-to value does not change the List. The
// copy is shallow: if T is a pointer, slice or map, the copy still refers to
// the same underlying data.
func (l *List[T]) FirstOrNil() (first *T) {
	if l.IsEmpty() {
		return nil
	}
	elem := l.elements[0]
	return &elem
}

// All iterates over the elements in order.
func (l *List[T]) All() iter.Seq[T] {
	return func(yield func(T) bool) {
		for _, v := range l.slice() {
			if !yield(v) {
				return // caller broke out of the loop
			}
		}
	}
}

// Enumerate yields index and element, like a slice range.
func (l *List[T]) Enumerate() iter.Seq2[int, T] {
	return func(yield func(int, T) bool) {
		for i, v := range l.slice() {
			if !yield(i, v) {
				return
			}
		}
	}
}

// Backward iterates from last to first.
func (l *List[T]) Backward() iter.Seq2[int, T] {
	return func(yield func(int, T) bool) {
		elems := l.slice()
		for i := len(elems) - 1; i >= 0; i-- {
			if !yield(i, elems[i]) {
				return
			}
		}
	}
}

// ToSet returns a Set holding the distinct elements of the List, or an
// empty Set if the List is empty. hashFunc and eqFunc define how the Set
// hashes and compares elements, see HashFunc and EqFunc.
func (l *List[T]) ToSet(hashFunc HashFunc[T], eqFunc EqFunc[T]) (set *Set[T]) {
	return SetOfSlice(hashFunc, eqFunc, l.slice())
}

// ToSequence returns a Sequence of the elements of the List, in order, or an
// empty Sequence if the List is empty. The Sequence reads the List's backing
// array each time it is evaluated instead of copying it.
func (l *List[T]) ToSequence() (sequence Sequence[T]) {
	return SequenceOfSlice(l.slice())
}
