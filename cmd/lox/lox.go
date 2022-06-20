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
	run(string(bs))
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

func run(text string) {
	s := lox.NewScanner(text)
	tokens := s.ScanTokens()
	for _, token := range tokens {
		fmt.Println(token)
	}
}
