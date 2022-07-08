package lox

import (
	"fmt"
)

type typePair [2]Type

type choicePoint struct {
	options []Type

	topRefID  int
	optionIdx int
	t2        Type
	trail     []*RefType
	stack     []typePair
}

type unifier struct {
	c       *TypeChecker
	stack   []typePair
	choices []*choicePoint
	errors  []typeError

	t2 Type
}

func (u *unifier) push(t1, t2 Type) {
	u.stack = append(u.stack, typePair{t1, t2})
}

func (u *unifier) err(t1, t2 Type) {
	u.errors = append(u.errors, typeError{t1, t2})
	if len(u.choices) > 0 {
		u.backtrack()
	} else {
		u.c.errors = append(u.c.errors, u.errors...)
		u.errors = nil
	}
}

func (u *unifier) match(t1, t2 Type) {
	u.t2 = t2
	t1.accept(u)
}

func (u *unifier) unify(t1, t2 Type) {
	u.stack = []typePair{{t1, t2}}
	for len(u.stack) > 0 {
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
		if _, ok := t2.(NilType); ok {
			// Nil matches with anything.
			// The case for t1.(NilType) is handled in its visitNilType method.
			continue
		}
		u.match(t1, t2)
	}
}

func (u *unifier) pushChoicePoint(options []Type) {
	// TODO: split stack in environments so that we don't need to copy everything?
	stack := make([]typePair, len(u.stack))
	copy(stack, u.stack)

	choice := &choicePoint{
		options:  options,
		topRefID: u.c.refID,
		stack:    stack,
		t2:       u.t2,
	}
	u.choices = append(u.choices, choice)
}

func (u *unifier) backtrack() {
	n := len(u.choices)
	if n == 0 {
		panic("no more choices")
	}
	u.unwindTrail()
	choice := u.choices[n-1]
	choice.optionIdx++
	// If this is the last option, pop choice point.
	if choice.optionIdx == len(choice.options)-1 {
		u.choices = u.choices[:n-1]
	}
	option := choice.options[choice.optionIdx]
	// Reset state.
	// TODO: remove this coupling between unifier and type checker.
	u.c.refID = choice.topRefID
	// Reset stack.
	l1, l2 := len(u.stack), len(choice.stack)
	copy(u.stack, choice.stack)
	if l1 < l2 {
		u.stack = append(u.stack, choice.stack[l1:]...)
	} else {
		u.stack = u.stack[:l2]
	}
	// Push new option.
	u.push(option, choice.t2)
}

func (u *unifier) bindRef(x *RefType, t Type) {
	if x.Value != nil {
		panic(fmt.Sprintf("compiler error: expecting to be called on an unbound ref, got %v", TypePrint(x)))
	}
	x.Value = t
	u.trail(x)
}

// Adds x to the trail, indicating that it was bound during the current choice point.
func (u *unifier) trail(x *RefType) {
	n := len(u.choices)
	if n == 0 {
		return
	}
	choice := u.choices[n-1]
	if choice.topRefID < x.id {
		// Unconditional ref: x is newer than current choice point, so it will
		// be recreated if we backtrack. There's no need to add it to the trail.
		return
	}
	choice.trail = append(choice.trail, x)
}

// Reset bindings for all refs that were bound during the current choice point.
func (u *unifier) unwindTrail() {
	n := len(u.choices)
	if n == 0 {
		return
	}
	choice := u.choices[n-1]
	for _, x := range choice.trail {
		x.Value = nil
	}
	choice.trail = nil
}

// ----

func (u *unifier) visitNilType(t1 NilType) {
	// Nil unifies with anything.
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
	y, ok := u.t2.(*RefType)
	if !ok {
		u.bindRef(x, u.t2)
		return
	}
	if x.id < y.id {
		u.bindRef(y, x)
	} else if x.id > y.id {
		u.bindRef(x, y)
	}
}

func (u *unifier) visitUnionType(t1 *UnionType) {
	u.pushChoicePoint(t1.Types)
	u.push(t1.Types[0], u.t2)
}
