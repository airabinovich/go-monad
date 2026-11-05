package monad

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

// assertEmptySet checks that s is empty but not nil, so it is safe to chain.
func assertEmptySet[T any](t *testing.T, s Set[T], msgAndArgs ...any) {
	t.Helper()
	assert.Equal(t, 0, s.Size(), msgAndArgs...)
}

func assertSetsAreEqual[T comparable](t *testing.T, s1 Set[T], s2 Set[T], msgAndArgs ...any) {
	t.Helper()
	assert.True(t, s1.Equal(s2), msgAndArgs...)
}

func TestSetOf(t *testing.T) {
	assertSetsAreEqual(t, SetOfSlice(HashComparable[int](), EqComparable[int](), []int{1, 2, 3}), SetOf(HashComparable[int](), EqComparable[int](), 1, 2, 3))
	assertSetsAreEqual(t, SetOf(HashComparable[int](), EqComparable[int](), 1, 2), SetOf(HashComparable[int](), EqComparable[int](), 1, 2, 1, 2), "duplicates must be removed")
	assertEmptySet(t, SetOf[int](HashComparable[int](), EqComparable[int]()))
}

func TestSetOfSlice(t *testing.T) {
	set := SetOfSlice(HashComparable[string](), EqComparable[string](), []string{"a", "b", "a"})
	assert.Equal(t, 2, set.Size(), "duplicates must be removed")
	assert.True(t, set.Contains("a"))
	assert.True(t, set.Contains("b"))

	assertEmptySet(t, SetOfSlice[int](HashComparable[int](), EqComparable[int](), nil))
	assertEmptySet(t, SetOfSlice(HashComparable[int](), EqComparable[int](), []int{}))
}

func TestSetSize(t *testing.T) {
	hashFunc := func(i int) (s uint64) {
		if i > 5 {
			return 1
		}
		return uint64(i)
	}
	set := SetOf(hashFunc, EqComparable[int](), 1, 2, 3, 8, 9, 0)
	assert.Equal(t, 6, set.Size(), "size must me independent of position in the set")
}

func TestSetZeroValueIsEmpty(t *testing.T) {
	var s Set[string]
	assert.True(t, s.IsEmpty())
	assert.False(t, s.First().IsPresent())
	assertEmptySet(t, s.Map(HashComparable[string](), EqComparable[string](), strconv.Quote), "methods on the zero value must return non-nil Sets")
	assertEmptySet(t, s.Filter(func(e string) (keep bool) { return true }))
	assertEmptySet(t, s.FlatMap(HashComparable[string](), EqComparable[string](), func(e string) (single Set[string]) { return SetOf(HashComparable[string](), EqComparable[string](), e) }))
	assertEmptyList(t, s.ToList())
}

func TestSetIsEmpty(t *testing.T) {
	assert.False(t, SetOf(HashComparable[int](), EqComparable[int](), 0).IsEmpty(), "a Set holding the zero value is not empty")
	assert.True(t, SetOf[int](HashComparable[int](), EqComparable[int]()).IsEmpty())
}

func TestSetFirst(t *testing.T) {
	set := SetOf(HashComparable[int](), EqComparable[int](), 1, 2, 3)
	first := set.First().GetOrNil()
	if assert.NotNil(t, first) {
		assert.True(t, set.Contains(*first), "First must return an element of the Set")
	}
	assert.False(t, SetOf[int](HashComparable[int](), EqComparable[int]()).First().IsPresent())
}

func TestSetFirstOrNil(t *testing.T) {
	set := SetOf(HashComparable[int](), EqComparable[int](), 1, 2, 3)
	first := set.FirstOrNil()
	if assert.NotNil(t, first) {
		assert.True(t, set.Contains(*first), "First must return an element of the Set")
	}
	assert.False(t, SetOf[int](HashComparable[int](), EqComparable[int]()).First().IsPresent())
}

func TestSetMap(t *testing.T) {
	assertSetsAreEqual(
		t,
		SetOf(HashComparable[string](), EqComparable[string](), "1", "2", "3"),
		SetOf(HashComparable[int](), EqComparable[int](), 1, 2, 3).
			Map(HashComparable[string](), EqComparable[string](), strconv.Itoa))

	isEven := func(i int) (even bool) { return i%2 == 0 }
	assertSetsAreEqual(
		t,
		SetOf(HashComparable[bool](), EqComparable[bool](), true, false),
		SetOf(HashComparable[int](), EqComparable[int](), 1, 2, 3, 4).
			Map(HashComparable[bool](), EqComparable[bool](), isEven),
		"equal results must be merged")

	called := false
	empty := SetOf[int](HashComparable[int](), EqComparable[int]()).
		Map(HashComparable[string](), EqComparable[string](), func(i int) (s string) {
			called = true
			return strconv.Itoa(i)
		})
	assert.False(t, called, "Map must not call f on an empty Set")
	assertEmptySet(t, empty)
}

