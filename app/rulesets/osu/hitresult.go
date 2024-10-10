package osu

import (
	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/framework/math/vector"
)

type HitResult int64

const (
	Ignore     = HitResult(0)
	SliderMiss = HitResult(1 << iota)
	Miss
	Hit50
	Hit100
	Hit300
	SliderStart
	SliderPoint
	SliderRepeat
	LegacySliderEnd
	SliderEnd
	SliderFinish // For lazer health processor
	SpinnerSpin
	SpinnerPoints
	SpinnerBonus
	MuAddition
	KatuAddition
	GekiAddition
	PositionalMiss
	Additions   = MuAddition | KatuAddition | GekiAddition
	Hit50m      = Hit50 | MuAddition
	Hit100m     = Hit100 | MuAddition
	Hit300m     = Hit300 | MuAddition
	Hit100k     = Hit100 | KatuAddition
	Hit300k     = Hit300 | KatuAddition
	Hit300g     = Hit300 | GekiAddition
	BaseHits    = Hit50 | Hit100 | Hit300
	BaseHitsM   = BaseHits | Miss
	HitValues   = Hit50 | Hit100 | Hit300 | GekiAddition | KatuAddition
	SliderHits  = SliderStart | SliderPoint | SliderRepeat | LegacySliderEnd | SliderEnd
	SpinnerHits = SpinnerSpin | SpinnerPoints | SpinnerBonus
	RawHits     = SliderHits | SpinnerHits
)

func (r HitResult) IsBonus() bool {
	v := r & (^Additions)

	return v&(SpinnerPoints|SpinnerBonus) != 0
}

func (r HitResult) AffectsAccV1() bool {
	v := r & (^Additions)

	return v&(BaseHitsM) != 0
}

func (r HitResult) AffectsAccLZ() bool {
	v := r & (^Additions)

	return v&(BaseHitsM|SliderHits|SliderMiss) != 0
}

func (r HitResult) ScoreValue() int64 {
	v := r & (^Additions)
	switch v {
	case Hit50:
		return 50
	case Hit100, SpinnerPoints:
		return 100
	case Hit300:
		return 300
	case SliderStart, SliderRepeat, SliderEnd:
		return 30
	case SliderPoint:
		return 10
	case SpinnerBonus:
		return 1100
	}

	return 0
}

func (r HitResult) ScoreValueV2() int64 {
	if r&SpinnerBonus > 0 {
		return 500
	}

	return r.ScoreValue()
}

func (r HitResult) ScoreValueLazer() int64 {
	v := r & (^Additions)
	switch v {
	case Hit50:
		return 50
	case Hit100:
		return 100
	case SliderEnd:
		return 150
	case Hit300:
		return 300
	case SliderStart, SliderPoint, SliderRepeat:
		return 30
	case SpinnerPoints, LegacySliderEnd:
		return 10
	case SpinnerBonus:
		return 50
	}

	return 0
}

func (r HitResult) ScoreValueMod(mod difficulty.Modifier) int64 {
	if mod.Active(difficulty.Lazer) {
		return r.ScoreValueLazer()
	} else if mod.Active(difficulty.ScoreV2) {
		return r.ScoreValueV2()
	}

	return r.ScoreValue()
}

type ComboResult uint8

const (
	Reset = ComboResult(iota)
	Hold
	Increase
)

type JudgementResult struct {
	HitResult HitResult
	MaxResult HitResult

	ComboResult ComboResult

	Time     int64
	Position vector.Vector2f

	Number int64
	object HitObject

	fromSliderFinish bool
}

func createJudgementResult(result HitResult, maxResult HitResult, comboResult ComboResult, time int64, position vector.Vector2f, obj HitObject) JudgementResult {
	nm := int64(-1)
	if obj != nil {
		nm = obj.GetNumber()
	}

	return JudgementResult{
		HitResult:   result,
		MaxResult:   maxResult,
		ComboResult: comboResult,
		Time:        time,
		Position:    position,
		Number:      nm,
		object:      obj,
	}
}

func createJudgementResultF(result HitResult, maxResult HitResult, comboResult ComboResult, time int64, position vector.Vector2f, obj HitObject, sliderFinish bool) JudgementResult {
	jResult := createJudgementResult(result, maxResult, comboResult, time, position, obj)

	jResult.fromSliderFinish = sliderFinish

	return jResult
}
