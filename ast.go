package lox

//go:generate go run ./cmd/gen_ast -spec ./cmd/gen_ast/expr.spec -dest expr.go
//go:generate go run ./cmd/gen_ast -spec ./cmd/gen_ast/stmt.spec -dest stmt.go
//go:generate go run ./cmd/gen_ast -spec ./cmd/gen_ast/type.spec -dest types.go
