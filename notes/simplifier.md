# Simplifying unification results

Unification may produce a graph of terms that is more
abstract than necessary, with Refs pointing to a single
value or having a lone constraint.

Diagram of admissible edges:

                    .-.
                    v |
                .----------.<------.
        .-------| Function |-----. |
        |       '----------'     | |
        v            ^           v |
    .--------.       |         .-----.
    | "Atom" |<------|---------| Ref |
    '--------'       |   .---->'-----'
        ^            |   |       |^ |
        |       .------------.   |'-'
        '-------| Constraint |<--' 
                '------------'
                

Disconsidering the case of bound refs, that can be replaced
by their bound values, we're left with

                    .-.
                    v |
                .----------.        
        .-------| Function |------. 
        |       '----------'      | 
        v            ^            v 
    .--------.       |         .-----.
    | "Atom" |       |   .---->| Ref |
    '--------'       |   |     '-----'
        ^            |   |        |  
        |       .------------.    |  
        '-------| Constraint |<---'
                '------------'
                
To simplify a term, we need to

1. List all reachable refs from the root solution.
2. List all constraints reachable from these reachable refs
3. Repeat until no new ref/constraint is added
4. Remove any ref from constraints that is not reachable
  - Note that a reachable ref in a constraint is not only a ref that is
    within it: it is a ref reachable referenced by another reachable ref,
    directly or indirectly via a function or chain of functions.
5. Bind any constraint that is constant among all constraints for a ref
6. Split independent constraints

## Examples

### Nothing to simplify

    - root: f([x, y], x)
    - constraints:
      - &c1 [{x: Num, y: Num}]
      - &c2 [{x: Str, y: Str}]
    - unbound_refs:
      x: [*c1, *c2]
      y: [*c1, *c2]

### Constant unbound ref

x receives always the same element:

    - root: f([x, y], x)
    - constraints:
      - &c1 [{x: Num, y: Num}]
      - &c2 [{x: Num, y: Str}]
    - unbound_refs:
      x: [*c1, *c2]
      y: [*c1, *c2]

Which becomes
   
    - root: f([Num, y], Num)
    - constraints:
      - &c1 [{y: Num}]
      - &c2 [{y: Str}]
    - unbound_refs:
      y: [*c1, *c2]

### Single constraint

A special case of the previous one is when a ref has a single constraint

    - root: f([x, y], x)
    - constraints:
      - &c1 [{x: Num, y: Str}]
    - unbound_refs:
      x: [*c1]
      y: [*c1]

Becomes

    - root: f([Num, Str], Num)

### Free reference

A subtle case happens when a ref is not listed in a constraint. In that case,
it keeps being free.

    - root: f([x, y], x)
    - constraints:
      - &c1 [{x: Num, y: Str}]
      - &c2 [{x: Str}]
    - unbound_refs:
      x: [*c1, *c2]
      y: [*c1, *c2]

In this case, nothing changes.

# STOP. Do we really need to associate constraints to a ref?

Can't we just... return all constraints for a given unification, i.e., the
available solutions to it?

## Logic model

    type(+, func([num, num], num)).
    type(+, func([str, str], str)).

    type(-, func([num, num], num)).
    type(-, func([num], num)).

    type(==, func([_, _], bool)).

    type(or, func([T1, _], T1)).
    type(or, func([_, T2], T2)).

    call(Func, Args, Result) :-
        type(Func, func(Args, Result)).

## Sample Lox program

    fun add3(x, y, z) {
        return x + y + z;
    }

    print add3(100, 200, 300);
    print add3("a", "b", "c");

### Typing the program

    type(add3, func([X, Y, Z], Ret)) :-
        call(+, [X, Y], R1),
        call(+, [R1, Z], R2),
        Ret = R2.

    type(add3_line4, Ret) :-
        call(add3, [num, num, num], Ret).

    type(add3_line5, Ret) :-
        call(add3, [str, str, str], Ret).

### Execution

    type(add3_line4, Ret).
    call(add3, [num, num, num], Ret).
    type(add3, func([num, num, num], Ret)).
    call(+, [num, num], R1),      call(+, [R1, Z], R2), Ret = R2.
    type(+, func([num, num], R1), call(+, [R1, Z], R2), Ret = R2.
    R1 = num,                     call(+, [R1, Z], R2), Ret = R2.
    call(+, [num, Z], R2),       Ret = R2. {R1: num}
    type(+, func([num, Z], R2)), Ret = R2. {R1: num}
    Z = num, R2 = num,           Ret = R2. {R1: num}

    {R1: num, Z: num, R2: num, Ret: num}

## Another program

    fun p123(x, f) {
        var a = 123;
        return f(x, a);
    }

    fun add(x, y) {
        return x + y;
    }

    print p123(10000, add); # line 10
    print p123("str", add); # line 11

### Typing

    type(p123, func([X, F], Ret)) :-
        A = num,
        call(F, [X, A], Ret).

    type(add, [X, Y], Ret) :-
        call(+, [X, Y], Ret).

    type(p123_line10, Ret) :-
        call(p123, [num, add], Ret).

    type(p123_line11, Ret) :-
        call(p123, [str, add], Ret).

### Execution

    type(p123_line10, Ret).
    call(p123, [num, add], Ret).
    type(p123, func([num, add], Ret)).
    X=num, F=add, A=num, call(F, [X, A], Ret).
    call(add, [num, num], Ret).       {X: num, F: add, A: num}
    type(add, func([num, num], Ret)). {X: num, F: add, A: num}
    call(+, [num, num], Ret).         {X: num, F: add, A: num}

    {X: num, F: add, A: num, Ret: num}


    type(p123_line11, Ret).
    call(p123, [str, add], Ret).
    type(p123, func([str, add], Ret)).
    X=str, F=add, A=num, call(F, [X, A], Ret).
    call(add, [str, num], Ret).       {X: str, F: add, A: num}
    type(add, func([str, num], Ret)). {X: str, F: add, A: num}
    call(+, [str, num], Ret).         {X: str, F: add, A: num}

    ERROR!

