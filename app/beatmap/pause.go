package beatmap

import (
	"strconv"
)

type Pause struct {
	StartTime float64
	EndTime   float64
}

func NewPause(data []string) *Pause {
	pause := &Pause{}
	pause.StartTime, _ = strconv.ParseFloat(data[1], 64)
	pause.EndTime, _ = strconv.ParseFloat(data[2], 64)
	return pause
}

func (pause *Pause) GetStartTime() float64 {
	return pause.StartTime
}

func (pause *Pause) GetEndTime() float64 {
	return pause.EndTime
}

func (pause *Pause) Length() float64 {
	return pause.EndTime - pause.StartTime
}
