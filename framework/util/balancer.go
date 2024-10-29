package util

import (
	"github.com/wieku/danser-go/framework/goroutines"
	"sync"
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
