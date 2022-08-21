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
               ret = Nil.`,
			clauses_(clause_(1, "foo", func_(types_(), refi_(6)),
				binding_(refi_(6), nil_))),
		},
		{
			dedent.Dedent(`
            fun foo() {}
            fun bar() {}
            `),
			`Type("foo", Fun([], ret)) :-
               ret = Nil.
            Type("bar", Fun([], ret)) :-
               ret = Nil.`,
			clauses_(
				clause_(1, "foo", func_(types_(), refi_(6)),
					binding_(refi_(6), nil_)),
				clause_(2, "bar", func_(types_(), refi_(8)),
					binding_(refi_(8), nil_))),
		},
		{
			"fun answer() { return 42; }",
			`Type("answer", Fun([], ret)) :-
               ret = Number.`,
			clauses_(clause_(1, "answer", func_(types_(), refi_(6)),
				binding_(refi_(6), num_))),
		},
		{
			dedent.Dedent(`
            fun foo() {
              var a = "test";
              return a;
            }`),
			`Type("foo", Fun([], ret)) :-
               _a = String, % implicit
               ret = _a.`,
			clauses_(clause_(1, "foo", func_(types_(), refi_(6)),
				binding_(refi_(6), brefi_(7, str_)))),
		},
		{
			dedent.Dedent(`
            fun foo() {
                foo;
                print foo;
            }`),
			`Type("foo", Fun([], ret)) :-
               Type("foo", r1),
               Type("foo", r2),
               ret = Nil.`,
			clauses_(clause_(1, "foo", func_(types_(), refi_(6)),
				binding_(refi_(6), nil_))),
		},
		{
			"fun id(x) { return x; }",
			`Type("id", Fun([_x], ret)) :-
               ret = _x.`,
			clauses_(clause_(1, "id", func_(types_(refi_(6)), refi_(7)),
				binding_(refi_(7), refi_(6)))),
		},
		{
			dedent.Dedent(`
            fun f(x) {
              var a;
              a = x;
              return a;
            }`),
			`Type("f", Fun([_x], ret)) :-
               _a = _x,
               ret = _a.`,
			clauses_(clause_(1, "f", func_(types_(refi_(6)), refi_(7)),
				binding_(refi_(8), refi_(6)),
				binding_(refi_(7), refi_(8)))),
		},
		{
			dedent.Dedent(`
            fun f(x) {
              var a;
              a = 20;
              return a = x;
            }`),
			`Type("f", Fun([_x], ret)) :-
               _a = Number,
               _a = _x,
               ret = _x.`,
			clauses_(clause_(1, "f", func_(types_(refi_(6)), refi_(7)),
				binding_(refi_(8), num_),
				binding_(refi_(8), refi_(6)),
				binding_(refi_(7), refi_(6)))),
		},
		{
			dedent.Dedent(`
            fun foo() {
              var a = 20;
              a = "str";
              return a;
            }`),
			`Type("foo", Fun([], ret)) :-
               _a = Number, % implicit
               _a = String,
               ret = _a.`,
			clauses_(clause_(1, "foo", func_(types_(), refi_(6)),
				binding_(brefi_(7, num_), str_),
				binding_(refi_(6), brefi_(7, num_)))),
		},
		{
			dedent.Dedent(`
            fun callWith43(f, x) {
                var a = 43;
                return f(a, x);
            }`),
			`Type("callWith43", Fun([_f, _x], ret)) :-
               _a = Number, % implicit
               _f = Fun([_a, _x], r1),
               ret = r1.`,
			clauses_(clause_(1, "callWith43", func_(types_(refi_(6), refi_(7)), refi_(8)),
				unify_(refi_(6), func_(types_(brefi_(9, num_), refi_(7)), refi_(10))),
				binding_(refi_(8), refi_(10)))),
		},
		{
			dedent.Dedent(`
            // Boolean primitives as functions. "l_" stands for "lambda calculus".
            fun l_true(x, y) { return x; }
            fun l_false(x, y) { return y; }
            fun l_if(cond, l_then, l_else) { return cond(l_then, l_else)(); }

            // Test it out.
            fun l_20() { return 20; }
            fun l_30() { return 30; }
            fun main() { print l_if(l_false, l_20, l_30); }
            `),
			`Type("l_true", Fun([_x, _y], ret)) :-
               ret = _x.
             Type("l_false", Fun([_x, _y], ret)) :-
               ret = _y.
             Type("l_if", Fun([_cond, _then, _else], ret)) :-
               _cond = Fun([_then, _else], r1),
               r1 = Fun([], r2),
               ret = r2.
             Type("l_20", Fun([], ret)) :-
               ret = Number.
             Type("l_30", Fun([], ret)) :-
               ret = Number.
             Type("main", Fun([], ret)) :-
               Type("l_if", Fun([_l_false, _l_20, _l_30], r1)),
               Type("l_false", _l_false),
               Type("l_20", _l_20),
               Type("l_30", _l_30),
               ret = r1.`,
			clauses_(
				clause_(1, "l_true", func_(types_(refi_(6), refi_(7)), refi_(8)),
					binding_(refi_(8), refi_(6))),
				clause_(2, "l_false", func_(types_(refi_(10), refi_(11)), refi_(12)),
					binding_(refi_(12), refi_(11))),
				clause_(3, "l_if", func_(types_(refi_(14), refi_(15), refi_(16)), refi_(17)),
					unify_(refi_(14), func_(types_(refi_(15), refi_(16)), refi_(18))),
					unify_(refi_(18), func_(types_(), refi_(19))),
					binding_(refi_(17), refi_(19))),
				clause_(4, "l_20", func_(types_(), refi_(21)),
					binding_(refi_(21), num_)),
				clause_(5, "l_30", func_(types_(), refi_(23)),
					binding_(refi_(23), num_)),
				clause_(6, "main", func_(types_(), refi_(25)),
					unify_(
						brefi_(13, func_(types_(refi_(14), refi_(15), refi_(16)), refi_(17))),
						func_(
							types_(
								brefi_(9, func_(types_(refi_(10), refi_(11)), refi_(12))),
								brefi_(20, func_(types_(), refi_(21))),
								brefi_(22, func_(types_(), refi_(23)))),
							refi_(26))),
					binding_(refi_(25), nil_)),
			),
		},
		{
			`fun foo() { return "a" + "b"; }`,
			`Type("foo", Fun([], ret)) :-
               Type("+", _plus),
               _plus = Fun([String, String], r1),
               ret = r1.`,
			clauses_(clause_(1, "foo", func_(types_(), refi_(6)),
				unify_(
					brefi_(1, func_(types_(num_, num_), num_)),
					func_(types_(str_, str_), refi_(7))),
				binding_(refi_(6), refi_(7)))),
		},
		{
			dedent.Dedent(`
            fun outer() {
              var a = 10;
              fun inner() {
                return a;
              }
              a = 20;
              inner();
            }`),
			`
            :- dynamic _a.
            Type("outer", Fun([], ret)) :-
              _a = Number, % implicit
              _a = Number,
              Type("inner", Fun([], r1)),
              ret = Nil.
            Type("inner", Fun([], ret)) :-
              ret = _a.`,
			clauses_(
				clause_(2, "inner", func_(types_(), refi_(9)),
					binding_(refi_(9), brefi_(7, num_))),
				clause_(1, "outer", func_(types_(), refi_(6)),
					binding_(brefi_(7, num_), num_),
					unify_(brefi_(8, func_(types_(), refi_(9))), func_(types_(), refi_(10))),
					binding_(refi_(6), nil_))),
		},
		{
			dedent.Dedent(`
            fun f1() {
              fun g() { return 10; }
              return g();
            }
            fun g() { return 30; }
            fun f2() {
              var a = h();
              fun g() { return a; }
              return g();
            }
            fun h() { return 40; }
            `),
			`
            Type("f1", Fun([], ret)) :-
              Type("f1/g", r1),
              r1 = Fun([], r2),
              ret = r2.
            Type("f1/g", Fun([], ret)) :-
              ret = Number.,

            Type("g", Fun([], ret)) :-
              ret = Number.

            :- dynamic _a.
            Type("f2", Fun([], ret)) :-
              Type("h", r1),
              _a = r1, % implicit
              Type("f2/g", r1),
              r1 = Fun([], r2),
              ret = r2.
            Type("f2/g", Fun([], ret)) :-
              ret = _a.

            Type("h", Fun([], ret)) :-
              ret = Number.`,
			clauses_(
				clause_(2, "g", func_(types_(), refi_(8)),
					binding_(refi_(8), num_)),
				clause_(1, "f1", func_(types_(), refi_(6)),
					unify_(brefi_(7, func_(types_(), refi_(8))), func_(types_(), refi_(9))),
					binding_(refi_(6), refi_(9))),
				clause_(3, "g", func_(types_(), refi_(11)),
					binding_(refi_(11), num_)),
				clause_(5, "g", func_(types_(), refi_(18)),
					binding_(refi_(18), brefi_(14, refi_(16)))),
				clause_(4, "f2", func_(types_(), refi_(13)),
					call_(6, refi_(15)),
					unify_(refi_(15), func_(types_(), refi_(16))),
					unify_(brefi_(17, func_(types_(), refi_(18))), func_(types_(), refi_(19))),
					binding_(refi_(13), refi_(19))),
				clause_(6, "h", func_(types_(), refi_(21)),
					binding_(refi_(21), num_)),
			),
		},
        {
            dedent.Dedent(`
            fun f() { g(); h(); }
            fun g() { h(); }
            fun h() {}`),
            ``,
            clauses_(
                clause_(1, "f", func_(types_(), refi_(6)),
                    call_(2, refi_(7)),
                    unify_(refi_(7), func_(types_(), refi_(8))),
                    call_(3, refi_(9)),
                    unify_(refi_(9), func_(types_(), refi_(10))),
                    binding_(refi_(6), nil_)),
                clause_(2, "g", func_(types_(), refi_(12)),
                    call_(3, refi_(13)),
                    unify_(refi_(13), func_(types_(), refi_(14))),
                    binding_(refi_(12), nil_)),
                clause_(3, "h", func_(types_(), refi_(16)),
                    binding_(refi_(16), nil_)),
            ),
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
