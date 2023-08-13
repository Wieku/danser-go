package animation

import (
	color2 "github.com/wieku/danser-go/framework/math/color"
	"github.com/wieku/danser-go/framework/math/mutils"
	"github.com/wieku/danser-go/framework/math/vector"
)

type TransformationType int64
type TransformationStatus int64

const (
	Fade = TransformationType(1 << iota)
	Rotate
	Scale
	ScaleVector
	Move
	MoveX
	MoveY
	Color3
	Color4
	HorizontalFlip
	VerticalFlip
	Additive
)

const (
	NotStarted = TransformationStatus(1 << iota)
	Going
	Ended
)

func timeClamp(start, end, time float64) float64 {
	if time < start {
		return 0.0
	}

	if time >= end {
		return 1.0
	}

	return mutils.Clamp((time-start)/(end-start), 0, 1)
}

type Transformation struct {
	transformationType TransformationType
	startValues        [4]float64
	endValues          [4]float64
	easing             func(float64) float64
	startTime, endTime float64

	repetitions int
	loopDelay   float64

	id int64
}

func NewBooleanTransform(transformationType TransformationType, startTime, endTime float64) *Transformation {
	if transformationType&(HorizontalFlip|VerticalFlip|Additive) == 0 {
		panic("Wrong TransformationType used!")
	}

	return &Transformation{transformationType: transformationType, startTime: startTime, endTime: endTime}
}

func NewSingleTransform(transformationType TransformationType, easing func(float64) float64, startTime, endTime, startValue, endValue float64) *Transformation {
	if transformationType&(Fade|Rotate|Scale|MoveX|MoveY) == 0 {
		panic("Wrong TransformationType used!")
	}

	transformation := &Transformation{transformationType: transformationType, startTime: startTime, endTime: endTime, easing: easing}
	transformation.startValues[0] = startValue
	transformation.endValues[0] = endValue

	return transformation
}

func NewVectorTransform(transformationType TransformationType, easing func(float64) float64, startTime, endTime, startValueX, startValueY, endValueX, endValueY float64) *Transformation {
	if transformationType&(ScaleVector|Move) == 0 {
		panic("Wrong TransformationType used!")
	}

	transformation := &Transformation{transformationType: transformationType, startTime: startTime, endTime: endTime, easing: easing}
	transformation.startValues[0] = startValueX
	transformation.startValues[1] = startValueY
	transformation.endValues[0] = endValueX
	transformation.endValues[1] = endValueY

	return transformation
}

func NewVectorTransformV(transformationType TransformationType, easing func(float64) float64, startTime, endTime float64, start, end vector.Vector2d) *Transformation {
	if transformationType&(ScaleVector|Move) == 0 {
		panic("Wrong TransformationType used!")
	}

	transformation := &Transformation{transformationType: transformationType, startTime: startTime, endTime: endTime, easing: easing}
	transformation.startValues[0] = start.X
	transformation.startValues[1] = start.Y
	transformation.endValues[0] = end.X
	transformation.endValues[1] = end.Y

	return transformation
}

func NewColorTransform(transformationType TransformationType, easing func(float64) float64, startTime, endTime float64, start, end color2.Color) *Transformation {
	if transformationType&(Color3|Color4) == 0 {
		panic("Wrong TransformationType used!")
	}

	transformation := &Transformation{transformationType: transformationType, startTime: startTime, endTime: endTime, easing: easing}
	transformation.startValues[0] = float64(start.R)
	transformation.startValues[1] = float64(start.G)
	transformation.startValues[2] = float64(start.B)
	transformation.startValues[3] = float64(start.A)

	transformation.endValues[0] = float64(end.R)
	transformation.endValues[1] = float64(end.G)
	transformation.endValues[2] = float64(end.B)
	transformation.endValues[3] = float64(end.A)

	return transformation
}

func (t *Transformation) GetStatus(time float64) TransformationStatus {
	if time < t.startTime {
		return NotStarted
	} else if time >= t.endTime {
		return Ended
	}

	return Going
}

func (t *Transformation) getProgress(time float64) float64 {
	return t.easing(timeClamp(t.startTime, t.endTime, time))
}

func (t *Transformation) GetSingle(time float64) float64 {
	return t.startValues[0] + t.getProgress(time)*(t.endValues[0]-t.startValues[0])
}

func (t *Transformation) GetDouble(time float64) (float64, float64) {
	progress := t.getProgress(time)
	return t.startValues[0] + progress*(t.endValues[0]-t.startValues[0]), t.startValues[1] + progress*(t.endValues[1]-t.startValues[1])
}

func (t *Transformation) GetVector(time float64) vector.Vector2d {
	return vector.NewVec2d(t.GetDouble(time))
}

func (t *Transformation) GetBoolean(time float64) bool {
	return t.startTime == t.endTime || time >= t.startTime && time < t.endTime
}

func (t *Transformation) GetColor(time float64) color2.Color {
	progress := t.getProgress(time)

	return color2.NewRGBA(
		float32(t.startValues[0]+progress*(t.endValues[0]-t.startValues[0])),
		float32(t.startValues[1]+progress*(t.endValues[1]-t.startValues[1])),
		float32(t.startValues[2]+progress*(t.endValues[2]-t.startValues[2])),
		float32(t.startValues[3]+progress*(t.endValues[3]-t.startValues[3])),
	)
}

func (t *Transformation) GetStartTime() float64 {
	return t.startTime
}

func (t *Transformation) GetEndTime() float64 {
	return t.endTime
}

func (t *Transformation) GetTotalEndTime() float64 {
	return t.endTime + float64(t.repetitions)*t.loopDelay
}

func (t *Transformation) GetType() TransformationType {
	return t.transformationType
}

func (t *Transformation) Clone(startTime, endTime float64) *Transformation {
	return &Transformation{
		transformationType: t.transformationType,
		startValues:        [4]float64{t.startValues[0], t.startValues[1], t.startValues[2], t.startValues[3]},
		endValues:          [4]float64{t.endValues[0], t.endValues[1], t.endValues[2], t.endValues[3]},
		easing:             t.easing,
		startTime:          startTime,
		endTime:            endTime,
	}
}

func (t *Transformation) SetLoop(runs int, delay float64) {
	t.repetitions = max(0, runs-1)
	t.loopDelay = delay
}

func (t *Transformation) IsLoop() bool {
	return t.repetitions > 0
}

func (t *Transformation) UpdateLoop() {
	if t.repetitions > 0 {

		t.startTime += t.loopDelay
		t.endTime += t.loopDelay
		t.repetitions--
	}
}

func (t *Transformation) SetID(id int64) {
	t.id = id
}

func (t *Transformation) GetID() int64 {
	return t.id
}
