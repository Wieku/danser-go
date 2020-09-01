package buffer

import "github.com/go-gl/gl/v3.3-core/gl"

type DrawMode uint32

const (
	StaticDraw  = DrawMode(gl.STATIC_DRAW)
	DynamicDraw = DrawMode(gl.DYNAMIC_DRAW)
	StreamDraw  = DrawMode(gl.STREAM_DRAW)
	StaticRead  = DrawMode(gl.STATIC_READ)
	DynamicRead = DrawMode(gl.DYNAMIC_READ)
	StreamRead  = DrawMode(gl.STREAM_READ)
	StaticCopy  = DrawMode(gl.STATIC_COPY)
	DynamicCopy = DrawMode(gl.DYNAMIC_COPY)
	StreamCopy  = DrawMode(gl.STREAM_COPY)
)
