package oppai

import (
	"fmt"
)

// Warnings ...
const Warnings = true

func info(s string) {
	if Warnings {
		fmt.Println(s)
	}
}
