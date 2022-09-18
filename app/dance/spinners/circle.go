package spinners

import (
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/framework/math/math32"
	"github.com/wieku/danser-go/framework/math/vector"
)

type CircleMover struct {
	start float64
	id    int
}

func NewCircleMover() *CircleMover {
	return &CircleMover{}
}

func (c *CircleMover) Init(start, _ float64, id int) {
	c.start = start
	c.id = id
}

func (c *CircleMover) GetPositionAt(time float64) vector.Vector2f {
	spS := settings.CursorDance.Spinners[c.id%len(settings.CursorDance.Spinners)]
	return vector.NewVec2fRad(rpms*float32(time-c.start)*2*math32.Pi, float32(spS.Radius)).Add(center.AddS(float32(spS.CenterOffsetX), float32(spS.CenterOffsetY)))
}
