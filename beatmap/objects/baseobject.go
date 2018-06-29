package objects

import (
	om "github.com/wieku/danser/bmath"
	"strconv"
)

type BaseObject interface {
	GetBasicData() *basicData
	Update(time int64) bool
	GetPosition() om.Vector2d
}

type basicData struct {
	StartPos, EndPos   om.Vector2d
	StartTime, EndTime int64
	StackOffset        om.Vector2d
	StackIndex         int64
	Number			   int64
	SliderPoint		   bool
}

func commonParse(data []string) *basicData {
	x, _ := strconv.ParseFloat(data[0], 64)
	y, _ := strconv.ParseFloat(data[1], 64)
	time, _ := strconv.ParseInt(data[2], 10, 64)
	return &basicData{StartPos: om.NewVec2d(x, y), StartTime: time, Number: -1}
}
