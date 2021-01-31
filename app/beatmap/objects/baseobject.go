package objects

import (
	"github.com/wieku/danser-go/app/audio"
	"github.com/wieku/danser-go/framework/graphics/batch"
	color2 "github.com/wieku/danser-go/framework/math/color"
	"github.com/wieku/danser-go/framework/math/vector"
	"strconv"
	"strings"
)

type Renderable interface {
	Draw(time int64, color color2.Color, batch *batch.QuadBatch) bool
	DrawApproach(time int64, color color2.Color, batch *batch.QuadBatch)
}

func commonParse(data []string, extraIndex int) *HitObject {
	x, _ := strconv.ParseFloat(data[0], 32)
	y, _ := strconv.ParseFloat(data[1], 32)
	time, _ := strconv.ParseInt(data[2], 10, 64)
	objType, _ := strconv.ParseInt(data[3], 10, 64)

	startPos := vector.NewVec2f(float32(x), float32(y))

	hitObject := &HitObject{
		StartPosRaw: startPos,
		EndPosRaw: startPos,
		StartTime: time,
		EndTime: time,
		HitObjectID: -1,
		NewCombo: (objType & 4) == 4,
	}

	hitObject.BasicHitSound = parseExtras(data, extraIndex)

	return hitObject
}

func parseExtras(data []string, extraIndex int) audio.HitSoundInfo {
	info := audio.HitSoundInfo{}
	if extraIndex < len(data) {
		extras := strings.Split(data[extraIndex], ":")
		sampleSet, _ := strconv.ParseInt(extras[0], 10, 64)
		additionSet, _ := strconv.ParseInt(extras[1], 10, 64)
		index, _ := strconv.ParseInt(extras[2], 10, 64)
		if len(extras) > 3 {
			volume, _ := strconv.ParseInt(extras[3], 10, 64)
			info.CustomVolume = float64(volume) / 100.0
		} else {
			info.CustomVolume = 0
		}

		info.SampleSet = int(sampleSet)
		info.AdditionSet = int(additionSet)
		info.CustomIndex = int(index)
	}

	return info
}
