package components

import (
	"github.com/wieku/danser-go/app/bmath"
	"github.com/wieku/danser-go/app/bmath/difficulty"
	"github.com/wieku/danser-go/app/rulesets/osu"
	"github.com/wieku/danser-go/app/skin"
	"github.com/wieku/danser-go/framework/graphics/sprite"
	"github.com/wieku/danser-go/framework/math/animation"
	"github.com/wieku/danser-go/framework/math/animation/easing"
	"github.com/wieku/danser-go/framework/math/vector"
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

	switch result {
	case osu.HitResults.Hit100:
		tex = "hit100"
	case osu.HitResults.Hit50:
		tex = "hit50"
	case osu.HitResults.Miss:
		tex = "hit0"
	}

	if tex == "" {
		return
	}

	frames := skin.GetFrames(tex, true)

	sprite := sprite.NewAnimation(frames, skin.GetInfo().GetFrameTime(len(frames)), false, -float64(time), position, bmath.Origin.Centre)

	fadeIn := float64(time + difficulty.ResultFadeIn)
	postEmpt := float64(time + difficulty.PostEmpt)
	fadeOut := postEmpt + float64(difficulty.ResultFadeOut)

	sprite.AddTransformUnordered(animation.NewSingleTransform(animation.Fade, easing.Linear, float64(time), fadeIn, 0.0, 1.0))
	sprite.AddTransformUnordered(animation.NewSingleTransform(animation.Fade, easing.Linear, postEmpt, fadeOut, 1.0, 0.0))

	if len(frames) == 1 {
		sprite.AddTransformUnordered(animation.NewSingleTransform(animation.Scale, easing.Linear, float64(time), float64(time+difficulty.ResultFadeIn*0.8), 0.6, 1.1))
		sprite.AddTransformUnordered(animation.NewSingleTransform(animation.Scale, easing.Linear, fadeIn, float64(time+difficulty.ResultFadeIn*1.2), 1.1, 0.9))
		sprite.AddTransformUnordered(animation.NewSingleTransform(animation.Scale, easing.Linear, float64(time+difficulty.ResultFadeIn*1.2), float64(time+difficulty.ResultFadeIn*1.4), 0.9, 1.0))

		if result == osu.HitResults.Miss {
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

func (results *HitResults) Draw(batch *sprite.SpriteBatch, _ float64) {
	batch.ResetTransform()

	scale := results.diff.CircleRadius / 64
	batch.SetScale(scale, scale)

	results.manager.Draw(int64(results.lastTime), batch)

	batch.ResetTransform()
}
