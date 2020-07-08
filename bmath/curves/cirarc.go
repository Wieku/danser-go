package curves

import (
	math2 "github.com/wieku/danser-go/bmath"
	"math"
	"sort"
)

type CirArc struct {
	pt1, pt2, pt3                  math2.Vector2d
	centre                         math2.Vector2d
	startAngle, totalAngle, r, dir float64
	Unstable                       bool
	lines []Linear
	sections []float64
	ApproxLength     float64
}

func NewCirArc(pt1, pt2, pt3 math2.Vector2d) *CirArc {
	arc := &CirArc{pt1: pt1, pt2: pt2, pt3: pt3, lines: make([]Linear, 0), sections: make([]float64, 1)}

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


	segments := int(arc.r*arc.totalAngle*0.125)

	//ang := 2*math.Asin(BEZIER_QUANTIZATION/(2*arc.r))

	previous := pt1

	for p := 1; p < segments; p++  {
		currentPoint := math2.NewVec2dRad(arc.startAngle+arc.dir*(float64(p)/float64(segments))*arc.totalAngle, arc.r).Add(arc.centre)

		arc.lines = append(arc.lines, NewLinear(previous, currentPoint))
		arc.sections = append(arc.sections, previous.Dst(currentPoint))
		if len(arc.sections) > 1 {
			arc.sections[len(arc.sections)-1] += arc.sections[len(arc.sections)-2]
		}
		previous = currentPoint

	}

	arc.ApproxLength = 0.0

	for _, l := range arc.lines  {
		arc.ApproxLength += l.GetLength()
	}

	return arc
}

func (ln CirArc) NPointAt(t float64) math2.Vector2d {
	return math2.NewVec2dRad(ln.startAngle+ln.dir*t*ln.totalAngle, ln.r).Add(ln.centre)
}

func (ln *CirArc) PointAt(t float64) math2.Vector2d {
	//return math2.NewVec2dRad(ln.startAngle+ln.dir*t*ln.totalAngle, ln.r).Add(ln.centre)
	desiredWidth := ln.ApproxLength * t

	lineI := sort.SearchFloat64s(ln.sections[:len(ln.sections)-2], desiredWidth)

	//println(lineI, len(ln.sections), len(ln.lines))
	line := ln.lines[lineI]

	point := line.PointAt((desiredWidth-ln.sections[lineI])/(ln.sections[lineI+1]-ln.sections[lineI]))
	return point
}

func (ln *CirArc) GetLength() float64 {
	return ln.ApproxLength/*ln.r*ln.totalAngle*/
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

func (ln *CirArc) GetLines() []Linear {
	return ln.lines
}
