package oppai

import (
	"fmt"
	"math"
	"sort"
)

// Warnings ...
const Warnings = true

func info(s string) {
	if Warnings {
		fmt.Println(s)
	}
}

func roundOppai(x float64) int { return int(math.Floor((x) + 0.5)) }

func reverseSortFloat64s(x []float64) {
	sort.Float64s(x)
	n := len(x)
	for i := 0; i < n/2; i++ {
		j := n - i - 1
		x[i], x[j] = x[j], x[i]
	}
}
