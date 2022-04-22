package utils

import (
	"golang.org/x/exp/constraints"
	"strconv"
)

func Humanize[T constraints.Integer](number T) string {
	stringified := strconv.FormatInt(int64(number), 10)

	a := len(stringified) % 3
	if a == 0 {
		a = 3
	}

	humanized := stringified[0:a]

	for i := a; i < len(stringified); i += 3 {
		humanized += "," + stringified[i:i+3]
	}

	return humanized
}
