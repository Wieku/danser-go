package difficulty

import (
	"math"
	"testing"
)

func TestLazerScoreMultipliers(t *testing.T) {
	tests := []struct {
		name string
		mods Modifier
		want float64
	}{
		{"HardRock", HardRock, 1.09},
		{"HiddenHardRock", Hidden | HardRock, 1.04 * 1.09},
		{"DoubleTime", DoubleTime, 1.23},
		{"HalfTime", HalfTime, 0.55},
		{"Flashlight", Flashlight, 1.2},
		{"Classic", Classic, 0.985},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			diff := NewDifficulty(5, 5, 5, 5)
			diff.SetMods(Lazer | test.mods)

			if got := diff.GetScoreMultiplier(); math.Abs(got-test.want) > 1e-12 {
				t.Fatalf("got %.12f, want %.12f", got, test.want)
			}
		})
	}
}
