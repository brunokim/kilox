# Challenge: implement `break` and `continue`

## Desired state

Test loop:

```
for (var i = 0; i < 10; i = i+1) {
    if (i < 4) {
        continue;
    }
    print i;
    if (i > 6) {
        break;
    }
}
```

becomes

```
// Option 1
{
    var i = 0;
    while (i < 10) {
        if (i < 4) {
            i = i + 1; // Put inc before continue
            continue;
        }
        print i;
        if (i > 6) {
            break;
        }
        i = i + 1;
    }
}
```

or 

```
// Option 2
{
    var i = 0;
    while (i < 10) {
        if (i < 4) {
            goto inc;
        }
        print i;
        if (i > 6) {
            goto break;
        }
        inc:
        i = i + 1;
    }
    break:
}
```

and should print

```
4
5
6
```

## Option #1

- The parser needs to keep track of the current for-loop's increment expression, to insert it before every `continue` statement.
  - It must also handle nested for-loops, for-while loops and while-for loops...
  - Keep a stack of current parsing loops.
- When the interpreter finds a `break`, it enters a state of `shouldBreak`.
  - when visiting a block, the block loop is finished.
  - when visiting a while-loop, the loop is broken and reset `shouldBreak=false`.
- When the interpreter finds a `continue`, it enters a state of `shouldContinue`.
  - when visiting a block, the block loop is finished.
  - when visiting a while-loop, the state is reset to `shouldContinue=false`.
- The solution should work when function calls are implemented! You can only break from within the same function.
  - The `shouldBreak` and `shouldContinue` states of the interpreter should be scoped to
    the current function, and stacked.
      - Can we reuse environments? They are stacked on blocks, not functions...

## Tests

- `while (true) break;`
- `var x = 0; while (true) { if (x < 10) { x = x + 1; continue }; break; } print x;`
- `var x = 0; while (x < 10) { if (x > 5) break; x = x + 2; } print x;`
- ```
  var x = 0;
  while (x < 10) {
      var start = x;
      while (true) {
          x = x + 1;
          print x;
          if (x - start >= 4) {
              break;
          }
      }
      if (x > 6) {
          continue;
      }
      x = x - 3;
  }
```
