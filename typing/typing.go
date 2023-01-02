package typing

import (
	"fmt"

	"github.com/brunokim/kilox"
	"github.com/brunokim/kilox/ordered"
)

func types(ts ...lox.Type) []lox.Type {
	return ts
}

func func_(params []lox.Type, result lox.Type) lox.FunctionType {
	return lox.FunctionType{
		Params: params,
		Return: result,
	}
}

var (
	nil_  = lox.NilType{}
	num_  = lox.NumberType{}
	bool_ = lox.BoolType{}
	str_  = lox.StringType{}
)

// ----

// Constraint is an instance of valid ref bindings.
type Constraint struct {
	*ordered.Map[*lox.RefType, lox.Type]
}

func NewConstraint() Constraint {
	return Constraint{ordered.MakeMap[*lox.RefType, lox.Type]()}
}

func Constraint1(x1 *lox.RefType, t1 lox.Type) Constraint {
	return Constraint{ordered.MakeMap1(x1, t1)}
}

// Walk ref chain until finding an unbound ref, or another type.
func deref(t lox.Type) lox.Type {
	for {
		x, ok := t.(*lox.RefType)
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
	t1, t2 lox.Type
}

func (err typeError) Error() string {
	return fmt.Sprintf("%v != %v", err.t1, err.t2)
}

// ----

type transformRef func(x *lox.RefType, cnstrs []Constraint) lox.Type

func Ground(t lox.Type) (lox.Type, bool) {
	isGround := true
	mapUnboundRefs(t, func(x *lox.RefType, cnstrs []Constraint) lox.Type {
		isGround = false
		return nil
	})
	return t, isGround
}

func Copy(t lox.Type, newRef func() *lox.RefType) lox.Type {
	table := make(map[*lox.RefType]*lox.RefType)
	transform := func(x *lox.RefType, constraints []Constraint) lox.Type {
		y, ok := table[x]
		if !ok {
			y = newRef()
			table[x] = y
		}
		return y
	}
	return mapUnboundRefs(t, transform)
}

// ----

func mapUnboundRefs(t lox.Type, f transformRef) lox.Type {
	m := refMapper{
		transform: f,
		seen:      make(map[*lox.RefType]struct{}),
	}
	m.visit(t)
	return m.state
}

type refMapper struct {
	transform transformRef
	state     lox.Type
	seen      map[*lox.RefType]struct{}
}

func (m *refMapper) visit(t lox.Type) lox.Type {
	t.Accept(m)
	return m.state
}

func (m *refMapper) VisitNilType(t lox.NilType)       { m.state = t }
func (m *refMapper) VisitBoolType(t lox.BoolType)     { m.state = t }
func (m *refMapper) VisitNumberType(t lox.NumberType) { m.state = t }
func (m *refMapper) VisitStringType(t lox.StringType) { m.state = t }

func (m *refMapper) VisitFunctionType(t lox.FunctionType) {
	params := make([]lox.Type, len(t.Params))
	for i, param := range t.Params {
		params[i] = m.visit(param)
	}
	result := m.visit(t.Return)
	m.state = lox.FunctionType{params, result}
}

func (m *refMapper) VisitRefType(t *lox.RefType) {
	if _, ok := m.seen[t]; ok {
		m.state = t
		return
	}
	m.seen[t] = struct{}{}
	if t.Value != nil {
		m.visit(t.Value)
	} else {
		m.state = m.transform(t, nil)
		// TODO
	}
}

func (m *refMapper) VisitConstraints(cnstrs []Constraint) []Constraint {
	constraints := make([]Constraint, len(cnstrs))
	for i, constraint := range cnstrs {
		constraints[i] = NewConstraint()
		for _, entry := range constraint.Entries() {
			ref := m.visit(entry.Key).(*lox.RefType)
			value := m.visit(entry.Value)
			constraints[i].Put(ref, value)
		}
	}
	return constraints
}
