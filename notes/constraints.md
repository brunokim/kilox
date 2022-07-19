# Constraints

In Prolog, an **attributed variable** is an unbound variable that has some
associated data. Before this variable is bound, a registered predicate is
called to evaluate if unification may proceed.

For example, let's say that we want `X` to express a range of possible
integer values. If it is bound to another variable that also represent a range,
the unification succeeds if their ranges overlap, and the resulting variable
has as new range the intersection of both of them. If the range reduces to a
single number, the var is now ground; if the intersection is empty, unification
fails.

This can be implemented as follows:

    main :-
        set_attr(X, range(0, 4)),
        set_attr(Y, range(1, 11)),
        set_attr(Z, range(4, 20)),

        (X = Y -> write_term(get_attr(X))),
        (X = Z -> write_term(get_attr(X))
        ; write_term("X and Z don't unify")),
        (Y = Z -> write_term(get_attr(Y))).

    verify_attributes(range/2, X, Y) :-
        get_attr(X, range(MinX, MaxX)),
        (var(Y), get_attr(Y, range(MinY, MaxY)) ->
            % If both variables have ranges, compute their interesection.
            intersection([MinX, MaxX], [MinY, MaxY], [Min, Max]),
            (Min + 1 #= Max ->
                % Range contains a single element, bind it.
                Y = Min
            ;   % New range is associated to Y.
                set_attr(Y, range(Min, Max))
            )
        ;   % Y is not a variable or doesn't have an attribute, so must
            % be between min and max of range.
            between(MinX, Y, MaxX)
        ).


## Type constraints

For our type checker, we need something simpler: the attributes consist only
of possible bindings, or **constraints**, for a Ref. There's no need for custom
code to be executed before unification; the only supported behavior is verifying
which combinations of constraints are satisfied.

For example:

    + : (t, t) -> t,
        t ~ [{t: Number}, {t: String}]
    - : t ~ [
            {t: (Number, Number) -> Number},
            {t: (Number) -> Number},
        ]

Some of the power comes from sharing the same constraints among multiple variables,
to represent their co-dependence.

    f : (t1, t2) -> t,
        t1 ~ [
            &c1 {t1: Number, t2: String, t: Number},
            &c2 {t1: String, t2: Number, t: Bool},
        ],
        t2 ~ [*c1, *c2],
        t ~ [*c1, *c2]

## Example

Consider the following function

    fun add3(x, y, z) {
        return x + y + z;
    }

We have the following types:

    add3 : (_1, _2, _3) -> _4
    x    : _1
    y    : _2
    z    : _3
    _ret : _4

The first "+" call (x+y) has type

    "+" : (_1, _2) -> _5

And is unified with
    
    "+" : (_6, _6) -> _6,
          _6 ~ [{_6: Number}, {_6: String}]

The unification stack becomes

    _1 = _6*
    --------
    _2 = _6*
    --------
    _5 = _6*
    --------
     stack

Now, since `_6` is constrained, we bind it to the first constraint and push
the remaining one to the choice stack, together with the current state.

    _1 = Number
    -----------  -----------------------------
    _2 = Number   constraints: [{_6: String}]
    -----------   curr_stack:  [(3 elems)]   
    _5 = Number   trail:       [_6]           
    -----------  -----------------------------
       stack                 choice

We bind all elements successfully and store that they're bound in the trail.

                 -------------------------------
                  constraints: [{_6: String}]
                  curr_stack:  [(3 elems)]   
                  trail:       [_6, _1, _2, _5]
    -----------  -------------------------------
       stack                 choice

Since unification was successful, we store this as a constraint. We don't associate to
the vars yet because we have more work to do, and they shouldn't appear to be constrained.

    c1 = {_6: Number, _1: Number, _2: Number, _5: Number}

Now, we still have a constraint in the choice point. We unwind the trail, restore the
stack, and pop the constraint.

    _1 = String
    -----------  -----------------------------
    _2 = String   constraints: []            
    -----------   curr_stack:  [(3 elems)]   
    _5 = String   trail:       [_6]           
    -----------  -----------------------------
       stack                 choice

We bind all elements successfully, so we create another constraint. and append it to
the affected variables:

    c2 = {_6: String, _1: String, _2: String, _5: String}

Now the choice stack is empty and we can finish unification. All affected variables
receive the constraints they took part in.

    _6 ~ [c1, c2]
    _1 ~ [c1, c2]
    _2 ~ [c1, c2]
    _5 ~ [c1, c2]

IMPORTANT: if there's only one constraint, then that's the final binding for the
affected variables!

The variable `_5` is the return value from this call, and is used in the subsequent
call to "+":

    (_5*, _3) -> _7,
    _5 ~ [c1, c2]

Which must be unified with another copy of "+"'s type:

    "+" : (_8, _8) -> _8,
          _8 ~ [{_8: Number}, {_8: String}]

We push their elements into the unification stack:

    _5* = _8*
    ---------
     _3 = _8*
    ---------
     _7 = _8*
    ---------
      stack

