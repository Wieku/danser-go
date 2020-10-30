package spinners

import (
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/framework/math/math32"
	"github.com/wieku/danser-go/framework/math/vector"
)

type HeartMover struct {
	start int64
}

func NewHeartMover() *HeartMover {
	return &HeartMover{}
}

func (c *HeartMover) Init(start, end int64) {
	c.start = start
}

func (c *HeartMover) GetPositionAt(time int64) vector.Vector2f {
	rad := rpms * float32(time-c.start) * 2 * math32.Pi
	x := math32.Pow(math32.Sin(rad), 3)
	y := (13*math32.Cos(rad) - 5*math32.Cos(2*rad) - 2*math32.Cos(3*rad) - math32.Cos(4*rad)) / 16
	return vector.NewVec2f(x, y).Mult(vector.NewVec2f(float32(settings.Dance.SpinnerRadius), -float32(settings.Dance.SpinnerRadius))).Add(center)
}
