package objects

import (
	"strconv"
	"github.com/wieku/danser-go/bmath"
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
	pause.objData.Number = -1
	return pause
}

func (self *Pause) GetBasicData() *basicData {
	return self.objData
}

func (self *Pause) SetTiming(timings *Timings) {

}

func (self *Pause) SetDifficulty(preempt, fadeIn float64) {

}

func (self *Pause) Update(time int64) bool {
	return time >= self.objData.EndTime
}

func (self *Pause) GetPosition() bmath.Vector2d {
	return self.objData.StartPos
}
