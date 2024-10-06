package common

import (
	"github.com/EdlinOrg/prominentcolor"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/wieku/danser-go/app/beatmap"
	"github.com/wieku/danser-go/app/graphics/gui/drawables"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/app/storyboard"
	"github.com/wieku/danser-go/framework/assets"
	"github.com/wieku/danser-go/framework/bass"
	"github.com/wieku/danser-go/framework/goroutines"
	"github.com/wieku/danser-go/framework/graphics/batch"
	"github.com/wieku/danser-go/framework/graphics/effects"
	"github.com/wieku/danser-go/framework/graphics/texture"
	"github.com/wieku/danser-go/framework/graphics/viewport"
	color2 "github.com/wieku/danser-go/framework/math/color"
	"github.com/wieku/danser-go/framework/math/math32"
	"github.com/wieku/danser-go/framework/math/mutils"
	"github.com/wieku/danser-go/framework/math/scaling"
	"github.com/wieku/danser-go/framework/math/vector"
	"log"
	"math"
	"path/filepath"
)

type Background struct {
	lastTime float64

	scaling scaling.Scaling

	background *texture.TextureSingle
	storyboard *storyboard.Storyboard
	triangles  *drawables.Triangles

	blur           *effects.BlurEffect
	blurVal        float64
	blurredTexture texture.Texture
	forceRedraw    bool

	blurActive bool

	parallaxPosition vector.Vector2d
	parallaxScale    float64
}

func NewBackground(loadDefault bool) *Background {
	bg := new(Background)
	bg.blurVal = -1
	bg.blur = effects.NewBlurEffect(int(settings.Graphics.GetWidth()), int(settings.Graphics.GetHeight()))
	bg.blurActive = settings.Playfield.Background.Blur.Enabled

	if loadDefault {
		image, err := assets.GetPixmap("assets/textures/background-1.png")
		if err != nil {
			panic(err)
		}

		bg.background = texture.LoadTextureSingle(image.RGBA(), 0)

		bg.triangles = drawables.NewTriangles(bg.getColors(image))

		if image != nil {
			image.Dispose()
		}
	} else {
		bg.triangles = drawables.NewTriangles(nil)
	}

	parallax := settings.Playfield.Background.Parallax

	if parallax.Enabled && settings.DIVIDES == 1 && math.Abs(parallax.Amount) > 0.0001 { // avoid scaling the background during fade in
		bg.parallaxScale = math.Abs(parallax.Amount)
	}

	return bg
}

func (bg *Background) SetBeatmap(beatMap *beatmap.BeatMap, loadDefault, loadStoryboards bool) {
	bgLoadFunc := func() {
		image, err := texture.NewPixmapFileString(filepath.Join(settings.General.GetSongsDir(), beatMap.Dir, beatMap.Bg))
		if err != nil && loadDefault {
			image, err = assets.GetPixmap("assets/textures/background-1.png")
			if err != nil {
				panic(err)
			}
		}

		bg.triangles.SetColors(bg.getColors(image))

		goroutines.CallNonBlockMain(func() {
			if bg.background != nil { // Dispose old background texture
				bg.background.Dispose()
				bg.background = nil
			}

			if image != nil {
				bg.background = texture.LoadTextureSingle(image.RGBA(), 0)
				image.Dispose()
			}

			bg.forceRedraw = true
		})
	}

	if settings.RECORD {
		bgLoadFunc()
	} else {
		goroutines.Run(bgLoadFunc)
	}

	if loadStoryboards {
		bg.storyboard = storyboard.NewStoryboard(beatMap)

		if bg.storyboard == nil {
			log.Println("Storyboard not found!")
		}
	}
}

func (bg *Background) SetTrack(track bass.ITrack) {
	bg.triangles.SetTrack(track)
}

func (bg *Background) Update(time float64, x, y float64) {
	if bg.lastTime == 0 {
		bg.lastTime = time
	}

	if bg.storyboard != nil {
		if settings.RECORD || !bg.storyboard.HasVisuals() { // Use update thread if we only have sounds
			bg.storyboard.Update(time)
		} else {
			if !bg.storyboard.IsThreadRunning() {
				bg.storyboard.StartThread()
			}
			bg.storyboard.UpdateTime(time)
		}
	}

	bg.triangles.SetDensity(settings.Playfield.Background.Triangles.Density)
	bg.triangles.SetScale(settings.Playfield.Background.Triangles.Scale)
	bg.triangles.Update(time)

	pX := 0.0
	pY := 0.0

	parallax := settings.Playfield.Background.Parallax

	parallaxTarget := 0.0

	if parallax.Enabled && settings.DIVIDES == 1 && math.Abs(parallax.Amount) > 0.0001 {
		parallaxTarget = math.Abs(parallax.Amount)

		if !math.IsNaN(x) && !math.IsNaN(y) {
			pX = mutils.Clamp(x, -1, 1) * parallax.Amount
			pY = mutils.Clamp(y, -1, 1) * parallax.Amount
		}
	}

	delta := math.Abs(time - bg.lastTime)

	p := math.Pow(1-parallax.Speed, delta/100)

	bg.parallaxPosition.X = pX*(1-p) + p*bg.parallaxPosition.X
	bg.parallaxPosition.Y = pY*(1-p) + p*bg.parallaxPosition.Y
	bg.parallaxScale = parallaxTarget*(1-p) + p*bg.parallaxScale

	bg.lastTime = time
}

