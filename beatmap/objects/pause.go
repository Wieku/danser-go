package objects

import (
	"strconv"
	"danser/bmath"
	"danser/render"
)

type Pause struct {
	objData *basicData
}

func NewPause(data []string) *Pause {
	pause := &Pause{}
	pause.objData = &basicData{}
	pause.objData.StartTime, _ = strconv.ParseInt(data[1], 10, 64)
	pause.objData.EndTime, _ = strconv.ParseInt(data[2], 10, 64)
	pause.objData.StartPos = bmath.NewVec2d(512/2, 384/2)
	pause.objData.EndPos = pause.objData.StartPos
	return pause
}

func (self Pause) GetBasicData() *basicData {
	return self.objData
}

func (self *Pause) Update(time int64, cursor *render.Cursor) bool {

	cursor.SetPos(self.objData.StartPos)
	//io.MouseMoveVec(self.objData.StartPos)

	return time >= self.objData.EndTime
}