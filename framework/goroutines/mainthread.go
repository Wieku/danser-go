package goroutines

import (
	"errors"
	"runtime"
)

// CallQueueCap is the capacity of the call queue. This means how many calls to CallNonBlock will not
// block until some call finishes.
//
// The default value is 16 and should be good for 99% usecases.
var CallQueueCap = 16

var (
	callQueue chan func()
)

func init() {
	runtime.LockOSThread()
}

func checkRun() {
	if callQueue == nil {
		panic(errors.New("mainthread: did not call Run"))
	}
}

// RunMain enables mainthread package functionality. To use mainthread package, put your main function
// code into the run function (the argument to Run) and simply call Run from the real main function.
//
// RunMain returns when run (argument) function finishes.
func RunMain(run func()) {
	callQueue = make(chan func(), CallQueueCap)

	done := make(chan struct{})
	go func() {
		runtime.LockOSThread()
		run()
		done <- struct{}{}
	}()

	for {
		select {
		case f := <-callQueue:
			f()
		case <-done:
			return
		}
	}
}

// CallNonBlock queues function f on the main thread and returns immediately. Does not wait until f
// finishes.
func CallNonBlockMain(f func()) {
	checkRun()
	callQueue <- f
}

// Call queues function f on the main thread and blocks until the function f finishes.
func CallMain(f func()) {
	checkRun()

	done := make(chan uint8)
	//wg := &sync.WaitGroup{}
	//wg.Add(1)

	callQueue <- func() {
		f()
		done <- 1
		//wg.Done()
	}

	//wg.Wait()
	<-done
}
