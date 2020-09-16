package buffer

type MemoryChunk struct {
	Offset int
	Data   []float32
}

type StreamingBuffer interface {
	Capacity() int
	Bind()
	Unbind()
	SetData(offset int, data []float32)
	Map(size int) MemoryChunk
	Unmap(offset int, size int)
	Dispose()
}
