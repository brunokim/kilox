# Choices and options

## Generic functions

To support generic functions, the unifier must have a **backtracking**
capability, to visit different binding possibilities. For example, the
"+" function has two possible types:

    + : (Number, Number) -> Number
    + : (String, String) -> String

The "-" function can be called in its binary or unary versions, even:

    - : (Number, Number) -> Number
    - : (Number) -> Number

We'd also want for user defined functions to be generic:

    // id : t -> t
    fun id(x) { return x; }

    // add : (t, t) -> t {t = Number ; t = String}
    fun add(x, y) {
        return x + y;
    }

## Constraints

One way to solve this would be using **constraints**, which are a tool
to augment the unification capability.

In a logic language they can be implemented with attributed variables:
variables that have associated data and code, which is invoked whenever
the variable is unified.

Attributed variables are powerful, but perhaps too powerful for Lox's
type checker, since it would require expressing and executing arbitrary
code. Instead, variables (`Ref`s) and composite types (currently only
`Function`s) will store a set of alternative bindings, called `options`.

## Options

For example, "+"'s type can be expressed as

    + : (t1, t2) -> t3 [
        {t1 = Number, t2 = Number, t3 = Number},
        {t1 = String, t2 = String, t3 = String},
    ]

or, more succintly:

    + : (t, t) -> t [{t = Number}, {t = String}]

The `[]` list contains the alternatives to consider for this type.
Each alternative is a set of bindings, referring to the refs present in
the params and return.

In Go, this translates to
   
    type Bindings map[*RefType]Type

    t := &RefType{}
    typePlus := FunctionType{
        Params: []Type{ t, t },
        Return: t,
        options: []Bindings{
            Bindings{t: NumberType{}},
            Bindings{t: StringType{}}
        },
    }

## Choices

When presented with a type containings options, the unifier must
store the alternatives and restore to them in case of a unification
failure. This is done using **choices**, or choice points: data structures
that store the unifier state at the point where an option is taken,
and the alternatives that weren't visited yet.

    type Choice struct {
        // State at the moment it was chosen.
        topRefID int
        stack []typePair

        // Alternatives not yet visited.
        options []Bindings

        // Variables bound since this choice was made.
        trail []*RefType
    }

Choices are stacked as alternatives are explored. This is what allows
a combinatorial exploration of the solution space. For example, say
that we have a generic function that returns any one of its 3 parameters:

    // f : (t1, t2, t3) -> t [{t: t1}, {t: t2}, {t: t3}]
    fun f(a, b, c) {
        var r = random();
        if (r < 1/3) {
            return a;
        }
        if (r < 2/3) {
            return b;
        }
        return c;
    }

In the following program, the checker will explore all 3 * 3 * 3 = 27 possibilities
for f's type in the 4th line before settling that one of the parameters must be
a number for the program to typecheck, for a total of 19 possibilities.

    var a = f(true, "str", 100);  // a : t1 [{t1: Bool}, {t1: String}, {t1: Number}]
    var b = f(true, "str", 100);  // b : t2 [{t2: Bool}, {t2: String}, {t2: Number}]
    var c = f(true, "str", 100);  // c : t3 [{t3: Bool}, {t3: String}, {t3: Number}]
    var x = f(a, b, c) + 1;
    // f(a, b, c): (t1, t2) -> t3 [
    //   {t1: Number}, {t2: Number}, {t3: Number},
    //   {t1: Number}, {t2: Number}, {t3: String},
    //   {t1: Number}, {t2: Number}, {t3: Bool},
    //   {t1: Number}, {t2: String}, {t3: Number},
    //   {t1: Number}, {t2: String}, {t3: String},
    //   {t1: Number}, {t2: String}, {t3: Bool},
    //   {t1: Number}, {t2: Bool}, {t3: Number},
    //   {t1: Number}, {t2: Bool}, {t3: String},
    //   {t1: Number}, {t2: Bool}, {t3: Bool},
    //   {t1: String}, {t2: Number}, {t3: Number},
    //   {t1: String}, {t2: Number}, {t3: String},
    //   {t1: String}, {t2: Number}, {t3: Bool},
    //   {t1: String}, {t2: String}, {t3: Number},
    //   {t1: String}, {t2: Bool}, {t3: Number},
    //   {t1: Bool}, {t2: Number}, {t3: Number},
    //   {t1: Bool}, {t2: Number}, {t3: String},
    //   {t1: Bool}, {t2: Number}, {t3: Bool},
    //   {t1: Bool}, {t2: String}, {t3: Number},
    //   {t1: Bool}, {t2: Bool}, {t3: Number},
    // ]

