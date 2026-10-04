package putils

type LogisticSum struct {
	stepCalc bool

	midpointOffset, multiplier, maxValue float64

	previousDivider float64
	first           bool

	recalculateCheck func(previous, current float64) bool
	fullDivider      func(strains []float64) float64

	value float64

	strains []float64
}

func NewLogisticSum(stepCalc bool, midpointOffset, multiplier, maxValue float64, recalculateCheck func(previous, current float64) bool, fullDivider func(strains []float64) float64) *LogisticSum {
	return &LogisticSum{
		stepCalc:         stepCalc,
		midpointOffset:   midpointOffset,
		multiplier:       multiplier,
		maxValue:         maxValue,
		recalculateCheck: recalculateCheck,
		fullDivider:      fullDivider,
		first:            true,
	}
}

func (lsum *LogisticSum) AddStrain(strain float64) {
	lsum.strains = append(lsum.strains, strain)
}

func (lsum *LogisticSum) ProcessLastStrain(divider float64) {
	if !lsum.stepCalc {
		return
	}

	if lsum.first || lsum.recalculateCheck(lsum.previousDivider, divider) {
		lsum.first = false
		lsum.previousDivider = divider
		lsum.value = lsum.calculateFull()
	} else if len(lsum.strains) > 0 && lsum.previousDivider != 0 {
		lsum.value += lsum.calculateStrain(lsum.strains[len(lsum.strains)-1])
	}
}

func (lsum *LogisticSum) UpdateDivider(divider float64) {
	if !lsum.stepCalc {
		return
	}

	if lsum.first || lsum.recalculateCheck(lsum.previousDivider, divider) {
		lsum.first = false
		lsum.previousDivider = divider
		lsum.value = lsum.calculateFull()
	}
}

func (lsum *LogisticSum) calculateStrain(strain float64) float64 {
	return Logistic(strain/lsum.previousDivider, lsum.midpointOffset, lsum.multiplier, lsum.maxValue)
}

func (lsum *LogisticSum) calculateFull() (sum float64) {
	if len(lsum.strains) == 0 || lsum.previousDivider == 0 {
		return
	}

	for _, s := range lsum.strains {
		sum += lsum.calculateStrain(s)
	}

	return
}

func (lsum *LogisticSum) GetValue() float64 {
	if lsum.stepCalc {
		return lsum.value
	}

	if len(lsum.strains) == 0 {
		return 0
	}

	lsum.previousDivider = lsum.fullDivider(lsum.strains)

	return lsum.calculateFull()
}
