# go-monad

Functional programming monads and collections for Go, built on generic methods (Go 1.27+).

## Design decisions

- Target Go 1.27. Use generic methods for any operation that changes the type parameter, so calls chain fluently. Don't fall back to top-level functions for these.
- Library code uses the standard library only. The only third-party dependency is `github.com/stretchr/testify`, and it is used only in tests. Dependencies are vendored on purpose.
- Type parameters are `any`. Types that need equality or hashing take that behavior as functions from the caller, so non-comparable types work too. Provide ready-made functions and constructors for the `comparable` case so the common path stays short.
- The API is not stable yet. Known limitations are documented in the doc comment and pinned by a test, not hidden.
- Types are thin: no overhead beyond what the behavior needs. Avoid copying caller data when it is safe not to, and say in the doc comment when a result shares memory with its input.
- Mutable types (`List`, `Set`) use pointer receivers. Immutable types use value receivers.
- `Sequence` is lazy. Transformations never evaluate anything. Only accessors evaluate, they stop as early as they can, and each call evaluates again from the source.
- No type is safe for concurrent use, and none synchronizes. This is intentional for now and may change.
- Panicking on misuse is a temporary solution. Whether to return errors instead is to be evaluated.

## API conventions

- Constructors are top-level functions named `<Type>Of`, with a suffix for alternative inputs.
- Outside the file that defines a type, including tests, create values only through its constructors. Never use a composite literal, `make` or a bare `nil` for them. The only exception is `var x T` in tests that check the zero value.
- Outside the file that defines a type, including tests, never read or write its unexported fields. Use the exported API.
- Exported API uses Go naming, even when the functional-programming name is usually lowercase.
- Chaining must be safe. Methods never return a nil collection: an empty result is a non-nil empty value, including when the receiver is nil or a zero value.
- The zero value and a nil receiver behave as the empty case wherever that is possible. Where it is not possible, panic and document it. Never fail silently, return wrong results, or return a value that panics later.
- Callbacks are never invoked on an empty value.
- Invalid arguments that can only be a programming error (nil function arguments) panic with a clear message. Do this even when the receiver is empty, so the bug shows up early.

## Style

- Name every return value, as documentation, including in function literals.
- Never use bare `return` in a function with results. A plain `return` in a function without results is fine.
- Keep comments and documentation to the minimum: state only what the code cannot show. Never restate the name, the signature or the obvious. Don't repeat a rule documented on the type in every method, and never leave commented-out code.
- Document mainly the exported functions, methods, constants and types. A doc comment starts with its name. Unexported code gets a comment only for a non-obvious why.
- A doc comment states what is not obvious from the signature: the empty case, when it panics, the order of results when the type is unordered, and when the result shares memory with the input.
- Code must be `gofmt`-clean and pass `go vet`.

## Testing

- Write tests with `github.com/stretchr/testify`: `assert` by default, `require` only when the rest of the test can't run after a failure.
- Name tests `Test<Type><Method>` (`Test<Constructor>` for constructors), one per exported function or method, so a missing test is easy to spot.
- Pass the expected value before the actual one, and use the most specific assertion (`Equal`, `Same`, `ElementsMatch` for unspecified order, `Nil`, `True`/`False`).
- Add a message to an assertion when the check alone doesn't explain the rule being tested.
- Test every method for its present case, its empty case, its nil and zero-value receiver, and that callbacks are not called on empty values.
- Every documented panic has a test that asserts it with `assert.Panics`. A test that expects a panic must make sure it comes from the call under test, not from setup code.
- Check empty results with the shared helpers (`assertEmptyList`, `assertEmptySet`, `assertEmptySequence`), never with `IsEmpty()` alone.
- Test `Sequence` laziness with call counters: callbacks must not run before an accessor is called, accessors that need few elements must stop early, and each accessor call must evaluate again.
- Keep statement coverage at 100%.

Before finishing a change, run:

```bash
gofmt -l . && go vet ./... && go test -cover ./...
```
