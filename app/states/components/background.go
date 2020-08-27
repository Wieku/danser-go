package components

import (
	"github.com/EdlinOrg/prominentcolor"
	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/wieku/danser-go/app/animation"
	"github.com/wieku/danser-go/app/animation/easing"
	"github.com/wieku/danser-go/app/audio"
	"github.com/wieku/danser-go/app/beatmap"
	"github.com/wieku/danser-go/app/bmath"
	"github.com/wieku/danser-go/app/render"
	"github.com/wieku/danser-go/app/render/batches"
	"github.com/wieku/danser-go/app/render/effects"
	"github.com/wieku/danser-go/app/render/gui/drawables"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/app/storyboard"
	"github.com/wieku/danser-go/app/utils"
	"github.com/wieku/danser-go/framework/graphics/texture"
	"log"
	"math"
	"math/rand"
	"path/filepath"
)

type Background struct {
	blur       *effects.BlurEffect
	scale      bmath.Vector2d
	position   bmath.Vector2d
	parallax   float64
	background *texture.TextureRegion
	storyboard *storyboard.Storyboard
	lastTime   int64
	bMap       *beatmap.BeatMap
	pulse      *animation.Glider
	triangles  *drawables.Triangles

	lastBeatLength float64
	lastBeatStart  float64
	lastBeatProg   int64

	rotation   *animation.Glider
	deltaSum   float64
	lastRandom float64
}

func NewBackground(beatMap *beatmap.BeatMap, parallax float64, useStoryboard bool) *Background {
	bg := new(Background)
	bg.bMap = beatMap
	bg.pulse = animation.NewGlider(1.0)
	bg.rotation = animation.NewGlider(0.0)
	bg.pulse.SetEasing(easing.OutQuad)
	bg.rotation.SetEasing(easing.InOutSine)
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
			bg.scale = bmath.NewVec2d(1, winscl/imScl).Scl(1 /*+parallax*/)
		} else {
			bg.scale = bmath.NewVec2d(imScl/winscl, 1).Scl(1 /*+parallax*/)
		}
	}
	bg.parallax = parallax

	imag, _ := utils.LoadImage(filepath.Join(settings.General.OsuSongsDir, beatMap.Dir, beatMap.Bg))

	cItems, _ := prominentcolor.KmeansWithAll(5, imag, prominentcolor.ArgumentDefault, prominentcolor.DefaultSize, prominentcolor.GetDefaultMasks())
	newCol := make([]bmath.Color, len(cItems))

	for i := 0; i < len(cItems); i++ {
		newCol[i] = bmath.Color{float64(cItems[i].Color.R) / 255, float64(cItems[i].Color.G) / 255, float64(cItems[i].Color.B) / 255, 1}
	}
	bg.triangles = drawables.NewTriangles(newCol)
	return bg
}

func (bg *Background) SetTrack(track *audio.Music) {
	bg.triangles.SetTrack(track)
}

