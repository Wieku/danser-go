package components

import (
	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/wieku/danser-go/beatmap"
	"github.com/wieku/danser-go/bmath"
	"github.com/wieku/danser-go/input"
	"github.com/wieku/danser-go/render"
	"github.com/wieku/danser-go/render/batches"
	"github.com/wieku/danser-go/render/effects"
	"github.com/wieku/danser-go/render/texture"
	"github.com/wieku/danser-go/settings"
	"github.com/wieku/danser-go/storyboard"
	"github.com/wieku/danser-go/utils"
	"log"
	"path/filepath"
)

type Background struct {
	blur       *effects.BlurEffect
	scale      bmath.Vector2d
	position      bmath.Vector2d
	parallax float64
	background *texture.TextureRegion
	storyboard *storyboard.Storyboard
	lastTime int64
}

func NewBackground(beatMap *beatmap.BeatMap, parallax float64, useStoryboard bool) *Background {
	bg := new(Background)
	bg.blur = effects.NewBlurEffect(int(settings.Graphics.GetWidth()), int(settings.Graphics.GetHeight()))

	var err error
	bg.background, err = utils.LoadTextureToAtlas(render.Atlas, filepath.Join(settings.General.OsuSongsDir, beatMap.Dir, beatMap.Bg))
	if err != nil {
		log.Println(err)
	}

	if settings.Playfield.StoryboardEnabled && useStoryboard {
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
			bg.scale = bmath.NewVec2d(1, winscl/imScl).Scl(1+parallax)
		} else {
			bg.scale = bmath.NewVec2d(imScl/winscl, 1).Scl(1+parallax)
		}
	}
	bg.parallax = parallax
	return bg
}

func (bg *Background) Update(time int64) {
	if bg.storyboard != nil {
		bg.storyboard.Update(time)
	}
	x, y := input.Win.GetCursorPos()
	x = -(x-settings.Graphics.GetWidthF()/2)/settings.Graphics.GetWidthF()*bg.parallax
	y = (y-settings.Graphics.GetHeightF()/2)/settings.Graphics.GetHeightF()*bg.parallax

	delta := float64(time - bg.lastTime)
	bg.position.X = bg.position.X + delta*0.005*(x-bg.position.X)
	bg.position.Y = bg.position.Y + delta*0.005*(y-bg.position.Y)

	bg.lastTime = time
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
				batch.SetTranslation(bg.position)
			}
			batch.DrawUnit(*bg.background)
		}

		if bg.storyboard != nil {
			batch.SetScale(1, 1)
			if !settings.Playfield.BlurEnable {
				batch.SetColor(bgAlpha, bgAlpha, bgAlpha, 1)
				batch.SetTranslation(bmath.NewVec2d(320*bg.position.X, 240*bg.position.Y))
				batch.SetScale(1+bg.parallax, 1+bg.parallax)
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
			batch.SetTranslation(bg.position)
			batch.SetScale(1+bg.parallax, 1+bg.parallax)
			batch.DrawUnit(texture.GetRegion())
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