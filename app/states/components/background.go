package components

import (
	"github.com/EdlinOrg/prominentcolor"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/wieku/danser-go/app/beatmap"
	"github.com/wieku/danser-go/app/bmath"
	"github.com/wieku/danser-go/app/render"
	"github.com/wieku/danser-go/app/render/batches"
	"github.com/wieku/danser-go/app/render/effects"
	"github.com/wieku/danser-go/app/render/gui/drawables"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/app/storyboard"
	"github.com/wieku/danser-go/app/utils"
	"github.com/wieku/danser-go/framework/bass"
	"github.com/wieku/danser-go/framework/graphics/texture"
	"github.com/wieku/danser-go/framework/graphics/viewport"
	"github.com/wieku/danser-go/framework/math/scaling"
	"log"
	"math"
	"path/filepath"
)

type Background struct {
	blur       *effects.BlurEffect
	scale      bmath.Vector2d
	position   bmath.Vector2d
	background *texture.TextureRegion
	storyboard *storyboard.Storyboard
	lastTime   float64
	bMap       *beatmap.BeatMap
	triangles  *drawables.Triangles
}

func NewBackground(beatMap *beatmap.BeatMap) *Background {
	bg := new(Background)
	bg.bMap = beatMap
	bg.blur = effects.NewBlurEffect(int(settings.Graphics.GetWidth()), int(settings.Graphics.GetHeight()))

	var err error
	bg.background, err = utils.LoadTextureToAtlas(render.Atlas, filepath.Join(settings.General.OsuSongsDir, beatMap.Dir, beatMap.Bg))
	if err != nil {
		log.Println(err)
	}

	if settings.Playfield.Background.LoadStoryboards {
		bg.storyboard = storyboard.NewStoryboard(beatMap)

		if bg.storyboard == nil {
			log.Println("Storyboard not found!")
		}
	}

	imag, _ := utils.LoadImage(filepath.Join(settings.General.OsuSongsDir, beatMap.Dir, beatMap.Bg))

	cItems, _ := prominentcolor.KmeansWithAll(5, imag, prominentcolor.ArgumentDefault, prominentcolor.DefaultSize, prominentcolor.GetDefaultMasks())
	newCol := make([]bmath.Color, len(cItems))

	for i := 0; i < len(cItems); i++ {
		newCol[i] = bmath.Color{float64(cItems[i].Color.R) / 255, float64(cItems[i].Color.G) / 255, float64(cItems[i].Color.B) / 255, 1}
	}
	bg.triangles = drawables.NewTriangles(newCol)
	return bg
}

func (bg *Background) SetTrack(track *bass.Track) {
	bg.triangles.SetTrack(track)
}

func (bg *Background) Update(time float64, x, y float64) {
	if bg.lastTime == 0 {
		bg.lastTime = time
	}

	if bg.storyboard != nil {
		bg.storyboard.Update(int64(time))
	}

	bg.triangles.Update(time)

	x = bmath.ClampF64(x, -1, 1) * settings.Playfield.Background.Parallax.Amount
	y = bmath.ClampF64(y, -1, 1) * settings.Playfield.Background.Parallax.Amount

	delta := math.Abs(time - bg.lastTime)

	p := math.Pow(1-settings.Playfield.Background.Parallax.Speed, delta/100)

	bg.position.X = x*(1-p) + p*bg.position.X
	bg.position.Y = y*(1-p) + p*bg.position.Y

	bg.lastTime = time
}

func project(pos bmath.Vector2d, camera mgl32.Mat4) bmath.Vector2d {
	res := camera.Mul4x1(mgl32.Vec4{pos.X32(), pos.Y32(), 0.0, 1.0})
	return bmath.NewVec2d((float64(res[0])/2+0.5)*settings.Graphics.GetWidthF(), float64((res[1])/2+0.5)*settings.Graphics.GetWidthF())
}

