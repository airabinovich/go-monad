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

// Set is an unordered collection of distinct values of type T. The HashFunc
// and EqFunc it is created with let it hold non-comparable elements.
//
// A nil *Set and the zero value behave like an empty Set, except that Add
// panics on both and Filter panics on the zero value, because they lack the
// functions to hash and compare. Filter on a nil Set returns an empty Set that
// also lacks them, so Add panics on that result. The constructors, Map and
// FlatMap panic on a nil HashFunc or EqFunc. These panics are a known
// limitation while the API is unstable.
type Set[T any] struct {
	elements map[uint64]*List[T]
	size     uint32 // kept in sync by Add, the only method that modifies the Set
	hashFunc HashFunc[T]
	eq       EqFunc[T]
}

// HashComparable returns a HashFunc for any comparable type. It panics if an
// element is an interface value whose dynamic type is not comparable.
func HashComparable[T comparable]() (hashFunc HashFunc[T]) {
	return func(elem T) uint64 {
		return maphash.Comparable(hashSeed, elem)
	}
}

// EqComparable returns an EqFunc for any comparable type.
func EqComparable[T comparable]() (eqFunc EqFunc[T]) {
	return func(elem T, other T) bool {
		return elem == other
	}
}

// SetOf returns a Set of the distinct elems, using HashComparable and EqComparable.
func SetOf[T comparable](elems ...T) (set *Set[T]) {
	return SetOfSliceFunc(HashComparable[T](), EqComparable[T](), elems)
}

// SetOfSlice is like SetOf, taking a slice.
func SetOfSlice[T comparable](elems []T) (set *Set[T]) {
	return SetOfSliceFunc(HashComparable[T](), EqComparable[T](), elems)
}

// SetOfFunc returns a Set of the distinct elems. It panics if hashFunc or eq is
// nil.
func SetOfFunc[T any](hashFunc HashFunc[T], eq EqFunc[T], elems ...T) (set *Set[T]) {
	return SetOfSliceFunc(hashFunc, eq, elems)
}

// SetOfSliceFunc is like SetOfFunc, taking a slice. It panics on a nil hashFunc
// or eq even when elems is empty.
func SetOfSliceFunc[T any](hashFunc HashFunc[T], eq EqFunc[T], elems []T) (set *Set[T]) {
	if hashFunc == nil || eq == nil {
		panic("hashFunc or eqFunc is nil")
	}
	set = &Set[T]{
		elements: make(map[uint64]*List[T], len(elems)),
		hashFunc: hashFunc,
		eq:       eq,
	}
	for _, elem := range elems {
		set.Add(elem)
	}
	return set
}

// buckets makes the read-only methods safe for a nil receiver.
func (s *Set[T]) buckets() (underlying map[uint64]*List[T]) {
	if s == nil {
		return nil
	}
	return s.elements
}

// String returns the elements formatted with %v, separated by commas and
// wrapped in curly braces, in unspecified order. A nil Set is formatted as {}.
func (s *Set[T]) String() (str string) {
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

// Add adds element unless an equal one is present. The Set is modified in
// place and returned. It panics on a nil Set, the zero value, or a Set without
// HashFunc or EqFunc.
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
		s.size++
		return s
	}
	for existingElement := range bucket.All() {
		if s.eq(element, existingElement) {
			return s
		}
	}
	bucket.Add(element)
	s.size++
	return s
}

// Contains reports whether the Set contains element.
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

// Size returns the number of elements in constant time. More than 2^32-1
// elements are not supported.
func (s *Set[T]) Size() (size int) {
	if s == nil {
		return 0
	}
	return int(s.size)
}

// IsEmpty reports whether the Set has no elements.
func (s *Set[T]) IsEmpty() (empty bool) {
	return s.Size() == 0
}

// Equal reports whether both Sets contain the same elements. The Sets may use
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
// results. hashFunc and eqFunc are those of the result and must not be nil.
func (s *Set[T]) Map[U any](hashFunc HashFunc[U], eqFunc EqFunc[U], f func(T) U) (mapped *Set[U]) {
	mapped = SetOfFunc[U](hashFunc, eqFunc)
	for _, elems := range s.buckets() {
		for elem := range elems.All() {
			mapped.Add(f(elem))
		}
	}
	return mapped
}

// FlatMap applies f to each element and returns the union of the Sets
// produced by f. A nil Set returned by f is treated as empty. hashFunc and
// eqFunc are those of the result and must not be nil.
func (s *Set[T]) FlatMap[U any](hashFunc HashFunc[U], eqFunc EqFunc[U], f func(T) *Set[U]) (mapped *Set[U]) {
	mapped = SetOfFunc(hashFunc, eqFunc)
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

// Filter returns a new Set with the elements that satisfy predicate. See Set
// for its behavior on a nil Set and the zero value.
func (s *Set[T]) Filter(predicate func(T) bool) (filtered *Set[T]) {
	if s == nil {
		return &Set[T]{elements: map[uint64]*List[T]{}}
	}
	filtered = SetOfFunc(s.hashFunc, s.eq)
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
// empty Option if the Set is empty. The element may differ between calls.
func (s *Set[T]) First() (first Option[T]) {
	for _, elems := range s.buckets() {
		for elem := range elems.All() {
			return OptionOf(&elem)
		}
	}
	return OptionOf[T](nil)
}

// FirstOrNil returns a pointer to a copy of an arbitrary element, or nil if
// the Set is empty. The element may differ between calls.
func (s *Set[T]) FirstOrNil() (first *T) {
	for _, elems := range s.buckets() {
		for elem := range elems.All() {
			return &elem
		}
	}
	return nil
}

// ToList returns a List of the elements in unspecified order.
func (s *Set[T]) ToList() (list *List[T]) {
	result := ListOf[T]()
	for _, elems := range s.buckets() {
		for elem := range elems.All() {
			result.Add(elem)
		}
	}
	return result
}

// ToSequence returns a Sequence of the elements in unspecified order. See
// SequenceOfSet.
func (s *Set[T]) ToSequence() (sequence Sequence[T]) {
	return SequenceOfSet(s)
}
