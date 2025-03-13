package curves

import (
	. "github.com/wieku/danser-go/framework/math/math87"
	"github.com/wieku/danser-go/framework/math/vector"
	"math"
)

const osuPi float32 = 3.14159274

type CirArc struct {
	pt1, pt2, pt3 vector.Vector2f
	centre        vector.Vector2f //nolint:misspell
	startAngle    float64
	totalAngle    float64
	dir           float64
	r             float32

	centreS   vector.Vector2f //nolint:misspell
	tInitialS float64
	tFinalS   float64
	rS        float32

	Unstable bool
}

func NewCirArc(a, b, c vector.Vector2f) *CirArc {
	arc := &CirArc{pt1: a, pt2: b, pt3: c, dir: 1}

	if vector.IsStraightLine32(a, b, c) {
		arc.Unstable = true
	}

	d := 2 * (a.X*(b.Y-c.Y) + b.X*(c.Y-a.Y) + c.X*(a.Y-b.Y))
	aSq := a.LenSq()
	bSq := b.LenSq()
	cSq := c.LenSq()

	arc.centre = vector.NewVec2f(
		(aSq*(b.Y-c.Y)+bSq*(c.Y-a.Y)+cSq*(a.Y-b.Y))/d,
		(aSq*(c.X-b.X)+bSq*(a.X-c.X)+cSq*(b.X-a.X))/d) //nolint:misspell

	arc.r = a.Dst(arc.centre)
	arc.startAngle = a.Copy64().AngleRV(arc.centre.Copy64())

	endAngle := c.Copy64().AngleRV(arc.centre.Copy64())

	for endAngle < arc.startAngle {
		endAngle += 2 * math.Pi
	}

	arc.totalAngle = endAngle - arc.startAngle

	aToC := c.Sub(a)
	aToC = vector.NewVec2f(aToC.Y, -aToC.X)

	if aToC.Dot(b.Sub(a)) < 0 {
		arc.dir = -arc.dir
		arc.totalAngle = 2*math.Pi - arc.totalAngle
	}

	arc.arcStable(a, b, c)

	return arc
}

func (arc *CirArc) arcStable(a, b, c vector.Vector2f) {
	a1 := a.Copy64()
	b1 := b.Copy64()
	c1 := c.Copy64()

	d := float32(2 * (a1.X*(b1.Y-c1.Y) + b1.X*(c1.Y-a1.Y) + c1.X*(a1.Y-b1.Y)))

	aSq := float64(a.LenSq87())
	bSq := float64(b.LenSq87())
	cSq := float64(c.LenSq87())

	arc.centreS = vector.NewVec2f(
		float32((aSq*(b1.Y-c1.Y)+bSq*(c1.Y-a1.Y)+cSq*(a1.Y-b1.Y))/float64(d)),
		float32((aSq*(c1.X-b1.X)+bSq*(a1.X-c1.X)+cSq*(b1.X-a1.X))/float64(d))) //nolint:misspell

	arc.rS = a.Dst87(arc.centreS)

	arc.tInitialS = ctAt(a, arc.centreS)
	tMid := ctAt(b, arc.centreS)
	arc.tFinalS = ctAt(c, arc.centreS)

	for tMid < arc.tInitialS {
		tMid += 2 * float64(osuPi)
	}

	for arc.tFinalS < arc.tInitialS {
		arc.tFinalS += 2 * float64(osuPi)
	}

	if tMid > arc.tFinalS {
		arc.tFinalS -= 2 * float64(osuPi)
	}
}

func ctAt(pt, centre vector.Vector2f) float64 {
	return math.Atan2(float64(pt.Y)-float64(centre.Y), float64(pt.X)-float64(centre.X))
}

func (arc *CirArc) PointAt(t float32) vector.Vector2f {
	return vector.NewVec2dRad(arc.startAngle+arc.dir*float64(t)*arc.totalAngle, float64(arc.r)).Copy32().Add(arc.centre)
}

func (arc *CirArc) PointAtL(t float64) vector.Vector2f {
	theta := arc.startAngle + arc.dir*t*arc.totalAngle
	return vector.NewVec2f(float32(math.Cos(theta))*arc.r+arc.centre.X, float32(math.Sin(theta))*arc.r+arc.centre.Y)
}

func (arc *CirArc) PointAtS(t float64) vector.Vector2f {
	theta := arc.tFinalS*t + arc.tInitialS*(1-t)
	return vector.NewVec2f(Add87(float32(math.Cos(theta)*float64(arc.rS)), arc.centreS.X), Add87(float32(math.Sin(theta)*float64(arc.rS)), arc.centreS.Y))
}

func (arc *CirArc) GetLength() float32 {
	return float32(float64(arc.r) * arc.totalAngle)
}

func (arc *CirArc) GetStartAngle() float32 {
	return arc.pt1.AngleRV(arc.PointAt(1.0 / arc.GetLength()))
}

func (arc *CirArc) GetEndAngle() float32 {
	return arc.pt3.AngleRV(arc.PointAt((arc.GetLength() - 1.0) / arc.GetLength()))
}
