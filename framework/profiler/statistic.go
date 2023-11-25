package profiler

var counters = make([]int64, size)
var pastCounters = make([]int64, size)

func ResetStats() {
	pastCounters, counters = counters, pastCounters

	for i := StatisticType(0); i < size; i++ {
		counters[i] = 0
	}
}

func IncrementStat(typ StatisticType) {
	counters[typ]++
}

func AddStat(typ StatisticType, amount int64) {
	counters[typ] += amount
}

func GetStat(typ StatisticType) int64 {
	return counters[typ]
}

func GetPreviousStat(typ StatisticType) int64 {
	return pastCounters[typ]
}
