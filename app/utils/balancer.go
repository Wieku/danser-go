package utils

import (
	"sync"
)

func Balance(threads int, candidates []interface{}, work func(a interface{}) interface{}) []interface{} {

	var result []interface{}

	channel := make(chan interface{}, threads)
	channelB := make(chan interface{}, len(candidates))
	var wg sync.WaitGroup

	for i := 0; i < threads; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				candidate, ok := <-channel
				if !ok {
					break
				}

				channelB <- work(candidate)
			}
		}()
	}

	for _, candidate := range candidates {
		channel <- candidate
	}

	close(channel)
	wg.Wait()

	for len(channelB) > 0 {
		beatmap := <-channelB
		result = append(result, beatmap)
	}

	return result
}
