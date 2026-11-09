package monad

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOptionOf(t *testing.T) {
	elem := new(42)
	assert.Same(t, elem, OptionOf(elem).GetOrNil(), "OptionOf must keep the given pointer")
	assert.Nil(t, OptionOf[int](nil).GetOrNil(), "OptionOf(nil) must be empty")
}

func TestOptionZeroValueIsEmpty(t *testing.T) {
	var o Option[string]
	assert.Nil(t, o.GetOrNil())
	assert.False(t, o.IsPresent())
}

func TestOptionIsPresent(t *testing.T) {
	assert.True(t, OptionOf(new(0)).IsPresent(), "a pointer to the zero value is present")
	assert.False(t, OptionOf[int](nil).IsPresent())
}

func TestOptionGetOrNil(t *testing.T) {
	elem := new(0)
	assert.Same(t, elem, OptionOf(elem).GetOrNil(), "GetOrNil must return the held pointer, not a copy")
	assert.Nil(t, OptionOf[int](nil).GetOrNil())
}

func TestOptionGetOrDefault(t *testing.T) {
	assert.Equal(t, 1, OptionOf(new(1)).GetOrDefault(2))
	assert.Equal(t, 2, OptionOf[int](nil).GetOrDefault(2))
}

func TestOptionMap(t *testing.T) {
	assert.Equal(t, "42", OptionOf(new(42)).Map(strconv.Itoa).GetOrDefault("none"))

	called := false
	empty := OptionOf[int](nil).Map(func(i int) (s string) {
		called = true
		return strconv.Itoa(i)
	})
	assert.False(t, called, "Map must not call f on an empty Option")
	assert.False(t, empty.IsPresent())
}

func TestOptionFlatMap(t *testing.T) {
	parse := func(s string) (parsed Option[int]) {
		i, err := strconv.Atoi(s)
		if err != nil {
			return OptionOf[int](nil)
		}
		return OptionOf(&i)
	}

	assert.Equal(t, 7, OptionOf(new("7")).FlatMap(parse).GetOrDefault(-1))
	assert.False(t, OptionOf(new("x")).FlatMap(parse).IsPresent(), "f returning empty must yield empty")

	called := false
	empty := OptionOf[string](nil).FlatMap(func(s string) (parsed Option[int]) {
		called = true
		return parse(s)
	})
	assert.False(t, called, "FlatMap must not call f on an empty Option")
	assert.False(t, empty.IsPresent())
}

func TestOptionFilter(t *testing.T) {
	isEven := func(i int) (even bool) { return i%2 == 0 }

	elem := new(4)
	assert.Same(t, elem, OptionOf(elem).Filter(isEven).GetOrNil(), "a matching value must be kept as is")
	assert.False(t, OptionOf(new(3)).Filter(isEven).IsPresent())

	called := false
	empty := OptionOf[int](nil).Filter(func(i int) (keep bool) {
		called = true
		return true
	})
	assert.False(t, called, "Filter must not call predicate on an empty Option")
	assert.False(t, empty.IsPresent())
}

func TestOptionToList(t *testing.T) {
	assert.Equal(t, ListOf(42), OptionOf(new(42)).ToList())
	assertEmptyList(t, OptionOf[int](nil).ToList())
}

func TestOptionChaining(t *testing.T) {
	got := OptionOf(new(3)).
		Map(func(i int) (doubled int) { return i * 2 }).
		Map(strconv.Itoa).
		FlatMap(func(s string) (exclaimed Option[string]) { r := s + "!"; return OptionOf(&r) }).
		GetOrDefault("none")
	assert.Equal(t, "6!", got)
}

func TestOptionToSequence(t *testing.T) {
	assert.Equal(t, ListOf(42), OptionOf(new(42)).ToSequence().ToList())
	assertEmptySequence(t, OptionOf[int](nil).ToSequence())
}

func TestOptionString(t *testing.T) {
	assert.Equal(t, "42", OptionOf(new(42)).String())
	assert.Equal(t, "0", OptionOf(new(0)).String())
	assert.Equal(t, "<nil>", OptionOf[int](nil).String())
}

func TestOptionToResult(t *testing.T) {
	elem := new(42)
	called := false
	present := OptionOf(elem).ToResult(func() (err error) {
		called = true
		return errBoom
	})
	assert.False(t, called, "ToResult must not call ifEmpty on a present Option")
	assertSuccess(t, present, 42)
	assert.Same(t, elem, present.GetOrNil(), "ToResult must keep the given pointer")

	assertFailure(t, OptionOf[int](nil).ToResult(func() (err error) { return errBoom }), errBoom)
	assertEmptySuccess(t, OptionOf[int](nil).ToResult(func() (err error) { return nil }))
}
