Var(Name: string)                               // Ret, _x, _y
Atom(Name: string)                              // fun, number
Functor(Name: string, Args: []Term)             // type(foo, X), =(string, _y)
List(Head: Term, Tail: Term)                    // [a, b, c]
Clause(Head: FunctorTerm, Body: []FunctorTerm)  // foo(X) :- =(X, number).

