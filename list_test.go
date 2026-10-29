package monad

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

// assertEmptyList checks that l is empty but not nil, so it is safe to chain.
func assertEmptyList[T comparable](t *testing.T, l List[T], msgAndArgs ...any) {
	t.Helper()
	assert.NotNil(t, l, msgAndArgs...)
	assert.Empty(t, l, msgAndArgs...)
}

func TestListOf(t *testing.T) {
	assert.Equal(t, ListOfSlice([]int{1, 2, 3}), ListOf(1, 2, 3))
	assertEmptyList(t, ListOf[int]())

	elems := []int{1, 2, 3}
	assert.Same(t, &elems[0], &ListOf(elems...)[0], "ListOf(s...) must not copy s")
}

func TestListOfSlice(t *testing.T) {
	elems := []int{1, 2, 3}
	list := ListOfSlice(elems)
	assert.Equal(t, ListOf(1, 2, 3), list)
	assert.Same(t, &elems[0], &list[0], "ListOfSlice must not copy elems")

	assertEmptyList(t, ListOfSlice[int](nil))
	assertEmptyList(t, ListOfSlice([]int{}))
}

func TestListZeroValueIsEmpty(t *testing.T) {
	var l List[string]
	assert.True(t, l.IsEmpty())
	assert.False(t, l.First().IsPresent())
	assertEmptyList(t, l.Map(strconv.Quote), "methods on the zero value must return non-nil Lists")
	assertEmptyList(t, l.Filter(func(s string) (keep bool) { return true }))
	assertEmptyList(t, l.FlatMap(func(s string) (pair List[string]) { return ListOf(s, s) }))
}

func TestListIsEmpty(t *testing.T) {
	assert.False(t, ListOf(0).IsEmpty(), "a List holding the zero value is not empty")
	assert.True(t, ListOf[int]().IsEmpty())
}

func TestListFirst(t *testing.T) {
	list := ListOf(1, 2, 3)
	assert.Same(t, &list[0], list.First().GetOrNil(), "First must point into the List, not copy")
	assert.False(t, ListOf[int]().First().IsPresent())
}

func TestListFirstOrNil(t *testing.T) {
	list := ListOf(1, 2, 3)
	assert.Same(t, &list[0], list.FirstOrNil(), "First must point into the List, not copy")
}

func TestListMap(t *testing.T) {
	assert.Equal(t, ListOf("1", "2", "3"), ListOf(1, 2, 3).Map(strconv.Itoa))

	called := false
	empty := ListOf[int]().Map(func(i int) (s string) {
		called = true
		return strconv.Itoa(i)
	})
	assert.False(t, called, "Map must not call f on an empty List")
	assertEmptyList(t, empty)
}

func TestListFlatMap(t *testing.T) {
	repeat := func(i int) (repeated List[int]) {
		elems := make([]int, i)
		for j := range elems {
			elems[j] = i
		}
		return ListOfSlice(elems)
	}

	assert.Equal(t, ListOf(1, 2, 2, 3, 3, 3), ListOf(1, 0, 2, 3).FlatMap(repeat), "results must be concatenated in order")
	assertEmptyList(t, ListOf(0, 0).FlatMap(repeat), "f returning only empty Lists must yield empty")

	called := false
	empty := ListOf[int]().FlatMap(func(i int) (repeated List[int]) {
		called = true
		return repeat(i)
	})
	assert.False(t, called, "FlatMap must not call f on an empty List")
	assertEmptyList(t, empty)
}

func TestListFilter(t *testing.T) {
	isEven := func(i int) (even bool) { return i%2 == 0 }

	assert.Equal(t, ListOf(2, 4), ListOf(1, 2, 3, 4).Filter(isEven))
	assertEmptyList(t, ListOf(1, 3).Filter(isEven))

	source := ListOf(1, 2, 3, 4)
	source.Filter(isEven)
	assert.Equal(t, ListOf(1, 2, 3, 4), source, "Filter must not modify the source List")

	called := false
	empty := ListOf[int]().Filter(func(i int) (keep bool) {
		called = true
		return true
	})
	assert.False(t, called, "Filter must not call predicate on an empty List")
	assertEmptyList(t, empty)
}

func TestListChaining(t *testing.T) {
	got := ListOf(1, 2, 3, 4).
		Filter(func(i int) (even bool) { return i%2 == 0 }).
		Map(strconv.Itoa).
		FlatMap(func(s string) (pair List[string]) { return ListOf(s, s+"!") }).
		First().
		GetOrDefault("none")
	assert.Equal(t, "2", got)
}

func TestListToSet(t *testing.T) {
	assert.Equal(t, SetOf(1, 2, 3), ListOf(1, 2, 2, 3, 1).ToSet(), "duplicates must be removed")
	assertEmptySet(t, ListOf[int]().ToSet())

	var l List[int]
	assertEmptySet(t, l.ToSet(), "ToSet on the zero value must return a non-nil Set")
}

func TestListToSequence(t *testing.T) {
	list := ListOf(1, 2, 2, 3)
	sequence := list.ToSequence()
	assert.Equal(t, list, sequence.ToList(), "order and duplicates must be kept")

	list[0] = 10
	assert.Equal(t, 10, sequence.First().GetOrDefault(-1), "the List must be read when evaluated, not copied")

	assertEmptySequence(t, ListOf[int]().ToSequence())
	var l List[int]
	assertEmptySequence(t, l.ToSequence(), "ToSequence on the zero value must return an empty Sequence")
}
