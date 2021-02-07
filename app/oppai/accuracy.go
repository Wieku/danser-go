package oppai

import (
	"math"
)

/* ------------------------------------------------------------- */
/* acc calc                                                      */

// Accuracy ...
type Accuracy struct {
	// if N300 = -1 it will be calculated from the object count
	N300, N100, N50, NMisses int
}

/**
 * @param nobjects the total number of hits (n300 + n100 + n50 +
 *                 nmiss). if -1, n300 must have been set and
 *                 will be used to deduce this value.
 * @return the accuracy value (0.0-1.0)
 */
func (acc *Accuracy) value(nobjects int) float64 {
	if nobjects < 0 && acc.N300 < 0 {
		//panic("either nobjects or n300 must be specified")
	}

	n300x := acc.N300
	if acc.N300 < 0 {
		n300x = nobjects - acc.N100 - acc.N50 - acc.NMisses
	}

	if nobjects < 0 {
		nobjects = n300x + acc.N100 + acc.N50 + acc.NMisses
	}
	if nobjects == 0 {
		return 0
	}

	res := (float64(acc.N50)*50.0 +
		float64(acc.N100)*100.0 +
		float64(n300x)*300.0) /
		(float64(nobjects) * 300.0)

	return math.Max(0.0, math.Min(res, 1.0))
}

func (acc *Accuracy) Value() float64 {
	return acc.value(-1)
}
