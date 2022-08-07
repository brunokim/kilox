# Forward declaration

## Program

    fun add_or_sub(ch, x, y) {
        var fn;
        if (ch == '+')
            fn = add;
        else
            fn = sub;
        return fn(x, y);
    }

    fun add(a, b) {
        return a + b;
    }

    fun sub(a, b) {
        return a - b;
    }

## Logic program

    type(add_or_sub, fun([Ch, X, Y], Ret)) :-
        type(==, fun([Ch, str], _)),
        (   Ch = str,
            type(add, Fn)
        ;   type(sub, Fn)
        ),
        Fn = fun([X, Y], R1),
        Ret = R1.

The `if` clause becomes a disjunction of terms. The condition is handled specially on the
`Ch = str` equality that sits on a single branch. We are working with broad classes here,
so it's not possible to say that `Ch = literal("+")` or similar.

    type(add, fun([A, B], Ret)) :-
        type(+, fun([A, B], R1)),
        Ret = R1.

    type(sub, fun([A, B], Ret)) :-
        type(-, fun([A, B], R1)),
        Ret = R1.
    
    % Facts
    type(+, fun([num, num], num)).
    type(+, fun([str, str], str)).

    type(-, fun([num], num)).
    type(-, fun([num, num], num)).

    type(==, fun([T1, T2], bool)).

The logic program only works because the whole program is compiled before execution,
so we can forward reference `type(add, ...)` and `type(sub, ...)` in `type(add_or_sub, ...)`.

## Expected types

    ?- type(add_or_sub, T).
    T = fun([str, num, num], num) ;
    T = fun([str, str, str], str) ;
    T = fun([_, num, num], num).

    ?- type(add, T).
    T = fun([num, num], num) ;
    T = fun([str, str], str).

    ?- type(sub, T).
    T = fun([num, num], num).

# Multiple returns

## Program

    fun oneof(x, y, z) {
        var r = random();
        if (r < 1/3) return x;
        if (r < 2/3) return y;
        return z;
    }

## Logic program

    type(oneof, fun([X, Y, Z], Ret)) :-
        type(random, fun([], R1)),
        R = R1,
        type(<, fun([R, num], _)),
        (   Ret = X
        ;   type(<, fun([R, num], _)),
            (   Ret = Y
            ;   Ret = Z
            )
        ).

    % Facts
    type(random, fun([], num)).

    type(<, fun([num, num], bool)).

## Expected types

    ?- type(oneof, T).
    T = fun([X, _, _], X) ;
    T = fun([_, Y, _], Y) ;
    T = fun([_, _, Z], Z).

## WAM compilation

In this pseudocode, `&Reg` means a register address, and `@Addr` means
an environment address.

    % type(oneof, fun([X, Y, Z], Ret)) :-
    allocate   5         % perm vars: X, Y, Z, Ret, R
    get_atom   X0, oneof
    get_struct X1, fun/2
    unify_var  &Params
    unify_var  @Ret

    get_list   &Params
    unify_var  @X
    unify_var  &Params

    get_list   &Params
    unify_var  @Y
    unify_var  &Params

    get_list   &Params
    unify_var  @Z
    unify_atom []

    %   type(random, fun([], R1)),
    put_atom   X0, random
    put_struct X1, fun/2
    unify_atom []
    unify_var  &R1
    call       type/2

    %   R = R1,
    get_var &R1, @R

    %   type(<, fun([R, num], _)),
    put_atom   <, X0
    put_struct fun/2, X1
    unify_var  &Params
    unify_var  &_

    put_list   &Params
    unify_val  @R
    unify_var  &Params

    put_list   &Params
    unify_atom num
    unify_atom []

    call       type/2

    %   (   Ret = X
    try      :else1
    get_val  @Ret, @X
    goto     :end1
    
    :else1
    trust
    %   ;   type(<, fun([R, num], _)),
    put_atom   <, X0
    put_struct fun/2, X1
    unify_var  &Params
    unify_var  &_

    put_list   &Params
    unify_val  @R
    unify_var  &Params

    put_list   &Params
    unify_atom num
    unify_atom []

    call       type/2

    %       (   Ret = Y
    try     :else2
    get_val @Ret, @Y
    goto    :end2

    :else2
    trust
    %       ;   Ret = Z
    get_val @Ret, @Z
    goto    :end2

    %       )
    :end2

    %   ).
    :end1
    deallocate
    proceed