func (bg *Background) Draw(time int64, batch *batches.SpriteBatch, blurVal, bgAlpha float64, camera mgl32.Mat4) {
	if bgAlpha < 0.01 {
		return
	}

	batch.Begin()
	batch.ResetTransform()
	batch.SetAdditive(false)

	if settings.Playfield.Background.Blur.Enabled {
		batch.SetColor(1, 1, 1, 1)

		bg.blur.SetBlur(blurVal, blurVal)
		bg.blur.Begin()
	} else {
		batch.SetColor(bgAlpha, bgAlpha, bgAlpha, 1)
	}

	bgScaling := scaling.Fill

	if bg.storyboard != nil && !bg.storyboard.IsWideScreen() {
		v1 := project(bmath.NewVec2d(256-320, 192+240), camera)
		v2 := project(bmath.NewVec2d(256+320, 192-240), camera)

		viewport.PushScissorPos(int(v1.X32()), int(v1.Y32()), int(v2.X32()-v1.X32()), int(v2.Y32()-v1.Y32()))

		bgScaling = scaling.Fit
	}

	if bg.background != nil && (bg.storyboard == nil || !bg.storyboard.BGFileUsed()) {
		batch.SetCamera(mgl32.Ortho(float32(-settings.Graphics.GetWidthF()/2), float32(settings.Graphics.GetWidthF()/2), float32(settings.Graphics.GetHeightF()/2), float32(-settings.Graphics.GetHeightF()/2), 1, -1))
		size := bgScaling.Apply(float32(bg.background.Width), float32(bg.background.Height), float32(settings.Graphics.GetWidthF()), float32(settings.Graphics.GetHeightF())).Scl(0.5)

		if !settings.Playfield.Background.Blur.Enabled {
			batch.SetTranslation(bg.position.Mult(bmath.NewVec2d(settings.Graphics.GetSizeF()).Scl(0.5)))
			size = size.Scl(float32(1 + settings.Playfield.Background.Parallax.Amount))
		}

		batch.SetScale(size.X64(), size.Y64())
		batch.DrawUnit(*bg.background)
	}

	if bg.storyboard != nil {
		batch.SetScale(1, 1)
		batch.SetTranslation(bmath.NewVec2d(0, 0))

		cam := camera
		if !settings.Playfield.Background.Blur.Enabled {
			scale := float32(1 + settings.Playfield.Background.Parallax.Amount)
			cam = mgl32.Translate3D(bg.position.X32(), -bg.position.Y32(), 0).Mul4(mgl32.Scale3D(scale, scale, 1)).Mul4(cam)
		}

		batch.SetCamera(cam)

		bg.storyboard.Draw(time, batch)
	}

	batch.SetCamera(mgl32.Ortho(float32(-settings.Graphics.GetWidthF()/2), float32(settings.Graphics.GetWidthF()/2), float32(settings.Graphics.GetHeightF()/2), float32(-settings.Graphics.GetHeightF()/2), 1, -1))
	batch.SetTranslation(bg.position.Mult(bmath.NewVec2d(settings.Graphics.GetSizeF()).Scl(0.5)))
	batch.SetScale(1+settings.Playfield.Background.Parallax.Amount, 1+settings.Playfield.Background.Parallax.Amount)
	//bg.triangles.Draw(float64(time), batch)

	if bg.storyboard != nil && !bg.storyboard.IsWideScreen() {
		viewport.PopScissor()
	}

	if settings.Playfield.Background.Blur.Enabled {
		batch.End()

		blurredTexture := bg.blur.EndAndProcess()

		batch.Begin()
		batch.SetColor(1, 1, 1, bgAlpha)
		batch.SetCamera(mgl32.Ortho(-1, 1, -1, 1, 1, -1))
		batch.SetTranslation(bg.position)
		batch.SetScale(1+settings.Playfield.Background.Parallax.Amount, 1+settings.Playfield.Background.Parallax.Amount)
		batch.DrawUnit(blurredTexture.GetRegion())
	}

	batch.End()
	batch.SetColor(1, 1, 1, 1)
	batch.ResetTransform()
}

func (bg *Background) DrawOverlay(time int64, batch *batches.SpriteBatch, bgAlpha float64, camera mgl32.Mat4) {
	if bgAlpha < 0.01 || bg.storyboard == nil {
		return
	}

	if !bg.storyboard.IsWideScreen() {
		v1 := project(bmath.NewVec2d(256-320, 192+240), camera)
		v2 := project(bmath.NewVec2d(256+320, 192-240), camera)

		viewport.PushScissorPos(int(v1.X32()), int(v1.Y32()), int(v2.X32()-v1.X32()), int(v2.Y32()-v1.Y32()))
	}

	batch.Begin()

	batch.SetColor(bgAlpha, bgAlpha, bgAlpha, 1)
	batch.ResetTransform()
	batch.SetAdditive(false)

	cam := mgl32.Translate3D(bg.position.X32(), -bg.position.Y32(), 0).Mul4(mgl32.Scale3D(float32(1+settings.Playfield.Background.Parallax.Amount), float32(1+settings.Playfield.Background.Parallax.Amount), 1)).Mul4(camera)
	batch.SetCamera(cam)

	bg.storyboard.DrawOverlay(time, batch)

	batch.End()

	if bg.storyboard != nil && !bg.storyboard.IsWideScreen() {
		viewport.PopScissor()
	}
	batch.SetColor(1, 1, 1, 1)
	batch.ResetTransform()
}

func (bg *Background) GetStoryboard() *storyboard.Storyboard {
	return bg.storyboard
}
