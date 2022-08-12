package typing_test

import (
	"testing"

	"github.com/brunokim/lox/typing"

	"github.com/google/go-cmp/cmp"
	"github.com/lithammer/dedent"
)

func TestBuildClauses(t *testing.T) {
	tests := []struct {
		text string
		doc  string
		want []typing.TypeClause
	}{
		{
			"fun foo() {}",
			`Type("foo", Fun([], ret)) :-
               _foo = Fun([], ret),
               ret = Nil.`,
			clauses_(clause_("foo", func_(types_(), refi_(1)),
				binding_(refi_(2), func_(types_(), refi_(1))),
				binding_(refi_(1), nil_))),
		},
		{
			dedent.Dedent(`
            fun foo() {}
            fun bar() {}
            `),
			`Type("foo", Fun([], ret)) :-
               _foo = Fun([], ret),
               ret = Nil.
            Type("bar", Fun([], ret)) :-
               _bar = Fun([], ret),
               ret = Nil.`,
			clauses_(
				clause_("foo", func_(types_(), refi_(1)),
					binding_(refi_(2), func_(types_(), refi_(1))),
					binding_(refi_(1), nil_)),
				clause_("bar", func_(types_(), refi_(3)),
					binding_(refi_(4), func_(types_(), refi_(3))),
					binding_(refi_(3), nil_))),
		},
		{
			"fun answer() { return 42; }",
			`Type("foo", Fun([], ret)) :-
               _foo = Fun([], ret),
               ret = Number.`,
			clauses_(clause_("answer", func_(types_(), refi_(1)),
				binding_(refi_(2), func_(types_(), refi_(1))),
				binding_(refi_(1), num_))),
		},
		{
			dedent.Dedent(`
            fun foo() {
              var a = "test";
              return a;
            }`),
			`Type("foo", Fun([], ret)) :-
               _foo = Fun([], ret),
               _a = String, % implicit
               ret = _a.`,
			clauses_(clause_("foo", func_(types_(), refi_(1)),
				binding_(refi_(2), func_(types_(), refi_(1))),
				binding_(refi_(1), brefi_(3, str_)))),
		},
		{
			dedent.Dedent(`
            fun foo() {
                foo;
                print foo;
            }`),
			`Type("foo", Fun([], ret)) :-
               _foo = Fun([], ret),
               % foo;
               % print foo;
               ret = Nil.`,
			clauses_(clause_("foo", func_(types_(), refi_(1)),
				binding_(refi_(2), func_(types_(), refi_(1))),
				binding_(refi_(1), nil_))),
		},
		{
			"fun id(x) { return x; }",
			`Type("id", Fun([_x], ret)) :-
               _foo = Fun([_x], ret),
               ret = _x.`,
			clauses_(clause_("id", func_(types_(refi_(1)), refi_(2)),
				binding_(refi_(3), func_(types_(refi_(1)), refi_(2))),
				binding_(refi_(2), refi_(1)))),
		},
		{
			dedent.Dedent(`
            fun f(x) {
              var a;
              a = x;
              return a;
            }`),
			`Type("f", Fun([_x], ret)) :-
               _f = Fun([_x], ret),
               _a = _x,
               ret = _a.`,
			clauses_(clause_("f", func_(types_(refi_(1)), refi_(2)),
				binding_(refi_(3), func_(types_(refi_(1)), refi_(2))),
				binding_(refi_(4), refi_(1)),
				binding_(refi_(2), refi_(4)))),
		},
		{
			dedent.Dedent(`
            fun f(x) {
              var a;
              a = 10;
              return a = x;
            }`),
			`Type("f", Fun([_x], ret)) :-
               _f = Fun([_x], ret),
               _a = Number,
               _a = _x,
               ret = _x.`,
			clauses_(clause_("f", func_(types_(refi_(1)), refi_(2)),
				binding_(refi_(3), func_(types_(refi_(1)), refi_(2))),
				binding_(refi_(4), num_),
				binding_(refi_(4), refi_(1)),
				binding_(refi_(2), refi_(1)))),
		},
		{
			dedent.Dedent(`
            fun foo() {
              var a = 10;
              a = "str";
              return a;
            }`),
			`Type("foo", Fun([], ret)) :-
               _foo = Fun([], ret),
               _a = Number, % implicit
               _a = String,
               ret = _a.`,
			clauses_(clause_("foo", func_(types_(), refi_(1)),
				binding_(refi_(2), func_(types_(), refi_(1))),
				binding_(brefi_(3, num_), str_),
				binding_(refi_(1), brefi_(3, num_)))),
		},
		{
			dedent.Dedent(`
            fun callWith32(f, x) {
                var a = 32;
                return f(a, x);
            }`),
			`Type("callWith32", Fun([_f, _x], ret)) :-
               _callWith32 = Fun([_f, _x], ret),
               _a = Number, % implicit
               _f = Fun([_a, _x], r1),
               ret = r1.`,
			clauses_(clause_("callWith32", func_(types_(refi_(1), refi_(2)), refi_(3)),
				binding_(refi_(4), func_(types_(refi_(1), refi_(2)), refi_(3))),
				unify_(refi_(1), func_(types_(brefi_(5, num_), refi_(2)), refi_(6))),
				binding_(refi_(3), refi_(6)))),
		},
		{
			dedent.Dedent(`
            //test:skip
            // Boolean primitives as functions.
            // "l_" stands for "lambda calculus".
            fun l_true(x, y) { return x; }
            fun l_false(x, y) { return y; }
            fun l_if(cond, l_then, l_else) { return cond(l_then, l_else)(); }

            // Test it out.
            fun l_10() { return 10; }
            fun l_20() { return 20; }
            fun main() { print l_if(l_false, l_10, l_20); }
            `),
			`Type("l_true", Fun([_x, _y], ret)) :-
               _l_true = Fun([_x, _y], ret),
               ret = _x.
             Type("l_false", Fun([_x, _y], ret)) :-
               _l_false = Fun([_x, _y], ret),
               ret = _y.
             Type("l_if", Fun([_cond, _then, _else], ret)) :-
               _l_if = Fun([_cond, _then, _else], ret),
               _cond = Fun([_then, _else], r1),
               r1 = Fun([], r2),
               ret = _r2.
             Type("l_10", Fun([], ret)) :-
               _l_10 = Fun([], ret),
               ret = Number.
             Type("l_20", Fun([], ret)) :-
               _l_20 = Fun([], ret),
               ret = Number.
             Type("main", Fun([], ret)) :-
               _main = Fun([], ret),
               Type("l_if", Fun([_l_false, _l_10, _l_20], r1)),
               Type("l_false", _l_false),
               Type("l_10", _l_10),
               Type("l_20", _l_20),
               ret = r1.`,
			nil,
		},
	}
	for _, test := range tests {
		t.Run(test.text, func(t *testing.T) {
			stmts := parse(t, test.text)
			clauses, err := typing.BuildClauses(stmts)
			if err != nil {
				t.Errorf("got err: %v", err)
				return
			}
			if diff := cmp.Diff(test.want, clauses, ignoreTypeFields); diff != "" {
				t.Errorf("(-want,+got):\n%s", diff)
			}
		})
	}
}
