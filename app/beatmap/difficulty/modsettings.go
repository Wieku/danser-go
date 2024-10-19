package difficulty

import "reflect"

var modConfigs map[Modifier]reflect.Type

func init() {
	modConfigs = map[Modifier]reflect.Type{
		HalfTime:         rfType[SpeedSettings](),
		Daycore:          rfType[SpeedSettings](),
		DoubleTime:       rfType[SpeedSettings](),
		Nightcore:        rfType[SpeedSettings](),
		Easy:             rfType[EasySettings](),
		Classic:          rfType[ClassicSettings](),
		Flashlight:       rfType[FlashlightSettings](),
		DifficultyAdjust: rfType[DiffAdjustSettings](),
	}
}

type SpeedSettings struct {
	SpeedChange float64 `json:"speed_change"`
	AdjustPitch bool    `json:"adjust_pitch"`
}

func NewSpeedSettings(rate float64, adjustPitch bool) SpeedSettings {
	return SpeedSettings{
		SpeedChange: rate,
		AdjustPitch: adjustPitch,
	}
}

type ClassicSettings struct {
	NoSliderHeadAccuracy bool `json:"no_slider_head_accuracy"`
	ClassicNoteLock      bool `json:"classic_note_lock"`
	AlwaysPlayTailSample bool `json:"always_play_tail_sample"`
	FadeHitCircleEarly   bool `json:"fade_hit_circle_early"`
	ClassicHealth        bool `json:"classic_health"`
}

func NewClassicSettings() ClassicSettings {
	return ClassicSettings{
		NoSliderHeadAccuracy: true,
		ClassicNoteLock:      true,
		AlwaysPlayTailSample: true,
		FadeHitCircleEarly:   true,
		ClassicHealth:        true,
	}
}

type EasySettings struct {
	Retries int `json:"retries"`
}

func NewEasySettings() EasySettings {
	return EasySettings{
		Retries: 2,
	}
}

type FlashlightSettings struct {
	FollowDelay    float64 `json:"follow_delay"`
	SizeMultiplier float64 `json:"size_multiplier"`
	ComboBasedSize bool    `json:"combo_based_size"`
}

func NewFlashlightSettings() FlashlightSettings {
	return FlashlightSettings{
		FollowDelay:    120,
		SizeMultiplier: 1,
		ComboBasedSize: true,
	}
}

type DiffAdjustSettings struct {
	ApproachRate      float64 `json:"approach_rate"`
	CircleSize        float64 `json:"circle_size"`
	DrainRate         float64 `json:"drain_rate"`
	OverallDifficulty float64 `json:"overall_difficulty"`
	ExtendedValues    bool    `json:"-"`
}

func NewDiffAdjustSettings(ar, cs, hp, od float64) DiffAdjustSettings {
	return DiffAdjustSettings{
		ApproachRate:      ar,
		CircleSize:        cs,
		DrainRate:         hp,
		OverallDifficulty: od,
	}
}
