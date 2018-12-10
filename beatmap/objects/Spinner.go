package objects

import (
	"danser/bmath"
	"strconv"
)

type Spinner struct {
	objData *basicData
	pos     bmath.Vector2d
	Timings *Timings
}

func NewSpinner(data []string) *Spinner {
	spinner := &Spinner{}
	spinner.objData = commonParse(data)
	endtime, _ := strconv.ParseInt(data[5], 10, 64)
	spinner.objData.EndTime = int64(endtime)
	spinner.pos = bmath.Vector2d{256,192}
	return spinner
}

func (self Spinner) GetBasicData() *basicData {
	return self.objData
}

func (self *Spinner) SetTiming(timings *Timings) {
	self.Timings = timings
}

func (self *Spinner) GetPosition() bmath.Vector2d {
	return self.pos
}

func (self *Spinner) Update(time int64) bool {
	return true
}
