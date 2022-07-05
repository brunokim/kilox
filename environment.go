package lox

import (
	"fmt"
)

type dynType int

const (
	staticEnvironment dynType = iota
	dynamicEnvironment
)

type Environment struct {
	enclosing *Environment
	locals    []interface{}
	dynamics  map[string]interface{}
}

func NewEnvironment(t dynType) *Environment {
	if t == dynamicEnvironment {
		return &Environment{
			dynamics: make(map[string]interface{}),
		}
	}
	return &Environment{}
}

func (env *Environment) isDynamic() bool {
	return env.dynamics != nil
}

func (env *Environment) Child(t dynType) *Environment {
	if t == dynamicEnvironment && !env.isDynamic() {
		panic("compiler error: dynamic environment can't be child of a static one")
	}
	child := NewEnvironment(t)
	child.enclosing = env
	return child
}

func (env *Environment) Define(name string, value interface{}) {
	if env.isDynamic() {
		env.dynamics[name] = value
	} else {
		env.locals = append(env.locals, value)
	}
}

func (env *Environment) Get(name Token) interface{} {
	for env != nil {
		v, ok := env.dynamics[name.Lexeme]
		if ok {
			return v
		}
		env = env.enclosing
	}
	panic(runtimeError{name, fmt.Sprintf("undefined variable %q", name.Lexeme)})
}

func (env *Environment) Set(name Token, value interface{}) {
	for env != nil {
		_, ok := env.dynamics[name.Lexeme]
		if ok {
			env.dynamics[name.Lexeme] = value
			return
		}
		env = env.enclosing
	}
	panic(runtimeError{name, fmt.Sprintf("undefined variable %q", name.Lexeme)})
}

// ----

func (env *Environment) ancestor(distance int) *Environment {
	for i := 0; i < distance; i++ {
		env = env.enclosing
		if env == nil || env.isDynamic() {
			panic("compiler error: invalid or dynamic environment reached when looking for static ancestor")
		}
	}
	return env
}

func (env *Environment) GetStatic(distance int, index int) interface{} {
	return env.ancestor(distance).locals[index]
}

func (env *Environment) SetStatic(distance int, index int, value interface{}) {
	env.ancestor(distance).locals[index] = value
}
