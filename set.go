package monad

import (
	"fmt"
	"hash/maphash"
	"strings"
)

var hashSeed = maphash.MakeSeed()

// HashFunc computes the hash of an element of a Set. Elements that are equal
// according to the Set's EqFunc must have the same hash.
type HashFunc[T any] func(T) uint64

// EqFunc reports whether two elements of a Set are equal.
type EqFunc[T any] func(T, T) bool

// Set is an unordered collection of distinct values of type T.
//
// It needs a HashFunc and an EqFunc.
// This way, the Set can hold non-comparable elements like slices or maps, or structs containing those.
//
// Always create a Set with SetOf, SetOfSlice or a method that returns a Set.
// Every function and method that returns a Set returns a non-nil one, even when
// it is empty, so calls can always be chained.
//
// A nil *Set and the zero value of Set behave like an empty Set with every
// method except Add. Add needs the HashFunc and EqFunc that only a created Set
// carries, so it panics on them. Filter on a nil Set returns an empty Set
// without these functions, so Add panics on that result as well.
//
// Set is not safe for concurrent use.
type Set[T any] struct {
	elements map[uint64]*List[T]
	hashFunc HashFunc[T]
	eq       EqFunc[T]
}

// HashComparable returns a HashFunc to be used for any comparable type in a Set.
// The returned function panics if the element is an interface value whose
// dynamic type is not comparable.
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
func SetOf[T any](hashFunc HashFunc[T], eq EqFunc[T], elems ...T) (set *Set[T]) {
	return SetOfSlice(hashFunc, eq, elems)
}

// SetOfSlice returns a Set holding the distinct elements of elems, or an
// empty Set if elems is empty or nil.
func SetOfSlice[T any](hashFunc HashFunc[T], eq EqFunc[T], elems []T) (set *Set[T]) {
	set = &Set[T]{elements: make(map[uint64]*List[T], len(elems)), hashFunc: hashFunc, eq: eq}
	for _, elem := range elems {
		set.Add(elem)
	}
	return set
}

// buckets returns the elements grouped by hash, or nil for a nil Set. It is the
// single place that makes the read-only methods safe for a nil receiver.
func (s *Set[T]) buckets() map[uint64]*List[T] {
	if s == nil {
		return nil
	}
	return s.elements
}

// String returns the elements formatted with %v, separated by commas and
// wrapped in curly braces, in unspecified order. A nil Set is formatted as {}.
func (s *Set[T]) String() string {
	builder := strings.Builder{}
	builder.WriteRune('{')
	listNr := 0
	for _, elems := range s.buckets() {
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

// Add adds element to the Set unless an equal element, according to the Set's
// EqFunc, is already present. The Set is modified in place and returned.
//
// Add panics if the Set is nil, is the zero value, or has no HashFunc or EqFunc,
// because it cannot hash or compare the element without them.
func (s *Set[T]) Add(element T) (set *Set[T]) {
	if s == nil || s.hashFunc == nil || s.eq == nil {
		panic("monad: Add called on a Set without HashFunc and EqFunc, create the Set with SetOf or SetOfSlice")
	}
	if s.elements == nil {
		s.elements = make(map[uint64]*List[T])
	}
	hash := s.hashFunc(element)
	bucket, ok := s.elements[hash]
	if !ok {
		s.elements[hash] = ListOf[T](element)
		return s
	}
	for existingElement := range bucket.All() {
		if s.eq(element, existingElement) {
			return s
		}
	}
	bucket.Add(element)
	return s
}

// Contains reports whether the Set contains the given element. A nil Set
// contains nothing.
func (s *Set[T]) Contains(element T) (exists bool) {
	if len(s.buckets()) == 0 {
		return false
	}
	if elements, ok := s.elements[s.hashFunc(element)]; ok {
		for elem := range elements.All() {
			if s.eq(element, elem) {
				return true
			}
		}
	}
	return false
}

// Size returns the number of elements held by the Set, or 0 for a nil Set.
func (s *Set[T]) Size() (size int) {
	elemCount := 0
	for _, elem := range s.buckets() {
		elemCount += elem.Len()
	}
	return elemCount
}

// IsEmpty reports whether the Set has no elements. A nil Set is empty.
func (s *Set[T]) IsEmpty() (empty bool) {
	return s.Size() == 0
}

// Equal reports whether both Sets contain exactly the same elements,
// regardless of the order. A nil Set is equal to an empty Set. Elements of s are
// looked up in other using other's HashFunc and EqFunc, so the two Sets may use
// different functions as long as they agree on which elements are equal.
func (s *Set[T]) Equal(other *Set[T]) (areEqual bool) {
	if s.Size() != other.Size() {
		return false
	}
	for _, elemsInS := range s.buckets() {
		for elemInS := range elemsInS.All() {
			if !other.Contains(elemInS) {
				return false
			}
		}
	}
	return true
}

// Map applies f to each element and returns a new Set with the distinct
// results. If the Set is empty, f is not called and an empty Set is
// returned.
func (s *Set[T]) Map[U any](hashFunc HashFunc[U], eqFunc EqFunc[U], f func(T) U) (mapped *Set[U]) {
	mapped = SetOf[U](hashFunc, eqFunc)
	for _, elems := range s.buckets() {
		for elem := range elems.All() {
			mapped.Add(f(elem))
		}
	}
	return mapped
}

// FlatMap applies f to each element and returns the union of the Sets
// produced by f. A nil Set returned by f is treated as empty. If the Set is
// empty, f is not called and an empty Set is returned.
func (s *Set[T]) FlatMap[U any](hashFunc HashFunc[U], eqFunc EqFunc[U], f func(T) *Set[U]) (mapped *Set[U]) {
	mapped = SetOf(hashFunc, eqFunc)
	for _, elems := range s.buckets() {
		for elem := range elems.All() {
			result := f(elem)
			for _, resultElements := range result.buckets() {
				for resultElement := range resultElements.All() {
					mapped.Add(resultElement)
				}
			}
		}
	}
	return mapped
}

// Filter returns a new Set with the elements that satisfy predicate, or an
// empty Set if none do. If the Set is empty, predicate is not called. Filter on
// a nil Set returns an empty Set without HashFunc and EqFunc, see Set.
func (s *Set[T]) Filter(predicate func(T) bool) (filtered *Set[T]) {
	if s == nil {
		return &Set[T]{elements: map[uint64]*List[T]{}}
	}
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
func (s *Set[T]) First() (first Option[T]) {
	for _, elems := range s.buckets() {
		for elem := range elems.All() {
			return OptionOf(&elem)
		}
	}
	return OptionOf[T](nil)
}

// FirstOrNil returns a pointer to a copy of an arbitrary element, or nil
// if the Set is empty. Changing the pointed-to value does not change the Set.
// Which element is returned is unspecified and may differ between calls.
func (s *Set[T]) FirstOrNil() (first *T) {
	for _, elems := range s.buckets() {
		for elem := range elems.All() {
			return &elem
		}
	}
	return nil
}

// ToList returns a List holding the elements of the Set in unspecified
// order, or an empty List if the Set is empty.
func (s *Set[T]) ToList() (list *List[T]) {
	result := ListOf[T]()
	for _, elems := range s.buckets() {
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
func (s *Set[T]) ToSequence() (sequence Sequence[T]) {
	return SequenceOfSet(s)
}
