package lox_test

import (
	"testing"

	"github.com/brunokim/lox"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestCheck(t *testing.T) {
	tests := []struct {
		text  string
		paths map[string]lox.Type
	}{
		{
			"var a = 1; print a;",
			map[string]lox.Type{
				"$.1.Expression": lox.NumberType{literalToken(lox.Number, "1", 1.0)},
			},
		},
	}
	for _, test := range tests {
		stmts := parseStmts(t, test.text)
		c := lox.NewTypeChecker()
		types, err := c.Check(stmts)
		if err != nil {
			t.Errorf("%q: got err: %v", test.text, err)
			continue
		}
		want := make(map[lox.Expr]lox.Type)
		for path, type_ := range test.paths {
			elem, err := walkPath(path, stmts)
			if err != nil {
				t.Fatalf("%q: invalid path %q for %v: %v", test.text, path, stmts, err)
			}
			want[elem.(lox.Expr)] = type_
		}
		opts := cmp.Options{
			cmpopts.IgnoreFields(lox.Token{}, "Line"),
		}
		if diff := cmp.Diff(want, types, opts); diff != "" {
			t.Errorf("%q: (-want,+got):%s", test.text, diff)
		}
	}
}
