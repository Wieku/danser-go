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
		if i == 0 {
			if s, ok := objs[i].(*objects.Slider); ok {
				points = append(points, s.GetStackedEndPositionMod(mover.mods), vector.NewVec2fRad(s.GetEndAngleMod(mover.mods), s.GetStackedEndPositionMod(mover.mods).Dst(objs[i+1].GetStackedStartPositionMod(mover.mods))*0.7).Add(s.GetStackedEndPositionMod(mover.mods)))
			}

			if s, ok := objs[i].(*objects.Circle); ok {
				points = append(points, s.GetStackedEndPositionMod(mover.mods), objs[i+1].GetStackedStartPositionMod(mover.mods).Sub(s.GetStackedEndPositionMod(mover.mods)).Scl(0.333).Add(s.GetStackedEndPositionMod(mover.mods)))
			}

			if s, ok := objs[i].(*objects.Spinner); ok {
				points = append(points, s.GetStackedEndPositionMod(mover.mods), objs[i+1].GetStackedStartPositionMod(mover.mods).Sub(s.GetStackedEndPositionMod(mover.mods)).Scl(0.333).Add(s.GetStackedEndPositionMod(mover.mods)))
			}

			timing = append(timing, int64(math.Max(objs[i].GetStartTime(), objs[i].GetEndTime())))
			endTime = math.Max(objs[i].GetStartTime(), objs[i].GetEndTime())

			continue
		}

		_, ok1 := objs[i].(*objects.Slider)
		_, ok2 := objs[i].(*objects.Spinner)

		ok := ok1 || ok2

		if ok || i == len(objs)-1 {
			if s, ok := objs[i].(*objects.Slider); ok {
				points = append(points, vector.NewVec2fRad(s.GetStartAngleMod(mover.mods), s.GetStackedStartPositionMod(mover.mods).Dst(objs[i-1].GetStackedEndPositionMod(mover.mods))*0.7).Add(s.GetStackedStartPositionMod(mover.mods)), s.GetStackedStartPositionMod(mover.mods))
			}

			if s, ok := objs[i].(*objects.Circle); ok {
				points = append(points, objs[i-1].GetStackedEndPositionMod(mover.mods).Sub(s.GetStackedStartPositionMod(mover.mods)).Scl(0.333).Add(s.GetStackedStartPositionMod(mover.mods)), s.GetStackedStartPositionMod(mover.mods))
			}

			if s, ok := objs[i].(*objects.Spinner); ok {
				points = append(points, objs[i-1].GetStackedEndPositionMod(mover.mods).Sub(s.GetStackedStartPositionMod(mover.mods)).Scl(0.333).Add(s.GetStackedStartPositionMod(mover.mods)), s.GetStackedStartPositionMod(mover.mods))
			}

			timing = append(timing, int64(objs[i].GetStartTime()))

			startTime = objs[i].GetStartTime()

			break
		} else if i < len(objs)-1 && i-1 > 0 {
			if _, ok := objs[i].(*objects.Circle); ok {
				pos1 := objs[i-1].GetStackedStartPositionMod(mover.mods)
				pos2 := objs[i].GetStackedStartPositionMod(mover.mods)
				pos3 := objs[i+1].GetStackedStartPositionMod(mover.mods)

				min := float32(25.0)
				max := float32(4000.0)
				if stream {
					max = 8000
				}

				sq1 := pos1.DstSq(pos2)
				sq2 := pos2.DstSq(pos3)

				if sq1 > max && sq2 > max && settings.Dance.Spline.RotationalForce {
					if stream {
						angle = 0
						stream = false
					} else {
						ang := int(math32.Abs(pos1.Sub(pos2).AngleR() - pos1.Sub(pos3).AngleR()))

						if ang == 0 {
							angle *= -1
						} else {
							angle = float32(ang) * 90 / 180 * math32.Pi
						}
					}
				} else if sq1 >= min && sq2 >= min && sq1 <= max && sq2 <= max && (settings.Dance.Spline.StreamWobble || settings.Dance.Spline.StreamHalfCircle) {
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
							timing = append(timing, int64((objs[i].GetStartTime()-objs[i-1].GetStartTime())*(3+float64(t))/6+objs[i-1].GetStartTime()))
						}
					} else {
						p4 := mid.Sub(pos1).Scl(scale).Rotate(angle).Add(mid)

						points = append(points, p4)
						timing = append(timing, int64((objs[i].GetStartTime()-objs[i-1].GetStartTime())/2+objs[i-1].GetStartTime()))
					}
				}
			}
		}

		if s, ok := objs[i].(*objects.Circle); ok {
			points = append(points, s.GetStackedEndPositionMod(mover.mods))
			timing = append(timing, int64(s.GetStartTime()))
		}
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
