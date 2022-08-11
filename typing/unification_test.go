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
		want   typing.Constraint
	}{
		{nil_, nil_, constr_()},
		{num_, num_, constr_()},
		{str_, str_, constr_()},
		{bool_, bool_, constr_()},
		{x, bool_, constr_(x, bool_)},
		{bool_, x, constr_(x, bool_)},
		{x, y, constr_(y, x)},
		{y, x, constr_(y, x)},
		{
			func_(types_(str_, num_), bool_),
			func_(types_(str_, num_), bool_),
			constr_(),
		},
		{nil_, num_, constr_()},
		{num_, nil_, constr_()},
		{nil_, str_, constr_()},
		{str_, nil_, constr_()},
		{nil_, x, constr_(x, nil_)},
		{x, nil_, constr_(x, nil_)},
		{
			// x = Nil, Num = x.
			func_(types_(x, num_), num_),
			func_(types_(nil_, x), num_),
			constr_(x, nil_),
		},
		{
			// Str = x, x = y, y = x.
			func_(types_(str_, x, y), num_),
			func_(types_(x, y, x), num_),
			constr_(x, str_, y, str_),
		},
		{
			// y = x, x = y, Str = y.
			func_(types_(y, x, str_), num_),
			func_(types_(x, y, x), num_),
			constr_(y, x, x, str_),
		},
	}
	for _, test := range tests {
		testName := fmt.Sprintf("%v=%v", test.t1, test.t2)
		t.Run(testName, func(t *testing.T) {
			got, err := typing.Unify(test.t1, test.t2)
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
			for _, x := range got.Keys() {
				x.Value = nil
			}
		})
	}
}
