package history

import "github.com/go-gl/gl/v3.3-core/gl"

var history = make(map[int][]uint32)

func Push(binding int) {
	var handle int32
	gl.GetIntegerv(uint32(binding), &handle)

	history[binding] = append(history[binding], uint32(handle))
}

func Pop(binding int) uint32 {
	handle := uint32(0)

	if stack, exists := history[binding]; exists && len(stack) > 0 {
		handle = stack[len(stack)-1]
		history[binding] = stack[:len(stack)-1]
	}

	return handle
}

func GetCurrent(binding int) uint32 {
	var handle int32
	gl.GetIntegerv(uint32(binding), &handle)
	return uint32(handle)
}
