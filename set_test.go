package monad

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

// assertEmptySet checks that s is empty but not nil, so it is safe to chain.
func assertEmptySet[T comparable](t *testing.T, s Set[T], msgAndArgs ...any) {
	t.Helper()
	assert.NotNil(t, s, msgAndArgs...)
	assert.Empty(t, s, msgAndArgs...)
}

func TestSetOf(t *testing.T) {
	assert.Equal(t, SetOfSlice([]int{1, 2, 3}), SetOf(1, 2, 3))
	assert.Equal(t, SetOf(1, 2), SetOf(1, 2, 1, 2), "duplicates must be removed")
	assertEmptySet(t, SetOf[int]())
}

func TestSetOfSlice(t *testing.T) {
	set := SetOfSlice([]string{"a", "b", "a"})
	assert.Len(t, set, 2, "duplicates must be removed")
	assert.Contains(t, set, "a")
	assert.Contains(t, set, "b")

	assertEmptySet(t, SetOfSlice[int](nil))
	assertEmptySet(t, SetOfSlice([]int{}))
}

func TestSetZeroValueIsEmpty(t *testing.T) {
	var s Set[string]
	assert.True(t, s.IsEmpty())
	assert.False(t, s.First().IsPresent())
	assertEmptySet(t, s.Map(strconv.Quote), "methods on the zero value must return non-nil Sets")
	assertEmptySet(t, s.Filter(func(e string) (keep bool) { return true }))
	assertEmptySet(t, s.FlatMap(func(e string) (single Set[string]) { return SetOf(e) }))
	assertEmptyList(t, s.ToList())
}

func TestSetIsEmpty(t *testing.T) {
	assert.False(t, SetOf(0).IsEmpty(), "a Set holding the zero value is not empty")
	assert.True(t, SetOf[int]().IsEmpty())
}

func TestSetFirst(t *testing.T) {
	set := SetOf(1, 2, 3)
	first := set.First().GetOrNil()
	if assert.NotNil(t, first) {
		assert.Contains(t, set, *first, "First must return an element of the Set")
	}
	assert.False(t, SetOf[int]().First().IsPresent())
}

func TestSetMap(t *testing.T) {
	assert.Equal(t, SetOf("1", "2", "3"), SetOf(1, 2, 3).Map(strconv.Itoa))

	isEven := func(i int) (even bool) { return i%2 == 0 }
	assert.Equal(t, SetOf(true, false), SetOf(1, 2, 3, 4).Map(isEven), "equal results must be merged")

	called := false
	empty := SetOf[int]().Map(func(i int) (s string) {
		called = true
		return strconv.Itoa(i)
	})
	assert.False(t, called, "Map must not call f on an empty Set")
	assertEmptySet(t, empty)
}

func TestSetFlatMap(t *testing.T) {
	divisors := func(n int) (found Set[int]) {
		elems := []int{}
		for d := 1; d <= n; d++ {
			if n%d == 0 {
				elems = append(elems, d)
			}
		}
		return SetOfSlice(elems)
	}

	assert.Equal(t, SetOf(1, 2, 3, 4, 6), SetOf(4, 6).FlatMap(divisors), "results must be merged into their union")
	assertEmptySet(t, SetOf(0).FlatMap(divisors), "f returning only empty Sets must yield empty")

	called := false
	empty := SetOf[int]().FlatMap(func(i int) (found Set[int]) {
		called = true
		return divisors(i)
	})
	assert.False(t, called, "FlatMap must not call f on an empty Set")
	assertEmptySet(t, empty)
}

func TestSetFilter(t *testing.T) {
	isEven := func(i int) (even bool) { return i%2 == 0 }

	assert.Equal(t, SetOf(2, 4), SetOf(1, 2, 3, 4).Filter(isEven))
	assertEmptySet(t, SetOf(1, 3).Filter(isEven))

	source := SetOf(1, 2, 3, 4)
	source.Filter(isEven)
	assert.Equal(t, SetOf(1, 2, 3, 4), source, "Filter must not modify the source Set")

	called := false
	empty := SetOf[int]().Filter(func(i int) (keep bool) {
		called = true
		return true
	})
	assert.False(t, called, "Filter must not call predicate on an empty Set")
	assertEmptySet(t, empty)
}

func TestSetToList(t *testing.T) {
	assert.ElementsMatch(t, ListOf(1, 2, 3), SetOf(1, 2, 3).ToList())
	assertEmptyList(t, SetOf[int]().ToList())
}

func TestSetChaining(t *testing.T) {
	got := SetOf(1, 2, 3, 4).
		Filter(func(i int) (even bool) { return i%2 == 0 }).
		Map(func(i int) (halved int) { return i / 2 }).
		FlatMap(func(i int) (pair Set[int]) { return SetOf(i, -i) }).
		ToList()
	assert.ElementsMatch(t, ListOf(1, -1, 2, -2), got)
}

func TestSetToSequence(t *testing.T) {
	set := SetOf(1, 2, 3)
	sequence := set.ToSequence()
	assert.ElementsMatch(t, ListOf(1, 2, 3), sequence.ToList())

	set[4] = struct{}{}
	assert.ElementsMatch(t, ListOf(1, 2, 3, 4), sequence.ToList(), "the Set must be read when evaluated, not copied")

	assertEmptySequence(t, SetOf[int]().ToSequence())
	var s Set[int]
	assertEmptySequence(t, s.ToSequence(), "ToSequence on the zero value must return an empty Sequence")
}
