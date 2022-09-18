package spinners

import (
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/framework/math/math32"
	"github.com/wieku/danser-go/framework/math/vector"
)

type HeartMover struct {
	start float64
	id    int
}

func NewHeartMover() *HeartMover {
	return &HeartMover{}
}

func (c *HeartMover) Init(start, _ float64, id int) {
	c.start = start
	c.id = id
}

func (c *HeartMover) GetPositionAt(time float64) vector.Vector2f {
	spS := settings.CursorDance.Spinners[c.id%len(settings.CursorDance.Spinners)]

	rad := rpms * float32(time-c.start) * 2 * math32.Pi
	x := math32.Pow(math32.Sin(rad), 3)
	y := (13*math32.Cos(rad) - 5*math32.Cos(2*rad) - 2*math32.Cos(3*rad) - math32.Cos(4*rad)) / 16
	return vector.NewVec2f(x, y).Mult(vector.NewVec2f(float32(spS.Radius), -float32(spS.Radius))).Add(center.AddS(float32(spS.CenterOffsetX), float32(spS.CenterOffsetY)))
}
