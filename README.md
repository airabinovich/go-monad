# go-monad

![logo](resources/logo.png)

Functional programming basic monads and collections for golang

## Project Status
This project is not yet intended for production use.
Expect big changes in the API and the features present.

## Overview

The package is called `monad` and has five types:

| Type | What it is | Mutable | Lazy |
|---|---|---|---|
| `List[T]` | Ordered collection, duplicates allowed | yes | no |
| `Set[T]` | Unordered collection of distinct elements | yes | no |
| `Option[T]` | A value that may be absent | no | no |
| `Result[T]` | A value or an error | no | no |
| `Sequence[T]` | Ordered elements that are computed on demand | no | yes |

`List` and `Set` are used through pointers (`*List[T]`, `*Set[T]`). `Option`, `Result` and `Sequence` are plain values.

All types share the same transformation methods, so they can be chained:

- `Map(f)` applies `f` to the held value(s).
- `FlatMap(f)` applies an `f` that returns the same kind of type and flattens the result.
- `Filter(predicate)` keeps what satisfies `predicate` (not available on `Result`).

`f` and `predicate` are never called when there is nothing to apply them to.

## List

An ordered collection. Create it with `ListOf(elems...)` or `ListOfSlice(slice)`. Both share the backing array of the slice you pass instead of copying it.

```go
names := monad.ListOf("ana", "bob", "cy").
    Filter(func(s string) bool { return len(s) > 2 }).
    Map(strings.ToUpper)

fmt.Println(names)              // [ANA,BOB]
fmt.Println(names.Len())        // 2
fmt.Println(names.Contains("ANA")) // true
```

| Method | Description |
|---|---|
| `Len()`, `IsEmpty()` | Number of elements, emptiness check |
| `Add(elem)` | Appends in place and returns the List |
| `Contains(elem)` | Membership, compared with `reflect.DeepEqual` |
| `Equal(other)` | Same elements in the same order |
| `Map`, `FlatMap`, `Filter` | Return a new List |
| `First()` | `Option` with a copy of the first element |
| `FirstOrNil()` | Pointer to a copy of the first element, or `nil` |
| `All()`, `Enumerate()`, `Backward()` | Iterators for `for ... range` |
| `ToSet(hashFunc, eqFunc)`, `ToSequence()` | Conversions |
| `String()` | Formatted as `[1,2,3]` |

## Set

A collection of distinct elements in unspecified order. Because elements may be of any type, you give the Set two functions: a `HashFunc` that computes a hash and an `EqFunc` that compares two elements. Elements that are equal must have the same hash.

For comparable types use the ready-made `HashComparable[T]()` and `EqComparable[T]()`:

```go
set := monad.SetOf(monad.HashComparable[int](), monad.EqComparable[int](), 1, 2, 2, 3)

fmt.Println(set.Size())        // 3
fmt.Println(set.Contains(2))   // true
```

With your own functions a Set can hold anything, for example case-insensitive strings:

```go
hash := func(s string) uint64 { return monad.HashComparable[string]()(strings.ToLower(s)) }
set := monad.SetOf(hash, strings.EqualFold, "Go", "GO", "Rust")

fmt.Println(set.Size()) // 2
```

| Method | Description |
|---|---|
| `Add(elem)` | Adds in place unless an equal element exists, returns the Set |
| `Contains(elem)`, `Size()`, `IsEmpty()` | Queries |
| `Equal(other)` | Same elements, regardless of order |
| `Map`, `FlatMap` | Also take the `HashFunc` and `EqFunc` of the resulting Set |
| `Filter(predicate)` | Returns a new Set with the same functions |
| `First()`, `FirstOrNil()` | An arbitrary element, as a copy |
| `ToList()`, `ToSequence()` | Conversions, order is unspecified |
| `String()` | Formatted as `{1,2,3}`, order is unspecified |

## Option

A value that is either present or absent, so you don't have to check for `nil`. `OptionOf` takes a pointer: `nil` means absent.

