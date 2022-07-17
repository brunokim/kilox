package lox

import (
	"fmt"
)

// Bindings is a representation of the values associated to a
// set of Refs.
type Bindings map[*RefType]Type

// Walk ref chain until finding an unbound ref, or another type.
func deref(t Type) Type {
	for {
		x, ok := t.(*RefType)
		if !ok {
			return t
		}
		if x.Value == nil {
			return x
		}
		t = x.Value
	}
}

// ----

type typeError struct {
	t1, t2 Type
}

func (err typeError) Error() string {
	return fmt.Sprintf("%v != %v", err.t1, err.t2)
}

func (c *TypeChecker) addError(err typeError) {
	c.errors = append(c.errors, err)
}

// ----

type transformRef func(x *RefType, opts []Bindings) Type

func Ground(t Type) (Type, bool) {
	isGround := true
	mapUnboundRefs(t, func(x *RefType, opts []Bindings) Type {
		isGround = false
		return nil
	})
	return t, isGround
}

func Copy(t Type, newRef func() *RefType) Type {
	table := make(map[*RefType]*RefType)
	transform := func(x *RefType, opts []Bindings) Type {
		y, ok := table[x]
		if !ok {
			y = newRef()
			y.options = opts
			table[x] = y
		}
		return y
	}
	return mapUnboundRefs(t, transform)
}

// ----

func mapUnboundRefs(t Type, f transformRef) Type {
	m := refMapper{transform: f}
	m.visit(t)
	return m.state
}

type refMapper struct {
	transform transformRef
	state     Type
}

func (m *refMapper) visit(t Type) Type {
	t.accept(m)
	return m.state
}

func (m *refMapper) visitNilType(t NilType)       { m.state = t }
func (m *refMapper) visitBoolType(t BoolType)     { m.state = t }
func (m *refMapper) visitNumberType(t NumberType) { m.state = t }
func (m *refMapper) visitStringType(t StringType) { m.state = t }

func (m *refMapper) visitFunctionType(t FunctionType) {
	params := make([]Type, len(t.Params))
	for i, param := range t.Params {
		params[i] = m.visit(param)
	}
	result := m.visit(t.Return)
	options := m.visitOptions(t.options)
	m.state = FunctionType{params, result, options}
}

func (m *refMapper) visitRefType(t *RefType) {
	if t.Value != nil {
		m.visit(t.Value)
	} else {
		options := m.visitOptions(t.options)
		m.state = m.transform(t, options)
	}
}

func (m *refMapper) visitOptions(opts []Bindings) []Bindings {
	options := make([]Bindings, len(opts))
	for i, option := range opts {
		options[i] = make(Bindings)
		for ref, value := range option {
			ref = m.visit(ref).(*RefType)
			value = m.visit(value)
			options[i][ref] = value
		}
	}
	return options
}
