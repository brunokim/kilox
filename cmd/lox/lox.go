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
	p := lox.NewParser(tokens)
	stmts, err := p.Parse()
	if err != nil {
		fmt.Println(err)
		return false
	}
	r.i.Interpret(stmts)
	return true
}
