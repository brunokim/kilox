package lox

type simplifier struct {
	currType Type
}

func simplifyType(t Type) Type {
	s := simplifier{}
	return s.simplify(t)
}

func (s *simplifier) simplify(t Type) Type {
	t.accept(s)
	return s.currType
}

func (s *simplifier) visitNilType(t NilType)           { s.currType = t }
func (s *simplifier) visitBoolType(t BoolType)         { s.currType = t }
func (s *simplifier) visitNumberType(t NumberType)     { s.currType = t }
func (s *simplifier) visitStringType(t StringType)     { s.currType = t }

func (s *simplifier) visitFunctionType(t FunctionType) {
    params := make([]Type, len(t.Params))
    for i, param := range t.Params {
        params[i] = s.simplify(param)
    }
    s.currType = FunctionType{
        Params: params,
        Return: s.simplify(t.Return),
    }
}

func (s *simplifier) visitRefType(t *RefType) {
    s.currType = t
}
