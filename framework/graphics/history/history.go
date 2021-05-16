package history

var history = make(map[int][]uint32)
var current = make(map[int]uint32)

func Push(binding int, new uint32) {
	history[binding] = append(history[binding], current[binding])
	current[binding] = new
}

func Pop(binding int) uint32 {
	handle := uint32(0)

	if stack, exists := history[binding]; exists && len(stack) > 0 {
		handle = stack[len(stack)-1]
		history[binding] = stack[:len(stack)-1]
	}

	current[binding] = handle

	return handle
}

func GetCurrent(binding int) uint32 {
	return current[binding]
}
