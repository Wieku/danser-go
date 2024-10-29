package goroutines

import (
	"errors"
	"github.com/wieku/danser-go/framework/profiler"
	"runtime"
)

// CallQueueCap is the capacity of the call queue. This means how many calls to CallNonBlock will not
// block until some call finishes.
var CallQueueCap = 100000

var (
	callQueue        chan func()
	mainLoopAdded    chan bool
	mainLoopCond     func() bool
	mainLoopFunc     func()
	mainLoopFinished chan bool
)

func init() {
	runtime.LockOSThread()
}

func checkRun() {
	if callQueue == nil {
		panic(errors.New("did not call RunMain"))
	}
}

// RunMain enables processing tasks on main thread. To use it, put your main function
// code into the run function (the argument to RunMain) and simply call RunMain from the real main function.
//
// RunMain returns when run (argument) function finishes and runCond argument for RunMainLoop returns false.
func RunMain(run func()) {
	callQueue = make(chan func(), CallQueueCap)

	mainLoopAdded = make(chan bool)
	mainLoopFinished = make(chan bool)

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
		case <-mainLoopAdded:
			goto mainLoop
		case <-done:
			return
		}
	}

mainLoop:

	for mainLoopCond() {
		profiler.Reset()

		profiler.StartGroup("goroutines.RunMain", profiler.PRoot)

		profiler.StartGroup("goroutines.RunMain", profiler.PSched)

		for sRun := len(callQueue) > 0; sRun; {
			select {
			case f := <-callQueue:
				f()
			default:
				sRun = false
			}
		}

		profiler.EndGroup()

		mainLoopFunc()

		profiler.EndGroup()
	}

	mainLoopFinished <- true
}

// RunMainLoop wires runFunc to the main thread and runs it in a loop as long as runCond returns true
//
// RunMainLoop returns when runCond returns false
func RunMainLoop(runCond func() bool, runFunc func()) {
	checkRun()

	mainLoopCond = runCond
	mainLoopFunc = runFunc
	mainLoopAdded <- true

	<-mainLoopFinished
}

// CallNonBlockMain queues function f on the main thread and returns immediately. Does not wait until f
// finishes.
func CallNonBlockMain(f func()) {
	checkRun()
	callQueue <- f
}

// CallMain queues function f on the main thread and blocks until the function f finishes.
func CallMain(f func()) {
	checkRun()

	done := make(chan uint8)

	callQueue <- func() {
		f()
		done <- 1
	}

	<-done
}