```go
n := 42
label := monad.OptionOf(&n).
    Filter(func(i int) bool { return i > 10 }).
    Map(strconv.Itoa).
    GetOrDefault("small")

fmt.Println(label) // 42
```

| Method | Description |
|---|---|
| `IsPresent()` | Whether there is a value |
| `GetOrDefault(def)` | The value, or `def` if absent |
| `GetOrNil()` | Pointer to the value, or `nil` |
| `Map`, `FlatMap`, `Filter` | Transform, empty stays empty |
| `ToList()`, `ToSequence()`, `ToResult(ifEmpty)` | Conversions |
| `String()` | The value, or `<nil>` |

## Result

The outcome of something that can fail: a value or an error. `ResultOf(value, err)` takes a pointer to the value and an error.

A Result that holds an error is a failure and hides any value it was created with. A Result with neither a value nor an error is an empty success.

```go
raw := "20"
parse := func(s string) monad.Result[int] {
    i, err := strconv.Atoi(s)
    if err != nil {
        return monad.ResultOf[int](nil, err)
    }
    return monad.ResultOf(&i, nil)
}

res := monad.ResultOf(&raw, nil).
    FlatMap(parse).
    Map(func(i int) int { return i * 2 })

fmt.Println(res.GetOrDefault(-1)) // 40
```

If `parse` fails, the later steps are skipped and the error is passed along.

| Method | Description |
|---|---|
| `IsSuccess()`, `IsFailure()` | Whether there is an error |
| `Error()` | The held error, or `nil` |
| `GetOrDefault(def)` | The value, or `def` on failure or when there is no value |
| `GetOrNil()` | Pointer to the value, or `nil` |
| `Map`, `FlatMap` | Transform, failures are passed along |
| `ToOption()` | Conversion |
| `String()` | The error message, or the value |

## Sequence

A lazy, ordered sequence. `Map`, `FlatMap` and `Filter` only describe the work. Nothing runs until you call an accessor (`IsEmpty`, `First`, `ToList`, `ToSet`) or range over `All()`. Accessors evaluate only as many elements as they need.

```go
seq := monad.SequenceOf(1, 2, 3, 4).
    Filter(func(i int) bool { return i%2 == 0 }).
    Map(strconv.Itoa) // nothing has run yet

fmt.Println(seq.ToList()) // [2,4]
```

Every accessor call evaluates the Sequence again from its source, so your functions may run more than once.

Create one with `SequenceOf(elems...)`, `SequenceOfSlice(slice)`, `SequenceOfSet(set)` or `SequenceOfSeq(iter.Seq)`. A Sequence reads its source when it is evaluated and does not copy it.

## Converting between types

| From \ To | List | Set | Option | Result | Sequence |
|---|---|---|---|---|---|
| **List** | | `ToSet` | `First` | | `ToSequence` |
| **Set** | `ToList` | | `First` | | `ToSequence` |
| **Option** | `ToList` | | | `ToResult` | `ToSequence` |
| **Result** | | | `ToOption` | | |
| **Sequence** | `ToList` | `ToSet` | `First` | | |

## Good to know

- **Nil and zero values:** a `nil` `*List` or `*Set` behaves like an empty one. The zero value of `Option` and `Sequence` is empty. The one exception is `Set.Add`, which panics on a nil or zero-value Set because it has no hash and equality functions; always create a Set with `SetOf` or `SetOfSlice`, or only for comparable types `SetOfComparable`.
- **Mutation:** `List.Add` and `Set.Add` change the collection in place. `Map`, `FlatMap` and `Filter` never modify their source.
- **Not concurrent:** none of the types is safe for concurrent use.
- **Copies:** `First` and `FirstOrNil` on `List`, `Set` and `Sequence` return a copy of the element, so changing it never changes the collection. The copy is shallow: if the element is a pointer, slice or map, it still refers to the same underlying data.
