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
	wasFirst bool
	rand     *rand.Rand

	lastPos  vector.Vector2f
	nextTime float64

	endTime float64
	mods    difficulty.Modifier
}

func NewExGonMover() MultiPointMover {
	return &ExGonMover{}
}

func (bm *ExGonMover) Reset(mods difficulty.Modifier) {
	bm.mods = mods
	bm.wasFirst = false
}

func (bm *ExGonMover) SetObjects(objs []objects.IHitObject) int {
	if !bm.wasFirst {
		bm.rand = rand.New(rand.NewSource((int64(objs[1].GetStartPosition().X)+1000*int64(objs[1].GetStartPosition().Y))*100 + int64(objs[1].GetStartTime())))

		bm.wasFirst = true
	}

	prev, next := objs[0], objs[1]

	bm.nextTime = prev.GetEndTime() + float64(settings.Dance.ExGon.Delay)
	bm.endTime = next.GetStartTime()

	return 2
}

func (bm *ExGonMover) Update(time float64) vector.Vector2f {
	if time >= bm.nextTime {
		bm.nextTime += float64(settings.Dance.ExGon.Delay)

		bm.lastPos = vector.NewVec2f(568, 426).Mult(vector.NewVec2f(float32(easing.InOutCubic(bm.rand.Float64())), float32(easing.InOutCubic(bm.rand.Float64())))).SubS(28, 21)
	}

	return bm.lastPos
}

func (bm *ExGonMover) GetEndTime() float64 {
	return bm.endTime
}
