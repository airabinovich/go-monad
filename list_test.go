package monad

import (
	"slices"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// assertEmptyList checks that l is empty and not nil, so it is safe to chain.
func assertEmptyList[T any](t *testing.T, l *List[T], msgAndArgs ...any) {
	t.Helper()
	require.NotNil(t, l, msgAndArgs...)
	assert.Empty(t, l.elements, msgAndArgs...)
}

// assertListsAreEqual checks that actual holds the same elements as expected, in the same order.
func assertListsAreEqual[T any](t *testing.T, expected *List[T], actual *List[T], msgAndArgs ...any) {
	t.Helper()
	assert.True(t, expected.Equal(actual), msgAndArgs...)
}

// assertListsAreNotEqual checks that actual differs from expected in elements or order.
func assertListsAreNotEqual[T any](t *testing.T, expected *List[T], actual *List[T], msgAndArgs ...any) {
	t.Helper()
	assert.False(t, expected.Equal(actual), msgAndArgs...)
}

func TestListOf(t *testing.T) {
	assertListsAreEqual(t, ListOfSlice([]int{1, 2, 3}), ListOf(1, 2, 3))
	assertEmptyList(t, ListOf[int]())

	elems := []int{1, 2, 3}
	list := ListOf(elems...)
	elems[0] = 10
	assertListsAreEqual(t, ListOf(10, 2, 3), list, "ListOf(s...) must not copy s")
}

func TestListOfSlice(t *testing.T) {
	elems := []int{1, 2, 3}
	list := ListOfSlice(elems)
	assertListsAreEqual(t, ListOf(1, 2, 3), list)
	elems[0] = 10
	assertListsAreEqual(t, ListOf(10, 2, 3), list, "ListOfSlice must not copy elems")

	assertEmptyList(t, ListOfSlice[int](nil))
	assertEmptyList(t, ListOfSlice([]int{}))
}

func TestListZeroValueIsEmpty(t *testing.T) {
	var l List[string]
	assert.True(t, l.IsEmpty())
	assert.False(t, l.First().IsPresent())
	assertEmptyList(t, l.Map(strconv.Quote), "methods on the zero value must return non-nil Lists")
	assertEmptyList(t, l.Filter(func(s string) (keep bool) { return true }))
	assertEmptyList(t, l.FlatMap(func(s string) (pair *List[string]) { return ListOf(s, s) }))
}

func TestListIsEmpty(t *testing.T) {
	assert.False(t, ListOf(0).IsEmpty(), "a List holding the zero value is not empty")
	assert.True(t, ListOf[int]().IsEmpty())
}

func TestListEqual(t *testing.T) {
	assertListsAreEqual(t, ListOf[int](), ListOf[int]())
	assertListsAreEqual(t, ListOf[int](1, 2, 3), ListOf[int](1, 2, 3))
	assertListsAreEqual(t, ListOf([]int{1}, []int{2}), ListOf([]int{1}, []int{2}), "elements must be compared with reflect.DeepEqual")

	var nilList *List[int]
	assertListsAreEqual(t, ListOf[int](), nilList, "a nil List equals an empty List")
	assertListsAreEqual(t, nilList, ListOf[int](), "a nil List equals an empty List")
	assertListsAreEqual(t, nilList, nilList)
}

func TestListNotEqual(t *testing.T) {
	assertListsAreNotEqual(t, ListOf[int](1), ListOf[int]())
	assertListsAreNotEqual(t, ListOf[int](1, 2), ListOf[int](2, 1), "order matters")
	assertListsAreNotEqual(t, ListOf[int](1, 2, 3), ListOf[int](1, 2, 4), "same length, different elements")
	assertListsAreNotEqual(t, ListOf[int](1, 1, 2), ListOf[int](1, 2, 2), "duplicate counts matter")

	var nilList *List[int]
	assertListsAreNotEqual(t, nilList, ListOf[int](1))
	assertListsAreNotEqual(t, ListOf[int](1), nilList)
}

func TestListFirst(t *testing.T) {
	list := ListOf(1, 2, 3)
	first := list.First().GetOrNil()
	require.NotNil(t, first)
	assert.Equal(t, 1, *first)

	*first = 10
	assert.Equal(t, 1, list.elements[0], "First must return a copy, changing it must not change the List")
	assert.False(t, ListOf[int]().First().IsPresent())
}

func TestListFirstOrNil(t *testing.T) {
	list := ListOf(1, 2, 3)
	first := list.FirstOrNil()
	require.NotNil(t, first)
	assert.Equal(t, 1, *first)

	*first = 10
	assert.Equal(t, 1, list.elements[0], "FirstOrNil must return a copy, changing it must not change the List")
	assert.NotSame(t, list.FirstOrNil(), list.FirstOrNil(), "every call must return a new pointer")
	assert.Nil(t, ListOf[int]().FirstOrNil(), "FirstOrNil on an empty List must return nil")

	var nilList *List[int]
	assert.Nil(t, nilList.FirstOrNil(), "FirstOrNil on a nil List must return nil")
}

func TestListMap(t *testing.T) {
	assertListsAreEqual(t, ListOf("1", "2", "3"), ListOf(1, 2, 3).Map(strconv.Itoa))

	called := false
	empty := ListOf[int]().Map(func(i int) (s string) {
		called = true
		return strconv.Itoa(i)
	})
	assert.False(t, called, "Map must not call f on an empty List")
	assertEmptyList(t, empty)
}

func TestListFlatMap(t *testing.T) {
	repeat := func(i int) (repeated *List[int]) {
		elems := make([]int, i)
		for j := range elems {
			elems[j] = i
		}
		return ListOfSlice(elems)
	}

	assert.Equal(t, ListOf(1, 2, 2, 3, 3, 3), ListOf(1, 0, 2, 3).FlatMap(repeat), "results must be concatenated in order")
	assertEmptyList(t, ListOf(0, 0).FlatMap(repeat), "f returning only empty Lists must yield empty")

	called := false
	empty := ListOf[int]().FlatMap(func(i int) (repeated *List[int]) {
		called = true
		return repeat(i)
	})
	assert.False(t, called, "FlatMap must not call f on an empty List")
	assertEmptyList(t, empty)
}

func TestListFilter(t *testing.T) {
	isEven := func(i int) (even bool) { return i%2 == 0 }

	assertListsAreEqual(t, ListOf(2, 4), ListOf(1, 2, 3, 4).Filter(isEven), "should contain only even numbers")
	assertEmptyList(t, ListOf(1, 3).Filter(isEven))

	source := ListOf(1, 2, 3, 4)
	source.Filter(isEven)
	assertListsAreEqual(t, ListOf(1, 2, 3, 4), source, "Filter must not modify the source List")

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
		FlatMap(func(s string) (pair *List[string]) { return ListOf(s, s+"!") }).
		First().
		GetOrDefault("none")
	assert.Equal(t, "2", got)
}

func TestListToSet(t *testing.T) {
	assertSetsAreEqual(
		t,
		SetOf(HashComparable[int](), EqComparable[int](), 1, 2, 3),
		ListOf(1, 2, 2, 3, 1).ToSet(HashComparable[int](), EqComparable[int]()),
		"duplicates must be removed",
	)
	assertEmptySet(t, ListOf[int]().ToSet(HashComparable[int](), EqComparable[int]()))

	var l List[int]
	assertEmptySet(t, l.ToSet(HashComparable[int](), EqComparable[int]()), "ToSet on the zero value must return an empty Set")
}

func TestListToSequence(t *testing.T) {
	list := ListOf(1, 2, 2, 3)
	sequence := list.ToSequence()
	assert.Equal(t, list, sequence.ToList(), "order and duplicates must be kept")

	list.elements[0] = 10
	assert.Equal(t, 10, sequence.First().GetOrDefault(-1), "the List must be read when evaluated, not copied")

	assertEmptySequence(t, ListOf[int]().ToSequence())
	var l List[int]
	assertEmptySequence(t, l.ToSequence(), "ToSequence on the zero value must return an empty Sequence")
}

func TestListLen(t *testing.T) {
	var nilList *List[int]
	assert.Equal(t, 0, nilList.Len())
	assert.Equal(t, 0, ListOf[int]().Len())
	assert.Equal(t, 3, ListOf(1, 2, 3).Len())
}

func TestListString(t *testing.T) {
	var nilList *List[int]
	assert.Equal(t, "[]", nilList.String())
	assert.Equal(t, "[]", ListOf[int]().String())
	assert.Equal(t, "[1]", ListOf(1).String())
	assert.Equal(t, "[1,2,3]", ListOf(1, 2, 3).String())
}

func TestListAdd(t *testing.T) {
	list := ListOf(1)
	assert.Same(t, list, list.Add(2), "Add must return the receiver")
	assertListsAreEqual(t, ListOf(1, 2), list)

	var zero List[int]
	zero.Add(1)
	assertListsAreEqual(t, ListOf(1), &zero, "Add must work on the zero value")

	var nilList *List[int]
	added := nilList.Add(1)
	assertListsAreEqual(t, ListOf(1), added, "Add on a nil List must return a new List")
	assert.Nil(t, nilList, "Add on a nil List must not modify the receiver")
}

func TestListContains(t *testing.T) {
	list := ListOf(1, 2, 3)
	assert.True(t, list.Contains(2))
	assert.False(t, list.Contains(4))
	assert.False(t, ListOf[int]().Contains(0))

	var nilList *List[int]
	assert.False(t, nilList.Contains(1))

	nested := ListOf([]int{1, 2})
	assert.True(t, nested.Contains([]int{1, 2}), "Contains must compare with reflect.DeepEqual")
}

func TestListIterators(t *testing.T) {
	list := ListOf("a", "b", "c")

	assert.Equal(t, []string{"a", "b", "c"}, slices.Collect(list.All()))
	assert.Empty(t, slices.Collect(ListOf[string]().All()))

	var indexes []int
	var values []string
	for i, v := range list.Enumerate() {
		indexes = append(indexes, i)
		values = append(values, v)
	}
	assert.Equal(t, []int{0, 1, 2}, indexes)
	assert.Equal(t, []string{"a", "b", "c"}, values)

	indexes, values = nil, nil
	for i, v := range list.Backward() {
		indexes = append(indexes, i)
		values = append(values, v)
	}
	assert.Equal(t, []int{2, 1, 0}, indexes)
	assert.Equal(t, []string{"c", "b", "a"}, values)

	var seen []string
	for v := range list.All() {
		seen = append(seen, v)
		break
	}
	assert.Equal(t, []string{"a"}, seen, "breaking out of the loop must stop the iteration")
}

func TestListNilIsEmpty(t *testing.T) {
	var l *List[string]
	assert.Equal(t, 0, l.Len())
	assert.True(t, l.IsEmpty())
	assert.Equal(t, "[]", l.String())
	assert.False(t, l.Contains("a"))
	assert.False(t, l.First().IsPresent())
	assert.Nil(t, l.FirstOrNil())
	assertEmptyList(t, l.Map(strconv.Quote), "methods on a nil List must return non-nil Lists")
	assertEmptyList(t, l.Filter(func(s string) (keep bool) { return true }))
	assertEmptyList(t, l.FlatMap(func(s string) (pair *List[string]) { return ListOf(s, s) }))
	assertEmptySet(t, l.ToSet(HashComparable[string](), EqComparable[string]()))
	assertEmptySequence(t, l.ToSequence())

	assert.Empty(t, slices.Collect(l.All()))
	for range l.Enumerate() {
		t.Fatal("Enumerate on a nil List must not yield any element")
	}
	for range l.Backward() {
		t.Fatal("Backward on a nil List must not yield any element")
	}
}

func TestListFlatMapTreatsNilAsEmpty(t *testing.T) {
	got := ListOf(1, 2, 3).FlatMap(func(i int) (res *List[int]) {
		if i == 2 {
			return nil
		}
		return ListOf(i)
	})
	assertListsAreEqual(t, ListOf(1, 3), got, "a nil List returned by f must be treated as empty")
}
