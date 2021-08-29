package util

import (
	"sync"
)

func Balance(workers int, candidates []interface{}, workerFunc func(a interface{}) interface{}) []interface{} {
	results := make([]interface{}, 0, len(candidates))

	channel := make(chan interface{}, workers)
	channelB := make(chan interface{}, len(candidates))

	var wg sync.WaitGroup

	for i := 0; i < workers; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for candidate := range channel {
				channelB <- workerFunc(candidate)
			}
		}()
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