Both variables are constrained, and we can start with either of them. Let's do it
with the left side first.

    Number = _8*
    ------------  --------------------------
        _3 = _8*   constraints: [c2]
    ------------   curr_stack:  [(3 elems)]
        _7 = _8*   trail: [_6, _1, _2, _5]
    ------------  --------------------------
       stack                choice

Now, we must push another choice point to handle `_8`'s constraints:

                     -----------------------------
                      constraints: [{_8: String}]
                      curr_stack: [(3 elems)]
    Number = Number   trail: [_8]
    ---------------  -----------------------------
        _3 = Number   constraints: [c2]
    ---------------   curr_stack:  [(3 elems)]
        _7 = Number   trail: [_6, _1, _2, _5]
    ---------------  -----------------------------
         stack                 choice

All unifications succeed

                     -----------------------------
                      constraints: [{_8: String}]
                      curr_stack: [(3 elems)]
                      trail: [_8, _3, _7]
                     -----------------------------
                      constraints: [c2]
                      curr_stack:  [(3 elems)]
                      trail: [_6, _1, _2, _5]
    ---------------  -----------------------------
         stack                 choice

We then store all bindings as a new constraint:

    c3 = [{_6: Number, _1: Number, _2: Number, _5: Number, _8: Number, _3: Number, _7: Number}]

(It's clear we'll need to dedicate some thought on simplification)

Now, we unwind the top choice point and test the next constraint.

                     -----------------------------
                      constraints: []             
                      curr_stack: [(3 elems)]
    Number = String   trail: [_8]
    ---------------  -----------------------------
        _3 = String   constraints: [c2]
    ---------------   curr_stack:  [(3 elems)]
        _7 = String   trail: [_6, _1, _2, _5]
    ---------------  -----------------------------
         stack                 choice

This unification fails, and we backtrack to the next available option. Since the top
choice is already empty, we pop it and rewind to the first choice's state.

    String = _8*
    ------------  --------------------------
        _3 = _8*   constraints: []  
    ------------   curr_stack:  [(3 elems)]
        _7 = _8*   trail: [_6, _1, _2, _5]
    ------------  --------------------------
       stack                choice

IMPORTANT: we must be able to restore a variable's constraint on backtracking! One
option is simply keeping it there even if it's bound.

Now, we push another choice point to handle `_8`'s constraints:

                     -----------------------------
                      constraints: [{_8: String}]
                      curr_stack: [(3 elems)]
    String = Number   trail: [_8]
    ---------------  -----------------------------
        _3 = Number   constraints: []  
    ---------------   curr_stack:  [(3 elems)]
        _7 = Number   trail: [_6, _1, _2, _5]
    ---------------  -----------------------------
         stack                 choice

This unification fails, so we backtrack to the next constraint available:

                     -----------------------------
                      constraints: []            
                      curr_stack: [(3 elems)]
    String = String   trail: [_8]
    ---------------  -----------------------------
        _3 = String   constraints: []  
    ---------------   curr_stack:  [(3 elems)]
        _7 = String   trail: [_6, _1, _2, _5]
    ---------------  -----------------------------
         stack                 choice

Now this unification succeeds

                     -----------------------------
                      constraints: []
                      curr_stack: [(3 elems)]
                      trail: [_8, _3, _7]
                     -----------------------------
                      constraints: []
                      curr_stack:  [(3 elems)]
                      trail: [_6, _1, _2, _5]
    ---------------  -----------------------------
         stack                 choice

We then store all bindings as a new constraint:

    c4 = [{_6: String, _1: String, _2: String, _5: String, _8: String, _3: String, _7: String}]

And set it as the constraint set for these variables

    _6: [c3, c4]
    _1: [c3, c4]
    _2: [c3, c4]
    _5: [c3, c4]
    _8: [c3, c4]
    _3: [c3, c4]
    _7: [c3, c4]

Finally, since the return value from the second "+" call is the return value of the `add3` function,
we have that `_4 = _7*`, which is bound in the order `_7*: _4`. The unbound variable must inherit all
the bound variable constraints. The constraints must also be updated to contain this new binding:

    c3[_4] = c3[_7];
    c3[_7] = _4;
    c4[_4] = c4[_7];
    c4[_7] = _4;
    _4 ~ [c3, c4]

The constraints are then simplified (how?) to contain only the relevant variables for `add3`:

    add3 : (_1, _2, _3) -> _4,
           _1 ~ [
             &c1 {_1: Number, _2: Number, _3: Number, _4: Number},
             &c2 {_1: String, _2: String, _3: String, _4: String},
           ],
           _2 ~ [*c1, *c2],
           _3 ~ [*c1, *c2],
           _4 ~ [*c1, *c2]

It'd be even better if we could simplify further:

    add3: (t, t, t) -> t,
          t ~ [{t: Number}, {t: String}]

But for now, the above procedure seems to work.
