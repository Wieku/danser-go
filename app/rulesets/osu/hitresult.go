package osu

import "github.com/wieku/danser-go/framework/math/vector"

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
	SliderEnd
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
	SliderHits  = SliderStart | SliderPoint | SliderRepeat | SliderEnd
	SpinnerHits = SpinnerSpin | SpinnerPoints | SpinnerBonus
	RawHits     = SliderHits | SpinnerHits
)

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

func (r HitResult) ScoreValueLazer() int64 {
	v := r & (^(Additions | SliderStart))
	switch v {
	case Hit50:
		return 50
	case Hit100:
		return 100
	case SliderEnd:
		return 150
	case Hit300:
		return 300
	case SliderRepeat:
		return 30
	case SliderPoint, SpinnerPoints:
		return 10
	case SpinnerBonus:
		return 50
	}

	return 0
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
}

func createJudgementResult(result HitResult, maxResult HitResult, comboResult ComboResult, time int64, position vector.Vector2f, obj HitObject) JudgementResult {
	return JudgementResult{
		HitResult:   result,
		MaxResult:   maxResult,
		ComboResult: comboResult,
		Time:        time,
		Position:    position,
		Number:      obj.GetNumber(),
		object:      obj,
	}
}
