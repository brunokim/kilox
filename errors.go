package lox

import (
	"strings"
)

type errors[T error] []T

func (errs errors[T]) Error() string {
	if len(errs) == 1 {
		return errs[0].Error()
	}
	msgs := make([]string, len(errs))
	for i, err := range errs {
		msgs[i] = err.Error()
	}
	return strings.Join(msgs, "\n")
}
