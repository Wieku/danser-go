package movers

import (
	"github.com/wieku/danser-go/app/beatmap/objects"
	"github.com/wieku/danser-go/app/bmath"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/framework/math/curves"
	"github.com/wieku/danser-go/framework/math/math32"
	"github.com/wieku/danser-go/framework/math/vector"
)

type SplineMover struct {
	curve              *curves.BSpline
	startTime, endTime int64
}

func NewSplineMover() MultiPointMover {
	return &SplineMover{}
}

func (mover *SplineMover) Reset() {

}

func (mover *SplineMover) SetObjects(objs []objects.IHitObject) int {
	points := make([]vector.Vector2f, 0)
	timing := make([]int64, 0)

	var endTime, startTime int64
	var angle float32
	var stream bool

	i := 0

	for ; i < len(objs); i++ {
		if i == 0 {
			if s, ok := objs[i].(*objects.Slider); ok {
				points = append(points, s.GetStackedEndPosition(), vector.NewVec2fRad(s.GetEndAngle(), s.GetStackedEndPosition().Dst(objs[i+1].GetStackedStartPosition())*0.7).Add(s.GetStackedEndPosition()))
			}

			if s, ok := objs[i].(*objects.Circle); ok {
				points = append(points, s.GetStackedEndPosition(), objs[i+1].GetStackedStartPosition().Sub(s.GetStackedEndPosition()).Scl(0.333).Add(s.GetStackedEndPosition()))
			}

			if s, ok := objs[i].(*objects.Spinner); ok {
				points = append(points, s.GetStackedEndPosition(), objs[i+1].GetStackedStartPosition().Sub(s.GetStackedEndPosition()).Scl(0.333).Add(s.GetStackedEndPosition()))
			}

			timing = append(timing, bmath.MaxI64(objs[i].GetStartTime(), objs[i].GetEndTime()))
			endTime = bmath.MaxI64(objs[i].GetStartTime(), objs[i].GetEndTime())

			continue
		}

		_, ok1 := objs[i].(*objects.Slider)
		_, ok2 := objs[i].(*objects.Spinner)

		ok := ok1 || ok2

		if ok || i == len(objs)-1 {
			if s, ok := objs[i].(*objects.Slider); ok {
				points = append(points, vector.NewVec2fRad(s.GetStartAngle(), s.GetStackedStartPosition().Dst(objs[i-1].GetStackedEndPosition())*0.7).Add(s.GetStackedStartPosition()), s.GetStackedStartPosition())
			}

			if s, ok := objs[i].(*objects.Circle); ok {
				points = append(points, objs[i-1].GetStackedEndPosition().Sub(s.GetStackedStartPosition()).Scl(0.333).Add(s.GetStackedStartPosition()), s.GetStackedStartPosition())
			}

			if s, ok := objs[i].(*objects.Spinner); ok {
				points = append(points, objs[i-1].GetStackedEndPosition().Sub(s.GetStackedStartPosition()).Scl(0.333).Add(s.GetStackedStartPosition()), s.GetStackedStartPosition())
			}

			timing = append(timing, objs[i].GetStartTime())

			startTime = objs[i].GetStartTime()

			break
		} else if i < len(objs)-1 && i-1 > 0 {
			if _, ok := objs[i].(*objects.Circle); ok {
				pos1 := objs[i-1].GetStackedStartPosition()
				pos2 := objs[i].GetStackedStartPosition()
				pos3 := objs[i+1].GetStackedStartPosition()

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
							timing = append(timing, (objs[i].GetStartTime()-objs[i-1].GetStartTime())*(3+int64(t))/6+objs[i-1].GetStartTime())
						}
					} else {
						p4 := mid.Sub(pos1).Scl(scale).Rotate(angle).Add(mid)

						points = append(points, p4)
						timing = append(timing, (objs[i].GetStartTime()-objs[i-1].GetStartTime())/2+objs[i-1].GetStartTime())
					}
				}
			}
		}

		if s, ok := objs[i].(*objects.Circle); ok {
			points = append(points, s.GetStackedEndPosition())
			timing = append(timing, s.GetStartTime())
		}
	}

	mover.startTime = startTime
	mover.endTime = endTime
	mover.curve = curves.NewBSpline(points, timing)

	return i + 1
}

func (mover *SplineMover) Update(time int64) vector.Vector2f {
	t := bmath.ClampF32(float32(time-mover.endTime)/float32(mover.startTime-mover.endTime), 0, 1)
	return mover.curve.PointAt(t)
}

func (mover *SplineMover) GetEndTime() int64 {
	return mover.startTime
}