func project(pos vector.Vector2d, camera mgl32.Mat4) vector.Vector2d {
	res := camera.Mul4x1(mgl32.Vec4{pos.X32(), pos.Y32(), 0.0, 1.0})
	return vector.NewVec2d((float64(res[0])/2+0.5)*settings.Graphics.GetWidthF(), float64((res[1])/2+0.5)*settings.Graphics.GetWidthF())
}

func (bg *Background) Draw(time float64, batch *batch.QuadBatch, blurVal, bgAlpha float64, camera mgl32.Mat4) {
	if bgAlpha < 0.01 {
		return
	}

	batch.Begin()

	needsRedraw := bg.forceRedraw || (bg.storyboard != nil && bg.storyboard.HasVisuals()) || !bg.blurActive || (settings.Playfield.Background.Triangles.Enabled && !settings.Playfield.Background.Triangles.DrawOverBlur)

	bg.forceRedraw = false

	if bg.blurActive != settings.Playfield.Background.Blur.Enabled {
		needsRedraw = true
		bg.blurActive = settings.Playfield.Background.Blur.Enabled
	}

	if math.Abs(bg.blurVal-blurVal) > 0.0001 {
		needsRedraw = true
		bg.blurVal = blurVal
	}

	var clipX, clipY, clipW, clipH int
	widescreen := true

	bg.scaling = scaling.Fill

	if bg.storyboard != nil && !bg.storyboard.IsWideScreen() {
		widescreen = false

		v1 := project(vector.NewVec2d(256-320, 192+240), camera)
		v2 := project(vector.NewVec2d(256+320, 192-240), camera)

		clipX, clipY, clipW, clipH = int(v1.X32()), int(v1.Y32()), int(v2.X32()-v1.X32()), int(v2.Y32()-v1.Y32())

		bg.scaling = scaling.Fit
	}

	if needsRedraw {
		batch.ResetTransform()
		batch.SetAdditive(false)

		opacity := 1.0
		if bg.storyboard != nil {
			opacity *= 1.0 - bg.storyboard.GetVideoAlpha()
		}

		if bg.blurActive {
			bg.blur.SetBlur(blurVal, blurVal)
			bg.blur.Begin()
		} else {
			opacity *= bgAlpha
		}

		batch.SetColor(opacity, opacity, opacity, 1)

		if !widescreen && !bg.blurActive {
			viewport.PushScissorPos(clipX, clipY, clipW, clipH)
		}

		if bg.background != nil && (bg.storyboard == nil || !bg.storyboard.BGFileUsed()) {
			batch.SetCamera(mgl32.Ortho(float32(-settings.Graphics.GetWidthF()/2), float32(settings.Graphics.GetWidthF()/2), float32(settings.Graphics.GetHeightF()/2), float32(-settings.Graphics.GetHeightF()/2), 1, -1))
			size := bg.scaling.Apply(float32(bg.background.GetWidth()), float32(bg.background.GetHeight()), float32(settings.Graphics.GetWidthF()), float32(settings.Graphics.GetHeightF())).Scl(0.5)

			if !bg.blurActive {
				batch.SetTranslation(bg.parallaxPosition.Mult(vector.NewVec2d(1, -1)).Mult(vector.NewVec2d(settings.Graphics.GetSizeF()).Scl(0.5)))
				size = size.Scl(float32(1 + bg.parallaxScale))
			}

			batch.SetScale(size.X64(), size.Y64())
			batch.DrawUnit(bg.background.GetRegion())
		}

		if bg.blurActive {
			batch.SetColor(1, 1, 1, 1)
		} else {
			batch.SetColor(bgAlpha, bgAlpha, bgAlpha, 1)
		}

		if bg.storyboard != nil {
			batch.SetScale(1, 1)
			batch.SetTranslation(vector.NewVec2d(0, 0))

			cam := camera
			if !bg.blurActive {
				scale := float32(1 + bg.parallaxScale)
				cam = mgl32.Translate3D(bg.parallaxPosition.X32(), bg.parallaxPosition.Y32(), 0).Mul4(mgl32.Scale3D(scale, scale, 1)).Mul4(cam)
			}

			batch.SetCamera(cam)

			bg.storyboard.Draw(time, batch)
		}

		if settings.Playfield.Background.Triangles.Enabled && !settings.Playfield.Background.Triangles.DrawOverBlur {
			bg.drawTriangles(batch, bgAlpha, bg.blurActive)
		}

		batch.Flush()
		batch.SetColor(1, 1, 1, 1)
		batch.ResetTransform()

		if !widescreen && !bg.blurActive {
			viewport.PopScissor()
		}

		if bg.blurActive {
			bg.blurredTexture = bg.blur.EndAndProcess()
		}
	}

	if !widescreen {
		viewport.PushScissorPos(clipX, clipY, clipW, clipH)
	}

	if bg.blurActive && bg.blurredTexture != nil {
		batch.ResetTransform()
		batch.SetAdditive(false)
		batch.SetColor(1, 1, 1, bgAlpha)
		batch.SetCamera(mgl32.Ortho(-1, 1, -1, 1, 1, -1))
		batch.SetTranslation(bg.parallaxPosition)
		batch.SetScale(1+bg.parallaxScale, 1+bg.parallaxScale)
		batch.DrawUnit(bg.blurredTexture.GetRegion())
		batch.Flush()
		batch.SetColor(1, 1, 1, 1)
		batch.ResetTransform()
	}

	if settings.Playfield.Background.Triangles.Enabled && settings.Playfield.Background.Triangles.DrawOverBlur {
		bg.drawTriangles(batch, bgAlpha, false)
	}

	batch.End()

	if !widescreen {
		viewport.PopScissor()
	}
}