(What will it do with that? I have no idea.)

## Procedure

Once a choice is stacked, all the bindings presented in the first option are executed,
and the remaining are stored in the choice. This must be done before dereferencing
those vars, because it may (or better, _should_) change depending on the option.
Then, the type may be unified as usual.

Once a type error is found, it must be _stored_ in the latest choice object and a new
option is chosen. The state is rewinded, including the variables that were bound since
the choice was made (the _trail_).

If all options are exhausted, then the errors are copied to the preceding choice and the
next one of its options is visited. If all choices are exhausted, then we must report
the errors that were found, or at least the ones that are more relevant (for example, that
got farther down the tree).

If a unification **succeeds**, it's still not time to stop. The successful bindings are
stored, the state is rewound and the next option is visited. At least in this case, it's
not necessary to keep errors for reporting. Once all successful branches are visited, then
we may write the type for the target function as the union of all its bindings.

## Example

Consider the function `add3`:

    fun add3(x, y, z) {
        return x + y + z;
    }

This becomes the following object:

    FunctionStmt{
        Params: []Token{{Lexeme: "x"}, {Lexeme: "y"}, {Lexeme: "z"}},
        Body: ReturnStmt{
            BinaryExpr{                             // #1
                Left: BinaryExpr{                   // #2
                    Left: VariableExpr{Name: "x"},  // #3
                    Operator: Token{Lexeme: "+"},
                    Right: VariableExpr{Name: "y"}, // #4
                },
                Operator: Token{Lexeme: "+"},
                Right: VariableExpr{Name: "z"},     // #5
            },
        },
    }

There are 5 expression to be typed: the 3 params, and the 2 calls to "+".
The params start with fresh types, as well as the return value:

    x  : _1
    y  : _2
    z  : _3
    ret: _4

We then retrieve and copy the type of `BinaryExpr` #1:

     + : (_5, _5) -> _5 [{_5: Number}, {_5: String}]

Then on to the `BinaryExpr` #2:

     + : (_6, _6) -> _6 [{_6: Number}, {_6: String}]

### First '+' call

We unify the type of #2 with the func made with the type of the args:

    (_6, _6) -> _6 [...] = (_1, _2) -> _7

`_7` is the return type of the call, still undefined. We proceed to start
unification, that at first already has a function with options. We bind the
first option and store the remaining:

                                                      choices               
                     stack                 .---------------------------.   
    .------------------------------------. | stack:   [ (*r1, *r2) ]   |   
    | t1: &r1 (Number, Number) -> Number | | options: [ {_6: String} ] |   
    | t2: &r2 (_1, _2) -> _7             | | trail:   [ ]              |   
    '------------------------------------' '---------------------------'   

The unification proceeds by stacking each function's element.

                    stack
    .------------------------------------.
    | t1: Number                         |
    | t2: _1                             |
    |------------------------------------|
    | t1: Number                         |            choices
    | t2: _2                             | .---------------------------.
    |------------------------------------| | stack:   [ (*r1, *r2) ]   |
    | t1: Number                         | | options: [ {_6: String} ] |
    | t2: _7                             | | trail:   [ ]              |
    '------------------------------------' '---------------------------'

Once all elements are unified, the stack is empty and the bound refs are
stored in the trail

                                                      choices
                                           .---------------------------.
                                           | stack:   [ (*r1, *r2) ]   |
                                           | options: [ {_6: String} ] |
                    stack                  | trail:   [ _1, _2, _7 ]   |
     ------------------------------------  '---------------------------'

Since there is still one choice, the bindings at the trail are saved as options:

    resultOptions = [{_1: Number, _2: Number, _7: Number}]

The stack is restored, all bindings are undone and the last option is popped.

                                                      choices               
                     stack                 .---------------------------.   
    .------------------------------------. | stack:   [ (*r1, *r2) ]   |   
    | t1: &r1 (String, String) -> String | | options: [ ]              |   
    | t2: &r2 (_1, _2) -> _7             | | trail:   [ ]              |   
    '------------------------------------' '---------------------------'   

