package typing

import (
	"fmt"

	"github.com/brunokim/kilox"
)

type typePair [2]lox.Type

type unifier struct {
	stack      []typePair
	err        error
	constraint Constraint

	t2 lox.Type
}

func Unify(t1, t2 lox.Type) (Constraint, error) {
	u := &unifier{
		constraint: NewConstraint(),
		stack:      []typePair{{t1, t2}},
	}
	for len(u.stack) > 0 {
		if err := u.unifyStep(); err != nil {
			return Constraint{}, err
		}
	}
	return u.constraint, nil
}

func (u *unifier) push(t1, t2 lox.Type) {
	u.stack = append(u.stack, typePair{t1, t2})
}

func (u *unifier) match(t1, t2 lox.Type) error {
	u.t2 = t2
	t1.Accept(u)
	return u.err
}

func (u *unifier) unifyStep() error {
	n := len(u.stack)
	var top typePair
	top, u.stack = u.stack[n-1], u.stack[:n-1]
	t1, t2 := deref(top[0]), deref(top[1])
	if x1, ok := t1.(*lox.RefType); ok {
		return u.match(x1, t2)
	}
	if x2, ok := t2.(*lox.RefType); ok {
		return u.match(x2, t1)
	}
	if _, ok := t2.(lox.NilType); ok {
		// Nil matches with anything.
		// The case for t1.(lox.NilType) is handled in its visitNilType method.
		return nil
	}
	return u.match(t1, t2)
}

func (u *unifier) bindRef(x *lox.RefType, t lox.Type) {
	if x.Value != nil {
		panic(fmt.Sprintf("compiler error: expecting to be called on an unbound ref, got %v", lox.PrintType(x)))
	}
	x.Value = t
	u.constraint.Put(x, t)
}

// ---- Type visitor

func (u *unifier) VisitNilType(t1 lox.NilType) {
	// Nil unifies with anything.
}

func (u *unifier) VisitBoolType(t1 lox.BoolType) {
	if _, ok := u.t2.(lox.BoolType); !ok {
		u.err = typeError{t1, u.t2}
	}
}

func (u *unifier) VisitNumberType(t1 lox.NumberType) {
	if _, ok := u.t2.(lox.NumberType); !ok {
		u.err = typeError{t1, u.t2}
	}
}

func (u *unifier) VisitStringType(t1 lox.StringType) {
	if _, ok := u.t2.(lox.StringType); !ok {
		u.err = typeError{t1, u.t2}
	}
}

func (u *unifier) VisitFunctionType(t1 lox.FunctionType) {
	t2, ok := u.t2.(lox.FunctionType)
	if !ok {
		u.err = typeError{t1, u.t2}
		return
	}
	if len(t1.Params) != len(t2.Params) {
		u.err = typeError{t1, t2}
		return
	}
	u.push(t1.Return, t2.Return)
	for i := len(t1.Params) - 1; i >= 0; i-- {
		u.push(t1.Params[i], t2.Params[i])
	}
}

func (u *unifier) VisitRefType(x *lox.RefType) {
	y, ok := u.t2.(*lox.RefType)
	if !ok {
		u.bindRef(x, u.t2)
		return
	}
	if x.ID < y.ID {
		u.bindRef(y, x)
	} else if x.ID > y.ID {
		u.bindRef(x, y)
	} else {
		// They are the same ref, do nothing.
	}
}
