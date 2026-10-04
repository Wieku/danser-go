package spinners

import (
	"strings"

	"github.com/wieku/danser-go/framework/math/vector"
)

const rpms = 0.00795

var center = vector.NewVec2f(256, 192)

type SpinnerMover interface {
	Init(start, end float64, id int, speed float64)
	GetPositionAt(time float64) vector.Vector2f
	GetSDelta(time float64) float32
}

type BaseMover struct {
	start, end float64
	speed      float64
	id         int
}

func (mover *BaseMover) Init(start, end float64, id int, speed float64) {
	mover.start = start
	mover.end = end
	mover.id = id
	mover.speed = speed
}

func (mover *BaseMover) GetSDelta(time float64) float32 {
	return float32((time - mover.start) / min(1, mover.speed))
}

func GetMoverByName(name string) SpinnerMover {
	switch strings.ToLower(name) {
	case "heart":
		return NewHeartMover()
	case "triangle":
		return NewTriangleMover()
	case "square":
		return NewSquareMover()
	case "cube":
		return NewCubeMover()
	default:
		return NewCircleMover()
	}
}

func GetMoverCtorByName(name string) func() SpinnerMover {
	return func() SpinnerMover {
		return GetMoverByName(name)
	}
}
