package statistic

var counters = make(map[StatisticType]int64)
var pastCounters = make(map[StatisticType]int64)

func Reset() {
	pastCounters = counters
	counters = make(map[StatisticType]int64)
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
