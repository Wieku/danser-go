package osu

import (
	"testing"

	"github.com/wieku/danser-go/app/beatmap/difficulty"
)

func TestLazerHitWindowIntegerBoundaries(t *testing.T) {
	diff := difficulty.NewDifficulty(5, 5, 10, 5)
	diff.SetMods(difficulty.Lazer)

	player := &difficultyPlayer{diff: diff}
	ruleset := &OsuRuleSet{}

	tests := []struct {
		delta float64
		want  HitResult
	}{
		{19, Hit300},
		{20, Hit100},
		{59, Hit100},
		{60, Hit50},
		{99, Hit50},
		{100, Miss},
	}

	for _, test := range tests {
		if got := ruleset.GetResultForDelta(player, test.delta); got != test.want {
			t.Errorf("delta %.0fms: got %v, want %v", test.delta, got, test.want)
		}
	}
}

func TestAccuracyDisplayMatchesOsuScorePage(t *testing.T) {
	score := &Score{Accuracy: 0.9928526757795051}
	if got := score.AccuracyDisplay(); got != 99.28 {
		t.Fatalf("got %.2f, want 99.28", got)
	}
}
