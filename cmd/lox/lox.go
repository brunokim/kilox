package main

import (
    "fmt"
    "github.com/brunokim/lox"
)

func main() {
    s := lox.NewScanner(`print "Hello, Lox!";`)
    tokens := s.ScanTokens()
    for _, token := range tokens {
        fmt.Println(token)
    }
}
