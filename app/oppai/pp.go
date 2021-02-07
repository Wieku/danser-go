package oppai

import (
	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/bmath"
	"math"
)

/* ------------------------------------------------------------- */
/* pp calc                                                       */

/* base pp value for stars, used internally by ppv2 */
func ppBase(stars float64) float64 {
	return math.Pow(5.0*math.Max(1.0, stars/0.0675)-4.0, 3.0) /
		100000.0
}

// PPv2Parameters : parameters to be passed to PPv2.
/* aim_stars, speed_stars, max_combo, nsliders, ncircles, nobjects,
* base_ar, base_od are required.
 */
//type PPv2Parameters struct {
//	Beatmap                      *Map
//	AimStars, SpeedStars         float64
//	MaxCombo                     int
//	NSliders, NCircles, NObjects int
//	BaseAR                       float32 // the base AR (before applying mods).
//	BaseOD                       float32 // the base OD (before applying mods).
//	Mode                         int     // gamemode
//	Mods                         int     // the mods bitmask, same as osu! api
//	Combo                        int     // Max combo achieved, if -1 it will default to maxCombo-nmiss
//	// if N300 = -1 it will default to nobjects - n100 - n50 - nmiss
//	N300, N100, N50, NMiss int
//	ScoreVer               int // scorev1 (1) or scorev2 (2).
//}

// PPv2 : structure to store ppv2 values
type PPv2 struct {
	Total, Aim, Speed, Acc float64
	ComputedAccuracy       Accuracy
}

func (pp *PPv2) PPv2x(aimStars, speedStars float64,
	maxCombo, nsliders, ncircles, nobjects,
	combo, n300, n100, n50, nmiss int, diff *difficulty.Difficulty,
	scoreVersion int) PPv2 {

	if maxCombo <= 0 {
		info("W: max_combo <= 0, changing to 1")
		maxCombo = 1
	}

	if combo < 0 {
		combo = maxCombo - nmiss
	}

	if n300 < 0 {
		n300 = nobjects - n100 - n50 - nmiss
	}

	/* accuracy -------------------------------------------- */
	var realAcc float64
	var accuracy float64

	pp.ComputedAccuracy = Accuracy{
		N300:    n300,
		N100:    n100,
		N50:     n50,
		NMisses: nmiss,
	}

	accuracy = pp.ComputedAccuracy.Value()
	realAcc = accuracy

	switch scoreVersion {
	case 1:
		/* scorev1 ignores sliders since they are free 300s
		and for some reason also ignores spinners */
		nspinners := nobjects - nsliders - ncircles

		realAcc = (&Accuracy{
			N300:    bmath.MaxI(n300-nsliders-nspinners, 0),
			N100:    n100,
			N50:     n50,
			NMisses: nmiss,
		}).Value()

		realAcc = math.Max(0.0, realAcc)
	case 2:
		ncircles = nobjects
	default:
		panic("unsupported score")
	}

	/* global values --------------------------------------- */
	nobjectsOver2k := float64(nobjects) / 2000.0
	lengthBonus := 0.95 + 0.4*math.Min(1.0, nobjectsOver2k)

	if nobjects > 2000 {
		lengthBonus += math.Log10(nobjectsOver2k) * 0.5
	}

	missPenalty := math.Pow(0.97, float64(nmiss))
	comboBreak := math.Pow(float64(combo), 0.8) /
		math.Pow(float64(maxCombo), 0.8)

	/* ar bonus -------------------------------------------- */
	arBonus := 1.0

	if diff.ARReal > 10.33 {
		arBonus += 0.3 * (diff.ARReal - 10.33)
	} else if diff.ARReal < 8.0 {
		arBonus += 0.01 * (8.0 - diff.ARReal)
	}

	/* aim pp ---------------------------------------------- */
	pp.Aim = ppBase(aimStars)
	pp.Aim *= lengthBonus
	pp.Aim *= missPenalty
	pp.Aim *= comboBreak
	pp.Aim *= arBonus

	hdBonus := 1.0
	if diff.Mods.Active(difficulty.Hidden) {
		hdBonus += 0.04 * (12.0 - diff.ARReal)
	}

	pp.Aim *= hdBonus

	if diff.Mods.Active(difficulty.Flashlight) {
		fl_bonus := 1.0 + 0.35*math.Min(1.0, float64(nobjects)/200.0)
		if nobjects > 200 {
			fl_bonus += 0.3 * math.Min(1, (float64(nobjects)-200.0)/300.0)
		}
		if nobjects > 500 {
			fl_bonus += (float64(nobjects) - 500.0) / 1200.0
		}
		pp.Aim *= fl_bonus
	}

	accBonus := 0.5 + accuracy/2.0
	od_squared := diff.ODReal * diff.ODReal
	odBonus := 0.98 + od_squared/2500.0

	pp.Aim *= accBonus
	pp.Aim *= odBonus

	/* speed pp -------------------------------------------- */
	pp.Speed = ppBase(speedStars)
	pp.Speed *= lengthBonus
	pp.Speed *= missPenalty
	pp.Speed *= comboBreak
	if diff.ARReal > 10.33 {
		pp.Speed *= arBonus
	}
	pp.Speed *= hdBonus

	pp.Speed *= 0.02 + accuracy
	pp.Speed *= 0.96 + float64(od_squared)/1600.0

	/* acc pp ---------------------------------------------- */
	pp.Acc = math.Pow(1.52163, diff.ODReal) *
		math.Pow(realAcc, 24.0) * 2.83

	pp.Acc *= math.Min(1.15, math.Pow(float64(ncircles)/1000.0, 0.3))

	if diff.Mods.Active(difficulty.Hidden) {
		pp.Acc *= 1.08
	}

	if diff.Mods.Active(difficulty.Flashlight) {
		pp.Acc *= 1.02
	}
	/* total pp -------------------------------------------- */
	finalMultiplier := 1.12

	if diff.Mods.Active(difficulty.NoFail) {
		finalMultiplier *= 0.90
	}

	if diff.Mods.Active(difficulty.SpunOut) {
		finalMultiplier *= 0.95
	}

	pp.Total = math.Pow(
		math.Pow(pp.Aim, 1.1)+math.Pow(pp.Speed, 1.1)+
			math.Pow(pp.Acc, 1.1),
		1.0/1.1) * finalMultiplier

	return *pp
}

