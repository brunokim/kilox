# Thinking about type errors

Consider the program

    fun process123(x, f) {
        var a = 123;
        return f(x, a);
    }

    fun add(x, y) {
        return x + y;
    }

    print process123(20000, add); // output: 20123
    print process123("str", add); // Fail to typecheck

It should fail to type check, because

    top level:
      - process123: [_1, _2, _3]
    process123 scope:
      - x: _1
      - f: _2
      - a: number("123")
      - _2 = [_1, number, _3]
      => process123: [_1, [_1, number, _3], _3]
    top level:
      - add: [_4, _5, _6]
    add scope:
      - x: _4
      - y: _5
      - [_4, _5, _6] = [number, number, number]
      ; [_4, _5, _6] = [string, string, string]
      => add: [number, number, number]
            ; [string, string, string]
    top level:
        {process123} = [number, {add}, _7]
        - [_1, [_1, number, _3], _3] = [number, [_4, _5, _6], _7],
          _4 = number, _5 = number, _6 = number.
        => {_1: number, _3: number, _7: number}
    top level:
        {process123} = [string, {add}, _7]
        - [_1, [_1, number, _3], _3] = [string, [_4, _5, _6], _7],
          _4 = number, _5 = number, _6 = number.
        ... _1: string,
            _4: _1,
            _5: number,
            _6: _3,
            _7: _3,
            _4 = number -> _1 = number -> string != number
            _5 = number,
            _6 = number -> _3 = number
        => {_1: string, _3: number, _7: number}
           error: string != number
                    ^           ^
                    _1          _4
                    ^           ^
                 process123     +:(int, int)->int
                 (def:line 1)   (def:builtin)
                 (call:line 11) (call:line 7)
                    ^
                 process123[1]
                    ^
                   "str"
                 (def:line 11)

        ... backtrack
            _4 = string -> _1 = string
            _5 = string -> number != string
            _6 = string -> _3 = string
        => {_1: string, _3: string, _7: string}
           error: number != string 
                   ^             ^
                   _5            +:(int,int)->int
                   ^             (def:builtin)
                process123[2]    (call:line 7)
                (def:line 1)
                (call:line 11)
                   ^
                [_1, number, _3]
                     ^
                 (def:line 1)
                 (call:line 11)
                     ^
                    "a" (process123)
                     ^
                    "123" (process123)


Inference chain