The unification proceeds as usual

                    stack
    .------------------------------------.
    | t1: String                         |
    | t2: _1                             |
    |------------------------------------|
    | t1: String                         |            choices
    | t2: _2                             | .---------------------------.
    |------------------------------------| | stack:   [ (*r1, *r2) ]   |
    | t1: String                         | | options: [ ]              |
    | t2: _7                             | | trail:   [ ]              |
    '------------------------------------' '---------------------------'


All elements are unified, the stack is empty and the bound refs are
stored in the trail

                                                      choices
                                           .---------------------------.
                                           | stack:   [ (*r1, *r2) ]   |
                                           | options: [ ]              |
                    stack                  | trail:   [ _1, _2, _7 ]   |
     ------------------------------------  '---------------------------'

These bindings are stored in the resulting options:

    resultOptions = [
        {_1: Number, _2: Number, _7: Number},
        {_1: String, _2: String, _7: String},
    ]

There's nothing to backtrack to, and the type of the `BinaryExpr` call is ready.

    x + y : (_1, _2) -> _7 [
        {_1: Number, _2: Number, _7: Number},
        {_1: String, _2: String, _7: String},
    ]

### Second '+' call

PROBLEM! We just need the value of \_7 right now! How do I retrieve it, and
keep the existing options?

I don't know yet, but let's say that we keep all options unchanged, but now
associated to the unbound ref.

    _7 : unbound [
        {_1: Number, _2: Number, _7: Number},
        {_1: String, _2: String, _7: String},
    ]

(This is starting to look a lot like attributed variables...)

Now!

Breath in, breath out.

Let's keep going to the next expression.

We unify the type of #1 with the func made with the type of the args:

    (_5, _5) -> _5 [...] = (_7 [...], _3) -> _8

At first we just "see" the function with options, so we put this in the stack.

                                                      choices               
                     stack                 .---------------------------.   
    .------------------------------------. | stack:   [ (*r1, *r2) ]   |   
    | t1: &r1 (Number, Number) -> Number | | options: [ {_5: String} ] |   
    | t2: &r2 (_7 [...], _3) -> _8       | | trail:   [ ]              |   
    '------------------------------------' '---------------------------'   

The unification proceeds by stacking each function's element.

                    stack
    .------------------------------------.
    | t1: Number                         |
    | t2: _7 [...]                       |
    |------------------------------------|
    | t1: Number                         |            choices
    | t2: _3                             | .---------------------------.
    |------------------------------------| | stack:   [ (*r1, *r2) ]   |
    | t1: Number                         | | options: [ {_5: String} ] |
    | t2: _8                             | | trail:   [ ]              |
    '------------------------------------' '---------------------------'

Now, we're unifying an atom with a ref with options! We must stack a new choice
and bind the first choice of \_7.

                    stack
    .------------------------------------.            choices
    | t1: Number                         | .---------------------------.
    | t2: Number                         | | stack:   [ 3 elems... ]   |
    |------------------------------------| | options: [ {_7: String} ] |
    | t1: Number                         | | trail:   [ _1, _2, _7 ]   |
    | t2: _3                             | |---------------------------|
    |------------------------------------| | stack:   [ (*r1, *r2) ]   |
    | t1: Number                         | | options: [ {_5: String} ] |
    | t2: _8                             | | trail:   [ ]              |
    '------------------------------------' '---------------------------'

We unify all elements, and the trail is stored in the latest choice

                                                      choices
                                           .---------------------------------.
                                           | stack:   [ 3 elems... ]         |
                                           | options: [ {_7: String} ]       |
                                           | trail:   [ _1, _2, _7, _3, _8 ] |
                                           |---------------------------------|
                                           | stack:   [ (*r1, *r2) ]         |
                                           | options: [ {_5: String} ]       |
                    stack                  | trail:   [ ]                    |
     ------------------------------------  '---------------------------------'

We save the successful bindings:

    resultOptions = [{_1: Number, _2: Number, _7: Number, _3: Number, _8: Number}]

Now on backtrack, we use the other option available at the top element of the
choice stack:

                    stack
    .------------------------------------.            choices
    | t1: Number                         | .---------------------------.
    | t2: String                         | | stack:   [ 3 elems... ]   |
    |------------------------------------| | options: [ ]              |
    | t1: Number                         | | trail:   [ _1, _2,_7 ]    |
    | t2: _3                             | |---------------------------|
    |------------------------------------| | stack:   [ (*r1, *r2) ]   |
    | t1: Number                         | | options: [ {_5: String} ] |
    | t2: _8                             | | trail:   [ ]              |
    '------------------------------------' '---------------------------'

