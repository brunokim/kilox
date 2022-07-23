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
				"$.1.Expression": num_,
			},
		},
		{
			"var a = true; var b = a; print b;",
			map[string]lox.Type{
				"$.1.Init":       bool_,
				"$.2.Expression": bool_,
			},
		},
		{
			"var a = 1; while (a < 4) { var b = a + 1; a = b; }",
			map[string]lox.Type{
				"$.1.Condition":                          func_(types(num_, num_), bool_),
				"$.1.Condition.Left":                     num_,
				"$.1.Body.Statements.0.Init":             func_(types(ref_(num_), ref_(num_)), ref_(num_)),
				"$.1.Body.Statements.0.Init.Left":        num_,
				"$.1.Body.Statements.1.Expression.Value": ref_(num_),
			},
		},
		{
			`
            fun add3(a, b, c) {
                return a + b + c;
            }
            print add3(1, 2, 3);
            print add3("x", "y", "z");
            `,
			map[string]lox.Type{
				"$.0.Body.0.Result.Left.Left":  ref_(num_),
				"$.0.Body.0.Result.Left.Right": ref_(num_),
				"$.0.Body.0.Result.Right":      ref_(num_),
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
			cmpopts.IgnoreFields(nil_, "Token"),
			cmpopts.IgnoreFields(num_, "Token"),
			cmpopts.IgnoreFields(bool_, "Token"),
			cmpopts.IgnoreFields(str_, "Token"),
			cmpopts.IgnoreFields(lox.RefType{}, "id", "constraints"),
		}
		if diff := cmp.Diff(want, types, opts); diff != "" {
			t.Errorf("%q: (-want,+got):%s", test.text, diff)
		}
	}
}
