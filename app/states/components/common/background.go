package common

import (
	"github.com/EdlinOrg/prominentcolor"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/wieku/danser-go/app/beatmap"
	"github.com/wieku/danser-go/app/bmath"
	"github.com/wieku/danser-go/app/graphics"
	"github.com/wieku/danser-go/app/graphics/gui/drawables"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/app/storyboard"
	"github.com/wieku/danser-go/app/utils"
	"github.com/wieku/danser-go/framework/bass"
	"github.com/wieku/danser-go/framework/graphics/effects"
	"github.com/wieku/danser-go/framework/graphics/sprite"
	"github.com/wieku/danser-go/framework/graphics/texture"
	"github.com/wieku/danser-go/framework/graphics/viewport"
	color2 "github.com/wieku/danser-go/framework/math/color"
	"github.com/wieku/danser-go/framework/math/scaling"
	"github.com/wieku/danser-go/framework/math/vector"
	"log"
	"math"
	"path/filepath"
)

type Background struct {
	blur           *effects.BlurEffect
	scale          vector.Vector2d
	position       vector.Vector2d
	background     *texture.TextureRegion
	storyboard     *storyboard.Storyboard
	lastTime       float64
	bMap           *beatmap.BeatMap
	triangles      *drawables.Triangles
	blurVal        float64
	blurredTexture texture.Texture
	scaling        scaling.Scaling
}

func NewBackground(beatMap *beatmap.BeatMap) *Background {
	bg := new(Background)
	bg.blurVal = -1
	bg.bMap = beatMap
	bg.blur = effects.NewBlurEffect(int(settings.Graphics.GetWidth()), int(settings.Graphics.GetHeight()))

	var err error
	bg.background, err = utils.LoadTextureToAtlas(graphics.Atlas, filepath.Join(settings.General.OsuSongsDir, beatMap.Dir, beatMap.Bg))
	if err != nil {
		log.Println(err)
	}

	if settings.Playfield.Background.LoadStoryboards {
		bg.storyboard = storyboard.NewStoryboard(beatMap)

		if bg.storyboard == nil {
			log.Println("Storyboard not found!")
		}
	}

	imag, err := texture.NewPixmapFileString(filepath.Join(settings.General.OsuSongsDir, beatMap.Dir, beatMap.Bg))

	newCol := make([]color2.Color, 0)

	if err == nil {
		cItems, _ := prominentcolor.KmeansWithAll(5, imag.NRGBA(), prominentcolor.ArgumentDefault, prominentcolor.DefaultSize, prominentcolor.GetDefaultMasks())
		newCol = make([]color2.Color, len(cItems))

		for i := 0; i < len(cItems); i++ {
			r, g, b := float32(cItems[i].Color.R)/255, float32(cItems[i].Color.G)/255, float32(cItems[i].Color.B)/255

			r = (1-r)*0.9 + r
			g = (1-g)*0.9 + g
			b = (1-b)*0.9 + b

			newCol[i] = color2.Color{r, g, b, 1}
		}
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
		if !bg.storyboard.IsThreadRunning() {
			bg.storyboard.StartThread()
		}
		bg.storyboard.UpdateTime(int64(time))
	}

	bg.triangles.Update(time)

	pX := 0.0
	pY := 0.0

	if settings.Playfield.Background.Parallax.Amount > 0.0001 && !math.IsNaN(x) && !math.IsNaN(y) && settings.DIVIDES == 1 {
		pX = bmath.ClampF64(x, -1, 1) * settings.Playfield.Background.Parallax.Amount
		pY = bmath.ClampF64(y, -1, 1) * settings.Playfield.Background.Parallax.Amount
	}

	delta := math.Abs(time - bg.lastTime)

	p := math.Pow(1-settings.Playfield.Background.Parallax.Speed, delta/100)

	bg.position.X = pX*(1-p) + p*bg.position.X
	bg.position.Y = pY*(1-p) + p*bg.position.Y

	bg.lastTime = time
}

func project(pos vector.Vector2d, camera mgl32.Mat4) vector.Vector2d {
	res := camera.Mul4x1(mgl32.Vec4{pos.X32(), pos.Y32(), 0.0, 1.0})
	return vector.NewVec2d((float64(res[0])/2+0.5)*settings.Graphics.GetWidthF(), float64((res[1])/2+0.5)*settings.Graphics.GetWidthF())
}

