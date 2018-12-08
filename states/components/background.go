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
	"github.com/go-gl/gl/v3.3-core/gl"
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

func project(pos bmath.Vector2d, camera mgl32.Mat4) bmath.Vector2d {
	res := camera.Mul4x1(mgl32.Vec4{pos.X32(), pos.Y32(), 0.0, 1.0})
	return bmath.NewVec2d((float64(res[0])/2+0.5)*settings.Graphics.GetWidthF(), float64((res[1])/2+0.5)*settings.Graphics.GetWidthF())
}

func (bg *Background) Draw(time int64, batch *batches.SpriteBatch, blurVal, bgAlpha float64, camera mgl32.Mat4) {
	batch.Begin()

	batch.SetColor(1, 1, 1, 1)
	batch.ResetTransform()
	batch.SetAdditive(false)

	if bg.background != nil || bg.storyboard != nil {
		if bg.storyboard != nil && !bg.storyboard.IsWideScreen() {
			v1 := project(bmath.NewVec2d(256-320, 192+240), camera)
			v2 := project(bmath.NewVec2d(256+320, 192-240), camera)
			gl.Enable(gl.SCISSOR_TEST)
			gl.Scissor(int32(v1.X32()), int32(v1.Y32()), int32(v2.X32()-v1.X32()), int32(v2.Y32()-v1.Y32()))
		}
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
		if bg.storyboard != nil && !bg.storyboard.IsWideScreen() {
			gl.Disable(gl.SCISSOR_TEST)
		}
	}

	batch.End()
}

func (bg *Background) GetStoryboard() *storyboard.Storyboard {
	return bg.storyboard
}