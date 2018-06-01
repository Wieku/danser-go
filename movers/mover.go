package movers

import (
	math2 "danser/bmath"
	"danser/beatmap/objects"
	"math"
)

type Mover interface {
	Reset()
	SetObjects(end, start objects.BaseObject)
	Update(time int64/*, cursor *render.Cursor*/)
}

func getMPoint(l1, l2, p math2.Vector2d) math2.Vector2d {
	a1 := (l2.Y - l1.Y) / (l2.X - l1.X)
	b1 := l1.Y - a1 * l1.X

	if a1 == 0 {
		return math2.NewVec2d(p.X, l1.Y)
	}

	if math.IsInf(a1, 0) {
		return math2.NewVec2d(l1.X, p.Y)
	}

	a2 := -1/a1
	b2 := p.Y - a2*p.X

	x1 := (b1-b2)/(a2-a1)
	y1 := a1*x1+b1

	return math2.NewVec2d(x1, y1)
}

func sliderMult(o1, o2 objects.BaseObject) float64 {
	_, ok1 := o1.(*objects.Slider)
	_, ok2 := o2.(*objects.Slider)
	if ok1 || ok2 {
		return BEZIER_SLIDER_AGGRESSIVENESS
	}
	return 1.0
}