package osu

import (
	"github.com/wieku/danser-go/app/beatmap"
	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/rulesets/osu/performance/pp220930"
)

type scoreProcessor interface {
	Init(beatMap *beatmap.BeatMap, player *difficultyPlayer)
	AddResult(result JudgementResult)
	ModifyResult(result HitResult, src HitObject) HitResult
	GetScore() int64
	GetCombo() int64
	GetAccuracy() float64
}

type Score struct {
	Score        int64
	Accuracy     float64
	Grade        Grade
	Combo        uint
	PerfectCombo bool
	Count300     uint
	CountGeki    uint
	Count100     uint
	CountKatu    uint
	Count50      uint
	CountMiss    uint
	CountSB      uint
	PP           pp220930.PPv2Results

	scoredObjects uint
}

func (s *Score) AddResult(result JudgementResult) {
	bResult := result.HitResult & BaseHitsM

	if bResult > 0 {
		switch bResult {
		case Hit300:
			s.Count300++
		case Hit100:
			s.Count100++
		case Hit50:
			s.Count50++
		case Miss:
			s.CountMiss++
		}

		s.scoredObjects++
	}

	if result.ComboResult == Reset && result.HitResult != Miss { // skips missed slider "ends" as they don't reset combo
		s.CountSB++
	}
}

func (s *Score) CalculateGrade(mods difficulty.Modifier) {
	baseGrade := s.gradeV1()

	if mods&(difficulty.Hidden|difficulty.Flashlight) > 0 {
		switch baseGrade {
		case S:
			baseGrade = SH
		case SS:
			baseGrade = SSH
		}
	}

	s.Grade = baseGrade
}

func (s *Score) gradeV1() Grade {
	ratio := 1.0
	if s.scoredObjects > 0 {
		ratio = float64(s.Count300) / float64(s.scoredObjects)
	}

	if s.Count300 == s.scoredObjects {
		return SS
	} else if ratio > 0.9 && float64(s.Count50)/float64(s.scoredObjects) < 0.01 && s.CountMiss == 0 {
		return S
	} else if ratio > 0.8 && s.CountMiss == 0 || ratio > 0.9 {
		return A
	} else if ratio > 0.7 && s.CountMiss == 0 || ratio > 0.8 {
		return B
	} else if ratio > 0.6 {
		return C
	}

	return D
}
