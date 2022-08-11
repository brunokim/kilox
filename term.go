// Generated file, do not modify
// Invocation: gen_ast -spec ./cmd/gen_ast/term.spec -dest term.go
package lox

type Term interface {
	Accept(v termVisitor)
}

type termVisitor interface {
	VisitVarTerm(t VarTerm)
	VisitAtomTerm(t AtomTerm)
	VisitFunctorTerm(t FunctorTerm)
	VisitListTerm(t ListTerm)
	VisitClauseTerm(t ClauseTerm)
}

type VarTerm struct {
	Name string
}

type AtomTerm struct {
	Name string
}

type FunctorTerm struct {
	Name string
	Args []Term
}

type ListTerm struct {
	Head Term
	Tail Term
}

type ClauseTerm struct {
	Head FunctorTerm
	Body []FunctorTerm
}

func (t VarTerm) Accept(v termVisitor) {
	v.VisitVarTerm(t)
}

func (t AtomTerm) Accept(v termVisitor) {
	v.VisitAtomTerm(t)
}

func (t FunctorTerm) Accept(v termVisitor) {
	v.VisitFunctorTerm(t)
}

func (t ListTerm) Accept(v termVisitor) {
	v.VisitListTerm(t)
}

func (t ClauseTerm) Accept(v termVisitor) {
	v.VisitClauseTerm(t)
}
