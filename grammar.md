# Lox grammar

Top-level

    program     ::= declaration* eof;
    declaration ::= classDecl
                  | funDecl
                  | varDecl
                  | statement
                  ;

Declaration
    
    classDecl ::= "class" identifier "{" method* "}" ;
    funDecl   ::= "fun" identifier function ;
    varDecl   ::= "var" identifier ( "=" expression )? ";" ;
    statement ::= exprStmt
                | printStmt
                | ifStmt
                | block
                | whileStmt
                | forStmt
                | breakStmt
                | continueStmt
                | returnStmt
                ;

Statements

    exprStmt  ::= expression ";" ;
    printStmt ::= "print" expression ";" ;
    ifStmt    ::= "if" "(" expression ")" statement ("else" statement)? ;
    block     ::= "{" declaration* "}" ;
    whileStmt ::= "while" "(" expression ")" statement ;
    forStmt   ::= "for" "(" forInit expression? ";" expression? ")" statement ;
    forInit   ::= varDecl | exprStmt | ";" ;

Sub-statements

    breakStmt    ::= "break" ";" ;
    continueStmt ::= "continue" ";" ;
    returnStmt   ::= "return" expression? ";"

Expressions

    expression ::= assignment ;
    assignment ::= identifier "=" assignment ;
                 | logic_or
                 ;
    logic_or   ::= logic_and ("or" logic_and)* ;
    logic_and  ::= equality ("and" equality)* ;
    equality   ::= comparison (("!="|"==") comparison)* ;
    comparison ::= term ((">"|"<"|">="|"<=") term)* ;
    term       ::= factor (("-"|"+") factor)* ;
    factor     ::= unary (("/"|"*") unary)* ;
    unary      ::= ("!"|"-") unary
                 | call
                 ;
    call       ::= primary ( "(" arguments? ")" | "." identifier )* ;
    arguments  ::= expression ( "," expression )* ;
    primary    ::= number | string | "true" | "false" | "nil"
                 | "(" expression ")"
                 | anonFunction
                 | identifier
                 ;

Functions

    anonFunction ::= "fun" function ;
    method       ::= identifier function ;
    function     ::= "(" parameters? ")" block ;
    parameters   ::= identifier ("," identifier)* ;


Tokens (handled in `scanner.go`)

    identifier  ::= alpha alphanum* ;
    number      ::= digit+ ('.' digit+)? ;
    string      ::= '"' string_char* '"' ;
    string_char ::= unicode - ["\\] | "\\\\" | "\\\"" ;
    alpha       ::= [a-zA-Z_] ;
    digit       ::= [0-9] ;
    alphanum    ::= alpha | digit ;

