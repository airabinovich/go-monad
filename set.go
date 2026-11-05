package monad

import (
	"fmt"
	"hash/maphash"
	"strings"
)

var hashSeed = maphash.MakeSeed()

// HashFunc is a function to compute a hash for an element in a Set
type HashFunc[T any] func(T) uint64

// EqFunc is a function that checks equality between two elements in a Set
type EqFunc[T any] func(T, T) bool

// Set is an unordered collection of distinct values of type T.
//
// It needs a HashFunc and an EqFunc.
// This way, the Set can hold non-comparable elements like slices or maps, or structs containing those.
//
// The zero value of Set is empty and lacks the mentioned functions, never construct directly.
// Use any constructor instead.
//
// Every function and method returns a non-nil Set, even when it is empty,
// so calls can always be chained. Set is not safe for concurrent use.
type Set[T any] struct {
	elements map[uint64]*List[T]
	hashFunc HashFunc[T]
	eq       EqFunc[T]
}

// HashComparable returns a HashFunc to be used for any comparable type in a Set
func HashComparable[T comparable]() HashFunc[T] {
	return func(elem T) uint64 {
		return maphash.Comparable(hashSeed, elem)
	}
}

// EqComparable returns an EqFunc to be used for any comparable type in a Set
func EqComparable[T comparable]() EqFunc[T] {
	return func(elem T, other T) bool {
		return elem == other
	}
}

// SetOf returns a Set holding the distinct elems, or an empty Set if no
// elems are given.
func SetOf[T any](hashFunc HashFunc[T], eq EqFunc[T], elems ...T) (set Set[T]) {
	return SetOfSlice(hashFunc, eq, elems)
}

// SetOfSlice returns a Set holding the distinct elements of elems, or an
// empty Set if elems is empty or nil.
func SetOfSlice[T any](hashFunc HashFunc[T], eq EqFunc[T], elems []T) (set Set[T]) {
	set = Set[T]{elements: make(map[uint64]*List[T], len(elems)), hashFunc: hashFunc, eq: eq}
	for _, elem := range elems {
		set.Add(elem)
	}
	return set
}

// String returns a string representation of the set
func (s Set[T]) String() string {
	builder := strings.Builder{}
	builder.WriteRune('{')
	listNr := 0
	for _, elems := range s.elements {
		elemsInList := elems.Len()
		if listNr > 0 && elemsInList > 0 {
			builder.WriteRune(',')
		}
		listNr++
		for elemNr, elem := range elems.Enumerate() {
			builder.WriteString(fmt.Sprintf("%v", elem))
			if elemNr < elemsInList-1 {
				builder.WriteRune(',')
			}
		}
	}
	builder.WriteRune('}')
	return builder.String()
}

// Add adds the given element to the current set and returns it
func (s Set[T]) Add(element T) (set Set[T]) {
	hash := s.hashFunc(element)
	if _, ok := s.elements[hash]; !ok {
		s.elements[hash] = ListOf[T](element)
	} else {
		alreadyInSet := false
		for existingElement := range s.elements[hash].All() {
			if s.eq(element, existingElement) {
				alreadyInSet = true
				break
			}
		}
		if !alreadyInSet {
			s.elements[hash].Add(element)
		}
	}
	return s
}

// Contains returns true if the Set contains the given element, false otherwise
func (s Set[T]) Contains(element T) (exists bool) {
	hash := s.hashFunc(element)
	if elements, ok := s.elements[hash]; ok {
		for elem := range elements.All() {
			if s.eq(element, elem) {
				return true
			}
		}
	}
	return false
}

// Size returns the amount of elements held by the Set
func (s Set[T]) Size() (size int) {
	elemCount := 0
	for _, elem := range s.elements {
		elemCount += elem.Len()
	}
	return elemCount
}

// IsEmpty reports whether the Set has no elements.
func (s Set[T]) IsEmpty() (empty bool) {
	return s.Size() == 0
}

// Equal returns true if the two sets are equal.
//
// Two sets are equal when they contain exactly the same elements,
// regardless of the order.
func (s Set[T]) Equal(other Set[T]) (areEqual bool) {
	if len(s.elements) != len(other.elements) {
		return false
	}
	if s.IsEmpty() && other.IsEmpty() {
		return true
	}
	for _, elemsInS := range s.elements {
		for elemInS := range elemsInS.All() {
			if !other.Contains(elemInS) {
				return false
			}
		}
	}
	for _, elemsInOther := range other.elements {
		for elemInOther := range elemsInOther.All() {
			if !s.Contains(elemInOther) {
				return false
			}
		}
	}
	return true
}

// Map applies f to each element and returns a new Set with the distinct
// results. If the Set is empty, f is not called and an empty Set is
// returned.
func (s Set[T]) Map[U any](hashFunc HashFunc[U], eqFunc EqFunc[U], f func(T) U) (mapped Set[U]) {
	mapped = SetOf[U](hashFunc, eqFunc)
	for _, elems := range s.elements {
		for elem := range elems.All() {
			mapped.Add(f(elem))
		}
	}
	return mapped
}

// FlatMap applies f to each element and returns the union of the Sets
// produced by f. If the Set is empty, f is not called and an empty Set is
// returned.
func (s Set[T]) FlatMap[U any](hashFunc HashFunc[U], eqFunc EqFunc[U], f func(T) Set[U]) (mapped Set[U]) {
	mapped = SetOf(hashFunc, eqFunc)
	for _, elems := range s.elements {
		for elem := range elems.All() {
			result := f(elem)
			for _, resultElements := range result.elements {
				for resultElement := range resultElements.All() {
					mapped.Add(resultElement)
				}
			}
		}
	}
	return mapped
}

// Filter returns a new Set with the elements that satisfy predicate, or an
// empty Set if none do. If the Set is empty, predicate is not called.
func (s Set[T]) Filter(predicate func(T) bool) (filtered Set[T]) {
	filtered = SetOf(s.hashFunc, s.eq)
	for _, elems := range s.elements {
		for elem := range elems.All() {
			if predicate(elem) {
				filtered.Add(elem)
			}
		}
	}
	return filtered
}

// First returns an Option holding a copy of an arbitrary element, or an
// empty Option if the Set is empty. Which element is returned is
// unspecified and may differ between calls.
func (s Set[T]) First() (first Option[T]) {
	for _, elems := range s.elements {
		for elem := range elems.All() {
			return OptionOf(&elem)
		}
	}
	return OptionOf[T](nil)
}

// FirstOrNil returns a pointer to an arbitrary element, or nil
// if the Set is empty. Which element is returned is
// unspecified and may differ between calls.
func (s Set[T]) FirstOrNil() (first *T) {
	for _, elems := range s.elements {
		for elem := range elems.All() {
			return &elem
		}
	}
	return nil
}

// ToList returns a List holding the elements of the Set in unspecified
// order, or an empty List if the Set is empty.
func (s Set[T]) ToList() (list *List[T]) {
	result := ListOf[T]()
	for _, elems := range s.elements {
		for elem := range elems.All() {
			result.Add(elem)
		}
	}
	return result
}

// ToSequence returns a Sequence of the elements of the Set in unspecified
// order, or an empty Sequence if the Set is empty. The Sequence reads the
// Set each time it is evaluated instead of copying it, and the order may
// differ between evaluations.
func (s Set[T]) ToSequence() (sequence Sequence[T]) {
	return SequenceOfSet(s)
}