func (bg *Background) Draw(time int64, batch *sprite.SpriteBatch, blurVal, bgAlpha float64, camera mgl32.Mat4) {
	if bgAlpha < 0.01 {
		return
	}

	needsRedraw := bg.storyboard != nil || !settings.Playfield.Background.Blur.Enabled

	if math.Abs(bg.blurVal-blurVal) > 0.001 {
		needsRedraw = true
		bg.blurVal = blurVal
	}

	if needsRedraw {
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

		bg.scaling = scaling.Fill

		if bg.storyboard != nil && !bg.storyboard.IsWideScreen() {
			v1 := project(vector.NewVec2d(256-320, 192+240), camera)
			v2 := project(vector.NewVec2d(256+320, 192-240), camera)

			viewport.PushScissorPos(int(v1.X32()), int(v1.Y32()), int(v2.X32()-v1.X32()), int(v2.Y32()-v1.Y32()))

			bg.scaling = scaling.Fit
		}

		if bg.background != nil && (bg.storyboard == nil || !bg.storyboard.BGFileUsed()) {
			batch.SetCamera(mgl32.Ortho(float32(-settings.Graphics.GetWidthF()/2), float32(settings.Graphics.GetWidthF()/2), float32(settings.Graphics.GetHeightF()/2), float32(-settings.Graphics.GetHeightF()/2), 1, -1))
			size := bg.scaling.Apply(float32(bg.background.Width), float32(bg.background.Height), float32(settings.Graphics.GetWidthF()), float32(settings.Graphics.GetHeightF())).Scl(0.5)

			if !settings.Playfield.Background.Blur.Enabled {
				batch.SetTranslation(bg.position.Mult(vector.NewVec2d(1, -1)).Mult(vector.NewVec2d(settings.Graphics.GetSizeF()).Scl(0.5)))
				size = size.Scl(float32(1 + settings.Playfield.Background.Parallax.Amount))
			}

			batch.SetScale(size.X64(), size.Y64())
			batch.DrawUnit(*bg.background)
		}

		if bg.storyboard != nil {
			batch.SetScale(1, 1)
			batch.SetTranslation(vector.NewVec2d(0, 0))

			cam := camera
			if !settings.Playfield.Background.Blur.Enabled {
				scale := float32(1 + settings.Playfield.Background.Parallax.Amount)
				cam = mgl32.Translate3D(bg.position.X32(), bg.position.Y32(), 0).Mul4(mgl32.Scale3D(scale, scale, 1)).Mul4(cam)
			}

			batch.SetCamera(cam)

			bg.storyboard.Draw(time, batch)
		}

		//batch.SetCamera(mgl32.Ortho(float32(-settings.Graphics.GetWidthF()/2), float32(settings.Graphics.GetWidthF()/2), float32(settings.Graphics.GetHeightF()/2), float32(-settings.Graphics.GetHeightF()/2), 1, -1))
		//batch.SetTranslation(bg.position.Mult(vector.NewVec2d(settings.Graphics.GetSizeF()).Scl(0.5)))
		//batch.SetScale(1+settings.Playfield.Background.Parallax.Amount, 1+settings.Playfield.Background.Parallax.Amount)
		//bg.triangles.Draw(float64(time), batch)

		if bg.storyboard != nil && !bg.storyboard.IsWideScreen() {
			viewport.PopScissor()
		}

		batch.End()
		batch.SetColor(1, 1, 1, 1)
		batch.ResetTransform()

		if settings.Playfield.Background.Blur.Enabled {
			bg.blurredTexture = bg.blur.EndAndProcess()
		}
	}

	if settings.Playfield.Background.Blur.Enabled && bg.blurredTexture != nil {
		batch.Begin()
		batch.ResetTransform()
		batch.SetAdditive(false)
		batch.SetColor(1, 1, 1, bgAlpha)
		batch.SetCamera(mgl32.Ortho(-1, 1, -1, 1, 1, -1))
		batch.SetTranslation(bg.position)
		batch.SetScale(1+settings.Playfield.Background.Parallax.Amount, 1+settings.Playfield.Background.Parallax.Amount)
		batch.DrawUnit(bg.blurredTexture.GetRegion())
		batch.End()
		batch.SetColor(1, 1, 1, 1)
		batch.ResetTransform()
	}
}

func (bg *Background) DrawOverlay(time int64, batch *sprite.SpriteBatch, bgAlpha float64, camera mgl32.Mat4) {
	if bgAlpha < 0.01 || bg.storyboard == nil {
		return
	}

	if !bg.storyboard.IsWideScreen() {
		v1 := project(vector.NewVec2d(256-320, 192+240), camera)
		v2 := project(vector.NewVec2d(256+320, 192-240), camera)

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
