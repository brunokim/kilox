package lox

//go:generate go run ./cmd/gen_ast -spec ./cmd/gen_ast/expr.spec -dest expr.go -extensions typename
//go:generate go run ./cmd/gen_ast -spec ./cmd/gen_ast/stmt.spec -dest stmt.go
//go:generate go run ./cmd/gen_ast -spec ./cmd/gen_ast/type.spec -dest type.go
//go:generate go run ./cmd/gen_ast -spec ./cmd/gen_ast/term.spec -dest term.go

// ---- String

func (e *BinaryExpr) String() string     { return PrintExpr(e) }
func (e *GroupingExpr) String() string   { return PrintExpr(e) }
func (e *LiteralExpr) String() string    { return PrintExpr(e) }
func (e *UnaryExpr) String() string      { return PrintExpr(e) }
func (e *VariableExpr) String() string   { return PrintExpr(e) }
func (e *AssignmentExpr) String() string { return PrintExpr(e) }
func (e *CallExpr) String() string       { return PrintExpr(e) }
func (e *FunctionExpr) String() string   { return PrintExpr(e) }
func (e *GetExpr) String() string        { return PrintExpr(e) }
func (e *SetExpr) String() string        { return PrintExpr(e) }
func (e *ThisExpr) String() string       { return PrintExpr(e) }

func (s ExpressionStmt) String() string { return PrintStmts(s) }
func (s PrintStmt) String() string      { return PrintStmts(s) }
func (s VarStmt) String() string        { return PrintStmts(s) }
func (s IfStmt) String() string         { return PrintStmts(s) }
func (s BlockStmt) String() string      { return PrintStmts(s) }
func (s LoopStmt) String() string       { return PrintStmts(s) }
func (s BreakStmt) String() string      { return PrintStmts(s) }
func (s ContinueStmt) String() string   { return PrintStmts(s) }
func (s FunctionStmt) String() string   { return PrintStmts(s) }
func (s ReturnStmt) String() string     { return PrintStmts(s) }
func (s ClassStmt) String() string      { return PrintStmts(s) }

func (t NilType) String() string      { return PrintType(t) }
func (t BoolType) String() string     { return PrintType(t) }
func (t NumberType) String() string   { return PrintType(t) }
func (t StringType) String() string   { return PrintType(t) }
func (t FunctionType) String() string { return PrintType(t) }
func (t *RefType) String() string     { return PrintType(t) }

func (t AtomTerm) String() string    { return PrintTerm(t) }
func (t VarTerm) String() string     { return PrintTerm(t) }
func (t FunctorTerm) String() string { return PrintTerm(t) }
func (t ListTerm) String() string    { return PrintTerm(t) }
func (t ClauseTerm) String() string  { return PrintTerm(t) }
