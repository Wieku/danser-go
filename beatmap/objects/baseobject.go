package objects

import (
	om "github.com/wieku/danser/bmath"
	"strconv"
	"github.com/wieku/danser/render"
)

type BaseObject interface {
	GetBasicData() *basicData
	Update(time int64, cursor *render.Cursor) bool
}

type basicData struct {
	StartPos, EndPos   om.Vector2d
	StartTime, EndTime int64
	StackOffset        om.Vector2d
	StackIndex         int64
	Number			   int64
}

func commonParse(data []string) *basicData {
	x, _ := strconv.ParseFloat(data[0], 64)
	y, _ := strconv.ParseFloat(data[1], 64)
	time, _ := strconv.ParseInt(data[2], 10, 64)
	return &basicData{StartPos: om.NewVec2d(x, y), StartTime: time, Number: -1}
}
