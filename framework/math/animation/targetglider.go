package animation

import (
	"github.com/wieku/danser-go/framework/math/mutils"
	"math"
)

type TargetGlider struct {
	value       float64
	targetValue float64
	decimals    int

	firstTime bool
	lastTime  float64
}

func NewTargetGlider(startValue float64, decimals int) *TargetGlider {
	return &TargetGlider{
		value:       startValue,
		targetValue: startValue,
		decimals:    mutils.Clamp(decimals, 0, 5),
		firstTime:   true,
	}
}

func (glider *TargetGlider) Update(time float64) {
	if glider.firstTime {
		glider.lastTime = time
		glider.firstTime = false
	}

	glider.UpdateDelta(time - glider.lastTime)
}

func (glider *TargetGlider) UpdateDelta(delta float64) {
	delta60 := delta / 16.66667

	if math.Abs(glider.value-glider.targetValue) < 0.5/math.Pow(10, float64(glider.decimals)) {
		glider.value = glider.targetValue
	} else {
		glider.value = glider.targetValue + (glider.value-glider.targetValue)*math.Pow(0.75-float64(glider.decimals)*0.125, delta60)
	}

	glider.lastTime += delta
}

func (glider *TargetGlider) GetValue() float64 {
	return glider.value
}

func (glider *TargetGlider) SetValue(value float64) {
	glider.value = value
	glider.targetValue = value
}

func (glider *TargetGlider) SetTarget(value float64) {
	glider.targetValue = value
}

func (glider *TargetGlider) SetDecimals(decimals int) {
	glider.decimals = mutils.Clamp(decimals, 0, 5)
}