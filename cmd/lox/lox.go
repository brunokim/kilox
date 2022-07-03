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
	r := newRunner()
	if len(os.Args) == 2 {
		r.runFile(os.Args[1])
	} else {
		r.runPrompt()
	}
}

type runner struct {
	i *lox.Interpreter
}

func newRunner() *runner {
	return &runner{
		i: lox.NewInterpreter(),
	}
}

func (r *runner) runFile(path string) {
	bs, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}
	if !r.run(string(bs)) {
		os.Exit(65)
	}
}

func (r *runner) runPrompt() {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("> ")
	for scanner.Scan() {
		r.run(scanner.Text())
		fmt.Print("> ")
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

func (r *runner) run(text string) bool {
	s := lox.NewScanner(text)
	tokens, err := s.ScanTokens()
	if err != nil {
		fmt.Println(err)
		return false
	}
	stmts, err1 := lox.NewParser(tokens).Parse()
	if err1 != nil {
		expr, err2 := lox.NewParser(tokens).ParseExpression()
		if err2 != nil {
			fmt.Println(err1)
			return false
		}
		stmts = []lox.Stmt{lox.PrintStmt{expr}}
	}
	resolver := lox.NewResolver(r.i)
	err = resolver.Resolve(stmts)
	if err != nil {
		fmt.Println(err)
		return false
	}
	err = r.i.Interpret(stmts)
	if err != nil {
		fmt.Println(err)
		return false
	}
	return true
}
