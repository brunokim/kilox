package errlist

import (
	"strings"
)

type Of[T error] []T

func (errs Of[T]) Error() string {
	if len(errs) == 1 {
		return errs[0].Error()
	}
	msgs := make([]string, len(errs))
	for i, err := range errs {
		msgs[i] = err.Error()
	}
	return strings.Join(msgs, "\n")
}
