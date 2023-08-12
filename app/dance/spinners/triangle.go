package spinners

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/framework/math/math32"
	"github.com/wieku/danser-go/framework/math/vector"
)

var indicesTriangle = []mgl32.Vec3{{-0.86602540378, -0.5, 0}, {0.86602540378, -0.5, 0}, {0, 1, 0}}

type TriangleMover struct {
	start float64
	id    int
}

func NewTriangleMover() *TriangleMover {
	return &TriangleMover{}
}

func (c *TriangleMover) Init(start, _ float64, id int) {
	c.start = start
	c.id = id
}

func (c *TriangleMover) GetPositionAt(time float64) vector.Vector2f {
	spS := settings.CursorDance.Spinners[c.id%len(settings.CursorDance.Spinners)]

	mat := mgl32.Rotate3DZ(float32(time-c.start) / 2000 * 2 * math32.Pi).Mul3(mgl32.Scale2D(float32(spS.Radius), float32(spS.Radius)))

	startIndex := (int64(max(0, time-c.start)) / 10) % 3

	pt1 := indicesTriangle[startIndex]

	pt2 := indicesTriangle[0]
	if startIndex < 2 {
		pt2 = indicesTriangle[startIndex+1]
	}

	pt1 = mat.Mul3x1(pt1)
	pt2 = mat.Mul3x1(pt2)

	t := float32(int64(time-c.start)%10) / 10

	return vector.NewVec2f((pt2.X()-pt1.X())*t+pt1.X(), (pt2.Y()-pt1.Y())*t+pt1.Y()).Add(center.AddS(float32(spS.CenterOffsetX), float32(spS.CenterOffsetY)))
}
