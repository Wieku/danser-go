package curves

import (
	math2 "danser/bmath"
	"math"
)

type CirArc struct {
	pt1, pt2, pt3                  math2.Vector2d
	centre                         math2.Vector2d
	startAngle, totalAngle, r, dir float64
	Unstable                       bool
}

func NewCirArc(pt1, pt2, pt3 math2.Vector2d) CirArc {
	arc := &CirArc{pt1: pt1, pt2: pt2, pt3: pt3}

	aSq := pt2.DstSq(pt3)
	bSq := pt1.DstSq(pt3)
	cSq := pt1.DstSq(pt2)

	if math.Abs(aSq) < 0.001 || math.Abs(bSq) < 0.001 || math.Abs(cSq) < 0.001 {
		arc.Unstable = true
	}

	s := aSq * (bSq + cSq - aSq)
	t := bSq * (aSq + cSq - bSq)
	u := cSq * (aSq + bSq - cSq)

	sum := s + t + u

	if math.Abs(sum) < 0.001 {
		arc.Unstable = true
	}

	centre := pt1.Scl(s).Add(pt2.Scl(t)).Add(pt3.Scl(u)).Scl(1 / sum)

	dA := pt1.Sub(centre)
	dC := pt3.Sub(centre)

	r := dA.Len()

	start := math.Atan2(dA.Y, dA.X)
	end := math.Atan2(dC.Y, dC.X)

	for end < start {
		end += 2 * math.Pi
	}

	dir := 1
	totalAngle := end - start

	aToC := pt3.Sub(pt1)
	aToC = math2.NewVec2d(aToC.Y, -aToC.X)
	if aToC.Dot(pt2.Sub(pt1)) < 0 {
		dir = -dir
		totalAngle = 2*math.Pi - totalAngle
	}

	arc.totalAngle = totalAngle
	arc.dir = float64(dir)
	arc.startAngle = start
	arc.centre = centre
	arc.r = r

	return *arc
}

func (ln CirArc) PointAt(t float64) math2.Vector2d {
	return math2.NewVec2dRad(ln.startAngle+ln.dir*t*ln.totalAngle, ln.r).Add(ln.centre)
}

func (ln CirArc) GetLength() float64 {
	return ln.r * ln.totalAngle
}

func (ln CirArc) GetStartAngle() float64 {
	return ln.pt1.AngleRV(ln.PointAt(1.0 / ln.GetLength()))
}

func (ln CirArc) GetEndAngle() float64 {
	return ln.pt3.AngleRV(ln.PointAt((ln.GetLength() - 1.0) / ln.GetLength()))
}

func (ln CirArc) GetPoints(num int) []math2.Vector2d {
	t0 := 1 / float64(num-1)

	points := make([]math2.Vector2d, num)
	t := 0.0
	for i := 0; i < num; i += 1 {
		points[i] = ln.PointAt(t)
		t += t0
	}

	return points
}
