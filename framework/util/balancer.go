package util

import (
	"github.com/wieku/danser-go/framework/goroutines"
	"sync"
	"time"
)

func Balance[T, B any](workers int, candidates []T, workerFunc func(a T) (B, bool)) []B {
	receive := make(chan B, workers)

	goroutines.Run(func() {
		BalanceChan(workers, candidates, receive, workerFunc)
		close(receive)
	})

	results := make([]B, 0, len(candidates))

	for ret := range receive {
		results = append(results, ret)
	}

	return results
}

func BalanceChan[T, B any](workers int, candidates []T, receive chan<- B, workerFunc func(a T) (B, bool)) {
	queue := make(chan T, workers)

	wg := &sync.WaitGroup{}

	for i := 0; i < workers; i++ {
		wg.Add(1)

		goroutines.Run(func() {
			defer wg.Done()

			for candidate := range queue {
				result, ok := workerFunc(candidate)

				if ok {
					receive <- result
				}
			}
		})
	}

	for _, candidate := range candidates {
		queue <- candidate
	}

	close(queue)
	wg.Wait()
}

func BalanceChanWatchdog[T, B any](workers int, candidates []T, receive chan<- B, timeout time.Duration, alertFunc func(workerId int, a T), workerFunc func(a T) (B, bool)) {
	queue := make(chan T, workers)

	wg := &sync.WaitGroup{}

	for i := 0; i < workers; i++ {
		wg.Add(1)

		processed := make(chan T)

		goroutines.Run(func() {
			defer wg.Done()

			for candidate := range queue {
				processed <- candidate

				result, ok := workerFunc(candidate)

				if ok {
					receive <- result
				}
			}

			close(processed)
		})

		goroutines.Run(func() {
			watchdog := time.NewTimer(timeout)

			runW := true
			var current T

			for runW {
				select {
				case <-watchdog.C:
					alertFunc(i, current)
				case current, runW = <-processed:
					if !runW {
						watchdog.Stop()
						break
					}

					watchdog.Reset(timeout)
				}
			}

		})
	}

	for _, candidate := range candidates {
		queue <- candidate
	}

	close(queue)
	wg.Wait()
}
