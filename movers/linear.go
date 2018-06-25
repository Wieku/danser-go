package movers

import (
	math2 "github.com/wieku/danser/bmath"
	"github.com/wieku/danser/bmath/curves"
	"github.com/wieku/danser/beatmap/objects"
	"math"
	"github.com/wieku/danser/render"
)

type LinearMover struct {
	bz curves.Bezier
	beginTime, endTime int64
}

func NewLinearMover() Mover {
	return &LinearMover{}
}

func (bm *LinearMover) Reset() {
}

func (bm *LinearMover) SetObjects(end, start objects.BaseObject) {
	endPos := end.GetBasicData().EndPos
	endTime := end.GetBasicData().EndTime
	startPos := start.GetBasicData().StartPos
	startTime := start.GetBasicData().StartTime

	bm.bz = curves.NewBezier([]math2.Vector2d{endPos, startPos})
	bm.endTime = endTime
	bm.beginTime = startTime
}

func (bm LinearMover) Update(time int64, cursor *render.Cursor) {
	t := float64(time - bm.endTime)/float64(bm.beginTime - bm.endTime)
	t = math.Max(0.0, math.Min(1.0, t))
	cursor.SetPos(bm.bz.NPointAt(t))
}