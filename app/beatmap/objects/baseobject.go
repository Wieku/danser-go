package objects

import (
	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/framework/graphics/batch"
	color2 "github.com/wieku/danser-go/framework/math/color"
	"github.com/wieku/danser-go/framework/math/vector"
	"strconv"
	"strings"
)

type BaseObject interface {
	GetBasicData() *basicData
	Update(time int64) bool
	SetTiming(timings *Timings)
	UpdateStacking()
	SetDifficulty(difficulty *difficulty.Difficulty)
	GetPosition() vector.Vector2f
}

type Renderable interface {
	Draw(time int64, color color2.Color, batch *batch.QuadBatch) bool
	DrawApproach(time int64, color color2.Color, batch *batch.QuadBatch)
}

type basicData struct {
	StartPos, EndPos   vector.Vector2f
	StartTime, EndTime int64
	StackOffset        vector.Vector2f
	StackIndex         int64
	Number             int64
	SliderPoint        bool
	SliderPointStart   bool
	SliderPointEnd     bool
	NewCombo           bool
	ComboNumber        int64
	ComboSet           int64

	sampleSet    int
	additionSet  int
	customIndex  int
	customVolume float64
}

func commonParse(data []string) *basicData {
	x, _ := strconv.ParseFloat(data[0], 32)
	y, _ := strconv.ParseFloat(data[1], 32)
	time, _ := strconv.ParseInt(data[2], 10, 64)
	objType, _ := strconv.ParseInt(data[3], 10, 64)
	return &basicData{StartPos: vector.NewVec2f(float32(x), float32(y)), StartTime: time, Number: -1, NewCombo: (objType & 4) == 4}
}

func (bData *basicData) parseExtras(data []string, extraIndex int) {
	if extraIndex < len(data) {
		extras := strings.Split(data[extraIndex], ":")
		sampleSet, _ := strconv.ParseInt(extras[0], 10, 64)
		additionSet, _ := strconv.ParseInt(extras[1], 10, 64)
		index, _ := strconv.ParseInt(extras[2], 10, 64)
		if len(extras) > 3 {
			volume, _ := strconv.ParseInt(extras[3], 10, 64)
			bData.customVolume = float64(volume) / 100.0
		} else {
			bData.customVolume = 0
		}

		bData.sampleSet = int(sampleSet)
		bData.additionSet = int(additionSet)
		bData.customIndex = int(index)
	}
}
