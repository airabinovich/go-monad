package monad

import (
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// collidingHash puts every element in the same bucket.
func collidingHash(int) (h uint64) { return 0 }

// assertEmptySet checks that s is not nil and holds no elements.
func assertEmptySet[T any](t *testing.T, s *Set[T], msgAndArgs ...any) {
	t.Helper()
	require.NotNil(t, s, msgAndArgs...)
	assert.Equal(t, 0, s.Size(), msgAndArgs...)
}

func assertSetsAreEqual[T any](t *testing.T, s1 *Set[T], s2 *Set[T], msgAndArgs ...any) {
	t.Helper()
	assert.True(t, s1.Equal(s2), msgAndArgs...)
}

func assertSetsAreNotEqual[T any](t *testing.T, s1 *Set[T], s2 *Set[T], msgAndArgs ...any) {
	t.Helper()
	assert.False(t, s1.Equal(s2), msgAndArgs...)
}

func TestSetOf(t *testing.T) {
	assertSetsAreEqual(t, SetOfSliceFunc(HashComparable[int](), EqComparable[int](), []int{1, 2, 3}), SetOf(1, 2, 3))
	assertSetsAreEqual(t, SetOf(1, 2), SetOf(1, 2, 1, 2), "duplicates must be removed")
	assertEmptySet(t, SetOfFunc[int](HashComparable[int](), EqComparable[int]()))
}

func TestSetOfFunc(t *testing.T) {
	assertSetsAreEqual(t,
		SetOfFunc[int](HashComparable[int](), EqComparable[int](), 1, 2, 3),
		SetOf(1, 2, 3))
	assertSetsAreEqual(t,
		SetOfFunc[string](HashComparable[string](), EqComparable[string](), "a", "b", "c"),
		SetOf("a", "b", "c"))
}

func TestSetOfSlice(t *testing.T) {
	assertSetsAreEqual(t, SetOf(1, 2), SetOfSlice([]int{1, 2, 1}), "duplicates must be removed")
	assertEmptySet(t, SetOfSlice[int](nil))
}

func TestSetOfSliceFunc(t *testing.T) {
	set := SetOfSliceFunc(HashComparable[string](), EqComparable[string](), []string{"a", "b", "a"})
	assert.Equal(t, 2, set.Size(), "duplicates must be removed")
	assert.True(t, set.Contains("a"))
	assert.True(t, set.Contains("b"))

	assertEmptySet(t, SetOfSliceFunc[int](HashComparable[int](), EqComparable[int](), nil))
	assertEmptySet(t, SetOfSliceFunc(HashComparable[int](), EqComparable[int](), []int{}))
}

func TestSetSize(t *testing.T) {
	hashFunc := func(i int) (s uint64) {
		if i > 5 {
			return 1
		}
		return uint64(i)
	}
	set := SetOfFunc(hashFunc, EqComparable[int](), 1, 2, 3, 8, 9, 0)
	assert.Equal(t, 6, set.Size(), "Size must count elements, not hash buckets")
}

func TestSetZeroValue(t *testing.T) {
	var s Set[string]
	hashFunc, eqFunc := HashComparable[string](), EqComparable[string]()

	assert.True(t, s.IsEmpty())
	assert.False(t, s.Contains("a"))
	assert.Equal(t, "{}", s.String())
	assertEmptySet(t, s.Map(hashFunc, eqFunc, strconv.Quote), "Map takes its own functions, so it works on the zero value")
	assertEmptySet(t, s.FlatMap(hashFunc, eqFunc, func(e string) (single *Set[string]) {
		return SetOfFunc(hashFunc, eqFunc, e)
	}))
	assertEmptyList(t, s.ToList())
	assertEmptySequence(t, s.ToSequence())
}

// The zero value has no HashFunc and EqFunc, so the operations that need to
// build a Set with the receiver's functions panic. This is a known limitation.
func TestSetZeroValuePanics(t *testing.T) {
	var zero Set[int]
	called := false
	assert.Panics(t, func() {
		zero.Filter(func(int) (keep bool) { called = true; return true })
	}, "Filter on the zero value must panic")
	assert.False(t, called, "Filter must not call predicate on an empty Set")
	assert.Panics(t, func() { zero.Add(1) }, "Add on the zero value must panic")
}

// A nil *Set is accepted by the read-only methods, but Add and Filter cannot work on it.
func TestSetNilPanics(t *testing.T) {
	var nilSet *Set[int]
	assert.Panics(t, func() { nilSet.Add(1) }, "Add on a nil Set must panic")

	called := false
	assert.Panics(t, func() {
		nilSet.Filter(func(int) (keep bool) { called = true; return true })
	}, "Filter on a nil Set must panic")
	assert.False(t, called, "Filter must not call predicate on an empty Set")
}

func TestHashComparable(t *testing.T) {
	hash := HashComparable[string]()
	assert.Equal(t, hash("a"), hash("a"), "equal values must have the same hash")
	assert.NotEqual(t, hash("a"), hash("b"))

	hashAny := HashComparable[any]()
	assert.Panics(t, func() { hashAny([]int{1}) }, "an interface holding a non-comparable value must panic")
}

func TestEqComparable(t *testing.T) {
	eq := EqComparable[string]()
	assert.True(t, eq("a", "a"))
	assert.False(t, eq("a", "b"))
}

func TestSetNilFunctionsPanic(t *testing.T) {
	hashFunc, eqFunc := HashComparable[int](), EqComparable[int]()

	assert.Panics(t, func() { SetOfFunc[int](nil, eqFunc) }, "nil HashFunc")
	assert.Panics(t, func() { SetOfFunc[int](hashFunc, nil) }, "nil EqFunc")
	assert.Panics(t, func() { SetOfFunc[int](nil, nil, 1) }, "nil HashFunc and EqFunc")
	assert.Panics(t, func() { SetOfSliceFunc[int](nil, eqFunc, nil) }, "nil HashFunc, even with no elements")
	assert.Panics(t, func() { SetOfSliceFunc[int](hashFunc, nil, []int{1}) }, "nil EqFunc")

	assert.Panics(t, func() {
		SetOf(1).Map(nil, eqFunc, func(i int) (same int) { return i })
	}, "Map with a nil HashFunc")
	assert.Panics(t, func() {
		SetOf(1).Map(hashFunc, nil, func(i int) (same int) { return i })
	}, "Map with a nil EqFunc")
	assert.Panics(t, func() {
		SetOf[int]().Map(nil, nil, func(i int) (same int) { return i })
	}, "Map must validate its functions even on an empty Set")
	assert.Panics(t, func() {
		SetOf(1).FlatMap(nil, eqFunc, func(i int) (single *Set[int]) { return SetOf(i) })
	}, "FlatMap with a nil HashFunc")
	assert.Panics(t, func() {
		SetOf(1).FlatMap(hashFunc, nil, func(i int) (single *Set[int]) { return SetOf(i) })
	}, "FlatMap with a nil EqFunc")
}

func TestSetNilIsEmpty(t *testing.T) {
	var s *Set[string]
	hashFunc, eqFunc := HashComparable[string](), EqComparable[string]()

	assert.Equal(t, 0, s.Size())
	assert.True(t, s.IsEmpty())
	assert.False(t, s.Contains("a"))
	assert.False(t, s.First().IsPresent())
	assert.Nil(t, s.FirstOrNil())
	assert.Equal(t, "{}", s.String())
	assertEmptyList(t, s.ToList())
	assertEmptySequence(t, s.ToSequence())
	assertEmptySet(t, s.Map(hashFunc, eqFunc, strconv.Quote), "methods on a nil Set must return non-nil Sets")
	assertEmptySet(t, s.FlatMap(hashFunc, eqFunc, func(e string) (single *Set[string]) { return SetOfFunc(hashFunc, eqFunc, e) }))

	assertSetsAreEqual(t, s, SetOfFunc(hashFunc, eqFunc), "a nil Set equals an empty Set")
	assertSetsAreEqual(t, SetOfFunc(hashFunc, eqFunc), s, "a nil Set equals an empty Set")
	assertSetsAreEqual(t, s, s)
	assertSetsAreNotEqual(t, s, SetOfFunc(hashFunc, eqFunc, "a"))
	assertSetsAreNotEqual(t, SetOfFunc(hashFunc, eqFunc, "a"), s)
}

func TestSetFlatMapTreatsNilAsEmpty(t *testing.T) {
	got := SetOf(1, 2, 3).
		FlatMap(HashComparable[int](), EqComparable[int](), func(i int) (res *Set[int]) {
			if i == 2 {
				return nil
			}
			return SetOf(i)
		})
	assertSetsAreEqual(t, SetOf(1, 3), got, "a nil Set returned by f must be treated as empty")
}

func TestSetIsEmpty(t *testing.T) {
	assert.False(t, SetOf(0).IsEmpty(), "a Set holding the zero value is not empty")
	assert.True(t, SetOfFunc[int](HashComparable[int](), EqComparable[int]()).IsEmpty())
}

func TestSetFirst(t *testing.T) {
	set := SetOf(1, 2, 3)
	first := set.First().GetOrNil()
	if assert.NotNil(t, first) {
		assert.True(t, set.Contains(*first), "First must return an element of the Set")
	}
	assert.False(t, SetOfFunc[int](HashComparable[int](), EqComparable[int]()).First().IsPresent())
}

func TestSetFirstOrNil(t *testing.T) {
	set := SetOf(1, 2, 3)
	first := set.FirstOrNil()
	if assert.NotNil(t, first) {
		assert.True(t, set.Contains(*first), "FirstOrNil must return an element of the Set")
	}
	assert.Nil(t, SetOfFunc[int](HashComparable[int](), EqComparable[int]()).FirstOrNil(), "FirstOrNil on an empty Set must return nil")
}

func TestSetMap(t *testing.T) {
	assertSetsAreEqual(
		t,
		SetOfFunc(HashComparable[string](), EqComparable[string](), "1", "2", "3"),
		SetOf(1, 2, 3).
			Map(HashComparable[string](), EqComparable[string](), strconv.Itoa))

	isEven := func(i int) (even bool) { return i%2 == 0 }
	assertSetsAreEqual(
		t,
		SetOfFunc(HashComparable[bool](), EqComparable[bool](), true, false),
		SetOf(1, 2, 3, 4).
			Map(HashComparable[bool](), EqComparable[bool](), isEven),
		"equal results must be merged")

	called := false
	empty := SetOfFunc[int](HashComparable[int](), EqComparable[int]()).
		Map(HashComparable[string](), EqComparable[string](), func(i int) (s string) {
			called = true
			return strconv.Itoa(i)
		})
	assert.False(t, called, "Map must not call f on an empty Set")
	assertEmptySet(t, empty)
}

func TestSetFlatMap(t *testing.T) {
	divisors := func(n int) (found *Set[int]) {
		var elems = make([]int, 0, n)
		for d := 1; d <= n; d++ {
			if n%d == 0 {
				elems = append(elems, d)
			}
		}
		return SetOfSliceFunc(HashComparable[int](), EqComparable[int](), elems)
	}

	assertSetsAreEqual(
		t,
		SetOf(1, 2, 3, 4, 6),
		SetOf(4, 6).
			FlatMap(HashComparable[int](), EqComparable[int](), divisors),
		"results must be merged into their union")
	assertEmptySet(
		t,
		SetOf(0).
			FlatMap(HashComparable[int](), EqComparable[int](), divisors),
		"f returning only empty Sets must yield empty")

	called := false
	empty := SetOfFunc(HashComparable[int](), EqComparable[int]()).
		FlatMap(HashComparable[int](), EqComparable[int](), func(i int) (found *Set[int]) {
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
		SetOf(2, 4),
		SetOf(1, 2, 3, 4).
			Filter(isEven),
	)
	assertEmptySet(t, SetOf(1, 3).Filter(isEven))

	source := SetOf(1, 2, 3, 4)
	source.Filter(isEven)
	assertSetsAreEqual(t, SetOf(1, 2, 3, 4), source, "Filter must not modify the source Set")

	filtered := source.Filter(isEven)
	filtered.Add(6)
	assert.False(t, source.Contains(6), "the result of Filter must not share storage with the source Set")

	called := false
	empty := SetOfFunc[int](HashComparable[int](), EqComparable[int]()).Filter(func(i int) (keep bool) {
		called = true
		return true
	})
	assert.False(t, called, "Filter must not call predicate on an empty Set")
	assertEmptySet(t, empty)
}

func TestSetToList(t *testing.T) {
	set := SetOf(1, 2, 3)
	assert.ElementsMatch(t, []int{1, 2, 3}, slices.Collect(set.ToList().All()), "ToList returns the elements in unspecified order")
	assertEmptyList(t, SetOfFunc[int](HashComparable[int](), EqComparable[int]()).ToList())
}

func TestSetChaining(t *testing.T) {
	got := SetOf(1, 2, 3, 4).
		Filter(func(i int) (even bool) { return i%2 == 0 }).
		Map(HashComparable[int](), EqComparable[int](), func(i int) (halved int) { return i / 2 }).
		FlatMap(HashComparable[int](), EqComparable[int](), func(i int) (opposite *Set[int]) {
			return SetOf(i, -i)
		})
	assertSetsAreEqual(t, SetOf(1, -1, 2, -2), got)
}

func TestSetToSequence(t *testing.T) {
	set := SetOf(1, 2, 3)
	sequence := set.ToSequence()
	assertSequenceElements(t, []int{1, 2, 3}, sequence)

	set.Add(4)
	assertSequenceElements(t, []int{1, 2, 3, 4}, sequence, "the Set must be read when evaluated, not copied")

	assertEmptySequence(t, SetOfFunc[int](HashComparable[int](), EqComparable[int]()).ToSequence())
	var s Set[int]
	assertEmptySequence(t, s.ToSequence(), "ToSequence on the zero value must return an empty Sequence")
	var nilSet *Set[int]
	assertEmptySequence(t, nilSet.ToSequence(), "ToSequence on a nil Set must return an empty Sequence")
}

func TestSetAdd(t *testing.T) {
	set := SetOfFunc(HashComparable[int](), EqComparable[int]())
	assert.Same(t, set, set.Add(1), "Add must return the Set")

	set.Add(1)
	assert.Equal(t, 1, set.Size(), "adding a duplicate must not change the Set")
	set.Add(2)
	assert.Equal(t, 2, set.Size())
	assert.True(t, set.Contains(2), "Add must modify the Set in place")
}

func TestSetAddWithCollidingHash(t *testing.T) {
	set := SetOfFunc(collidingHash, EqComparable[int](), 1, 2, 3)
	set.Add(2)
	assert.Equal(t, 3, set.Size(), "a duplicate with a colliding hash must not be added")
	set.Add(4)
	assert.Equal(t, 4, set.Size(), "a new element with a colliding hash must be added")
}

func TestSetAddUsesEqFunc(t *testing.T) {
	hashFunc := func(s string) (h uint64) { return HashComparable[string]()(strings.ToLower(s)) }
	set := SetOfFunc(hashFunc, strings.EqualFold, "Go")
	set.Add("GO")
	assert.Equal(t, 1, set.Size(), "elements equal according to the EqFunc must not be added twice")
	assert.True(t, set.Contains("gO"))
}

func TestSetContains(t *testing.T) {
	set := SetOf(1, 2)
	assert.True(t, set.Contains(1))
	assert.False(t, set.Contains(3))
	assert.False(t, SetOfFunc(HashComparable[int](), EqComparable[int]()).Contains(0))

	colliding := SetOfFunc(collidingHash, EqComparable[int](), 1, 2)
	assert.True(t, colliding.Contains(2))
	assert.False(t, colliding.Contains(3), "an element with the same hash is not necessarily in the Set")
}

func TestSetEqual(t *testing.T) {
	newSet := func(elems ...int) (set *Set[int]) {
		return SetOf(elems...)
	}

	assertSetsAreEqual(t, newSet(), newSet())
	assertSetsAreEqual(t, newSet(1, 2, 3), newSet(3, 2, 1), "order must not matter")
	assertSetsAreNotEqual(t, newSet(1, 2), newSet(1, 2, 3))
	assertSetsAreNotEqual(t, newSet(1, 2, 3), newSet(1, 2, 4), "same size, different elements")
	assertSetsAreNotEqual(t, newSet(), newSet(0))

	assertSetsAreEqual(
		t,
		newSet(1, 2, 3, 4),
		SetOfFunc(collidingHash, EqComparable[int](), 1, 2, 3, 4),
		"Sets with different hash functions must be compared by their elements")
}

func TestSetString(t *testing.T) {
	assert.Equal(t, "{}", SetOfFunc(HashComparable[int](), EqComparable[int]()).String())
	assert.Equal(t, "{1}", SetOf(1).String())
	assert.Equal(t, "{1,2,3}", SetOfFunc(collidingHash, EqComparable[int](), 1, 2, 3).String(), "elements in the same bucket keep insertion order")
}
