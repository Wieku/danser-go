package goroutines

import (
	"runtime"
	"strings"
)

func GetStackTrace(cutLines int) []string {
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
	stack = append([]string{stack[0]}, stack[3+cutLines:]...)

	return stack
}
