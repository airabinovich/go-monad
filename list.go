package monad

import (
	"fmt"
	"iter"
	"reflect"
	"slices"
	"strings"
)

// List is an ordered sequence of values of type T. The zero value and a nil
// *List behave like an empty List, except that Add on a nil List returns a new
// List.
type List[T any] struct {
	elements []T
}

// ListOf returns a List of elems. When called as ListOf(s...), the List shares
// s's backing array.
func ListOf[T any](elems ...T) (list *List[T]) {
	return ListOfSlice(elems)
}

// ListOfSlice returns a List of elems. The List shares elems's backing array.
func ListOfSlice[T any](elems []T) (list *List[T]) {
	if elems == nil {
		return &List[T]{}
	}
	return &List[T]{slices.Clip(elems)}
}

// slice makes the methods safe for a nil receiver.
func (l *List[T]) slice() (underlying []T) {
	if l == nil {
		return nil
	}
	return l.elements
}

// Len returns the number of elements in the List, or 0 for a nil List.
func (l *List[T]) Len() (length int) {
	return len(l.slice())
}

// String returns the elements formatted with %v, separated by commas and
// wrapped in square brackets, e.g. [1,2,3]. A nil List is formatted as [].
func (l *List[T]) String() (str string) {
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
func (l *List[T]) Equal(other *List[T]) (areEqual bool) {
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
	l.elements = append(l.elements, element)
	return l
}

// Contains reports whether element is in the List. Elements are compared with
// reflect.DeepEqual.
func (l *List[T]) Contains(element T) (contains bool) {
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
func (l *List[T]) All() (all iter.Seq[T]) {
	return func(yield func(T) bool) {
		for _, v := range l.slice() {
			if !yield(v) {
				return // caller broke out of the loop
			}
		}
	}
}

// Enumerate yields index and element, like a slice range.
func (l *List[T]) Enumerate() (enumerated iter.Seq2[int, T]) {
	return func(yield func(int, T) bool) {
		for i, v := range l.slice() {
			if !yield(i, v) {
				return
			}
		}
	}
}

// Backward iterates from last to first.
func (l *List[T]) Backward() (backwardEnumerated iter.Seq2[int, T]) {
	return func(yield func(int, T) bool) {
		elems := l.slice()
		for i, elem := range slices.Backward(elems) {
			if !yield(i, elem) {
				return
			}
		}
	}
}

// ToSet returns a Set holding the distinct elements of the List, or an
// empty Set if the List is empty. hashFunc and eqFunc define how the Set
// hashes and compares elements, see HashFunc and EqFunc.
func (l *List[T]) ToSet(hashFunc HashFunc[T], eqFunc EqFunc[T]) (set *Set[T]) {
	return SetOfSliceFunc(hashFunc, eqFunc, l.slice())
}

// ToSequence returns a Sequence of the elements of the List, in order, or an
// empty Sequence if the List is empty. The Sequence reads the List's backing
// array each time it is evaluated instead of copying it.
func (l *List[T]) ToSequence() (sequence Sequence[T]) {
	return SequenceOfSeq(l.All())
}
