package monad

import (
	"iter"
	"slices"
)

// Sequence is a lazy, ordered sequence of values of type T. Unlike a List,
// its elements are not directly accessible: transformations (Map, FlatMap,
// Filter) only describe work, and nothing runs until an accessor (IsEmpty,
// First, ToList, ToSet) requests elements. Accessors evaluate only as many
// elements as they need, and every accessor call evaluates the Sequence
// again from its source, so callbacks may run more than once. Unlike a Set,
// a Sequence may hold duplicates.
//
// Sequence wraps an iter.Seq, so it adds no overhead beyond a function
// value. The zero value of Sequence is empty. Sequence is not safe for
// concurrent use.
type Sequence[T comparable] struct {
	seq iter.Seq[T]
}

// SequenceOf returns a Sequence of elems, in order, or an empty Sequence if
// no elems are given. When called as SequenceOf(s...), the Sequence reads
// s's backing array each time it is evaluated instead of copying it.
func SequenceOf[T comparable](elems ...T) (sequence Sequence[T]) {
	return SequenceOfSlice(elems)
}

// SequenceOfSlice returns a Sequence of the elements of elems, in order, or
// an empty Sequence if elems is empty or nil. The Sequence reads elems's
// backing array each time it is evaluated instead of copying it.
func SequenceOfSlice[T comparable](elems []T) (sequence Sequence[T]) {
	return SequenceOfSeq(slices.Values(elems))
}

// SequenceOfSeq returns a Sequence of the values produced by seq, in order,
// or an empty Sequence if seq is nil or produces nothing. seq is called each
// time the Sequence is evaluated.
func SequenceOfSeq[T comparable](seq iter.Seq[T]) (sequence Sequence[T]) {
	return Sequence[T]{seq: seq}
}

// all returns the underlying iterator, or an empty one for the zero value.
func (s Sequence[T]) all() (seq iter.Seq[T]) {
	if s.seq == nil {
		return func(yield func(T) bool) {}
	}
	return s.seq
}

// Map returns a Sequence that applies f to each element, in order. f is not
// called until an accessor requests elements, and it is never called on an
// empty Sequence.
func (s Sequence[T]) Map[U comparable](f func(T) U) (mapped Sequence[U]) {
	source := s.all()
	return Sequence[U]{seq: func(yield func(U) bool) {
		for elem := range source {
			if !yield(f(elem)) {
				break
			}
		}
	}}
}

// FlatMap returns a Sequence that applies f to each element and yields the
// elements of each resulting Sequence, in order. f is not called until an
// accessor requests elements, and it is never called on an empty Sequence.
func (s Sequence[T]) FlatMap[U comparable](f func(T) Sequence[U]) (mapped Sequence[U]) {
	source := s.all()
	return Sequence[U]{seq: func(yield func(U) bool) {
		for elem := range source {
			for result := range f(elem).all() {
				if !yield(result) {
					return
				}
			}
		}
	}}
}

// Filter returns a Sequence of the elements that satisfy predicate, in
// order. predicate is not called until an accessor requests elements, and it
// is never called on an empty Sequence.
func (s Sequence[T]) Filter(predicate func(T) bool) (filtered Sequence[T]) {
	source := s.all()
	return Sequence[T]{seq: func(yield func(T) bool) {
		for elem := range source {
			if predicate(elem) && !yield(elem) {
				break
			}
		}
	}}
}

// IsEmpty reports whether the Sequence has no elements. It is an accessor:
// it evaluates the Sequence up to its first element.
func (s Sequence[T]) IsEmpty() (empty bool) {
	for range s.all() {
		return false
	}
	return true
}

// First returns an Option holding the first element, or an empty Option if
// the Sequence is empty. It is an accessor: it evaluates the Sequence up to
// its first element.
func (s Sequence[T]) First() (first Option[T]) {
	for elem := range s.all() {
		return OptionOf(&elem)
	}
	return OptionOf[T](nil)
}

// ToList returns a List of all elements, in order, including duplicates, or
// an empty List if the Sequence is empty. It is an accessor: it evaluates
// the whole Sequence.
func (s Sequence[T]) ToList() (list List[T]) {
	return ListOfSlice(slices.Collect(s.all()))
}

// ToSet returns a Set of the distinct elements, or an empty Set if the
// Sequence is empty. It is an accessor: it evaluates the whole Sequence.
func (s Sequence[T]) ToSet() (set Set[T]) {
	set = Set[T]{}
	if s.seq == nil {
		return set
	}
	for elem := range s.seq {
		set[elem] = struct{}{}
	}
	return set
}
