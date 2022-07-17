package lox_test

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/brunokim/lox"
)

func runLox(text string, experiments map[string]bool) (string, error) {
	s := lox.NewScanner(text)
	tokens, err := s.ScanTokens()
	if err != nil {
		return "", err
	}
	p := lox.NewParser(tokens)
	stmts, err := p.Parse()
	if err != nil {
		return "", err
	}
	i := lox.NewInterpreter()
	r := lox.NewResolver(i)
	err = r.Resolve(stmts)
	if err != nil {
		return "", err
	}
	if experiments["typing"] {
		c := lox.NewTypeChecker()
		_, err := c.Check(stmts)
		if err != nil {
			return "", err
		}
	}
	var b strings.Builder
	i.SetStdout(&b)
	err = i.Interpret(stmts)
	return b.String(), err
}

func parser(t *testing.T, text string) *lox.Parser {
	s := lox.NewScanner(text)
	tokens, err := s.ScanTokens()
	if err != nil {
		t.Fatalf("parser(%q): %v", text, err)
	}
	return lox.NewParser(tokens)
}

func parseStmts(t *testing.T, text string) []lox.Stmt {
	p := parser(t, text)
	stmts, err := p.Parse()
	if err != nil {
		t.Fatalf("parseStmts(%q): %v", text, err)
	}
	return stmts
}

func parseExpr(t *testing.T, text string) lox.Expr {
	p := parser(t, text)
	expr, err := p.ParseExpression()
	if err != nil {
		t.Fatalf("parseExpr(%q): %v", text, err)
	}
	return expr
}

// ---- scanner test

func token(tokenType lox.TokenType, text string) lox.Token {
	return literalToken(tokenType, text, nil)
}

func literalToken(tokenType lox.TokenType, text string, literal any) lox.Token {
	return lox.Token{
		TokenType: tokenType,
		Lexeme:    text,
		Literal:   literal,
	}
}

// ---- parser test

func literal(tokenType lox.TokenType, v any) *lox.LiteralExpr {
	return &lox.LiteralExpr{Token: literalToken(tokenType, "", v), Value: v}
}

func number(x float64) *lox.LiteralExpr {
	return &lox.LiteralExpr{Token: literalToken(lox.Number, "", x), Value: x}
}

func boolean(x bool) *lox.LiteralExpr {
	if x {
		return &lox.LiteralExpr{Token: token(lox.True, ""), Value: true}
	}
	return &lox.LiteralExpr{Token: token(lox.False, ""), Value: false}
}

func variableExpr(name string) *lox.VariableExpr {
	return &lox.VariableExpr{Name: token(lox.Identifier, name)}
}

// ---- interpreter test

func extractExpected(text string) (string, string) {
	wantOutput := extractComment(text, "output")
	wantError := extractComment(text, "error")
	return wantOutput, wantError
}

func extractComment(text, pattern string) string {
	commentRE := regexp.MustCompile("(?im)// " + pattern + ":(.*)$")

	var b strings.Builder
	matches := commentRE.FindAllStringSubmatch(text, -1)
	for _, match := range matches {
		b.WriteString(strings.TrimPrefix(match[1], " "))
		b.WriteRune('\n')
	}
	return b.String()
}

func extractExperiments(text string) map[string]bool {
	expStr := extractComment(text, "experiments")
	exps := make(map[string]bool)
	for _, exp := range strings.Split(expStr, ",") {
		exp = strings.TrimSpace(exp)
		if exp == "" {
			continue
		}
		// Last setting wins.
		switch exp[0] {
		case '+':
			exps[exp[1:]] = true
		case '-':
			exps[exp[1:]] = false
		default:
			exps[exp] = true
		}
	}
	return exps
}

// ---- type checker test

func types(ts ...lox.Type) []lox.Type {
	return ts
}

func ref_(value lox.Type) *lox.RefType {
	return &lox.RefType{Value: value}
}

func func_(params []lox.Type, result lox.Type) lox.FunctionType {
	return lox.FunctionType{
		Params: params,
		Return: result,
	}
}

var (
	nil_  = lox.NilType{}
	num_  = lox.NumberType{}
	bool_ = lox.BoolType{}
	str_  = lox.StringType{}
)

func walkPath(path string, obj any) (any, error) {
	steps := strings.Split(path, ".")
	if (steps[0] != "") && (steps[0] != "$") {
		return nil, fmt.Errorf("%q: at 0: invalid root object %q", path, steps[0])
	}
	value := reflect.ValueOf(obj)
	for i, step := range steps[1:] {
		fmt.Println("walk", step)
		// Dereference value until hitting a concrete type.
		for value.Kind() == reflect.Pointer || value.Kind() == reflect.Interface {
			value = value.Elem()
		}
		// If step is a valid integer, consider it an index.
		if idx, err := strconv.Atoi(step); err == nil {
			value, err = getIndex(value, idx)
			if err != nil {
				return nil, fmt.Errorf("%q at %d: %w", path, i, err)
			}
			continue
		}
		var err error
		value, err = getKey(value, step)
		if err != nil {
			return nil, fmt.Errorf("%q at %d: %w", path, i, err)
		}
	}
	return value.Interface(), nil
}

func getIndex(value reflect.Value, idx int) (reflect.Value, error) {
	switch value.Kind() {
	case reflect.Map:
		v := value.MapIndex(reflect.ValueOf(idx))
		if !v.IsValid() {
			return reflect.Value{}, fmt.Errorf("key %d not found in map", idx)
		}
		return v, nil
	case reflect.Array, reflect.Slice, reflect.String:
		if idx < 0 || idx >= value.Len() {
			return reflect.Value{}, fmt.Errorf("index %d is out of range [0,%d)", idx, value.Len())
		}
		return value.Index(idx), nil
	default:
		return reflect.Value{}, fmt.Errorf("can't access member %d of %v", idx, value.Type())
	}
}

func getKey(value reflect.Value, key string) (reflect.Value, error) {
	switch value.Kind() {
	case reflect.Map:
		v := value.MapIndex(reflect.ValueOf(key))
		if !v.IsValid() {
			return reflect.Value{}, fmt.Errorf("key %q not found in map with type %v", key, value.Type())
		}
		return v, nil
	case reflect.Struct:
		v := value.FieldByName(key)
		if !v.IsValid() {
			return reflect.Value{}, fmt.Errorf("field %q not found in struct with type %v", key, value.Type())
		}
		return v, nil
	default:
		return reflect.Value{}, fmt.Errorf("can't access member %q of %v", key, value.Type())
	}
}
