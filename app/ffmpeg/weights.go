package ffmpeg

import (
	"fmt"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/framework/math/animation/easing"
	"strconv"
	"strings"
)

var easings = []easing.Easing{
	easing.Linear,
	easing.InQuad,
	easing.OutQuad,
	easing.InOutQuad,
	easing.InCubic,
	easing.OutCubic,
	easing.InOutCubic,
	easing.InQuart,
	easing.OutQuart,
	easing.InOutQuart,
	easing.InQuint,
	easing.OutQuint,
	easing.InOutQuint,
	easing.InSine,
	easing.OutSine,
	easing.InOutSine,
	easing.InExpo,
	easing.OutExpo,
	easing.InOutExpo,
	easing.InCirc,
	easing.OutCirc,
	easing.InOutCirc,
	easing.InBack,
	easing.OutBack,
	easing.InOutBack,
}

func calculateWeights(bFrames int) []float32 {
	var weights []float32

	if settings.Recording.MotionBlur.BlendWeights.UseManualWeights {
		weightsSplit := strings.Split(settings.Recording.MotionBlur.BlendWeights.ManualWeights, " ")

		for _, s := range weightsSplit {
			v, err := strconv.ParseFloat(s, 32)
			if err != nil {
				panic(fmt.Sprintf("Failed to parse weight: %s", s))
			}

			weights = append(weights, float32(v))
		}
	} else {
		id := settings.Recording.MotionBlur.BlendWeights.AutoWeightsID
		if id < 0 || id > len(easings) {
			id = 0
		}

		if id == 0 {
			for i := 0; i < bFrames; i++ {
				weights = append(weights, 1)
			}
		} else {
			easeFunc := easings[id-1]

			for i := 0; i < bFrames; i++ {
				w := 1.0 + easeFunc(float64(i)/float64(bFrames-1)) * 100
				weights = append(weights, float32(w))
			}
		}
	}

	return weights
}
