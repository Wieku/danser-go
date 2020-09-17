package objects

import (
	"github.com/wieku/danser-go/app/bmath/difficulty"
	"github.com/wieku/danser-go/framework/math/vector"
	"strconv"
)

type Pause struct {
	objData *basicData
}

func NewPause(data []string) *Pause {
	pause := &Pause{}
	pause.objData = &basicData{}
	pause.objData.StartTime, _ = strconv.ParseInt(data[1], 10, 64)
	pause.objData.EndTime, _ = strconv.ParseInt(data[2], 10, 64)
	pause.objData.StartPos = vector.NewVec2f(512/2, 384/2)
	pause.objData.EndPos = pause.objData.StartPos
	pause.objData.Number = -1
	return pause
}

func (self *Pause) GetBasicData() *basicData {
	return self.objData
}

func (self *Pause) SetTiming(timings *Timings) {

}

func (self *Pause) UpdateStacking() {}

func (self *Pause) SetDifficulty(diff *difficulty.Difficulty) {

}

func (self *Pause) Update(time int64) bool {
	return time >= self.objData.EndTime
}

func (self *Pause) GetPosition() vector.Vector2f {
	return self.objData.StartPos
}
