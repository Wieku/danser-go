package skills

import "math"

type timedStrain struct {
	value     float64
	startTime float64
}

// getStrainPeaks reconstructs chronological 400ms peaks without changing skill difficulty values.
func getStrainPeaks(strains []timedStrain, strainDecay func(ms float64) float64) []float64 {
	const sectionLength = 400

	if len(strains) == 0 {
		return []float64{0}
	}

	sectionEnd := math.Ceil(strains[0].startTime/sectionLength) * sectionLength
	sectionPeak := 0.0
	previous := strains[0]
	peaks := make([]float64, 0)

	for _, strain := range strains {
		for strain.startTime > sectionEnd {
			peaks = append(peaks, sectionPeak)
			sectionPeak = previous.value * strainDecay(sectionEnd-previous.startTime)
			sectionEnd += sectionLength
		}

		sectionPeak = max(sectionPeak, strain.value)
		previous = strain
	}

	return append(peaks, sectionPeak)
}
