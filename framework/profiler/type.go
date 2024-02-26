package profiler

type StatisticType int

const (
	VAOBinds = StatisticType(iota)
	VBOBinds
	IBOBinds
	FBOBinds
	DrawCalls
	VerticesDrawn
	VertexUpload
	SpritesDrawn
	size
)