func (bg *Background) Update(time int64, x, y float64) {
	if bg.lastTime == 0 {
		bg.lastTime = time
	}
	if bg.storyboard != nil {
		bg.storyboard.Update(time)
	}
	//x, y := input.Win.GetCursorPos()
	x = -(x - settings.Graphics.GetWidthF()/2) / settings.Graphics.GetWidthF() * bg.parallax
	y = (y - settings.Graphics.GetHeightF()/2) / settings.Graphics.GetHeightF() * bg.parallax

	delta := float64(time - bg.lastTime)
	bg.position.X = bg.position.X + delta*0.000625*(x-bg.position.X)
	bg.position.Y = bg.position.Y + delta*0.000625*(y-bg.position.Y)

	bg.lastTime = time

	bg.deltaSum += delta

	if bg.deltaSum >= 0 {
		randomV := rand.Float64()

		bg.rotation.AddEvent(float64(time), float64(time)+math.Abs(randomV-bg.lastRandom)*4000, (randomV-0.5)*0.08)
		bg.deltaSum -= math.Abs(randomV-bg.lastRandom) * 4000
		bg.lastRandom = randomV
	}

	bTime := bg.bMap.Timings.Current.BaseBpm

	if bTime != bg.lastBeatLength {
		bg.lastBeatLength = bTime
		bg.lastBeatStart = float64(bg.bMap.Timings.Current.Time)
		bg.lastBeatProg = -1
	}

	if int64((float64(time)-bg.lastBeatStart)/bg.lastBeatLength) > bg.lastBeatProg {

		if bg.bMap.Timings.Current.Kiai {
			bg.pulse.AddEventS(float64(time), float64(time)+bg.lastBeatLength, 1.05, 1.0)
		}

		bg.lastBeatProg++
	}

	bg.pulse.Update(float64(time))
	bg.rotation.Update(float64(time))
	bg.triangles.Update(float64(time))
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

		if bg.background != nil && (bg.storyboard == nil || !bg.storyboard.BGFileUsed()) && bgAlpha > 0.01 {
			batch.SetCamera(mgl32.Ortho(-1*float32(settings.Graphics.GetAspectRatio()), 1*float32(settings.Graphics.GetAspectRatio()), -1, 1, 1, -1).Mul4( /*mgl32.Translate3D(bg.position.X32(), bg.position.Y32(), 0).Mul4*/ (mgl32.Scale3D(float32(bg.scale.X), float32(-bg.scale.Y), 0)) /*.Mul4(mgl32.HomogRotate3DZ(float32(bg.rotation.GetValue())))*/))
			batch.SetScale( /*bg.scale.X*bg.pulse.GetValue(), -bg.scale.Y*bg.pulse.GetValue()*/ settings.Graphics.GetAspectRatio(), 1)

			if !settings.Playfield.BlurEnable {
				bgAlpha1 := math.Max(0.0, bgAlpha /* - 0.15*/)
				batch.SetColor(bgAlpha1, bgAlpha1, bgAlpha1, 1)
				//batch.SetColor(bgAlpha, bgAlpha, bgAlpha, 1)
				//batch.SetTranslation(bg.position)
			}
			batch.DrawUnit(*bg.background)
			/*if bg.pulse.GetValue() > 1.0 {
			batch.SetColor(bgAlpha*0.6, bgAlpha*0.6, bgAlpha*0.6, 1)

			batch.SetScale(/*bg.scale.X*bg.pulse.GetValue(), -bg.scale.Y*bg.pulse.GetValue()*/ /*settings.Graphics.GetAspectRatio()*bg.pulse.GetValue(), bg.pulse.GetValue())
				batch.SetAdditive(true)
				batch.DrawUnit(*bg.background)
				batch.SetAdditive(false)
			}*/
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

		batch.SetCamera(mgl32.Ortho(float32(-settings.Graphics.GetWidthF()/2), float32(settings.Graphics.GetWidthF()/2), float32(settings.Graphics.GetHeightF()/2), float32(-settings.Graphics.GetHeightF()/2), 1, -1))
		if !settings.Playfield.BlurEnable {
			batch.SetColor(bgAlpha, bgAlpha, bgAlpha, 1)
		} else {
			batch.SetColor(1, 1, 1, 1)
		}
		//bg.triangles.Draw(float64(time), batch)

		if settings.Playfield.BlurEnable {
			batch.End()

			texture := bg.blur.EndAndProcess()
			batch.Begin()
			batch.SetColor(1, 1, 1, bgAlpha)
			batch.SetCamera(mgl32.Ortho(-1, 1, -1, 1, 1, -1))
			//batch.SetTranslation(bg.position)
			//batch.SetScale(1+bg.parallax, 1+bg.parallax)
			batch.SetTranslation(bmath.NewVec2d(0, 0))
			batch.SetScale(1, 1)
			//batch.SetScale(1+bg.parallax, 1+bg.parallax)
			batch.DrawUnit(texture.GetRegion())
		}
		if bg.storyboard != nil && !bg.storyboard.IsWideScreen() {
			gl.Disable(gl.SCISSOR_TEST)
		}
	}

	batch.End()
}

func (bg *Background) DrawOverlay(time int64, batch *batches.SpriteBatch, bgAlpha float64, camera mgl32.Mat4) {
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

		if bg.storyboard != nil {
			batch.SetScale(1, 1)

			batch.SetColor(bgAlpha, bgAlpha, bgAlpha, 1)
			batch.SetTranslation(bmath.NewVec2d(320*bg.position.X, 240*bg.position.Y))
			batch.SetScale(1+bg.parallax, 1+bg.parallax)

			batch.SetCamera(camera)

			bg.storyboard.DrawOverlay(time, batch)
			batch.Flush()
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