func (bg *Background) drawTriangles(batch *batch.QuadBatch, bgAlpha float64, blur bool) {
	batch.ResetTransform()
	cam := mgl32.Ortho(float32(-settings.Graphics.GetWidthF()/2), float32(settings.Graphics.GetWidthF()/2), float32(settings.Graphics.GetHeightF()/2), float32(-settings.Graphics.GetHeightF()/2), 1, -1)

	if !blur {
		batch.SetColor(bgAlpha, bgAlpha, bgAlpha, 1)

		subScale := float32(settings.Playfield.Background.Triangles.ParallaxMultiplier)
		scale := 1 + float32(bg.parallaxScale)*math32.Abs(subScale)
		cam = mgl32.Translate3D(bg.parallaxPosition.X32()*subScale, bg.parallaxPosition.Y32()*subScale, 0).Mul4(mgl32.Scale3D(scale, scale, 1)).Mul4(cam)
	} else {
		batch.SetColor(1, 1, 1, 1)
	}

	batch.SetCamera(cam)
	bg.triangles.Draw(bg.lastTime, batch)

	batch.SetAdditive(false)
	batch.ResetTransform()
}

func (bg *Background) DrawOverlay(time float64, batch *batch.QuadBatch, bgAlpha float64, camera mgl32.Mat4) {
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

	scale := float32(1 + bg.parallaxScale)
	cam := mgl32.Translate3D(bg.parallaxPosition.X32(), -bg.parallaxPosition.Y32(), 0).Mul4(mgl32.Scale3D(scale, scale, 1)).Mul4(camera)
	batch.SetCamera(cam)

	bg.storyboard.DrawOverlay(time, batch)

	batch.End()

	if !bg.storyboard.IsWideScreen() {
		viewport.PopScissor()
	}

	batch.SetColor(1, 1, 1, 1)
	batch.ResetTransform()
}

func (bg *Background) GetStoryboard() *storyboard.Storyboard {
	return bg.storyboard
}

func (bg *Background) getColors(image *texture.Pixmap) []color2.Color {
	newCol := make([]color2.Color, 0)

	if image != nil {
		cItems, err1 := prominentcolor.KmeansWithAll(10, image.NRGBA(), prominentcolor.ArgumentDefault, prominentcolor.DefaultSize, prominentcolor.GetDefaultMasks())

		if err1 == nil {
			for i := 0; i < len(cItems); i++ {
				if cItems[i].Color.R+cItems[i].Color.G+cItems[i].Color.B == 0 { //skip black colors as they're useless
					continue
				}

				newCol = append(newCol, color2.NewIRGB(uint8(cItems[i].Color.R), uint8(cItems[i].Color.G), uint8(cItems[i].Color.B)) /*.Lighten2(0.15)*/)
			}
		}
	}

	return newCol
}

func (bg *Background) HasBackground() bool {
	return bg.background != nil
}

func (bg *Background) SetDefaultColor(col color2.Color) {
	bg.triangles.SetDefaultColor(col)
}
