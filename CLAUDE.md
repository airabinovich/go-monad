# go-monad

Functional programming monads and collections for Go, built on generic methods (Go 1.27+).

## Rules

- Target Go 1.27. Use generic methods (`func (o Option[T]) Map[U comparable](...) Option[U]`) for any operation that changes the type parameter, so calls chain fluently. Don't fall back to top-level functions for these.
- Library code uses the standard library only. The only third-party dependency is `github.com/stretchr/testify`, and it is used only in tests.
- Type parameters are `comparable`, not `any`, so every type can convert to a `Set` (map keys must be comparable). Generic methods constrain their own parameters the same way (`Map[U comparable]`).
- Exported API uses Go naming (`Filter`, `GetOrElse`), even when the functional-programming name is usually lowercase.
- The zero value of every monad must be valid and represent its empty/neutral case.
- It must always be safe to chain calls. Functions and methods never return a nil collection: an empty result is a non-nil empty value (`List[T]{}` or `Set[T]{}`, not `nil`), including when the receiver is a nil zero value.
- Callbacks (`f`, `predicate`, ...) are never invoked on an empty value.
- Constructors are top-level functions named `<Type>Of` (e.g. `OptionOf`, `ListOf`, `SetOf`, `SequenceOf`), with a suffix for alternative inputs (e.g. `ListOfSlice`, `SetOfSlice`, `SetOfSeq`, `SequenceOfSlice`, `SequenceOfSeq`).
- Outside the file that defines a type, including tests, create values of that type only through its constructor functions (`OptionOf[T](nil)`, `ListOf(...)`). Never use a composite literal (`Option[T]{}`, `List[T]{...}`), `make`, or a bare `nil` for it. The only exception is `var x T` in tests that check the zero value.
- Keep overhead to a minimum: types are thin wrappers over Go built-ins (`List[T]` is a plain `[]T`, `Set[T]` is a plain `map[T]struct{}`, `Sequence[T]` wraps a single `iter.Seq[T]`). Wrap caller data (pointers, slices) without copying when the representation allows it, and say in the doc comment when the result shares memory with the input.
- `Sequence` is lazy. Transformations (`Map`, `FlatMap`, `Filter`, ...) only wrap the iterator and never evaluate anything. Only accessors (`IsEmpty`, `First`, `ToList`, `ToSet`, ...) evaluate it, they stop as soon as they have the elements they need, and each call evaluates again from the source. Its elements are never exposed directly, only through accessors. Unlike `Set`, it keeps duplicates.
- Types are not safe for concurrent use. Don't add synchronization.

## Style

- Name every return value, as documentation: `func (o Option[T]) GetOrElse(other T) (value T)`.
- Never use bare `return` in a function with results. Always return values explicitly, even with named results. A plain `return` in a function without results (e.g. to stop an iterator early) is fine.
- Every exported identifier has a doc comment starting with its name, stating behavior for both the present and the empty case, and the order of results when the type is unordered (`Set` order is unspecified).
- Code must be `gofmt`-clean and pass `go vet`.

## Testing

- Write tests with `github.com/stretchr/testify`: `assert` by default, `require` only when the rest of the test can't run after a failure.
- Pass the expected value before the actual one, and use the most specific assertion: `Equal` for values, `Same` for pointer identity, `ElementsMatch` for results in unspecified order, `Nil`, `True`/`False`.
- Add a message to an assertion when the check alone doesn't explain the rule being tested (e.g. `"Map must not call f on an empty Option"`).
- Every method is tested for its present case, its empty case, and that callbacks are not called on empty values.
- Check empty collection results with a helper that asserts both `NotNil` and `Empty` (`assertEmptyList`, `assertEmptySet`), never `IsEmpty()` alone. For a `Sequence`, use `assertEmptySequence`, which checks every accessor.
- Test `Sequence` laziness with call counters: callbacks must not run before an accessor is called, accessors that need few elements must stop early, and each accessor call must evaluate again.
- Keep statement coverage at 100%.

Before finishing a change, run:

```bash
gofmt -l . && go vet ./... && go test -cover ./...
```
