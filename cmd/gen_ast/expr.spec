// Lox expressions
*Binary(Left: Expr, Operator: Token, Right: Expr)        // a + b
*Grouping(Expression: Expr)                              // (a)
*Literal(Token: Token, Value: any)                       // 123, "abc"
*Unary(Operator: Token, Right: Expr)                     // -a
*Variable(Name: Token)                                   // a
*Assignment(Name: Token, Value: Expr)                    // a = 1
*Logic(Left: Expr, Operator: Token, Right: Expr)         // x and y
*Call(Callee: Expr, Paren: Token, Args: []Expr)          // f(a, 1, true)
*Function(Keyword: Token, Params: []Token, Body: []Stmt) // fun(x, y) { }
*Get(Object: Expr, Name: Token)                          // obj.field
*Set(Object: Expr, Name: Token, Value: Expr)             // obj.field = 1
