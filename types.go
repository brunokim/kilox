package lox

import (
	"fmt"
)

// Constraint is an instance of valid ref bindings.
type Constraint map[*RefType]Type

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

type transformRef func(x *RefType, cnstrs []Constraint) Type

func Ground(t Type) (Type, bool) {
	isGround := true
	mapUnboundRefs(t, func(x *RefType, cnstrs []Constraint) Type {
		isGround = false
		return nil
	})
	return t, isGround
}

func Copy(t Type, newRef func() *RefType) Type {
	table := make(map[*RefType]*RefType)
	transform := func(x *RefType, constraints []Constraint) Type {
		y, ok := table[x]
		if !ok {
			y = newRef()
			y.constraints = constraints
			table[x] = y
		}
		return y
	}
	return mapUnboundRefs(t, transform)
}

// ----

func mapUnboundRefs(t Type, f transformRef) Type {
	m := refMapper{
		transform: f,
		seen:      make(map[*RefType]struct{}),
	}
	m.visit(t)
	return m.state
}

type refMapper struct {
	transform transformRef
	state     Type
	seen      map[*RefType]struct{}
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
	m.state = FunctionType{params, result}
}

func (m *refMapper) visitRefType(t *RefType) {
	if _, ok := m.seen[t]; ok {
		m.state = t
		return
	}
	m.seen[t] = struct{}{}
	if t.Value != nil {
		m.visit(t.Value)
	} else {
		constraints := m.visitConstraints(t.constraints)
		m.state = m.transform(t, constraints)
	}
}

func (m *refMapper) visitConstraints(cnstrs []Constraint) []Constraint {
	constraints := make([]Constraint, len(cnstrs))
	for i, constraint := range cnstrs {
		constraints[i] = make(Constraint)
		for ref, value := range constraint {
			ref = m.visit(ref).(*RefType)
			value = m.visit(value)
			constraints[i][ref] = value
		}
	}
	return constraints
}
