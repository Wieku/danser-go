package objects

import (
	"github.com/wieku/danser-go/app/beatmap/difficulty"
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

func (pause *Pause) GetBasicData() *basicData {
	return pause.objData
}

func (pause *Pause) SetTiming(timings *Timings) {

}

func (pause *Pause) UpdateStacking() {}

func (pause *Pause) SetDifficulty(diff *difficulty.Difficulty) {

}

func (pause *Pause) Update(time int64) bool {
	return time >= pause.objData.EndTime
}

func (pause *Pause) GetPosition() vector.Vector2f {
	return pause.objData.StartPos
}
