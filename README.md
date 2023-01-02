# kilox

Kim's Lox implementation from Crafting Interpreters book.

## Go implementation

```sh
go run cmd/lox/lox.go
```

### Additions

- [x] Ignore params ending with underscore
- [x] Instance initializer
- [x] Class var
- [x] Class var initializer
- [x] `new` for class initialization
- [ ] typing (experimental)

## C implementation (ongoing)

```sh
pushd src
make build
popd
src/clox
```
