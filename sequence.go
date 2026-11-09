package monad

import (
	"iter"
	"slices"
)

// Sequence is a lazy, ordered sequence of values of type T that may hold
// duplicates. Transformations (Map, FlatMap, Filter) only describe work.
// Nothing runs until an accessor (IsEmpty, First, FirstOrNil, ToList, ToSet) or
// a range over All requests elements. Accessors evaluate only as many elements
// as they need, and each call evaluates again from the source, so callbacks may
// run more than once. The zero value is empty.
type Sequence[T any] struct {
	seq iter.Seq[T]
}

// SequenceOf returns a Sequence of elems, in order, or an empty Sequence if
// no elems are given. When called as SequenceOf(s...), the Sequence reads
// s's backing array each time it is evaluated instead of copying it.
func SequenceOf[T any](elems ...T) (sequence Sequence[T]) {
	return SequenceOfSlice(elems)
}

// SequenceOfSlice returns a Sequence of the elements of elems, in order, or
// an empty Sequence if elems is empty or nil. The Sequence reads elems's
// backing array each time it is evaluated instead of copying it.
func SequenceOfSlice[T any](elems []T) (sequence Sequence[T]) {
	return SequenceOfSeq(slices.Values(elems))
}

// SequenceOfSet returns a Sequence of the elements of the Set in unspecified
// order. The Set is not copied: it is read each time the Sequence is evaluated,
// so the order may differ between evaluations and later changes to the Set are
// visible.
func SequenceOfSet[T any](s *Set[T]) (sequence Sequence[T]) {
	return Sequence[T]{
		seq: func(yield func(T) bool) {
			for _, elems := range s.buckets() {
				for elem := range elems.All() {
					if !yield(elem) {
						return
					}
				}
			}
		},
	}
}

// SequenceOfSeq returns a Sequence of the values produced by seq. A nil seq
// gives an empty Sequence. seq is called each time the Sequence is evaluated.
func SequenceOfSeq[T any](seq iter.Seq[T]) (sequence Sequence[T]) {
	return Sequence[T]{seq: seq}
}

// All returns the underlying iterator, or an empty one for the zero value.
func (s Sequence[T]) All() (seq iter.Seq[T]) {
	if s.seq == nil {
		return func(yield func(T) bool) {}
	}
	return s.seq
}

// Map returns a Sequence that applies f to each element.
func (s Sequence[T]) Map[U any](f func(T) U) (mapped Sequence[U]) {
	source := s.All()
	return Sequence[U]{seq: func(yield func(U) bool) {
		for elem := range source {
			if !yield(f(elem)) {
				break
			}
		}
	}}
}

// FlatMap returns a Sequence that applies f to each element and yields the
// elements of each resulting Sequence.
func (s Sequence[T]) FlatMap[U any](f func(T) Sequence[U]) (mapped Sequence[U]) {
	source := s.All()
	return Sequence[U]{seq: func(yield func(U) bool) {
		for elem := range source {
			for result := range f(elem).All() {
				if !yield(result) {
					return
				}
			}
		}
	}}
}

// Filter returns a Sequence of the elements that satisfy predicate.
func (s Sequence[T]) Filter(predicate func(T) bool) (filtered Sequence[T]) {
	source := s.All()
	return Sequence[T]{seq: func(yield func(T) bool) {
		for elem := range source {
			if predicate(elem) && !yield(elem) {
				break
			}
		}
	}}
}

// IsEmpty reports whether the Sequence has no elements.
func (s Sequence[T]) IsEmpty() (empty bool) {
	for range s.All() {
		return false
	}
	return true
}

// First returns an Option holding the first element, or an empty Option if
// the Sequence is empty.
func (s Sequence[T]) First() (first Option[T]) {
	return OptionOf[T](s.FirstOrNil())
}

// FirstOrNil returns a pointer to the first element, or nil if the Sequence is
// empty.
func (s Sequence[T]) FirstOrNil() (first *T) {
	for elem := range s.All() {
		return &elem
	}
	return nil
}

// ToList returns a List of all elements, including duplicates.
func (s Sequence[T]) ToList() (list *List[T]) {
	return ListOfSlice(slices.Collect(s.All()))
}

// ToSet returns a Set of the distinct elements. hashFunc and eqFunc are
// passed to SetOfFunc, so they panic if nil.
func (s Sequence[T]) ToSet(hashFunc HashFunc[T], eqFunc EqFunc[T]) (set *Set[T]) {
	set = SetOfFunc[T](hashFunc, eqFunc)
	for elem := range s.All() {
		set.Add(elem)
	}
	return set
}
