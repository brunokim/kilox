package typing_test

import (
	"testing"

	"github.com/brunokim/lox"
	"github.com/brunokim/lox/typing"
	"github.com/brunokim/lox/valuepath"

	"github.com/google/go-cmp/cmp"
	"github.com/lithammer/dedent"
)

func TestCheck(t *testing.T) {
	tests := []struct {
		text  string
		paths map[string]lox.Type
	}{
		{
			dedent.Dedent(`
            var a = 1;
            print a;`),
			map[string]lox.Type{
				"$.1.Expression": num_, // line 2: a
			},
		},
		{
			dedent.Dedent(`
            var a = true;
            var b = a;
            print b;`),
			map[string]lox.Type{
				"$.1.Init":       bool_, // line 2: a
				"$.2.Expression": bool_, // line 3: b
			},
		},
		{
			dedent.Dedent(`
            var a = 1;
            while (a < 4) {
                var b = a + 1;
                a = b;
            }`),
			map[string]lox.Type{
				"$.1.Condition":                          func_(ts_(num_, num_), bool_),                     // line 2: a < 4
				"$.1.Condition.Left":                     num_,                                              // line 2: a
				"$.1.Body.Statements.0.Init":             func_(ts_(bref_(num_), bref_(num_)), bref_(num_)), // line 3: a + 1
				"$.1.Body.Statements.0.Init.Left":        num_,                                              // line 3: a
				"$.1.Body.Statements.1.Expression.Value": bref_(num_),                                       // line 4: b
			},
		},
		{
			dedent.Dedent(`
            //test:skip
            fun add3(a, b, c) {
                return a + b + c;
            }
            print add3(1, 2, 3);
            print add3("x", "y", "z");`),
			map[string]lox.Type{
				"$.0.Body.0.Result":            func_(ts_(uref_(), uref_()), uref_()),                  // line 2: a+b+c
				"$.0.Body.0.Result.Left":       func_(ts_(uref_(), uref_()), uref_()),                  // line 2: a+b
				"$.0.Body.0.Result.Left.Left":  uref_(),                                                // line 2: a
				"$.0.Body.0.Result.Left.Right": bref_(num_),                                            // line 2: b
				"$.0.Body.0.Result.Right":      bref_(num_),                                            // line 2: c
				"$.1.Expression.Callee":        func_(ts_(uref_(), bref_(num_), bref_(num_)), uref_()), // line 4: add3
				"$.2.Expression.Callee":        func_(ts_(uref_(), bref_(num_), bref_(num_)), uref_()), // line 5: add3
			},
		},
	}
	for _, test := range tests {
		t.Run(test.text, func(t *testing.T) {
			stmts := parse(t, test.text)
			c := typing.NewChecker()
			types, err := c.Check(stmts)
			if err != nil {
				t.Errorf("%q: got err: %v", test.text, err)
				return
			}
			want := make(map[lox.Expr]lox.Type)
			for path, type_ := range test.paths {
				elem, err := valuepath.Walk(path, stmts)
				if err != nil {
					t.Fatalf("%q: invalid path %q for %v: %v", test.text, path, stmts, err)
				}
				want[elem.(lox.Expr)] = type_
			}
			if diff := cmp.Diff(want, types, ignoreTypeFields); diff != "" {
				t.Errorf("(-want,+got):\n%s", diff)
			}
		})
	}
}