## Procedural execution, without comments

    t = Ref("T")
    call("type", 'oneof', Ref("T"))

    x0, x1 = 'oneof', t
    unify(x0, 'oneof')

    params = List()
    ret = Ref("Ret")
    t.value = Struct("fun", [
        [Ref("X"), Ref("Y"), Ref("Z")],
        Ref("Ret"),
    ])

    r1 = Ref("R1")
    call("type", 'random', Struct("fun", ['[]', r1]))
    // -> r1.value = 'num'

    r = r1

    call("type", '<', Struct("fun", [
        [r, 'num'],
        Ref(),
    ])
    // -> r.value = num

    pushChoice()
    unify(ret, x)
    // -> ret.value = x

    yield {'T': t"fun([X, Y, Z], X)"}

    popChoice()
    call("type", '<', Struct("fun", [
        [r, 'num'],
        Ref(),
    ])
    // -> r.value = num

    pushChoice()
    unify(ret, y)
    // -> ret.value = y

    yield {'T': t"fun([X, Y, Z], Y)"}

    popChoice()

    unify(ret, z)
    // -> ret.value = z

    yield {'T': t"fun([X, Y, Z], Z)"}

## Tree walking

    visitFunctionStmt:
      name: oneof
      params:
      - token: x
      - token: y
      - token: z
      action0: |
        x, y, z, ret = Ref(), Ref(), Ref(), Ref()
        bind("oneof", Fun([x, y, z], ret)
        pushScope(ret)
        bind("x", x)
        bind("y", y)
        bind("z", z)
      body:
      - visitVarStmt:
          value:
          - visitCallExpr:
              callee:
              - visitVariableExpr:
                  name: random
                  action: |
                    return get("random")
              args: []
              action: |
                ret = Ref()
                unify(callee, Fun([], ret))
                return ret
          target:
            name: r
            action: |
              bind("r", value)
      - visitIfStmt:
          condition:
            visitBinaryExpr:
              left:
                visitVariableExpr:
                  name: r
                  action: |
                    return get("r")
              right:
                visitBinaryExpr:
                  left:
                    visitLiteral:
                      value: 1
                      action: |
                        return 'num'
                  right:
                    visitLiteral:
                      value: 3
                      action: |
                        return 'num'
                  op:
                    name: /
                    action:
                      return get("/")
                  action: |
                    ret = Ref()
                    unify(op, Fun([left, right], ret))
                    return ret
              op:
                name: <
                action: |
                  return get("<")
              action: |
                ret = Ref()
                unify(op, Fun([left, right], ret))
                return ret
          then:
            action0: |
              pushChoice()
            visitReturnStmt:
              expression:
                visitVariableExpr:
                  name: x
                  action: |
                    return get("x")
              action: |
                unify(currentScope().ret, expression)
                yield currentScope().trailBindings

       - visitIfStmt:
          condition:
            visitBinaryExpr:
              left:
                visitVariableExpr:
                  name: r
                  action: |
                    return get("r")
              right:
                visitBinaryExpr:
                  left:
                    visitLiteral:
                      value: 2
                      action: |
                        return 'num'
                  right:
                    visitLiteral:
                      value: 3
                      action: |
                        return 'num'
                  op:
                    name: /
                    action:
                      return get("/")
                  action: |
                    ret = Ref()
                    unify(op, Fun([left, right], ret))
                    return ret
              op:
                name: <
                action: |
                  return get("<")
              action: |
                ret = Ref()
                unify(op, Fun([left, right], ret))
                return ret
          then:
            action0: |
              pushChoice()
            visitReturnStmt:
              expression:
                visitVariableExpr:
                  name: y
                  action: |
                    return get("y")
              action: |
                unify(currentScope().ret, expression)
                yield currentScope().trailBindings

        - visitReturnStmt:
          expression:
            visitVariableExpr:
              name: z
              action: |
                return get("z")
          action: |
            unify(currentScope().ret, expression)
            yield currentScope().trailBindings

## Tree walking actions

    def functionStmt():
      x, y, z, ret = Ref(), Ref(), Ref(), Ref()
      bind("oneof", Fun([x, y, z], ret)
      pushScope(ret)
      bind("x", x)
      bind("y", y)
      bind("z", z)
 
      def body():
        def expressionStmt():
          def expression():
            def value():
              def call():
                def callee():
                  return get("random")
                
                def args():
                  return []

                ret = Ref()
                unify(callee(), Fun(args(), ret))
                return ret

            bind("r", value())
          
          expression()
          yield from []

        def ifStmt1():
          def condition():
            def left():
              return get("r")

            def right():
              def left():
                return 'num'

              def right():
                return 'num'

              def op():
                return get("/")

              ret = Ref()
              unify(op(), Fun([left(), right()], ret))
              return ret

            def op():
              return get("<")

            ret = Ref()
            unify(op(), Fun([left, right], ret))
            return ret

          def then():
            def returnStmt():
              def expression():
                return get("x")

              unify(currScope().ret, expression())
              yield currScope().trailBindings
              backtrack()
           
            yield from returnStmt()

          condition()
          pushChoice([then])
          yield from tryChoice()

        def ifStmt2():
          def condition():
            def left():
              return get("r")

            def right():
              def left():
                return 'num'

              def right():
                return 'num'

              def op():
                return get("/")

              ret = Ref()
              unify(op(), Fun([left(), right()], ret))
              return ret

            def op():
              return get("<")

            ret = Ref()
            unify(op(), Fun([left, right], ret))
            return ret

          def then():
            def returnStmt():
              def expression():
                return get("y")

              unify(currScope().ret, expression())
              yield currScope().trailBindings
              backtrack()
           
            yield from returnStmt()

          condition()
          pushChoice([then])
          yield from tryChoice()

        def returnStmt():
          def expression():
            return get("z")
          
          unify(currScope().ret, expression())
          yield currScope().trailBindings
          backtrack()

        yield from expressionStmt()
        yield from ifStmt1()
        yield from ifStmt2()
        yield from returnStmt()


      constraints = body()
      try:
        first = next(constraints)
      except StopIteration:
        raise TypeError("couldn't find a solution")
      
      try:
        second = next(constraints)
        # There are 2 or more solutions 
        cs = [first, second]
        for i, constraint in constraints:
          if i == MAX_CONSTRAINTS:
            raise TypeError("too many solutions")
          cs.append(constraint)
        yield from cs

      except StopIteration:
        # There's a single solution
        applyConstraint(first.toConstraint())
        yield from []
      finally:
        popScope()


Still not clear...
