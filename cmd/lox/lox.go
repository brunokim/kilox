package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/brunokim/lox"
)

func main() {
	if len(os.Args) > 2 {
		fmt.Println("Usage: lox [script]")
		return
	}
	expr := lox.Binary{
		lox.Unary{
			lox.Token{lox.Minus, "-", nil, 1},
			lox.Literal{123},
		},
		lox.Token{lox.Star, "*", nil, 1},
		lox.Grouping{
			lox.Literal{45.67},
		},
	}
	fmt.Println(new(lox.ASTPrinter).Print(expr))
	if len(os.Args) == 2 {
		runFile(os.Args[1])
	} else {
		runPrompt()
	}
}

func runFile(path string) {
	bs, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}
	if !run(string(bs)) {
		os.Exit(65)
	}
}

func runPrompt() {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("> ")
	for scanner.Scan() {
		run(scanner.Text())
		fmt.Print("> ")
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

func run(text string) bool {
	s := lox.NewScanner(text)
	tokens := s.ScanTokens()
	if err := s.Err(); err != nil {
		log.Print(err)
		return false
	}
	for _, token := range tokens {
		fmt.Println(token)
	}
	return true
}
