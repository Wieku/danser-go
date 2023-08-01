package statistic

var counters = make([]int64, size)
var pastCounters = make([]int64, size)

func Reset() {
	pastCounters, counters = counters, pastCounters

	for i := StatisticType(0); i < size; i++ {
		counters[i] = 0
	}
}

func Increment(typ StatisticType) {
	counters[typ]++
}

func Add(typ StatisticType, amount int64) {
	counters[typ] += amount
}

func Get(typ StatisticType) int64 {
	return counters[typ]
}

func GetPrevious(typ StatisticType) int64 {
	return pastCounters[typ]
}
