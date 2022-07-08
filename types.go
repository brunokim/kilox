package lox

import (
	"fmt"
)

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

// Splice union types into list.
func flattenTypes(ts []Type) []Type {
	result := make([]Type, 0, len(ts))
	for _, t := range ts {
		t = deref(t)
		u, ok := t.(*UnionType)
		if !ok {
			result = append(result, t)
		} else {
			result = append(result, flattenTypes(u.Types)...)
		}
	}
	return result
}

// Construct a union of distinct types.
func unionTypes(ts ...Type) Type {
	if len(ts) == 0 {
		return NilType{}
	}
	if len(ts) == 1 {
		return ts[0]
	}
	u := new(UnionType)
	seen := make(map[string]struct{})
	for _, t := range flattenTypes(ts) {
		key := TypePrint(t)
		if _, ok := seen[key]; !ok {
			seen[key] = struct{}{}
			u.Types = append(u.Types, t)
		}
	}
	return u
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

func Ground(t Type) (Type, bool) {
	isGround := true
	t = mapUnboundRefs(t, func(x *RefType) Type {
		isGround = false
		return x
	})
	return t, isGround
}

func Copy(t Type, newRef func() *RefType) Type {
	table := make(map[*RefType]*RefType)
	return mapUnboundRefs(t, func(x *RefType) Type {
		y, ok := table[x]
		if !ok {
			y = newRef()
			table[x] = y
		}
		return y
	})
}

// ----

func mapUnboundRefs(t Type, f func(x *RefType) Type) Type {
	m := refMapper{transform: f}
	m.visit(t)
	return m.state
}

type refMapper struct {
	transform func(x *RefType) Type
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
	m.state = FunctionType{params, result}
}

func (m *refMapper) visitRefType(t *RefType) {
	if t.Value != nil {
		m.visit(t.Value)
	} else {
		m.state = m.transform(t)
	}
}

func (m *refMapper) visitUnionType(t *UnionType) {
	var types []Type
	for _, subtype := range t.Types {
		types = append(types, m.visit(subtype))
	}
	m.state = unionTypes(types...)
}
