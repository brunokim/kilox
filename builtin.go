package lox

import (
	"fmt"
	"math/rand"
	"time"
)

var builtin = NewEnvironment(dynamicEnvironment)

func init() {
	builtin.Define("clock", clockFunc{})
	builtin.Define("type", typeFunc{})
	builtin.Define("random", randomFunc{})
	builtin.Define("randomSeed", randomSeedFunc{})
}

// ----

type clockFunc struct{}

func (f clockFunc) Arity() int { return 0 }
func (f clockFunc) Call(i *Interpreter, args []any) any {
	return float64(time.Now().UnixMicro()) / 1e6
}
func (f clockFunc) String() string { return "<native fn clock>" }

// ----

type typeFunc struct{}

func (f typeFunc) Arity() int { return 1 }
func (f typeFunc) Call(i *Interpreter, args []any) any {
	arg := args[0]
	switch v := arg.(type) {
	case bool:
		return BoolType{}
	case float64:
		return NumberType{}
	case string:
		return StringType{}
	case instance:
		return v.class
	case class:
		return v.meta
	case metaClass:
		return metaType{}
	case metaType:
		return metaType{}
	case function:
		params := make([]Type, v.Arity())
		for i := 0; i < v.Arity(); i++ {
			params[i] = &RefType{id: i + 1}
		}
		return FunctionType{
			Params: params,
			Return: &RefType{id: v.Arity() + 1},
		}
	default:
		if arg == nil {
			return NilType{}
		}
		panic(runtimeError{Token{}, fmt.Sprintf("unhandled type(%[1]v) (%[1]T)", arg)})
	}
}
func (f typeFunc) String() string { return "<native fn type>" }

// ----

type randomFunc struct{}

func (f randomFunc) Arity() int { return 0 }
func (f randomFunc) Call(i *Interpreter, args []any) any {
	return rand.Float64()
}
func (f randomFunc) String() string { return "<native fn random>" }

// ----

type randomSeedFunc struct{}

func (f randomSeedFunc) Arity() int { return 1 }
func (f randomSeedFunc) Call(i *Interpreter, args []any) any {
	arg := args[0]
	seed, ok := arg.(float64)
	if !ok {
		panic(runtimeError{Token{}, fmt.Sprintf("unhandled randomSeed(%[1]v) (%[1]T)", arg)})
	}
	rand.Seed(int64(seed))
	return nil
}
func (f randomSeedFunc) String() string { return "<native fn randomSeed>" }
