package movers

import (
	"danser/bmath/curves"
	//"osubot/io"
	"danser/beatmap/objects"
	"math"
)

const INVERTABLE = false
const CIRFRAGMENT = 0.5

type CircularMover struct {
	ca curves.Curve
	beginTime, endTime int64
	invert float64
}

func NewCircularMover() *CircularMover {
	cm := &CircularMover{invert:-1}
	return cm
}

func (bm *CircularMover) Reset() {
	bm.invert = -1
}

func (bm *CircularMover) SetObjects(end, start objects.BaseObject) {
	endPos := end.GetBasicData().EndPos
	startPos := start.GetBasicData().StartPos
	bm.endTime = end.GetBasicData().EndTime
	bm.beginTime = start.GetBasicData().StartTime

	if INVERTABLE {
		bm.invert = -1 * bm.invert
	}

	if endPos == startPos {
		bm.ca = curves.NewLinear(endPos, startPos)
		return
	}

	p := endPos.Mid(startPos)
	p = p.Sub(endPos).Rotate(bm.invert*math.Pi/2).Scl(CIRFRAGMENT).Add(p)
	bm.ca = curves.NewCirArc(endPos, p, startPos)
}

func (bm CircularMover) Update(time int64/*, cursor *render.Cursor*/) {
	//cursor.SetPos(bm.ca.PointAt(float64(time - bm.endTime)/float64(bm.beginTime - bm.endTime)))
	//io.MouseMoveVec(bm.ca.PointAt(float64(time - bm.endTime)/float64(bm.beginTime - bm.endTime)))
}
