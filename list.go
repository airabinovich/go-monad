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

// Len returns the length of this list
func (l *List[T]) Len() int {
	if l == nil || l.elements == nil {
		return 0
	}
	return len(l.elements)
}

// String returns a string representation of this List
func (l *List[T]) String() string {
	if l == nil {
		return "[]"
	}
	builder := strings.Builder{}
	builder.WriteRune('[')
	length := l.Len()
	for i, elem := range l.elements {
		builder.WriteString(fmt.Sprintf("%v", elem))
		if i < length-1 {
			builder.WriteRune(',')
		}
	}
	builder.WriteRune(']')
	return builder.String()
}

// Equal returns true when 2 lists contains exactly the same elements and in the same order
func (l *List[T]) Equal(other *List[T]) bool {
	if l.IsEmpty() && other.IsEmpty() {
		return true
	}
	if l.Len() != other.Len() {
		return false
	}
	for i := 0; i < l.Len(); i++ {
		if !reflect.DeepEqual(l.elements[i], other.elements[i]) {
			return false
		}
	}
	return true
}

// Add adds the given element to the List and returns it
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

// Contains returns true if the given element is in the List, false otherwise
func (l *List[T]) Contains(element T) bool {
	if l == nil || l.elements == nil {
		return false
	}
	for _, elem := range l.elements {
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
	for _, elem := range l.elements {
		mappedElems = append(mappedElems, f(elem))
	}
	return ListOfSlice(mappedElems)
}

// FlatMap applies f to each element and returns the concatenation of the
// Lists produced by f, in order. If the List is empty, f is not called and
// an empty List is returned.
func (l *List[T]) FlatMap[U any](f func(T) *List[U]) (mapped *List[U]) {
	mappedElems := make([]U, 0, l.Len())
	for _, elem := range l.elements {
		mappedElems = append(mappedElems, f(elem).elements...)
	}
	return ListOfSlice(mappedElems)
}

// Filter returns a new List with the elements that satisfy predicate, in
// order, or an empty List if none do. If the List is empty, predicate is
// not called.
func (l *List[T]) Filter(predicate func(T) bool) (filtered *List[T]) {
	filteredElems := make([]T, 0, l.Len())
	for _, elem := range l.elements {
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

// First returns an Option holding a pointer to the first element,
// or an empty Option if the List is empty. The pointer holds the
// real element, not a copy.
func (l *List[T]) First() (first Option[T]) {
	return OptionOf(l.FirstOrNil())
}

// FirstOrNil returns a pointer to the first element in the List,
// or nil if the List is empty. The pointer holds the
// real element, not a copy
func (l *List[T]) FirstOrNil() (first *T) {
	if l.IsEmpty() {
		return nil
	}
	return &l.elements[0]
}

// All iterates over the elements in order.
func (l *List[T]) All() iter.Seq[T] {
	return func(yield func(T) bool) {
		for _, v := range l.elements {
			if !yield(v) {
				return // caller broke out of the loop
			}
		}
	}
}

// Enumerate yields index and element, like a slice range.
func (l *List[T]) Enumerate() iter.Seq2[int, T] {
	return func(yield func(int, T) bool) {
		for i, v := range l.elements {
			if !yield(i, v) {
				return
			}
		}
	}
}

// Backward iterates from last to first.
func (l *List[T]) Backward() iter.Seq2[int, T] {
	return func(yield func(int, T) bool) {
		for i := len(l.elements) - 1; i >= 0; i-- {
			if !yield(i, l.elements[i]) {
				return
			}
		}
	}
}

// ToSet returns a Set holding the distinct elements of the List, or an
// empty Set if the List is empty.
// hashFunc is the hashFunc for the set and eqFunc is the EqFunc to compare elements in the set
func (l *List[T]) ToSet(hashFunc HashFunc[T], eqFunc EqFunc[T]) (set Set[T]) {
	return SetOfSlice(hashFunc, eqFunc, l.elements)
}

// ToSequence returns a Sequence of the elements of the List, in order, or an
// empty Sequence if the List is empty. The Sequence reads the List's backing
// array each time it is evaluated instead of copying it.
func (l *List[T]) ToSequence() (sequence Sequence[T]) {
	return SequenceOfSlice(l.elements)
}
