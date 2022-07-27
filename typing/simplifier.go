package typing

import (
	"github.com/brunokim/lox"
)

type simplifier struct {
	currType lox.Type
}

func simplifyType(t lox.Type) lox.Type {
	s := simplifier{}
	return s.simplify(t)
}

func (s *simplifier) simplify(t lox.Type) lox.Type {
	t.Accept(s)
	return s.currType
}

func (s *simplifier) VisitNilType(t lox.NilType)       { s.currType = t }
func (s *simplifier) VisitBoolType(t lox.BoolType)     { s.currType = t }
func (s *simplifier) VisitNumberType(t lox.NumberType) { s.currType = t }
func (s *simplifier) VisitStringType(t lox.StringType) { s.currType = t }

func (s *simplifier) VisitFunctionType(t lox.FunctionType) {
	params := make([]lox.Type, len(t.Params))
	for i, param := range t.Params {
		params[i] = s.simplify(param)
	}
	s.currType = lox.FunctionType{
		Params: params,
		Return: s.simplify(t.Return),
	}
}

func (s *simplifier) VisitRefType(t *lox.RefType) {
	s.currType = t
}
