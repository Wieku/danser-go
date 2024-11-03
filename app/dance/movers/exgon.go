package movers

import (
	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/beatmap/objects"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/framework/math/animation/easing"
	"github.com/wieku/danser-go/framework/math/vector"
	"math/rand"
)

type ExGonMover struct {
	*basicMover

	wasFirst bool
	rand     *rand.Rand

	endPos vector.Vector2f

	lastPos  vector.Vector2f
	nextTime float64

	delay float64
}

func NewExGonMover() MultiPointMover {
	return &ExGonMover{basicMover: &basicMover{}}
}

func (mover *ExGonMover) Reset(diff *difficulty.Difficulty, id int) {
	mover.basicMover.Reset(diff, id)
	mover.wasFirst = false
}

func (mover *ExGonMover) SetObjects(objs []objects.IHitObject) int {
	config := settings.CursorDance.MoverSettings.ExGon[mover.id%len(settings.CursorDance.MoverSettings.ExGon)]
	mover.delay = float64(config.Delay)

	if !mover.wasFirst {
		mover.rand = rand.New(rand.NewSource((int64(objs[1].GetStartPosition().X)+1000*int64(objs[1].GetStartPosition().Y))*100 + int64(objs[1].GetStartTime())))

		mover.wasFirst = true
	}

	start, end := objs[0], objs[1]

	mover.nextTime = start.GetEndTime() + mover.delay

	mover.startTime = start.GetStartTime()
	mover.endTime = end.GetStartTime()

	mover.lastPos = start.GetStackedEndPositionMod(mover.diff)
	mover.endPos = end.GetStackedStartPositionMod(mover.diff)

	return 2
}

func (mover *ExGonMover) Update(time float64) vector.Vector2f {
	if mover.endTime-time < mover.delay {
		return mover.endPos
	}

	if time >= mover.nextTime {
		mover.nextTime += mover.delay

		mover.lastPos = vector.NewVec2f(568, 426).Mult(vector.NewVec2f(float32(easing.InOutCubic(mover.rand.Float64())), float32(easing.InOutCubic(mover.rand.Float64())))).SubS(28, 21)
	}

	return mover.lastPos
}
