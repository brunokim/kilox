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
			`Type(foo, Fun([], ret)) :-
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
			`Type(foo, Fun([], ret)) :-
               _foo = Fun([], ret),
               ret = Nil.
            Type(bar, Fun([], ret)) :-
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
			`Type(foo, Fun([], ret)) :-
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
			`Type(foo, Fun([], ret)) :-
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
			`Type(foo, Fun([], ret)) :-
               _foo = Fun([], ret),
               % foo;
               % print foo;
               ret = Nil.`,
			clauses_(clause_("foo", func_(types_(), refi_(1)),
				binding_(refi_(2), func_(types_(), refi_(1))),
				binding_(refi_(1), nil_))),
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
