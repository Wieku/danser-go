package score

func CalculateAccuracy(hits []int64) float64 {
	sum := int64(0)
	for _, value := range hits{
		sum += value
	}
	return 100 * float64(sum) / float64(300 * len(hits))
}

func CalculateRank(hits []int64) Rank {
	countall := len(hits)
	count300 := 0
	count100 := 0
	count50 := 0
	countmiss :=0
	for _, value := range hits{
		switch value {
		case 300:
			count300 += 1
			break
		case 100:
			count100 += 1
			break
		case 50:
			count50 += 1
			break
		case 0:
			countmiss += 1
			break
		}
	}
	if count300 == countall {
		// SS
		return SS
	}else if ((float64(count300) / float64(countall)) > 0.9) && ((float64(count50) / float64(countall)) < 0.01) && (countmiss == 0) {
		// S
		return S
	}else if ((float64(count300) / float64(countall)) > 0.9) || (((float64(count300) / float64(countall)) > 0.8) && (countmiss == 0)) {
		// A
		return A
	}else if ((float64(count300) / float64(countall)) > 0.8) || (((float64(count300) / float64(countall)) > 0.7) && (countmiss == 0)) {
		// B
		return B
	}else if ((float64(count300) / float64(countall)) > 0.6) {
		// C
		return C
	}else {
		// D
		return D
	}
}