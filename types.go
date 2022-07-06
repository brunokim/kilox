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

// Construct a union of distinct types.
func unionTypes(ts ...Type) Type {
	if len(ts) == 0 {
		return NilType{}
	}
	if len(ts) == 1 {
		return ts[0]
	}
	u := new(UnionType)
	seen := make(map[Type]struct{})
	for _, t := range ts {
		if _, ok := seen[t]; !ok {
			seen[t] = struct{}{}
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