// PPv2WithMods calculates the pp of the map with the mods passed and acc passed
//func (pp *PPv2) PPv2WithMods(aimStars, speedStars float64, b *Map, mods, n300, n100, n50, nmiss, combo int) {
//	pp.ppv2x(aimStars, speedStars, -1, b.NSliders, b.NCircles,
//		len(b.Objects), b.Mode, mods, combo, n300, n100, n50, nmiss,
//		b.AR, b.OD, 1, b, -1)
//}
//
//// PPv2ssWithMods calculates the pp of the map with the mods passed and 100% acc
//func (pp *PPv2) PPv2ssWithMods(aimStars, speedStars float64, b *Map, mods int) {
//	pp.ppv2x(aimStars, speedStars, -1, b.NSliders, b.NCircles,
//		len(b.Objects), b.Mode, mods, -1, -1, 0, 0, 0,
//		b.AR, b.OD, 1, b, -1)
//}
//
//// PPv2ss calculates the pp of the map with nomods and 100% acc
//func (pp *PPv2) PPv2ss(aimStars, speedStars float64, b *Map) {
//	pp.ppv2x(aimStars, speedStars, -1, b.NSliders, b.NCircles,
//		len(b.Objects), b.Mode, ModsNOMOD, -1, -1, 0, 0, 0,
//		b.AR, b.OD, 1, b, -1)
//}

// MultiAccPP returns the PP for every acc step
type MultiAccPP struct {
	P95   float64
	P98   float64
	P99   float64
	P99p5 float64
	P100  float64
}
