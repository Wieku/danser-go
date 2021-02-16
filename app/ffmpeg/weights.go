package ffmpeg

import (
	"fmt"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/framework/math/animation/easing"
	"math"
	"strconv"
	"strings"
)

var easings = []easing.Easing{
	flat,
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
	inBack,
	easing.OutBack,
	inOutBack,
	gauss,
	gaussSymmetric,
	pyramidSymmetric,
	semiCircle,
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

		easeFunc := easings[id]
		for i := 0; i < bFrames; i++ {
			w := 1.0 + easeFunc(float64(i)/float64(bFrames-1))*100
			weights = append(weights, float32(w))
		}
	}

	return weights
}

func flat(_ float64) float64 {
	return 1.0
}

func inBack(t float64) float64 {
	return easing.InBack(t) + 0.100004
}

func inOutBack(t float64) float64 {
	return easing.InOutBack(t) + 0.100004
}

func gauss(t float64) float64 {
	return math.Exp(-math.Pow(settings.Recording.MotionBlur.BlendWeights.GaussWeightsMult*(t-1), 2))
}

func gaussSymmetric(t float64) float64 {
	return math.Exp(-math.Pow(settings.Recording.MotionBlur.BlendWeights.GaussWeightsMult*(t*2-1), 2))
}

func pyramidSymmetric(t float64) float64 {
	return 1.0 - math.Abs(t*2-1)
}

func semiCircle(t float64) float64 {
	return math.Sqrt(1 - math.Pow(0.5-t, 2))
}
