package beatmap

import (
	"strconv"
)

type Pause struct {
	StartTime int64
	EndTime   int64
}

func NewPause(data []string) *Pause {
	pause := &Pause{}
	pause.StartTime, _ = strconv.ParseInt(data[1], 10, 64)
	pause.EndTime, _ = strconv.ParseInt(data[2], 10, 64)
	return pause
}

func (pause *Pause) GetStartTime() int64 {
	return pause.StartTime
}

func (pause *Pause) GetEndTime() int64 {
	return pause.EndTime
}
