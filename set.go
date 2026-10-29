package monad

import (
	"fmt"
	"maps"
	"strings"
)

// Set is an unordered collection of distinct values of type T. It is a plain
// map[T]struct{}, so it adds no overhead and supports len, range and key
// lookups directly. The zero value of Set is empty. Every function and
// method returns a non-nil Set, even when it is empty, so calls can always
// be chained. Set is not safe for concurrent use.
type Set[T comparable] map[T]struct{}

// SetOf returns a Set holding the distinct elems, or an empty Set if no
// elems are given.
func SetOf[T comparable](elems ...T) (set Set[T]) {
	return SetOfSlice(elems)
}

// SetOfSlice returns a Set holding the distinct elements of elems, or an
// empty Set if elems is empty or nil. Duplicates are removed by using the
// elements as map keys. It panics if an element is an interface value whose
// dynamic type is not comparable.
func SetOfSlice[T comparable](elems []T) (set Set[T]) {
	set = make(Set[T], len(elems))
	for _, elem := range elems {
		set[elem] = struct{}{}
	}
	return set
}

func (s Set[T]) String() string {
	builder := strings.Builder{}
	builder.WriteRune('{')
	length := len(s)
	i := 0
	for k, _ := range s {
		builder.WriteString(fmt.Sprintf("%v", k))
		if i < length-1 {
			builder.WriteRune(',')
		}
		i++
	}
	builder.WriteRune('}')
	return builder.String()
}

// Map applies f to each element and returns a Set with the distinct
// results. If the Set is empty, f is not called and an empty Set is
// returned.
func (s Set[T]) Map[U comparable](f func(T) U) (mapped Set[U]) {
	mapped = make(Set[U], len(s))
	for elem := range s {
		mapped[f(elem)] = struct{}{}
	}
	return mapped
}

// FlatMap applies f to each element and returns the union of the Sets
// produced by f. If the Set is empty, f is not called and an empty Set is
// returned.
func (s Set[T]) FlatMap[U comparable](f func(T) Set[U]) (mapped Set[U]) {
	mapped = Set[U]{}
	for elem := range s {
		for result := range f(elem) {
			mapped[result] = struct{}{}
		}
	}
	return mapped
}

// Filter returns a new Set with the elements that satisfy predicate, or an
// empty Set if none do. If the Set is empty, predicate is not called.
func (s Set[T]) Filter(predicate func(T) bool) (filtered Set[T]) {
	filtered = Set[T]{}
	for elem := range s {
		if predicate(elem) {
			filtered[elem] = struct{}{}
		}
	}
	return filtered
}

// IsEmpty reports whether the Set has no elements.
func (s Set[T]) IsEmpty() (empty bool) {
	return len(s) == 0
}

// First returns an Option holding a copy of an arbitrary element, or an
// empty Option if the Set is empty. Which element is returned is
// unspecified and may differ between calls.
func (s Set[T]) First() (first Option[T]) {
	for elem := range s {
		return OptionOf(&elem)
	}
	return OptionOf[T](nil)
}

// FirstOrNil returns a pointer to an arbitrary element, or nil
// if the Set is empty. Which element is returned is
// unspecified and may differ between calls.
func (s Set[T]) FirstOrNil() (first *T) {
	for elem := range s {
		res := elem
		return &res
	}
	return nil
}

// ToList returns a List holding the elements of the Set in unspecified
// order, or an empty List if the Set is empty.
func (s Set[T]) ToList() (list List[T]) {
	elems := make([]T, 0, len(s))
	for elem := range s {
		elems = append(elems, elem)
	}
	return ListOfSlice(elems)
}

// ToSequence returns a Sequence of the elements of the Set in unspecified
// order, or an empty Sequence if the Set is empty. The Sequence reads the
// Set each time it is evaluated instead of copying it, and the order may
// differ between evaluations.
func (s Set[T]) ToSequence() (sequence Sequence[T]) {
	return SequenceOfSeq(maps.Keys(s))
}
