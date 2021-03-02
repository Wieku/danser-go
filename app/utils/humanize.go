package utils

import "strconv"

func Humanize(number int64) string {
	stringified := strconv.FormatInt(number, 10)

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
