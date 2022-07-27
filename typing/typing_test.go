package typing_test

import (
	"fmt"

	"github.com/brunokim/lox"
	"github.com/brunokim/lox/typing"
)

var (
	nil_  = lox.NilType{}
	num_  = lox.NumberType{}
	bool_ = lox.BoolType{}
	str_  = lox.StringType{}
)

func ts_(ts ...lox.Type) []lox.Type {
	return ts
}

// Unbound ref
func uref_() *lox.RefType {
	return &lox.RefType{}
}

// Bound ref
func bref_(value lox.Type) *lox.RefType {
	return &lox.RefType{Value: value}
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
