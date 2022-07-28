package typing_test

import (
	"fmt"
	"testing"

	"github.com/brunokim/lox"
	"github.com/brunokim/lox/ordered"
	"github.com/brunokim/lox/typing"

	"github.com/google/go-cmp/cmp"
)

var (
	x = &lox.RefType{ID: 1000}
	y = &lox.RefType{ID: 2000}
)

func TestUnifier(t *testing.T) {
	tests := []struct {
		t1, t2 lox.Type
		want   []typing.Constraint
	}{
		{nil_, nil_, constrs_(constr_())},
		{num_, num_, constrs_(constr_())},
		{str_, str_, constrs_(constr_())},
		{bool_, bool_, constrs_(constr_())},
		{x, bool_, constrs_(constr_(x, bool_))},
		{bool_, x, constrs_(constr_(x, bool_))},
		{x, y, constrs_(constr_(y, x))},
		{y, x, constrs_(constr_(y, x))},
		{
			func_(ts_(str_, num_), bool_),
			func_(ts_(str_, num_), bool_),
			constrs_(constr_()),
		},
		{nil_, num_, constrs_(constr_())},
		{num_, nil_, constrs_(constr_())},
		{nil_, str_, constrs_(constr_())},
		{str_, nil_, constrs_(constr_())},
		{nil_, x, constrs_(constr_(x, nil_))},
		{x, nil_, constrs_(constr_(x, nil_))},
		{
			// x = Nil, Num = x.
			func_(ts_(x, num_), num_),
			func_(ts_(nil_, x), num_),
			constrs_(constr_(x, nil_)),
		},
		{
			// Str = x, x = y, y = x.
			func_(ts_(str_, x, y), num_),
			func_(ts_(x, y, x), num_),
			constrs_(constr_(x, str_, y, str_)),
		},
		{
			// y = x, x = y, Str = y.
			func_(ts_(y, x, str_), num_),
			func_(ts_(x, y, x), num_),
			constrs_(constr_(y, x, x, str_)),
		},
	}
	for _, test := range tests {
		testName := fmt.Sprintf("%v=%v", test.t1, test.t2)
		t.Run(testName, func(t *testing.T) {
			ctx := new(dummyUnificationCtx)
			u := typing.NewUnifier(ctx)
			got, err := u.Unify(test.t1, test.t2)
			if err != nil {
				t.Fatalf("got err: %v", err)
			}
			opts := cmp.Options{
				cmp.Transformer("Constraint", func(c typing.Constraint) []ordered.Entry[*lox.RefType, lox.Type] {
					return c.Entries()
				}),
			}
			if diff := cmp.Diff(test.want, got, opts); diff != "" {
				t.Errorf("(-want, +got)\n%s", diff)
			}
			// Reset shared refs.
			for _, constr := range got {
				for _, x := range constr.Keys() {
					x.Value = nil
				}
			}
		})
	}
}

type dummyUnificationCtx struct {
	id int
}

func (ctx *dummyUnificationCtx) GetRefID() int {
	ctx.id++
	return ctx.id
}

func (ctx *dummyUnificationCtx) SetRefID(id int) {
	ctx.id = id
}