func TestSetFlatMap(t *testing.T) {
	divisors := func(n int) (found Set[int]) {
		var elems = make([]int, 0, n)
		for d := 1; d <= n; d++ {
			if n%d == 0 {
				elems = append(elems, d)
			}
		}
		return SetOfSlice(HashComparable[int](), EqComparable[int](), elems)
	}

	assertSetsAreEqual(
		t,
		SetOf(HashComparable[int](), EqComparable[int](), 1, 2, 3, 4, 6),
		SetOf(HashComparable[int](), EqComparable[int](), 4, 6).
			FlatMap(HashComparable[int](), EqComparable[int](), divisors),
		"results must be merged into their union")
	assertEmptySet(
		t,
		SetOf(HashComparable[int](), EqComparable[int](), 0).
			FlatMap(HashComparable[int](), EqComparable[int](), divisors),
		"f returning only empty Sets must yield empty")

	called := false
	empty := SetOf(HashComparable[int](), EqComparable[int]()).
		FlatMap(HashComparable[int](), EqComparable[int](), func(i int) (found Set[int]) {
			called = true
			return divisors(i)
		},
		)
	assert.False(t, called, "FlatMap must not call f on an empty Set")
	assertEmptySet(t, empty)
}

func TestSetFilter(t *testing.T) {
	isEven := func(i int) (even bool) { return i%2 == 0 }

	assertSetsAreEqual(
		t,
		SetOf(HashComparable[int](), EqComparable[int](), 2, 4),
		SetOf(HashComparable[int](), EqComparable[int](), 1, 2, 3, 4).
			Filter(isEven),
	)
	assertEmptySet(t, SetOf(HashComparable[int](), EqComparable[int](), 1, 3).Filter(isEven))

	source := SetOf(HashComparable[int](), EqComparable[int](), 1, 2, 3, 4)
	source.Filter(isEven)
	assertSetsAreEqual(t, SetOf(HashComparable[int](), EqComparable[int](), 1, 2, 3, 4), source, "Filter must not modify the source Set")

	called := false
	empty := SetOf[int](HashComparable[int](), EqComparable[int]()).Filter(func(i int) (keep bool) {
		called = true
		return true
	})
	assert.False(t, called, "Filter must not call predicate on an empty Set")
	assertEmptySet(t, empty)
}

func TestSetToList(t *testing.T) {
	assertListsAreEqual(t, ListOf(1, 2, 3), SetOf(HashComparable[int](), EqComparable[int](), 1, 2, 3).ToList())
	assertEmptyList(t, SetOf[int](HashComparable[int](), EqComparable[int]()).ToList())
}

func TestSetChaining(t *testing.T) {
	got := SetOf(HashComparable[int](), EqComparable[int](), 1, 2, 3, 4).
		Filter(func(i int) (even bool) { return i%2 == 0 }).
		Map(HashComparable[int](), EqComparable[int](), func(i int) (halved int) { return i / 2 }).
		FlatMap(HashComparable[int](), EqComparable[int](), func(i int) (opposite Set[int]) {
			return SetOf(HashComparable[int](), EqComparable[int](), i, -i)
		})
	assertSetsAreEqual(t, SetOf(HashComparable[int](), EqComparable[int](), 1, -1, 2, -2), got)
}

func TestSetToSequence(t *testing.T) {
	set := SetOf(HashComparable[int](), EqComparable[int](), 1, 2, 3)
	sequence := set.ToSequence()
	assertSequenceContainsAll(t, []int{1, 2, 3}, sequence)
	assertSequenceDoesNotContainsAny(t, []int{4}, sequence)

	set.Add(4)
	assertSequenceContainsAll(t, []int{1, 2, 3, 4}, sequence, "the Set must be read when evaluated, not copied")

	assertEmptySequence(t, SetOf[int](HashComparable[int](), EqComparable[int]()).ToSequence())
	var s Set[int]
	assertEmptySequence(t, s.ToSequence(), "ToSequence on the zero value must return an empty Sequence")
}
