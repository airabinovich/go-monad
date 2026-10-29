package monad

import (
	"errors"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var errBoom = errors.New("boom")

// assertSuccess checks that r is a success holding the expected value.
func assertSuccess[T comparable](t *testing.T, r Result[T], expected T) {
	t.Helper()
	assert.True(t, r.IsSuccess(), "expected success")
	assert.False(t, r.IsFailure(), "expected no failure")
	assert.NoError(t, r.Error())
	got := r.GetOrNil()
	require.NotNil(t, got, "expected a value")
	assert.Equal(t, expected, *got)
}

// assertFailure checks that r is a failure holding the expected error.
func assertFailure[T comparable](t *testing.T, r Result[T], expected error) {
	t.Helper()
	assert.True(t, r.IsFailure(), "expected failure")
	assert.False(t, r.IsSuccess(), "expected no success")
	assert.ErrorIs(t, r.Error(), expected)
	assert.Nil(t, r.GetOrNil(), "a failure must not expose a value")
}

// assertEmptySuccess checks that r has neither a value nor an error.
func assertEmptySuccess[T comparable](t *testing.T, r Result[T]) {
	t.Helper()
	assert.True(t, r.IsSuccess(), "expected success")
	assert.NoError(t, r.Error())
	assert.Nil(t, r.GetOrNil(), "expected no value")
}

func TestResultOf(t *testing.T) {
	assertSuccess(t, ResultOf(new(42), nil), 42)
	assertSuccess(t, ResultOf(new(0), nil), 0)
	assertFailure(t, ResultOf[int](nil, errBoom), errBoom)
	assertEmptySuccess(t, ResultOf[int](nil, nil))

	both := ResultOf(new(42), errBoom)
	assert.True(t, both.IsFailure(), "an error makes the Result a failure")
	assert.ErrorIs(t, both.Error(), errBoom)
	assert.Nil(t, both.GetOrNil(), "the value must be hidden when there is an error")
}

func TestResultString(t *testing.T) {
	assert.Equal(t, "42", ResultOf(new(42), nil).String())
	assert.Equal(t, "0", ResultOf(new(0), nil).String())
	assert.Equal(t, "-7", ResultOf(new(-7), nil).String())
	assert.Equal(t, "hello", ResultOf(new("hello"), nil).String())
	assert.Equal(t, "boom", ResultOf[int](nil, errBoom).String())
	assert.Equal(t, "boom", ResultOf(new(42), errBoom).String(), "the error takes precedence over the value")
}

// Documents the current behavior: String dereferences a nil value.
// If String is changed to handle this case, update this test.
func TestResultStringWithoutValueNorErrorPanics(t *testing.T) {
	assert.Panics(t, func() { _ = ResultOf[int](nil, nil).String() })
}

func TestResultMap(t *testing.T) {
	assertSuccess(t, ResultOf(new(42), nil).Map(strconv.Itoa), "42")
	assertSuccess(t, ResultOf(new(0), nil).Map(strconv.Itoa), "0")
	assertSuccess(t, ResultOf(new(2), nil).Map(func(i int) (tripled int) { return i * 3 }), 6)

	called := false
	failed := ResultOf[int](nil, errBoom).Map(func(i int) (s string) {
		called = true
		return strconv.Itoa(i)
	})
	assert.False(t, called, "Map must not call f on a failure")
	assertFailure(t, failed, errBoom)

	called = false
	failedWithValue := ResultOf(new(42), errBoom).Map(func(i int) (s string) {
		called = true
		return strconv.Itoa(i)
	})
	assert.False(t, called, "Map must not call f when there is an error, even with a value")
	assertFailure(t, failedWithValue, errBoom)

	called = false
	empty := ResultOf[int](nil, nil).Map(func(i int) (s string) {
		called = true
		return strconv.Itoa(i)
	})
	assert.False(t, called, "Map must not call f when there is no value")
	assertEmptySuccess(t, empty)

	calls := 0
	ResultOf(new(1), nil).Map(func(i int) (same int) { calls++; return i })
	assert.Equal(t, 1, calls, "Map must call f exactly once")

	orig := ResultOf(new(1), nil)
	_ = orig.Map(func(i int) (shifted int) { return i + 100 })
	assertSuccess(t, orig, 1)
}

func TestResultFlatMap(t *testing.T) {
	parse := func(s string) (parsed Result[int]) {
		i, err := strconv.Atoi(s)
		if err != nil {
			return ResultOf[int](nil, err)
		}
		return ResultOf(&i, nil)
	}

	assertSuccess(t, ResultOf(new("42"), nil).FlatMap(parse), 42)

	notANumber := ResultOf(new("abc"), nil).FlatMap(parse)
	assert.True(t, notANumber.IsFailure(), "f returning a failure must yield a failure")
	assert.Error(t, notANumber.Error())
	assert.Nil(t, notANumber.GetOrNil())

	assertEmptySuccess(t, ResultOf(new("x"), nil).FlatMap(func(s string) (empty Result[int]) {
		return ResultOf[int](nil, nil)
	}))

	called := false
	failed := ResultOf[string](nil, errBoom).FlatMap(func(s string) (res Result[int]) {
		called = true
		return parse(s)
	})
	assert.False(t, called, "FlatMap must not call f on a failure")
	assertFailure(t, failed, errBoom)

	called = false
	failedWithValue := ResultOf(new("1"), errBoom).FlatMap(func(s string) (res Result[int]) {
		called = true
		return parse(s)
	})
	assert.False(t, called, "FlatMap must not call f when there is an error, even with a value")
	assertFailure(t, failedWithValue, errBoom)

	called = false
	empty := ResultOf[string](nil, nil).FlatMap(func(s string) (res Result[int]) {
		called = true
		return parse(s)
	})
	assert.False(t, called, "FlatMap must not call f when there is no value")
	assertEmptySuccess(t, empty)
}

func TestResultChaining(t *testing.T) {
	parse := func(s string) (parsed Result[int]) {
		i, err := strconv.Atoi(s)
		if err != nil {
			return ResultOf[int](nil, err)
		}
		return ResultOf(&i, nil)
	}

	got := ResultOf(new("20"), nil).
		FlatMap(parse).
		Map(func(i int) (doubled int) { return i * 2 }).
		Map(strconv.Itoa)
	assertSuccess(t, got, "40")

	broken := ResultOf(new("x"), nil).
		FlatMap(parse).
		Map(func(i int) (doubled int) { return i * 2 }).
		Map(strconv.Itoa)
	assert.True(t, broken.IsFailure(), "a failure must short-circuit the chain")
}

func TestResultIsFailureAndIsSuccess(t *testing.T) {
	assert.True(t, ResultOf(new(1), nil).IsSuccess())
	assert.False(t, ResultOf(new(1), nil).IsFailure())

	assert.True(t, ResultOf[int](nil, errBoom).IsFailure())
	assert.False(t, ResultOf[int](nil, errBoom).IsSuccess())

	assert.True(t, ResultOf[int](nil, nil).IsSuccess(), "no value and no error is still a success")
	assert.False(t, ResultOf[int](nil, nil).IsFailure())

	assert.True(t, ResultOf(new(1), errBoom).IsFailure(), "an error wins over a value")
	assert.False(t, ResultOf(new(1), errBoom).IsSuccess())
}

func TestResultGetOrNil(t *testing.T) {
	elem := new(42)
	assert.Same(t, elem, ResultOf(elem, nil).GetOrNil(), "GetOrNil must return the given pointer")
	assert.Nil(t, ResultOf[int](nil, errBoom).GetOrNil())
	assert.Nil(t, ResultOf(new(42), errBoom).GetOrNil(), "an error hides the value")
	assert.Nil(t, ResultOf[int](nil, nil).GetOrNil())
}

func TestResultGetOrDefault(t *testing.T) {
	assert.Equal(t, 42, ResultOf(new(42), nil).GetOrDefault(-1))
	assert.Equal(t, 0, ResultOf(new(0), nil).GetOrDefault(-1), "the zero value must not be replaced by the default")
	assert.Equal(t, -1, ResultOf[int](nil, errBoom).GetOrDefault(-1))
	assert.Equal(t, -1, ResultOf(new(42), errBoom).GetOrDefault(-1), "an error hides the value")
	assert.Equal(t, -1, ResultOf[int](nil, nil).GetOrDefault(-1))
}

func TestResultError(t *testing.T) {
	assert.Equal(t, errBoom, ResultOf[int](nil, errBoom).Error())
	assert.Equal(t, errBoom, ResultOf(new(1), errBoom).Error())
	assert.NoError(t, ResultOf(new(1), nil).Error())
	assert.NoError(t, ResultOf[int](nil, nil).Error())
}

func TestResultToOption(t *testing.T) {
	elem := new(42)
	present := ResultOf(elem, nil).ToOption()
	assert.True(t, present.IsPresent(), "a successful Result must yield a present Option")
	assert.Same(t, elem, present.GetOrNil(), "ToOption must keep the given pointer")

	assert.False(t, ResultOf[int](nil, errBoom).ToOption().IsPresent(), "a failure must yield an empty Option")
	assert.False(t, ResultOf(new(42), errBoom).ToOption().IsPresent(), "an error hides the value")
	assert.False(t, ResultOf[int](nil, nil).ToOption().IsPresent(), "no value must yield an empty Option")
}
