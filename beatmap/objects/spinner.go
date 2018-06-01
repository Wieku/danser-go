package objects

import (
	"strconv"
)
const rps = 8.0

type Spinner struct {

	objData *basicData
	clicked bool
}

func NewSpinner(data []string) *Spinner {
	spinner := &Spinner{clicked: false}
	spinner.objData = commonParse(data)
	spinner.objData.EndTime, _ = strconv.ParseInt(data[5], 10, 64)
	spinner.objData.EndPos = spinner.objData.StartPos
	return spinner
}

func (self Spinner) GetBasicData() *basicData {
	return self.objData
}

func (self *Spinner) Update(time int64/*, cursor *render.Cursor*/) bool {
	if !self.clicked {
		//io.MouseClick(io.LEFT)
		self.clicked = true
	}
	data := self.objData
	//len := 150.0 * math.Sin(float64(data.EndTime - time) * math.Pi / float64(data.EndTime - data.StartTime))

	//cursor.SetPos(bmath.NewVec2dRad(float64(time - data.StartTime)*2*math.Pi/(1000/rps), len).Add(data.StartPos))
	//io.MouseMoveVec(math2.NewVec2dRad(float64(time - data.StartTime)*2*math.Pi/(1000/rps), len).Add(data.StartPos))

	if time >= data.EndTime {
		//io.MouseUnClick(io.LEFT)
		self.clicked = false
		return true
	}

	return false
}