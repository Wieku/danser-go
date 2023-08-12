package spinners

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/framework/math/math32"
	"github.com/wieku/danser-go/framework/math/vector"
)

var cubeVertices = []mgl32.Vec4{
	{-1, -1, -1, 1},
	{-1, 1, -1, 1},
	{1, 1, -1, 1},
	{1, -1, -1, 1},
	{-1, -1, 1, 1},
	{-1, 1, 1, 1},
	{1, 1, 1, 1},
	{1, -1, 1, 1},
}

var cubeIndices = []int{0, 1, 2, 3, 0, 4, 5, 1, 5, 6, 2, 6, 7, 3, 7, 4}

type CubeMover struct {
	start float64
	id    int
}

func NewCubeMover() *CubeMover {
	return &CubeMover{}
}

func (c *CubeMover) Init(start, _ float64, id int) {
	c.start = start
	c.id = id
}

func (c *CubeMover) GetPositionAt(time float64) vector.Vector2f {
	spS := settings.CursorDance.Spinners[c.id%len(settings.CursorDance.Spinners)]

	radY := math32.Sin(float32(time-c.start)/9000*2*math32.Pi) * 3.0 / 18 * math32.Pi
	radX := math32.Sin(float32(time-c.start)/5000*2*math32.Pi) * 3.0 / 18 * math32.Pi

	scale := (1.0 + math32.Sin(float32(time-c.start)/4500*2*math32.Pi)*0.3) * float32(spS.Radius)

	mat := mgl32.HomogRotate3DY(radY).Mul4(mgl32.HomogRotate3DX(radX)).Mul4(mgl32.Scale3D(scale, scale, scale))

	startIndex := (int64(max(0, time-c.start)) / 4) % int64(len(cubeIndices))

	i1 := cubeIndices[startIndex]

	i2 := cubeIndices[0]
	if startIndex < int64(len(cubeIndices))-1 {
		i2 = cubeIndices[startIndex+1]
	}

	pt1 := cubeVertices[i1]
	pt2 := cubeVertices[i2]

	t := float32(int64(time-c.start)%4) / 4

	pt := mgl32.Vec4{(pt2.X()-pt1.X())*t + pt1.X(), (pt2.Y()-pt1.Y())*t + pt1.Y(), (pt2.Z()-pt1.Z())*t + pt1.Z(), 1.0}

	pt = mat.Mul4x1(pt)

	pt[0] *= 1 + pt[2]/scale/10
	pt[1] *= 1 + pt[2]/scale/10

	return vector.NewVec2f(pt.X(), pt.Y()).Add(center.AddS(float32(spS.CenterOffsetX), float32(spS.CenterOffsetY)))
}
