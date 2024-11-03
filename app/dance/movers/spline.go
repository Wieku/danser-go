package movers

import (
	"github.com/wieku/danser-go/app/beatmap/objects"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/framework/math/curves"
	"github.com/wieku/danser-go/framework/math/math32"
	"github.com/wieku/danser-go/framework/math/mutils"
	"github.com/wieku/danser-go/framework/math/vector"
)

const (
	streamEntryMin = 25
	streamEntryMax = 4000
	streamEscape   = 8000
)

type SplineMover struct {
	*basicMover

	curve *curves.Spline
}

func NewSplineMover() MultiPointMover {
	return &SplineMover{basicMover: &basicMover{}}
}

func (mover *SplineMover) SetObjects(objs []objects.IHitObject) int {
	config := settings.CursorDance.MoverSettings.Spline[mover.id%len(settings.CursorDance.MoverSettings.Spline)]

	points := make([]vector.Vector2f, 0)
	timing := make([]float64, 0)

	var angle float32
	var stream bool

	i := 0

	for ; i < len(objs); i++ {
		o := objs[i]

		if i == 0 {
			cEnd := o.GetStackedEndPositionMod(mover.diff)
			nStart := objs[i+1].GetStackedStartPositionMod(mover.diff)

			var wPoint vector.Vector2f

			switch s := o.(type) {
			case objects.ILongObject:
				wPoint = cEnd.Add(vector.NewVec2fRad(s.GetEndAngleMod(mover.diff), cEnd.Dst(nStart)*0.7))
			default:
				wPoint = cEnd.Lerp(nStart, 0.333)
			}

			points = append(points, cEnd, wPoint)
			timing = append(timing, max(o.GetStartTime(), o.GetEndTime()))

			mover.startTime = max(o.GetStartTime(), o.GetEndTime())

			continue
		}

		if _, ok := o.(objects.ILongObject); ok || i == len(objs)-1 {
			pEnd := objs[i-1].GetStackedEndPositionMod(mover.diff)
			cStart := o.GetStackedStartPositionMod(mover.diff)

			var wPoint vector.Vector2f

			switch s := o.(type) {
			case objects.ILongObject:
				wPoint = cStart.Add(vector.NewVec2fRad(s.GetStartAngleMod(mover.diff), cStart.Dst(pEnd)*0.7))
			default:
				wPoint = cStart.Lerp(pEnd, 0.333)
			}

			points = append(points, wPoint, cStart)
			timing = append(timing, o.GetStartTime())

			mover.endTime = o.GetStartTime()

			break
		} else if i > 1 && i < len(objs)-1 {
			pos1 := objs[i-1].GetStackedStartPositionMod(mover.diff)
			pos2 := o.GetStackedStartPositionMod(mover.diff)
			pos3 := objs[i+1].GetStackedStartPositionMod(mover.diff)

			min := float32(streamEntryMin)
			max := float32(streamEntryMax)
			if stream {
				max = streamEscape
			}

			sq1 := pos1.DstSq(pos2)
			sq2 := pos2.DstSq(pos3)

			if sq1 > max && sq2 > max && config.RotationalForce {
				if stream {
					angle = 0
					stream = false
				} else {
					ang := int(math32.Abs(pos1.AngleRV(pos2) - pos1.AngleRV(pos3)))

					if ang == 0 {
						angle *= -1
					} else {
						angle = float32(ang) * 90 / 180 * math32.Pi
					}
				}
			} else if sq1 >= min && sq1 <= max && sq2 >= min && sq2 <= max && (config.StreamWobble || config.StreamHalfCircle) {
				if stream {
					angle *= -1

					if math32.Abs(angle) < 0.01 {
						pp1 := points[len(points)-1]

						shoeF := pp1.X*pos2.Y + pos2.X*pos3.Y + pos3.X*pp1.Y
						shoeS := pp1.Y*pos2.X + pos2.Y*pos3.X + pos3.Y*pp1.X

						sig := (shoeF - shoeS) > 0

						angle = math32.Pi / 2
						if sig {
							angle *= -1
						}
					}
				} else {
					stream = true
				}
			} else {
				stream = false
				angle = 0
			}

			if math32.Abs(angle) > 0.01 {
				mid := pos1.Mid(pos2)

				scale := float32(1.0)
				if stream && !config.StreamHalfCircle {
					scale = float32(config.WobbleScale)
				}

				if stream && config.StreamHalfCircle {
					sign := -1
					if angle < 0 {
						sign = 1
					}

					for t := -2; t <= 2; t++ {
						p4 := mid.Sub(pos1).Scl(scale).Rotate(angle + float32(sign*t)*math32.Pi/6).Add(mid)

						points = append(points, p4)
						timing = append(timing, (o.GetStartTime()-objs[i-1].GetStartTime())*(3+float64(t))/6+objs[i-1].GetStartTime())
					}
				} else {
					p4 := mid.Sub(pos1).Scl(scale).Rotate(angle).Add(mid)

					points = append(points, p4)
					timing = append(timing, (o.GetStartTime()-objs[i-1].GetStartTime())/2+objs[i-1].GetStartTime())
				}
			}
		}

		points = append(points, o.GetStackedEndPositionMod(mover.diff))
		timing = append(timing, o.GetStartTime())
	}

	timeDiff := make([]float32, len(timing)-1)

	for j := 0; j < len(timeDiff); j++ {
		timeDiff[j] = float32(timing[j+1] - timing[j])
	}

	beziers := curves.SolveBSpline(points)
	beziersC := make([]curves.Curve, len(beziers))

	for j, b := range beziers {
		if timeDiff[j] > 600 {
			scl := timeDiff[j] / 2

			b.Points[1] = b.Points[0].Add(b.Points[1].Sub(b.Points[0]).Nor().Scl(scl))
			b.Points[2] = b.Points[3].Add(b.Points[2].Sub(b.Points[3]).Nor().Scl(scl))
		}

		beziersC[j] = b
	}

	mover.curve = curves.NewSplineW(beziersC, timeDiff)

	return i + 1
}

func (mover *SplineMover) Update(time float64) vector.Vector2f {
	t := mutils.Clamp((time-mover.startTime)/(mover.endTime-mover.startTime), 0, 1)
	return mover.curve.PointAt(float32(t))
}
