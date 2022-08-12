package typing_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/brunokim/lox"
	"github.com/brunokim/lox/typing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

var (
	nil_  = lox.NilType{}
	num_  = lox.NumberType{}
	bool_ = lox.BoolType{}
	str_  = lox.StringType{}
)

func types_(ts ...lox.Type) []lox.Type {
	return ts
}

// Unbound ref
func ref_() *lox.RefType {
	return &lox.RefType{}
}

// Bound ref
func bref_(value lox.Type) *lox.RefType {
	return &lox.RefType{Value: value}
}

// Unbound ref with known ID.
func refi_(id int) *lox.RefType {
	return &lox.RefType{ID: id}
}

// Bound ref with known ID.
func brefi_(id int, value lox.Type) *lox.RefType {
	return &lox.RefType{ID: id, Value: value}
}

func func_(params []lox.Type, result lox.Type) lox.FunctionType {
	return lox.FunctionType{
		Params: params,
		Return: result,
	}
}

func constr_(entries ...any) typing.Constraint {
	if len(entries)%2 != 0 {
		panic(fmt.Sprintf("expecting an even number of args to constr_"))
	}
	c := typing.NewConstraint()
	for i := 0; i < len(entries); i += 2 {
		ref := entries[i].(*lox.RefType)
		value := entries[i+1].(lox.Type)
		c.Put(ref, value)
	}
	return c
}

func constrs_(cs ...typing.Constraint) []typing.Constraint {
	return cs
}

func binding_(x *lox.RefType, value lox.Type) typing.BindingGoal {
	return typing.BindingGoal{x, value}
}

func unify_(t1, t2 lox.Type) typing.UnificationGoal {
	return typing.UnificationGoal{t1, t2}
}

func clause_(name string, head lox.FunctionType, body ...typing.Goal) typing.TypeClause {
	return typing.TypeClause{name, head, body}
}

func clauses_(cls ...typing.TypeClause) []typing.TypeClause {
	return cls
}

// ----

func shouldSkip(text string) bool {
	text = strings.TrimSpace(text)
	return strings.HasPrefix(text, "//test:skip")
}

func parse(t *testing.T, text string) []lox.Stmt {
	if shouldSkip(text) {
		t.Skip()
	}
	s := lox.NewScanner(text)
	tokens, err := s.ScanTokens()
	if err != nil {
		t.Fatalf("scanner: %v", err)
	}
	p := lox.NewParser(tokens)
	stmts, err := p.Parse()
	if err != nil {
		t.Fatalf("parser: %v", err)
	}
	return stmts
}

// ----

var ignoreTypeFields = cmp.Options{
	cmpopts.IgnoreFields(nil_, "Token"),
	cmpopts.IgnoreFields(num_, "Token"),
	cmpopts.IgnoreFields(bool_, "Token"),
	cmpopts.IgnoreFields(str_, "Token"),
	cmpopts.EquateEmpty(),
}
