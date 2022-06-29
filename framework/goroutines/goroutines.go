package goroutines

import (
	"runtime"
)

type CrashHandler func(err any, stackTrace []string)

//in goroutines package
var crashHandler CrashHandler

func SetCrashHandler(fn CrashHandler) {
	crashHandler = fn
}

// Run runs the goroutine using go's scheduler. Not safe for thread-related operations
func Run(fn func()) {
	runInternal(false, fn)
}

// RunOS runs the goroutine as a system thread
func RunOS(fn func()) {
	runInternal(true, fn)
}

func runInternal(osThread bool, fn func()) {
	go func() {
		defer func() {
			if crashHandler != nil {
				if err := recover(); err != nil {
					stackTrace := GetStackTrace(4)

					crashHandler(err, stackTrace)

					panic(err) //panic again if crashHandler didn't exit the program
				}
			}
		}()

		if osThread {
			runtime.LockOSThread()
		}

		fn()
	}()
}
