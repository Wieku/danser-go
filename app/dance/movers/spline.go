package movers

import (
	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/beatmap/objects"
	"github.com/wieku/danser-go/app/bmath"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/framework/math/curves"
	"github.com/wieku/danser-go/framework/math/math32"
	"github.com/wieku/danser-go/framework/math/vector"
	"math"
)

const (
	streamEntryMin = 25
	streamEntryMax = 4000
	streamEscape   = 8000
)

type SplineMover struct {
	curve              *curves.BSpline
	startTime, endTime float64
	mods               difficulty.Modifier
}

func NewSplineMover() MultiPointMover {
	return &SplineMover{}
}

func (mover *SplineMover) Reset(mods difficulty.Modifier) {
	mover.mods = mods
}

func (mover *SplineMover) SetObjects(objs []objects.IHitObject) int {
	points := make([]vector.Vector2f, 0)
	timing := make([]int64, 0)

	var endTime, startTime float64
	var angle float32
	var stream bool

	i := 0

	for ; i < len(objs); i++ {
		o := objs[i]

		if i == 0 {
			cEnd := o.GetStackedEndPositionMod(mover.mods)
			nStart := objs[i+1].GetStackedStartPositionMod(mover.mods)

			var wPoint vector.Vector2f

			switch s := o.(type) {
			case objects.ILongObject:
				wPoint = cEnd.Add(vector.NewVec2fRad(s.GetEndAngleMod(mover.mods), cEnd.Dst(nStart)*0.7))
			default:
				wPoint = cEnd.Lerp(nStart, 0.333)
			}

			points = append(points, cEnd, wPoint)
			timing = append(timing, int64(math.Max(o.GetStartTime(), o.GetEndTime())))

			endTime = math.Max(o.GetStartTime(), o.GetEndTime())

			continue
		}

		if _, ok := o.(objects.ILongObject); ok || i == len(objs)-1 {
			pEnd := objs[i-1].GetStackedEndPositionMod(mover.mods)
			cStart := o.GetStackedStartPositionMod(mover.mods)

			var wPoint vector.Vector2f

			switch s := o.(type) {
			case objects.ILongObject:
				wPoint = cStart.Add(vector.NewVec2fRad(s.GetStartAngleMod(mover.mods), cStart.Dst(pEnd)*0.7))
			default:
				wPoint = cStart.Lerp(pEnd, 0.333)
			}

			points = append(points, wPoint, cStart)
			timing = append(timing, int64(o.GetStartTime()))

			startTime = o.GetStartTime()

			break
		} else if i > 1 && i < len(objs)-1 {
			pos1 := objs[i-1].GetStackedStartPositionMod(mover.mods)
			pos2 := o.GetStackedStartPositionMod(mover.mods)
			pos3 := objs[i+1].GetStackedStartPositionMod(mover.mods)

			min := float32(streamEntryMin)
			max := float32(streamEntryMax)
			if stream {
				max = streamEscape
			}

			sq1 := pos1.DstSq(pos2)
			sq2 := pos2.DstSq(pos3)

			if sq1 > max && sq2 > max && settings.Dance.Spline.RotationalForce {
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
			} else if sq1 >= min && sq1 <= max && sq2 >= min && sq2 <= max && (settings.Dance.Spline.StreamWobble || settings.Dance.Spline.StreamHalfCircle) {
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
				if stream && !settings.Dance.Spline.StreamHalfCircle {
					scale = float32(settings.Dance.Spline.WobbleScale)
				}

				if stream && settings.Dance.Spline.StreamHalfCircle {
					sign := -1
					if angle < 0 {
						sign = 1
					}

					for t := -2; t <= 2; t++ {
						p4 := mid.Sub(pos1).Scl(scale).Rotate(angle + float32(sign*t)*math32.Pi/6).Add(mid)

						points = append(points, p4)
						timing = append(timing, int64((o.GetStartTime()-objs[i-1].GetStartTime())*(3+float64(t))/6+objs[i-1].GetStartTime()))
					}
				} else {
					p4 := mid.Sub(pos1).Scl(scale).Rotate(angle).Add(mid)

					points = append(points, p4)
					timing = append(timing, int64((o.GetStartTime()-objs[i-1].GetStartTime())/2+objs[i-1].GetStartTime()))
				}
			}
		}

		points = append(points, o.GetStackedEndPositionMod(mover.mods))
		timing = append(timing, int64(o.GetStartTime()))
	}

	mover.startTime = startTime
	mover.endTime = endTime
	mover.curve = curves.NewBSpline(points, timing)

	return i + 1
}

func (mover *SplineMover) Update(time float64) vector.Vector2f {
	t := bmath.ClampF32(float32(time-mover.endTime)/float32(mover.startTime-mover.endTime), 0, 1)
	return mover.curve.PointAt(t)
}

func (mover *SplineMover) GetEndTime() float64 {
	return mover.startTime
}
