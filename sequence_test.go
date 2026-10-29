package monad

import (
	"slices"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

// assertEmptySequence checks that s is empty and that its accessors return
// non-nil empty collections, so it is safe to chain.
func assertEmptySequence[T comparable](t *testing.T, s Sequence[T], msgAndArgs ...any) {
	t.Helper()
	assert.True(t, s.IsEmpty(), msgAndArgs...)
	assert.False(t, s.First().IsPresent(), msgAndArgs...)
	assertEmptyList(t, s.ToList(), msgAndArgs...)
	assertEmptySet(t, s.ToSet(), msgAndArgs...)
}

func TestSequenceOf(t *testing.T) {
	assert.Equal(t, ListOf(1, 2, 1, 3), SequenceOf(1, 2, 1, 3).ToList(), "order and duplicates must be kept")
	assertEmptySequence(t, SequenceOf[int]())
}

func TestSequenceOfSlice(t *testing.T) {
	elems := []int{1, 2, 3}
	sequence := SequenceOfSlice(elems)
	elems[0] = 10
	assert.Equal(t, ListOf(10, 2, 3), sequence.ToList(), "the slice must be read when evaluated, not copied")

	assertEmptySequence(t, SequenceOfSlice[int](nil))
	assertEmptySequence(t, SequenceOfSlice([]int{}))
}

func TestSequenceOfSeq(t *testing.T) {
	assert.Equal(t, ListOf(1, 2, 3), SequenceOfSeq(slices.Values([]int{1, 2, 3})).ToList())
	assertEmptySequence(t, SequenceOfSeq[int](nil))
	assertEmptySequence(t, SequenceOfSeq(slices.Values([]int{})))
}

func TestSequenceZeroValueIsEmpty(t *testing.T) {
	var s Sequence[string]
	assertEmptySequence(t, s)
	assertEmptySequence(t, s.Map(strconv.Quote), "methods on the zero value must return empty Sequences")
	assertEmptySequence(t, s.Filter(func(e string) (keep bool) { return true }))
	assertEmptySequence(t, s.FlatMap(func(e string) (single Sequence[string]) { return SequenceOf(e) }))
}

func TestSequenceIsEmpty(t *testing.T) {
	assert.False(t, SequenceOf(0).IsEmpty(), "a Sequence holding the zero value is not empty")
	assert.True(t, SequenceOf[int]().IsEmpty())

	calls := 0
	SequenceOf(1, 2, 3).Map(func(i int) (same int) { calls++; return i }).IsEmpty()
	assert.Equal(t, 1, calls, "IsEmpty must evaluate only the first element")
}

func TestSequenceFirst(t *testing.T) {
	assert.Equal(t, 1, SequenceOf(1, 2, 3).First().GetOrDefault(-1))
	assert.False(t, SequenceOf[int]().First().IsPresent())
}

func TestSequenceToSet(t *testing.T) {
	assert.Equal(t, SetOf(1, 2, 3), SequenceOf(1, 2, 2, 3, 1).ToSet(), "duplicates must be removed")
}

func TestSequenceMap(t *testing.T) {
	assert.Equal(t, ListOf("1", "2", "1"), SequenceOf(1, 2, 1).Map(strconv.Itoa).ToList())

	calls := 0
	mapped := SequenceOf(1, 2, 3).Map(func(i int) (s string) {
		calls++
		return strconv.Itoa(i)
	})
	assert.Equal(t, 0, calls, "Map must not call f before an accessor is called")
	assert.Equal(t, "1", mapped.First().GetOrDefault("none"))
	assert.Equal(t, 1, calls, "First must evaluate only the first element")
	mapped.ToList()
	assert.Equal(t, 4, calls, "every accessor call must evaluate the Sequence again")

	called := false
	empty := SequenceOf[int]().Map(func(i int) (s string) {
		called = true
		return strconv.Itoa(i)
	})
	assertEmptySequence(t, empty)
	assert.False(t, called, "Map must not call f on an empty Sequence")
}

func TestSequenceFlatMap(t *testing.T) {
	repeat := func(i int) (repeated Sequence[int]) {
		elems := make([]int, i)
		for j := range elems {
			elems[j] = i
		}
		return SequenceOfSlice(elems)
	}

	assert.Equal(t, ListOf(1, 2, 2, 3, 3, 3), SequenceOf(1, 0, 2, 3).FlatMap(repeat).ToList(), "results must be concatenated in order")
	assertEmptySequence(t, SequenceOf(0, 0).FlatMap(repeat), "f returning only empty Sequences must yield empty")

	calls := 0
	flat := SequenceOf(2, 3).FlatMap(func(i int) (repeated Sequence[int]) {
		calls++
		return repeat(i)
	})
	assert.Equal(t, 0, calls, "FlatMap must not call f before an accessor is called")
	assert.Equal(t, 2, flat.First().GetOrDefault(-1))
	assert.Equal(t, 1, calls, "First must stop inside the first inner Sequence")

	called := false
	empty := SequenceOf[int]().FlatMap(func(i int) (repeated Sequence[int]) {
		called = true
		return repeat(i)
	})
	assertEmptySequence(t, empty)
	assert.False(t, called, "FlatMap must not call f on an empty Sequence")
}

func TestSequenceFilter(t *testing.T) {
	isEven := func(i int) (even bool) { return i%2 == 0 }

	assert.Equal(t, ListOf(2, 4, 2), SequenceOf(1, 2, 3, 4, 2).Filter(isEven).ToList())
	assertEmptySequence(t, SequenceOf(1, 3).Filter(isEven))

	calls := 0
	filtered := SequenceOf(1, 2, 3, 4).Filter(func(i int) (even bool) {
		calls++
		return isEven(i)
	})
	assert.Equal(t, 0, calls, "Filter must not call predicate before an accessor is called")
	assert.Equal(t, 2, filtered.First().GetOrDefault(-1))
	assert.Equal(t, 2, calls, "First must stop at the first matching element")

	called := false
	empty := SequenceOf[int]().Filter(func(i int) (keep bool) {
		called = true
		return true
	})
	assertEmptySequence(t, empty)
	assert.False(t, called, "Filter must not call predicate on an empty Sequence")
}

func TestSequenceChaining(t *testing.T) {
	got := SequenceOf(1, 2, 3, 4).
		Filter(func(i int) (even bool) { return i%2 == 0 }).
		Map(strconv.Itoa).
		FlatMap(func(s string) (pair Sequence[string]) { return SequenceOf(s, s+"!") }).
		ToList()
	assert.Equal(t, ListOf("2", "2!", "4", "4!"), got)
}