This unification fails immediately with `Number != String`!
We pop the last element of the choice stack and use the last option
in the remaining choice:

                                                      choices               
                     stack                 .---------------------------.   
    .------------------------------------. | stack:   [ (*r1, *r2) ]   |   
    | t1: &r1 (String, String) -> String | | options: [ ]              |   
    | t2: &r2 (_7 [...], _3) -> _8       | | trail:   [ ]              |   
    '------------------------------------' '---------------------------'   

We expand the functions elementwise...

                    stack
    .------------------------------------.
    | t1: String                         |
    | t2: _7 [...]                       |
    |------------------------------------|
    | t1: String                         |            choices
    | t2: _3                             | .---------------------------.
    |------------------------------------| | stack:   [ (*r1, *r2) ]   |
    | t1: String                         | | options: [ ]              |
    | t2: _8                             | | trail:   [ ]              |
    '------------------------------------' '---------------------------'

And put a new choice for the options in `_7`.

                    stack
    .------------------------------------.            choices
    | t1: String                         | .---------------------------.
    | t2: Number                         | | stack:   [ 3 elems... ]   |
    |------------------------------------| | options: [ {_7: String} ] |
    | t1: String                         | | trail:   [ _1, _2, _7 ]   |
    | t2: _3                             | |---------------------------|
    |------------------------------------| | stack:   [ (*r1, *r2) ]   |
    | t1: String                         | | options: [ ]              |
    | t2: _8                             | | trail:   [ ]              |
    '------------------------------------' '---------------------------'

The first unification fails with `String != Number`, so we pop the last
option from the latest choice and rebind `_7` (and `_1` and `_2`, even
if they're not referenced here).

                    stack
    .------------------------------------.            choices
    | t1: String                         | .---------------------------.
    | t2: String                         | | stack:   [ 3 elems... ]   |
    |------------------------------------| | options: [ ]              |
    | t1: String                         | | trail:   [ _1, _2, _7 ]   |
    | t2: _3                             | |---------------------------|
    |------------------------------------| | stack:   [ (*r1, *r2) ]   |
    | t1: String                         | | options: [ ]              |
    | t2: _8                             | | trail:   [ ]              |
    '------------------------------------' '---------------------------'

All element unifications succeed, leaving us with the following bindings:

                                                      choices
                                           .---------------------------------.
                                           | stack:   [ 3 elems... ]         |
                                           | options: [ ]                    |
                                           | trail:   [ _1, _2, _7, _3, _8 ] |
                                           |---------------------------------|
                                           | stack:   [ (*r1, *r2) ]         |
                                           | options: [ ]                    |
                    stack                  | trail:   [ ]                    |
     ------------------------------------  '---------------------------------'


These bindings are stored as successful:

    resultOptions = [
        {_1: Number, _2: Number, _7: Number, _3: Number, _8: Number},
        {_1: String, _2: String, _7: String, _3: String, _8: String},
    ]

Since there's nothing to backtrack to, the type of the second `BinaryExpr` call is

    (x + y) + z : (_7, _3) -> _8 [
        {_1: Number, _2: Number, _7: Number, _3: Number, _8: Number},
        {_1: String, _2: String, _7: String, _3: String, _8: String},
    ]

### Function statement

The return of the second call is the return of the function, so `_4 = _8`.
This makes `_4` an attributed variable too, with the same constraints.

PROBLEM: how should we include `_4` in the above constraints?

PROBLEM: is `_8` considered a bound variable? If so, `_4` would point to `_8`;
otherwise, `_8` points to `_4`.

PROBLEM: if `_8` points to `_4`, and `_8` is the reference in the constraints,
then we should _unify_ variables to their values when popping an option, so
that `_4` can receive it, instead of simply binding them.

Now, the function type could be returned with the constraints attached to `_4`:

    add3 : (_1, _2, _3) -> (_4 [...])

But ideally the constraints should reside in the function type, and be
filtered down only to the referenced variables there.

    add3: (_1, _2_, _3) -> _4 [
        {_1: Number, _2: Number, _3: Number, _4: Number},
        {_1: String, _2: String, _3: String, _4: String},
    ]

PROBLEM: what if different variables in the function type point to the same
constraints?

PROBLEM: what if different variables in the function type point to _different_
constraints?

This makes me want to revisit everything by using constraints only in variables...
