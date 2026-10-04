package utils

import (
	"strconv"
	"strings"

	"golang.org/x/exp/constraints"
)

func Humanize[T constraints.Integer](number T) string {
	stringified := strconv.FormatInt(int64(number), 10)

	a := len(stringified) % 3
	if a == 0 {
		a = 3
	}

	var humanized strings.Builder
	humanized.WriteString(stringified[:a])

	for i := a; i < len(stringified); i += 3 {
		humanized.WriteRune(',')
		humanized.WriteString(stringified[i : i+3])
	}

	return humanized.String()
}
