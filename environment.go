package lox

import (
	"fmt"
	"strings"
)

type dynType int

const (
	staticEnvironment dynType = iota
	dynamicEnvironment
)

type Environment struct {
	enclosing *Environment
	locals    []any
	dynamics  map[string]any
}

func NewEnvironment(t dynType) *Environment {
	if t == dynamicEnvironment {
		return &Environment{
			dynamics: make(map[string]any),
		}
	}
	return &Environment{}
}

func (env *Environment) isDynamic() bool {
	return env.dynamics != nil
}

func (env *Environment) Debug() string {
	var b strings.Builder
	env.debug(&b)
	return b.String()
}

func (env *Environment) debug(b *strings.Builder) {
	if env.isDynamic() {
		b.WriteString("- type: dynamic\n")
		b.WriteString("  bindings:\n")
		for name, value := range env.dynamics {
			fmt.Fprintf(b, "    %[1]s: {type: %[2]T, value: %[2]v}\n", name, value)
		}
	} else {
		b.WriteString("- type: static\n")
		b.WriteString("  bindings:\n")
		for _, value := range env.locals {
			fmt.Fprintf(b, "    - {type: %[1]T, value: %[1]v}\n", value)
		}
	}
	if env.enclosing != nil {
		env.enclosing.debug(b)
	}
}

// ----

func (env *Environment) Child(t dynType) *Environment {
	if t == dynamicEnvironment && !env.isDynamic() {
		panic("compiler error: dynamic environment can't be child of a static one")
	}
	child := NewEnvironment(t)
	child.enclosing = env
	return child
}

func (env *Environment) Define(name string, value any) {
	if env.isDynamic() {
		env.dynamics[name] = value
	} else {
		env.locals = append(env.locals, value)
	}
}

func (env *Environment) Get(name Token) any {
	for env != nil {
		v, ok := env.dynamics[name.Lexeme]
		if ok {
			return v
		}
		env = env.enclosing
	}
	panic(runtimeError{name, fmt.Sprintf("undefined variable %q", name.Lexeme)})
}

func (env *Environment) Set(name Token, value any) {
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

func (env *Environment) GetStatic(distance int, index int) any {
	return env.ancestor(distance).locals[index]
}

func (env *Environment) SetStatic(distance int, index int, value any) {
	env.ancestor(distance).locals[index] = value
}
