package typing2

import (
	"github.com/brunokim/lox"
)

// ----

func atom(name string) lox.AtomTerm {
	return lox.AtomTerm{name}
}

func var_(name string) lox.VarTerm {
	return lox.VarTerm{name}
}

func functor(name string, terms ...lox.Term) lox.FunctorTerm {
	return lox.FunctorTerm{name, terms}
}

func list(terms ...lox.Term) lox.Term {
	return lox.TermsToList(terms...)
}

func clause(head lox.FunctorTerm, body ...lox.FunctorTerm) lox.ClauseTerm {
	return lox.ClauseTerm{head, body}
}

func clauses(clauses ...lox.ClauseTerm) []lox.ClauseTerm {
	return clauses
}

// ----

func RunClauses(clauses []lox.ClauseTerm) ([]lox.Bindings, error) {
	return nil, nil
}

func resolveBindings(term lox.Term, bindingsList []lox.Bindings) []lox.Term {
	return nil
}
