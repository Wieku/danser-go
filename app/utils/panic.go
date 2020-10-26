package utils

import (
	"runtime"
	"strings"
)

func GetPanicStackTrace() []string {
	buf := make([]byte, 1024)
	for {
		n := runtime.Stack(buf, true)
		if n < len(buf) {
			buf = buf[:n]
			break
		}

		buf = make([]byte, 2*len(buf))
	}

	stack := strings.Split(string(buf), "\n")
	stack = append([]string{stack[0]}, stack[7:]...)

	return stack
}
