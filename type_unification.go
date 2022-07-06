package lox

import (
	"fmt"
)

type typePair [2]Type

type unifier struct {
	c     *TypeChecker
	stack []typePair

	t2 Type
}

func (u *unifier) push(t1, t2 Type) {
	u.stack = append(u.stack, typePair{t1, t2})
}

func (u *unifier) err(t1, t2 Type) {
	u.c.addError(typeError{t1, t2})
}

func (u *unifier) match(t1, t2 Type) {
	fmt.Println("match", t1, t2)
	u.t2 = t2
	t1.accept(u)
}

func (u *unifier) unify(t1, t2 Type) {
	u.stack = []typePair{{t1, t2}}
	for len(u.stack) > 0 {
		fmt.Println(u.stack)
		n := len(u.stack)
		var top typePair
		top, u.stack = u.stack[n-1], u.stack[:n-1]
		t1, t2 = deref(top[0]), deref(top[1])
		if x1, ok := t1.(*RefType); ok {
			u.match(x1, t2)
			continue
		}
		if x2, ok := t2.(*RefType); ok {
			u.match(x2, t1)
			continue
		}
		u.match(t1, t2)
	}
}

// ----

func (u *unifier) visitNilType(t1 NilType) {
	if _, ok := u.t2.(NilType); !ok {
		u.err(t1, u.t2)
	}
}

func (u *unifier) visitBoolType(t1 BoolType) {
	if _, ok := u.t2.(BoolType); !ok {
		u.err(t1, u.t2)
	}
}

func (u *unifier) visitNumberType(t1 NumberType) {
	if _, ok := u.t2.(NumberType); !ok {
		u.err(t1, u.t2)
	}
}

func (u *unifier) visitStringType(t1 StringType) {
	if _, ok := u.t2.(StringType); !ok {
		u.err(t1, u.t2)
	}
}

func (u *unifier) visitFunctionType(t1 FunctionType) {
	t2, ok := u.t2.(FunctionType)
	if !ok {
		u.err(t1, u.t2)
		return
	}
	if len(t1.Params) != len(t2.Params) {
		u.err(t1, t2)
		return
	}
	u.push(t1.Return, t2.Return)
	for i := len(t1.Params) - 1; i >= 0; i-- {
		u.push(t1.Params[i], t2.Params[i])
	}
}

func (u *unifier) visitRefType(x *RefType) {
	if x.Value != nil {
		panic(fmt.Sprintf("compiler error: expecting to be called on an unbound ref, got %v", TypePrint(x)))
	}
	y, ok := u.t2.(*RefType)
	if !ok {
		x.Value = u.t2
		return
	}
	if x.id < y.id {
		y.Value = x
	} else if x.id > y.id {
		x.Value = y
	}
}

func (u *unifier) visitUnionType(t1 *UnionType) {
	panic("lox.(*unifier).visitUnionType is not implemented")
}
