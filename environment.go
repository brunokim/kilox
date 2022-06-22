package lox

import (
	"fmt"
)

type Environment struct {
	values map[string]interface{}
}

func NewEnvironment() *Environment {
	env := new(Environment)
	env.values = make(map[string]interface{})
	return env
}

func (env *Environment) Define(name string, value interface{}) {
	env.values[name] = value
}

func (env *Environment) Get(name Token) interface{} {
	v, ok := env.values[name.Lexeme]
	if ok {
		return v
	}
	panic(runtimeError{name, fmt.Sprintf("undefined variable %q", name.Lexeme)})
}
