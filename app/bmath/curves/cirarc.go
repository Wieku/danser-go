package curves

import (
	"github.com/wieku/danser-go/app/bmath"
	"github.com/wieku/danser-go/app/bmath/math32"
	"math"
)

type CirArc struct {
	pt1, pt2, pt3                  bmath.Vector2f
	centre                         bmath.Vector2f
	startAngle, totalAngle, r, dir float32
	Unstable                       bool
}

func NewCirArc(pt1, pt2, pt3 bmath.Vector2f) *CirArc {
	arc := &CirArc{pt1: pt1, pt2: pt2, pt3: pt3}

	aSq := pt2.DstSq(pt3)
	bSq := pt1.DstSq(pt3)
	cSq := pt1.DstSq(pt2)

	if math32.Abs(aSq) < 0.001 || math32.Abs(bSq) < 0.001 || math32.Abs(cSq) < 0.001 {
		arc.Unstable = true
	}

	s := aSq * (bSq + cSq - aSq)
	t := bSq * (aSq + cSq - bSq)
	u := cSq * (aSq + bSq - cSq)

	sum := s + t + u

	if math32.Abs(sum) < 0.001 {
		arc.Unstable = true
	}

	centre := pt1.Scl(s).Add(pt2.Scl(t)).Add(pt3.Scl(u)).Scl(1 / sum)

	dA := pt1.Sub(centre)
	dC := pt3.Sub(centre)

	r := dA.Len()

	start := math32.Atan2(dA.Y, dA.X)
	end := math32.Atan2(dC.Y, dC.X)

	for end < start {
		end += 2 * math.Pi
	}

	dir := 1
	totalAngle := end - start

	aToC := pt3.Sub(pt1)
	aToC = bmath.NewVec2f(aToC.Y, -aToC.X)
	if aToC.Dot(pt2.Sub(pt1)) < 0 {
		dir = -dir
		totalAngle = 2*math.Pi - totalAngle
	}

	arc.totalAngle = totalAngle
	arc.dir = float32(dir)
	arc.startAngle = start
	arc.centre = centre
	arc.r = r

	return arc
}

func (arc *CirArc) PointAt(t float32) bmath.Vector2f {
	return bmath.NewVec2fRad(arc.startAngle+arc.dir*t*arc.totalAngle, arc.r).Add(arc.centre)
}

func (arc *CirArc) GetLength() float32 {
	return arc.r * arc.totalAngle
}

func (arc *CirArc) GetStartAngle() float32 {
	return arc.pt1.AngleRV(arc.PointAt(1.0 / arc.GetLength()))
}

func (arc *CirArc) GetEndAngle() float32 {
	return arc.pt3.AngleRV(arc.PointAt((arc.GetLength() - 1.0) / arc.GetLength()))
}
