package util

import (
	"github.com/wieku/danser-go/framework/goroutines"
	"sync"
)

func Balance[T, B any](workers int, candidates []T, workerFunc func(a T) *B) []*B {
	results := make([]*B, 0, len(candidates))

	channel := make(chan T, workers)
	channelB := make(chan *B, len(candidates))

	var wg sync.WaitGroup

	for i := 0; i < workers; i++ {
		wg.Add(1)

		goroutines.Run(func() {
			defer wg.Done()

			for candidate := range channel {
				channelB <- workerFunc(candidate)
			}
		})
	}

	for _, candidate := range candidates {
		channel <- candidate
	}

	close(channel)
	wg.Wait()

	for len(channelB) > 0 {
		result := <-channelB

		if result != nil {
			results = append(results, result)
		}
	}

	close(channelB)

	return results
}
