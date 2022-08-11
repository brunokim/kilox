package lox

type Bindings map[VarTerm]Term

// ----

func TermsToList(terms ...Term) Term {
	return TermsToIncompleteList(terms, ListTerm{})
}

func TermsToIncompleteList(terms []Term, tail Term) Term {
	for i := len(terms) - 1; i >= 0; i-- {
		tail = ListTerm{terms[i], tail}
	}
	return tail
}

func ListToTerms(l ListTerm) ([]Term, Term) {
	var t Term = l

	var terms []Term
	for {
		l, ok := t.(ListTerm)
		if !ok || l == (ListTerm{}) {
			return terms, t
		}
		terms = append(terms, l.Head)
		t = l.Tail
	}
}

// ----

func TermToType(term Term) Type {
	return nil
}
