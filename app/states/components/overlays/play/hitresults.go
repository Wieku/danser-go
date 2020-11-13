package play

import (
	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/bmath"
	"github.com/wieku/danser-go/app/rulesets/osu"
	"github.com/wieku/danser-go/app/skin"
	"github.com/wieku/danser-go/framework/graphics/batch"
	"github.com/wieku/danser-go/framework/graphics/sprite"
	"github.com/wieku/danser-go/framework/math/animation"
	"github.com/wieku/danser-go/framework/math/animation/easing"
	"github.com/wieku/danser-go/framework/math/vector"
	"math"
	"math/rand"
)

type HitResults struct {
	manager  *sprite.SpriteManager
	lastTime float64
	diff     *difficulty.Difficulty
}

func NewHitResults(diff *difficulty.Difficulty) *HitResults {
	// Preload all frames to avoid stalling during gameplay
	skin.GetFrames("hit0", true)
	skin.GetFrames("hit50", true)
	skin.GetFrames("hit100", true)
	skin.GetFrames("hit100k", true)
	skin.GetFrames("hit300", true)
	skin.GetFrames("hit300k", true)
	skin.GetFrames("hit300g", true)

	return &HitResults{manager: sprite.NewSpriteManager(), diff: diff}
}

func (results *HitResults) AddResult(time int64, result osu.HitResult, position vector.Vector2d) {
	var tex string
	var particle string

	switch result & osu.BaseHitsM {
	case osu.Hit300:
		tex = "hit300"
		particle = "particle300"
	case osu.Hit100:
		tex = "hit100"
		particle = "particle100"
	case osu.Hit50:
		tex = "hit50"
		particle = "particle50"
	case osu.Miss:
		tex = "hit0"
	}

	switch result & osu.Additions {
	case osu.KatuAddition:
		tex += "k"
	case osu.GekiAddition:
		tex += "g"
	}

	if tex == "" {
		return
	}

	frames := skin.GetFrames(tex, true)

	particles := false

	if particle != "" && len(frames) > 0 {
		particleTex := skin.GetTextureSource(particle, skin.GetSourceFromTexture(frames[0]))

		if particleTex != nil {
			particles = true

			for i := 0; i < 150; i++ {
				fadeOut := 500 + 700*rand.Float64()
				direction := vector.NewVec2dRad(rand.Float64()*2*math.Pi, rand.Float64()*35)

				sp := sprite.NewSpriteSingle(particleTex, float64(time), position, bmath.Origin.Centre)
				sp.SetAdditive(true)
				sp.AddTransform(animation.NewSingleTransform(animation.Fade, easing.OutQuad, float64(time), float64(time)+fadeOut, 1.0, 0.0))
				sp.AddTransform(animation.NewVectorTransformV(animation.Move, easing.OutQuad, float64(time), float64(time)+fadeOut, position, position.Add(direction)))
				sp.ResetValuesToTransforms()
				sp.AdjustTimesToTransformations()
				sp.ShowForever(false)

				results.manager.Add(sp)
			}
		}
	}

	sprite := sprite.NewAnimation(frames, skin.GetInfo().GetFrameTime(len(frames)), false, float64(time)+1, position, bmath.Origin.Centre)

	fadeIn := float64(time + difficulty.ResultFadeIn)
	if particles {
		fadeIn = float64(time + 80)
	}

	postEmpt := float64(time + difficulty.PostEmpt)
	fadeOut := postEmpt + float64(difficulty.ResultFadeOut)

	sprite.AddTransformUnordered(animation.NewSingleTransform(animation.Fade, easing.Linear, float64(time), fadeIn, 0.0, 1.0))
	sprite.AddTransformUnordered(animation.NewSingleTransform(animation.Fade, easing.Linear, postEmpt, fadeOut, 1.0, 0.0))

	if len(frames) == 1 {
		if particles {
			sprite.AddTransformUnordered(animation.NewSingleTransform(animation.Scale, easing.Linear, float64(time), fadeOut, 0.9, 1.05))
		} else {
			sprite.AddTransformUnordered(animation.NewSingleTransform(animation.Scale, easing.Linear, float64(time), float64(time+difficulty.ResultFadeIn*0.8), 0.6, 1.1))
			sprite.AddTransformUnordered(animation.NewSingleTransform(animation.Scale, easing.Linear, fadeIn, float64(time+difficulty.ResultFadeIn*1.2), 1.1, 0.9))
			sprite.AddTransformUnordered(animation.NewSingleTransform(animation.Scale, easing.Linear, float64(time+difficulty.ResultFadeIn*1.2), float64(time+difficulty.ResultFadeIn*1.4), 0.9, 1.0))
		}

		if result == osu.Miss {
			rotation := rand.Float64()*0.3 - 0.15

			sprite.AddTransformUnordered(animation.NewSingleTransform(animation.Rotate, easing.Linear, float64(time), fadeIn, 0.0, rotation))
			sprite.AddTransformUnordered(animation.NewSingleTransform(animation.Rotate, easing.Linear, fadeIn, fadeOut, rotation, rotation*2))

			sprite.AddTransformUnordered(animation.NewSingleTransform(animation.MoveY, easing.Linear, float64(time), fadeOut, position.Y-5, position.Y+40))
		}
	}

	sprite.SortTransformations()
	sprite.AdjustTimesToTransformations()
	sprite.ResetValuesToTransforms()

	results.manager.Add(sprite)
}

func (results *HitResults) Update(time float64) {
	results.manager.Update(int64(time))
	results.lastTime = time
}

func (results *HitResults) Draw(batch *batch.QuadBatch, _ float64) {
	batch.ResetTransform()

	scale := results.diff.CircleRadius / 64
	batch.SetScale(scale, scale)

	results.manager.Draw(int64(results.lastTime), batch)

	batch.ResetTransform()
}
