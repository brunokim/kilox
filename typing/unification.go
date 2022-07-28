package typing

import (
	"fmt"

	"github.com/brunokim/lox"
	"github.com/brunokim/lox/errlist"
)

type typePair [2]lox.Type

type choicePoint struct {
	constraints []Constraint

	topRefID int
	trail    []*lox.RefType
	stack    []typePair

	x *lox.RefType
	t lox.Type
}

// Interface to decouple unifier from type checker.
type UnificationCtx interface {
	GetRefID() int
	SetRefID(id int)
}

type Unifier struct {
	ctx         UnificationCtx
	stack       []typePair
	choices     []*choicePoint
	errors      []typeError
	constraints []Constraint

	t2 lox.Type
}

func NewUnifier(ctx UnificationCtx) *Unifier {
	return &Unifier{ctx: ctx}
}

func (u *Unifier) reset() {
	u.stack = nil
	u.choices = nil
	u.errors = nil
	u.constraints = nil
}

func (u *Unifier) push(t1, t2 lox.Type) {
	u.stack = append(u.stack, typePair{t1, t2})
}

func (u *Unifier) err(t1, t2 lox.Type) {
	if len(u.constraints) == 0 {
		// Keep track of errors only if we don't have any solution.
		u.errors = append(u.errors, typeError{t1, t2})
	}
	u.backtrack()
}

func (u *Unifier) match(t1, t2 lox.Type) {
	u.t2 = t2
	t1.Accept(u)
}

func (u *Unifier) Unify(t1, t2 lox.Type) ([]Constraint, error) {
	u.reset()
	u.pushChoicePoint(nil, nil)
	u.stack = []typePair{{t1, t2}}
	for len(u.stack) > 0 || len(u.choices) > 0 {
		if len(u.stack) > 0 {
			u.unifyStep()
			continue
		}
		// Stack is empty, found a solution!
		u.constraints = append(u.constraints, u.constraint())
		u.errors = nil
		u.backtrack()
	}
	if len(u.errors) > 0 {
		return nil, errlist.Of[typeError](u.errors)
	}
	if len(u.constraints) == 1 {
		// Single solution, bind vars already.
		for _, entry := range u.constraints[0].Entries() {
			x, value := entry.Key, entry.Value
			x.Value = value
		}
	}
	return u.constraints, nil
}

func (u *Unifier) constraint() Constraint {
	constraint := NewConstraint()
	for _, choice := range u.choices {
		for _, x := range choice.trail {
			constraint.Put(x, x.Value)
		}
	}
	return constraint
}

func (u *Unifier) unifyStep() {
	n := len(u.stack)
	var top typePair
	top, u.stack = u.stack[n-1], u.stack[:n-1]
	t1, t2 := deref(top[0]), deref(top[1])
	if x1, ok := t1.(*lox.RefType); ok {
		u.match(x1, t2)
		return
	}
	if x2, ok := t2.(*lox.RefType); ok {
		u.match(x2, t1)
		return
	}
	if _, ok := t2.(lox.NilType); ok {
		// Nil matches with anything.
		// The case for t1.(lox.NilType) is handled in its visitNilType method.
		return
	}
	u.match(t1, t2)
}

func (u *Unifier) pushChoicePoint(x *lox.RefType, constraints []Constraint) {
	// TODO: split stack in environments so that we don't need to copy everything?
	stack := make([]typePair, len(u.stack))
	copy(stack, u.stack)

	choice := &choicePoint{
		constraints: constraints,
		topRefID:    u.ctx.GetRefID(),
		stack:       stack,
		x:           x,
		t:           u.t2,
	}
	u.choices = append(u.choices, choice)
}

func (u *Unifier) popConstraint() (*choicePoint, Constraint) {
	for i := len(u.choices) - 1; i >= 0; i-- {
		choice := u.choices[i]
		choice.unwindTrail()
		if len(choice.constraints) > 0 {
			// Pop constraint from choicepoint.
			constraint := choice.constraints[0]
			choice.constraints = choice.constraints[1:]
			return choice, constraint
		}
		// Pop choicepoint if it has no more constraints.
		u.choices = u.choices[:i]
	}
	return nil, Constraint{}
}

func (u *Unifier) backtrack() error {
	choice, constraint := u.popConstraint()
	if choice == nil {
		return fmt.Errorf("no more choices")
	}

	// Reset ctx state.
	u.ctx.SetRefID(choice.topRefID)

	// Reset stack.
	l1, l2 := len(u.stack), len(choice.stack)
	copy(u.stack, choice.stack)
	if l1 < l2 {
		u.stack = append(u.stack, choice.stack[l1:]...)
	} else {
		u.stack = u.stack[:l2]
	}

	// Set new constraints.
	u.applyConstraint(choice.x, choice.t, constraint)

	return nil
}

func (u *Unifier) peek() *choicePoint {
	n := len(u.choices)
	if n == 0 {
		return nil
	}
	return u.choices[n-1]
}

func (u *Unifier) unifyRef(x *lox.RefType, t lox.Type) {
	u.bindRef(x, t)
}

func (u *Unifier) applyConstraint(x *lox.RefType, t lox.Type, constraint Constraint) {
	for _, entry := range constraint.Entries() {
		y, value := entry.Key, entry.Value
		u.bindRef(y, value)
	}
	if x.Value == nil {
		// Variable constraints do not apply to itself (is this legal?).
		u.bindRef(x, t)
	} else {
		// Revisit x, now bound to a constraint.
		u.push(x, t)
	}
}

func (u *Unifier) bindRef(x *lox.RefType, t lox.Type) {
	if x.Value != nil {
		panic(fmt.Sprintf("compiler error: expecting to be called on an unbound ref, got %v", lox.PrintType(x)))
	}
	x.Value = t
	u.peek().addToTrail(x)
}

// Adds x to the trail, indicating that it was bound during the current choice point.
//
// NOTE: we usually don't store refs that would be recreated on backtrack (so called unconditional
// refs), but since we also use the trail to build the constraint solution at the end, we do
// trail all bound refs.
func (cp *choicePoint) addToTrail(x *lox.RefType) {
	if cp == nil {
		return
	}
	cp.trail = append(cp.trail, x)
}

// Reset bindings for all refs that were bound during the current choice point.
func (cp *choicePoint) unwindTrail() {
	if cp == nil {
		return
	}
	for _, x := range cp.trail {
		x.Value = nil
	}
	cp.trail = nil
}

// ---- Type visitor

func (u *Unifier) VisitNilType(t1 lox.NilType) {
	// Nil unifies with anything.
}

func (u *Unifier) VisitBoolType(t1 lox.BoolType) {
	if _, ok := u.t2.(lox.BoolType); !ok {
		u.err(t1, u.t2)
	}
}

func (u *Unifier) VisitNumberType(t1 lox.NumberType) {
	if _, ok := u.t2.(lox.NumberType); !ok {
		u.err(t1, u.t2)
	}
}

func (u *Unifier) VisitStringType(t1 lox.StringType) {
	if _, ok := u.t2.(lox.StringType); !ok {
		u.err(t1, u.t2)
	}
}

func (u *Unifier) VisitFunctionType(t1 lox.FunctionType) {
	t2, ok := u.t2.(lox.FunctionType)
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

func (u *Unifier) VisitRefType(x *lox.RefType) {
	y, ok := u.t2.(*lox.RefType)
	if !ok {
		u.unifyRef(x, u.t2)
		return
	}
	if x.ID < y.ID {
		u.unifyRef(y, x)
	} else if x.ID > y.ID {
		u.unifyRef(x, y)
	}
}
