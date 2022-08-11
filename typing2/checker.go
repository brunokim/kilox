package typing2

import (
	"fmt"

	"github.com/brunokim/lox"
	"github.com/brunokim/lox/errlist"
)

type typeError struct {
	t1, t2 lox.Type
}

func (err typeError) Error() string {
	return fmt.Sprintf("%v != %v", err.t1, err.t2)
}

// ----

type scope struct {
	enclosing  *scope
	currClause *lox.ClauseTerm
	hasReturn  bool
}

type Checker struct {
	errors  []typeError
	clauses []lox.ClauseTerm
	exprs   map[lox.Expr]lox.Term

	scope *scope

	currTerm lox.Term
}

func NewChecker() *Checker {
	return &Checker{
		exprs: make(map[lox.Expr]lox.Term),
	}
}

func (c *Checker) Build(stmts []lox.Stmt) ([]lox.ClauseTerm, error) {
	c.checkStmts(stmts)
	if len(c.errors) > 0 {
		return nil, errlist.Of[typeError](c.errors)
	}
	return c.clauses, nil
}

func (c *Checker) Check(stmts []lox.Stmt) (map[lox.Expr][]lox.Type, error) {
	clauses, err := c.Build(stmts)
	if err != nil {
		return nil, err
	}
	bindingsList, err := RunClauses(clauses)
	if err != nil {
		return nil, err
	}
	exprTypes := make(map[lox.Expr][]lox.Type)
	for expr, term := range c.exprs {
		terms := resolveBindings(term, bindingsList)
		types := make([]lox.Type, len(terms))
		for i, term := range terms {
			types[i] = lox.TermToType(term)
		}
		exprTypes[expr] = types
	}
	return exprTypes, nil
}

// ----

func (c *Checker) checkStmts(stmts []lox.Stmt) {
	for _, stmt := range stmts {
		c.checkStmt(stmt)
	}
}

func (c *Checker) checkStmt(stmt lox.Stmt) {
	stmt.Accept(c)
}

func (c *Checker) checkExpr(expr lox.Expr) lox.Term {
	expr.Accept(c)
	return c.currTerm
}

// ---- StmtVisitor

func (c *Checker) VisitExpressionStmt(s lox.ExpressionStmt) {
	panic(fmt.Sprintf("typing2.(*Checker).VisitExpressionStmt is not implemented"))
}

func (c *Checker) VisitPrintStmt(s lox.PrintStmt) {
	panic(fmt.Sprintf("typing2.(*Checker).VisitPrintStmt is not implemented"))
}

func (c *Checker) VisitVarStmt(s lox.VarStmt) {
	panic(fmt.Sprintf("typing2.(*Checker).VisitVarStmt is not implemented"))
}

func (c *Checker) VisitIfStmt(s lox.IfStmt) {
	panic(fmt.Sprintf("typing2.(*Checker).VisitIfStmt is not implemented"))
}

func (c *Checker) VisitBlockStmt(s lox.BlockStmt) {
	panic(fmt.Sprintf("typing2.(*Checker).VisitBlockStmt is not implemented"))
}

func (c *Checker) VisitLoopStmt(s lox.LoopStmt) {
	panic(fmt.Sprintf("typing2.(*Checker).VisitLoopStmt is not implemented"))
}

func (c *Checker) VisitBreakStmt(s lox.BreakStmt) {
	panic(fmt.Sprintf("typing2.(*Checker).VisitBreakStmt is not implemented"))
}

func (c *Checker) VisitContinueStmt(s lox.ContinueStmt) {
	panic(fmt.Sprintf("typing2.(*Checker).VisitContinueStmt is not implemented"))
}

func (c *Checker) VisitFunctionStmt(s lox.FunctionStmt) {
	defer func(old *scope) { c.scope = old }(c.scope)
	c.scope = &scope{enclosing: c.scope}

	params := make([]lox.Term, len(s.Params))
	for i, param := range s.Params {
		params[i] = var_("_" + param.Lexeme)
	}
	cl := clause(
		functor("type", atom(s.Name.Lexeme), functor("fun", list(params...), var_("Ret"))),
	)
	c.scope.currClause = &cl
	c.checkStmts(s.Body)

	if !c.scope.hasReturn {
		fmt.Println("no return")
		cl.Body = append(cl.Body, functor("=", atom("nil"), var_("Ret")))
	}
	c.clauses = append(c.clauses, cl)
}

func (c *Checker) VisitReturnStmt(s lox.ReturnStmt) {
	x := c.checkExpr(s.Result)
	cl := c.scope.currClause
	cl.Body = append(cl.Body, functor("=", x, var_("Ret")))
	c.scope.hasReturn = true
}

func (c *Checker) VisitClassStmt(s lox.ClassStmt) {
	panic(fmt.Sprintf("typing2.(*Checker).VisitClassStmt is not implemented"))
}

// ---- ExprVisitor

func (c *Checker) VisitBinaryExpr(e *lox.BinaryExpr) {
	panic(fmt.Sprintf("typing2.(*Checker).VisitBinaryExpr is not implemented"))
}

func (c *Checker) VisitGroupingExpr(e *lox.GroupingExpr) {
	panic(fmt.Sprintf("typing2.(*Checker).VisitGroupingExpr is not implemented"))
}

func (c *Checker) VisitLiteralExpr(e *lox.LiteralExpr) {
	switch e.Value.(type) {
	case bool:
		c.currTerm = atom("bool")
	case float64:
		c.currTerm = atom("number")
	case string:
		c.currTerm = atom("string")
	default:
		if e.Value == nil {
			c.currTerm = atom("nil")
		} else {
			panic(fmt.Sprintf("unhandled literal value type %[1]T (%[1]v)", e.Value))
		}
	}
}

func (c *Checker) VisitUnaryExpr(e *lox.UnaryExpr) {
	panic(fmt.Sprintf("typing2.(*Checker).VisitUnaryExpr is not implemented"))
}

func (c *Checker) VisitVariableExpr(e *lox.VariableExpr) {
	panic(fmt.Sprintf("typing2.(*Checker).VisitVariableExpr is not implemented"))
}

func (c *Checker) VisitAssignmentExpr(e *lox.AssignmentExpr) {
	panic(fmt.Sprintf("typing2.(*Checker).VisitAssignmentExpr is not implemented"))
}

func (c *Checker) VisitLogicExpr(e *lox.LogicExpr) {
	panic(fmt.Sprintf("typing2.(*Checker).VisitLogicExpr is not implemented"))
}

func (c *Checker) VisitCallExpr(e *lox.CallExpr) {
	panic(fmt.Sprintf("typing2.(*Checker).VisitCallExpr is not implemented"))
}

func (c *Checker) VisitFunctionExpr(e *lox.FunctionExpr) {
	panic(fmt.Sprintf("typing2.(*Checker).VisitFunctionExpr is not implemented"))
}

func (c *Checker) VisitGetExpr(e *lox.GetExpr) {
	panic(fmt.Sprintf("typing2.(*Checker).VisitGetExpr is not implemented"))
}

func (c *Checker) VisitSetExpr(e *lox.SetExpr) {
	panic(fmt.Sprintf("typing2.(*Checker).VisitSetExpr is not implemented"))
}

func (c *Checker) VisitThisExpr(e *lox.ThisExpr) {
	panic(fmt.Sprintf("typing2.(*Checker).VisitThisExpr is not implemented"))
}
