package objects

import (
	"danser/bmath"
	"danser/audio"
	"strconv"
	"strings"
)

type Circle struct {
	objData *basicData
	sample int
	Timings *Timings
	ownSampleSet int
}

func NewCircle(data []string) *Circle {
	circle := &Circle{}
	circle.objData = commonParse(data)
	f, _ := strconv.ParseInt(data[4], 10, 64)
	circle.sample = int(f)
	circle.objData.EndTime = circle.objData.StartTime
	circle.objData.EndPos = circle.objData.StartPos
	if len(data) > 5 {
		e, _ := strconv.ParseInt(strings.Split(data[5],":")[0], 10, 64)
		circle.ownSampleSet = int(e)
	} else {
		circle.ownSampleSet = 0
	}
	return circle
}

func DummyCircle(pos bmath.Vector2d, time int64) *Circle {
	circle := &Circle{objData:&basicData{}}
	circle.objData.StartPos = pos
	circle.objData.EndPos = pos
	circle.objData.EndTime = circle.objData.StartTime
	circle.objData.EndPos = circle.objData.StartPos
	return circle
}

func (self Circle) GetBasicData() *basicData {
	return self.objData
}

func (self *Circle) Update(time int64) bool {

	//cursor.SetPos(self.objData.StartPos)
	if self.ownSampleSet == 0 {
		audio.PlaySample(self.Timings.Current.SampleSet, self.sample)
	} else {
		audio.PlaySample(self.ownSampleSet, self.sample)
	}

	return true
}

func (self *Circle) SetTiming(timings *Timings) {
	self.Timings = timings
}