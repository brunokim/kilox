package lox

import (
	"fmt"
)

type Environment struct {
	enclosing *Environment
	values    map[string]interface{}
}

func NewEnvironment() *Environment {
	env := new(Environment)
	env.values = make(map[string]interface{})
	return env
}

func (env *Environment) Child() *Environment {
	child := NewEnvironment()
	child.enclosing = env
	return child
}

func (env *Environment) Define(name string, value interface{}) {
	env.values[name] = value
}

// ----

func (env *Environment) Get(name Token) interface{} {
	v, ok := env.values[name.Lexeme]
	if ok {
		return v
	}
	if env.enclosing != nil {
		return env.enclosing.Get(name)
	}
	panic(runtimeError{name, fmt.Sprintf("undefined variable %q", name.Lexeme)})
}

func (env *Environment) Set(name Token, value interface{}) {
	if _, ok := env.values[name.Lexeme]; ok {
		env.values[name.Lexeme] = value
		return
	}
	if env.enclosing != nil {
		env.enclosing.Set(name, value)
		return
	}
	panic(runtimeError{name, fmt.Sprintf("undefined variable %q", name.Lexeme)})
}

func (env *Environment) ancestor(distance int) *Environment {
	for i := 0; i < distance; i++ {
		env = env.enclosing
	}
	return env
}

func (env *Environment) GetAt(distance int, name string) interface{} {
	return env.ancestor(distance).values[name]
}

func (env *Environment) SetAt(distance int, name Token, value interface{}) {
	env.ancestor(distance).values[name.Lexeme] = value
}
