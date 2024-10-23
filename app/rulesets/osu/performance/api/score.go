package api

type PerfScore struct {
	Accuracy     float64
	MaxCombo     int
	CountGreat   int
	CountOk      int
	CountMeh     int
	CountMiss    int
	SliderBreaks int
	SliderEnd    int
}
