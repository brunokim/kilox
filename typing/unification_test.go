package typing_test

import (
	"fmt"
	"testing"

	"github.com/brunokim/lox"
	"github.com/brunokim/lox/typing"

	"github.com/google/go-cmp/cmp"
)

func TestUnifier(t *testing.T) {
	tests := []struct {
		t1, t2 lox.Type
		want   []typing.Constraint
	}{
		{lox.NilType{}, lox.NilType{}, nil},
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
			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("(-want, +got)\n%s", diff)
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
