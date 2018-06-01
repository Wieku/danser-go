package objects

import (
	"danser/bmath/sliders"
	m2 "danser/bmath"
	"strconv"
	"strings"
	"log"
	"danser/audio"
)

type Slider struct {
	objData     *basicData
	multiCurve  sliders.SliderAlgo
	Timings     *Timings
	pixelLength float64
	partLen 	float64
	repeat      int64
	clicked     bool
	sampleSets 	[]int
	samples 	[]int
	lastT		int64
	Pos 		m2.Vector2d
	divides 	int
	LasTTI int64
	End bool
}

func NewSlider(data []string) *Slider {
	slider := &Slider{clicked: false}
	slider.objData = commonParse(data)
	slider.pixelLength, _ = strconv.ParseFloat(data[7], 64)
	slider.repeat, _ = strconv.ParseInt(data[6], 10, 64)

	list := strings.Split(data[5], "|")
	points := []m2.Vector2d{slider.objData.StartPos}

	if list[0] == "C" {
		return nil
	}

	for i := 1; i < len(list); i++ {
		list2 := strings.Split(list[i], ":")
		x, _ := strconv.ParseFloat(list2[0], 64)
		y, _ := strconv.ParseFloat(list2[1], 64)
		points = append(points, m2.NewVec2d(x, y))
	}

	slider.multiCurve = sliders.NewSliderAlgo(list[0], points)

	slider.objData.EndTime = slider.objData.StartTime
	slider.objData.EndPos = slider.multiCurve.PointAt(float64(slider.repeat%2))
	slider.Pos = slider.objData.StartPos

	slider.samples = make([]int, slider.repeat+1)
	slider.sampleSets = make([]int, slider.repeat+1)
	slider.lastT = 1
	if len(data) > 8 {
		subData := strings.Split(data[8], "|")
		for i, v := range subData {
			f, _ := strconv.ParseInt(v, 10, 64)
			slider.samples[i] = int(f)
		}
	}

	if len(data) > 9 {
		subData := strings.Split(data[9], "|")
		for i, v := range subData {
			f, _ := strconv.ParseInt(strings.Split(v,":")[0], 10, 64)
			slider.sampleSets[i] = int(f)
		}
	}
	slider.End = false
	return slider
}

func (self Slider) GetBasicData() *basicData {
	return self.objData
}

func (self Slider) GetHalf() m2.Vector2d {
	return self.multiCurve.PointAt(0.5).Add(self.objData.StackOffset)
}

func (self Slider) GetStartAngle() float64 {
	return self.GetBasicData().StartPos.AngleRV(self.multiCurve.PointAt(0.02).Add(self.objData.StackOffset))
}

func (self Slider) GetEndAngle() float64 {
	return self.GetBasicData().EndPos.AngleRV(self.multiCurve.PointAt(0.98 - float64(1-self.repeat%2)*0.96).Add(self.objData.StackOffset))
}

func (self Slider) GetPartLen() float64 {
	return 20.0 / float64(self.Timings.GetSliderTime(self.pixelLength)) * self.pixelLength
}

func (self Slider) GetPointAt(time int64) m2.Vector2d {
	partLen := float64(self.Timings.GetSliderTimeS(time, self.pixelLength))
	times := int64(float64(time - self.objData.StartTime) / partLen) + 1

	ttime := float64(time) - float64(self.objData.StartTime) - float64(times-1) * partLen

	rt := float64(self.pixelLength) / self.multiCurve.Length

	var pos m2.Vector2d
	if (times%2) == 1 {
		pos = self.multiCurve.PointAt(rt*ttime/partLen)
	} else {
		pos = self.multiCurve.PointAt((1.0 - ttime/partLen)*rt)
	}

	return pos.Add(self.objData.StackOffset)
}

func (self Slider) endTime() int64 {
	return self.objData.StartTime + self.repeat * self.Timings.GetSliderTime(self.pixelLength)
}

func (self *Slider) SetTiming(timings *Timings) {
	self.Timings = timings
	if timings.GetSliderTimeS(self.objData.StartTime, self.pixelLength) < 0 {
		log.Println( self.objData.StartTime, self.pixelLength, "wuuuuuuuuuuuuuut")
	}
	self.objData.EndTime = self.objData.StartTime + timings.GetSliderTimeS(self.objData.StartTime, self.pixelLength) * self.repeat
}

func (self *Slider) GetCurve() []m2.Vector2d {
	t0 := 2 / self.pixelLength
	rt := float64(self.pixelLength) / self.multiCurve.Length
	points := make([]m2.Vector2d, int(self.pixelLength/2))
	t:= 0.0
	for i:=0; i < int(self.pixelLength/2); i+=1 {
		points[i] = self.multiCurve.PointAt(t*rt)
		t+=t0
	}
	return points
}

func (self *Slider) Update(time int64) bool {
	//TODO: CLEAN THIS
	if time < self.endTime() {
		sliderTime := self.Timings.GetSliderTime(self.pixelLength)
		pixLen := self.multiCurve.Length
		self.partLen = float64(sliderTime)
		self.objData.EndTime = self.objData.StartTime + sliderTime * self.repeat
		times := int64(float64(time - self.objData.StartTime) / self.partLen) + 1

		ttime := float64(time) - float64(self.objData.StartTime) - float64(times-1) * self.partLen

		if self.lastT != times {
			ss := self.sampleSets[times-1]
			if ss == 0 {
				ss = self.Timings.Current.SampleSet
			}
			audio.PlaySample(ss, self.samples[times-1])
			self.lastT = times
		}


		rt := float64(self.pixelLength) / pixLen

		var pos m2.Vector2d
		if (times%2) == 1 {
			pos = self.multiCurve.PointAt(rt*ttime/self.partLen)
		} else {
			pos = self.multiCurve.PointAt((1.0 - ttime/self.partLen)*rt)
		}
		self.Pos = pos
		//cursor.SetPos(pos.Add(self.objData.StackOffset))
		//io.MouseMoveVec(pos.Add(self.objData.StackOffset))

		if !self.clicked {
			//io.MouseClick(io.LEFT)
			ss := self.sampleSets[0]
			if ss == 0 {
				ss = self.Timings.Current.SampleSet
			}
			audio.PlaySample(ss, self.samples[0])
			self.clicked = true
		}

		return false
	}

	ss := self.sampleSets[self.repeat]
	if ss == 0 {
		ss = self.Timings.Current.SampleSet
	}
	audio.PlaySample(ss, self.samples[self.repeat])
	self.End = true
	self.clicked = false
	//io.MouseUnClick(io.LEFT)

	return true
}