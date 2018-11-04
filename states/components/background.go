package components

import (
	"github.com/wieku/danser/render/texture"
	"github.com/wieku/danser/storyboard"
	"github.com/wieku/danser/beatmap"
	"github.com/wieku/danser/utils"
	"github.com/wieku/danser/render"
	"path/filepath"
	"github.com/wieku/danser/settings"
	"log"
	"github.com/wieku/danser/bmath"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/wieku/danser/render/batches"
	"github.com/wieku/danser/render/effects"
)

type Background struct {
	blur       *effects.BlurEffect
	scale      bmath.Vector2d
	background *texture.TextureRegion
	storyboard *storyboard.Storyboard
}

func NewBackground(beatMap *beatmap.BeatMap) *Background {
	bg := new(Background)
	bg.blur = effects.NewBlurEffect(int(settings.Graphics.GetWidth()), int(settings.Graphics.GetHeight()))

	var err error
	bg.background, err = utils.LoadTextureToAtlas(render.Atlas, filepath.Join(settings.General.OsuSongsDir, beatMap.Dir, beatMap.Bg))
	if err != nil {
		log.Println(err)
	}

	if settings.Playfield.StoryboardEnabled {
		bg.storyboard = storyboard.NewStoryboard(beatMap)

		if bg.storyboard == nil {
			log.Println("Storyboard not found!")
		}
	}

	winscl := settings.Graphics.GetAspectRatio()

	if bg.background != nil {
		imScl := float64(bg.background.Width) / float64(bg.background.Height)

		condition := imScl < winscl
		if bg.storyboard != nil && !bg.storyboard.IsWideScreen() {
			condition = !condition
		}

		if condition {
			bg.scale = bmath.NewVec2d(1, winscl/imScl)
		} else {
			bg.scale = bmath.NewVec2d(imScl/winscl, 1)
		}
	}

	return bg
}

func (bg *Background) Update(time int64) {
	if bg.storyboard != nil {
		bg.storyboard.Update(time)
	}
}

func (bg *Background) Draw(time int64, batch *batches.SpriteBatch, blurVal, bgAlpha float64, camera mgl32.Mat4) {
	batch.Begin()

	batch.SetColor(1, 1, 1, 1)
	batch.ResetTransform()
	batch.SetAdditive(false)

	if bg.background != nil || bg.storyboard != nil {
		if settings.Playfield.BlurEnable {
			bg.blur.SetBlur(blurVal, blurVal)
			bg.blur.Begin()
		}

		if bg.background != nil && (bg.storyboard == nil || !bg.storyboard.BGFileUsed()) {
			batch.SetCamera(mgl32.Ortho(-1, 1, -1, 1, 1, -1))
			batch.SetScale(bg.scale.X, -bg.scale.Y)
			if !settings.Playfield.BlurEnable {
				batch.SetColor(1, 1, 1, bgAlpha)
			}
			batch.DrawUnit(*bg.background)
		}

		if bg.storyboard != nil {
			batch.SetScale(1, 1)
			if !settings.Playfield.BlurEnable {
				batch.SetColor(bgAlpha, bgAlpha, bgAlpha, 1)
			}
			batch.SetCamera(camera)
			bg.storyboard.Draw(time, batch)
			batch.Flush()
		}

		if settings.Playfield.BlurEnable {
			batch.End()

			texture := bg.blur.EndAndProcess()
			batch.Begin()
			batch.SetColor(1, 1, 1, bgAlpha)
			batch.SetCamera(mgl32.Ortho(-1, 1, -1, 1, 1, -1))
			batch.DrawUnscaled(texture.GetRegion())
		}

	}

	batch.End()
}

func (bg *Background) GetStoryboard() *storyboard.Storyboard {
	return bg.storyboard
}