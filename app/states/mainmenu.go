package states

import (
	"fmt"
	"github.com/EdlinOrg/prominentcolor"
	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/wieku/danser-go/app/animation"
	"github.com/wieku/danser-go/app/animation/easing"
	"github.com/wieku/danser-go/app/audio"
	"github.com/wieku/danser-go/app/beatmap"
	"github.com/wieku/danser-go/app/bmath"
	"github.com/wieku/danser-go/app/input"
	"github.com/wieku/danser-go/app/render"
	"github.com/wieku/danser-go/app/render/batches"
	"github.com/wieku/danser-go/app/render/effects"
	"github.com/wieku/danser-go/app/render/font"
	"github.com/wieku/danser-go/app/render/gui/drawables"
	"github.com/wieku/danser-go/app/render/sprites"
	"github.com/wieku/danser-go/app/render/texture"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/app/states/components"
	"github.com/wieku/danser-go/app/utils"
	"github.com/wieku/danser-go/framework/qpc"
	"log"
	"math"
	"path/filepath"
	"time"
)

type MainMenu struct {
	font *font.Font
	bMap *beatmap.BeatMap

	bloomEffect *effects.BloomEffect
	lastTime    int64
	progressMsF float64
	progressMs  int64
	batch       *batches.SpriteBatch

	cursor *render.Cursor

	background *components.Background

	LogoS1 *sprites.Sprite
	LogoS2 *sprites.Sprite

	WaveTex   *texture.TextureRegion
	BgScl     bmath.Vector2d
	vposition bmath.Vector2d

	fadeOut     float64
	fadeIn      float64
	entry       float64
	start       bool
	mus         bool
	musicPlayer *audio.Music

	visualiser *drawables.Visualiser
	triangles  *drawables.Triangles

	profiler  *utils.FPSCounter
	profilerU *utils.FPSCounter

	camera  *bmath.Camera
	scamera *bmath.Camera

	leftPulse  *sprites.Sprite
	rightPulse *sprites.Sprite

	cursorGlider *animation.Glider
	counter      float64
	fpsC         float64
	fpsU         float64

	mapFullName string

	Epi            *texture.TextureRegion
	epiGlider      *animation.Glider
	currentBeatVal float64
	lastBeatLength float64
	lastBeatStart  float64
	beatProgress   float64
	lastBeatProg   int64

	progress, lastProgress float64

	heartbeat *audio.Sample

	vol        float64
	volAverage float64
	waves      []*sprites.Sprite

	hovering bool
	hover    float64

	cookieSize float64
}

func NewMainMenu(beatMap *beatmap.BeatMap) *MainMenu {
	player := new(MainMenu)
	render.LoadTextures()
	render.SetupSlider()
	player.batch = batches.NewSpriteBatch()

	player.font = font.GetFont("Exo 2 Bold")

	player.bMap = beatMap
	player.mapFullName = fmt.Sprintf("%s - %s [%s]", beatMap.Artist, beatMap.Name, beatMap.Difficulty)
	log.Println("Playing:", player.mapFullName)

	var err error

	Logo, err := utils.LoadTextureToAtlas(render.Atlas, "assets/textures/coinbig.png")

	player.LogoS1 = sprites.NewSpriteSingle(Logo, 0, bmath.NewVec2d(0, 0), bmath.NewVec2d(0, 0))
	player.LogoS2 = sprites.NewSpriteSingle(Logo, 0, bmath.NewVec2d(0, 0), bmath.NewVec2d(0, 0))

	FlashTex, err := utils.LoadTextureToAtlas(render.Atlas, "assets/textures/flash.png")

	player.leftPulse = sprites.NewSpriteSingle(FlashTex, 0, bmath.NewVec2d((float64(FlashTex.Width)-settings.Graphics.GetWidthF())/2, 0), bmath.NewVec2d(0, 0))
	player.rightPulse = sprites.NewSpriteSingle(FlashTex, 0, bmath.NewVec2d((settings.Graphics.GetWidthF()-float64(FlashTex.Width))/2, 0), bmath.NewVec2d(0, 0))

	scal := settings.Graphics.GetHeightF() / float64(FlashTex.Height)

	player.leftPulse.SetScaleV(bmath.NewVec2d(1, scal))
	player.leftPulse.SetHFlip(true)
	player.leftPulse.SetAdditive(true)
	player.leftPulse.SetAlpha(0)

	player.rightPulse.SetScaleV(bmath.NewVec2d(1, scal))
	player.rightPulse.SetAdditive(true)
	player.rightPulse.SetAlpha(0)

	player.WaveTex, err = utils.LoadTextureToAtlas(render.Atlas, "assets/textures/coinwave.png")
	player.Epi, err = utils.LoadTextureToAtlas(render.Atlas, "assets/textures/warning.png")

	if err != nil {
		log.Println(err)
	}

	player.background = components.NewBackground(beatMap, 0.02, false)

	if settings.Graphics.GetWidthF() > settings.Graphics.GetHeightF() {
		player.cookieSize = 0.5 * settings.Graphics.GetHeightF()
	} else {
		player.cookieSize = 0.5 * settings.Graphics.GetWidthF()
	}

	player.camera = bmath.NewCamera()
	player.camera.SetOsuViewport(int(settings.Graphics.GetWidth()), int(settings.Graphics.GetHeight()), settings.Playfield.Scale, false)
	player.camera.Update()

	render.Camera = player.camera

	player.scamera = bmath.NewCamera()
	player.scamera.SetViewport(int(settings.Graphics.GetWidth()), int(settings.Graphics.GetHeight()), false)
	player.scamera.SetOrigin(bmath.NewVec2d(settings.Graphics.GetWidthF()/2, settings.Graphics.GetHeightF()/2))
	player.scamera.Update()

	player.bMap.Reset()

	player.heartbeat = audio.NewSample("assets/sounds/heartbeat.mp3")

	player.cursor = render.NewCursor()

	player.lastTime = -1

	log.Println("Music:", beatMap.Audio)

	player.fadeOut = 1.0
	player.fadeIn = 0.0
	player.hover = 1.0

	player.cursorGlider = animation.NewGlider(1)

	player.epiGlider = animation.NewGlider(0)
	player.epiGlider.AddEvent(0, 500, 1.0)
	player.epiGlider.AddEvent(2500, 3000, 0.0)

	musicPlayer := audio.NewMusic(filepath.Join(settings.General.OsuSongsDir, beatMap.Dir, beatMap.Audio))

	//mapStart := player.bMap.HitObjects[0].GetBasicData().StartTime

	go func() {
		player.entry = 1
		time.Sleep(time.Duration(settings.Playfield.LeadInTime * float64(time.Second)))

		player.start = true
		musicPlayer.Play()
		musicPlayer.SetTempo(settings.SPEED)
		musicPlayer.SetPitch(settings.PITCH)
		//musicPlayer.SetPosition(57)
	}()

	player.visualiser = drawables.NewVisualiser(player.cookieSize*0.66, player.cookieSize*2, bmath.NewVec2d(0, 0))

	imag, _ := utils.LoadImage(filepath.Join(settings.General.OsuSongsDir, beatMap.Dir, beatMap.Bg))

	cItems, _ := prominentcolor.KmeansWithAll(5, imag, prominentcolor.ArgumentDefault, prominentcolor.DefaultSize, prominentcolor.GetDefaultMasks())
	newCol := make([]bmath.Color, len(cItems))

	for i := 0; i < len(cItems); i++ {
		newCol[i] = bmath.Color{float64(cItems[i].Color.R) / 255, float64(cItems[i].Color.G) / 255, float64(cItems[i].Color.B) / 255, 1}
	}

	player.triangles = drawables.NewTriangles(newCol)

	player.profilerU = utils.NewFPSCounter(60, false)
	go func() {
		//var last = musicPlayer.GetPosition()
		var lastT = qpc.GetNanoTime()
		for {

			currtime := qpc.GetNanoTime()

			player.profilerU.PutSample(1000.0 / (float64(currtime-lastT) / 1000000.0))

			delta := float64(currtime-lastT) / 1000000.0

			lastT = currtime

			player.progressMsF = musicPlayer.GetPosition()*1000 + float64(settings.Audio.Offset)

			player.bMap.Update(int64(player.progressMsF))

			player.background.Update(int64(player.progressMsF), 0, 0)
			player.visualiser.Update(player.progressMsF)
			player.triangles.Update(player.progressMsF)

			if player.start && len(player.bMap.Queue) > 0 {
				player.cursorGlider.Update(player.progressMsF)
			}

			x, y := input.Win.GetCursorPos()

			player.cursor.SetScreenPos(bmath.NewVec2d(x, y).Copy32())
			player.cursor.Update(delta)

			player.vposition.X = -(x - settings.Graphics.GetWidthF()/2) / settings.Graphics.GetWidthF() * 0.04
			player.vposition.Y = -(y - settings.Graphics.GetHeightF()/2) / settings.Graphics.GetHeightF() * 0.04

			player.visualiser.Position = player.vposition.Scl(player.cookieSize * 0.66)
			player.LogoS1.SetPosition(player.vposition.Scl(player.cookieSize * 0.66))
			player.LogoS2.SetPosition(player.vposition.Scl(player.cookieSize * 0.66))

			xm := x - settings.Graphics.GetWidthF()/2 - player.vposition.X*player.cookieSize*0.66
			ym := y - settings.Graphics.GetHeightF()/2 - player.vposition.Y*player.cookieSize*0.66

			//log.Println(xm, ym, xm*xm+ym*ym)
			if xm*xm+ym*ym < player.cookieSize*0.66*player.cookieSize*0.66*player.hover*player.hover {
				player.hover = math.Min(1.1, player.hover+delta*0.001)
			} else {
				player.hover = math.Max(1.0, player.hover-delta*0.001)
			}

			bTime := player.bMap.Timings.Current.BaseBpm

			if player.start {
				if bTime != player.lastBeatLength {
					player.lastBeatLength = bTime
					player.lastBeatStart = float64(player.bMap.Timings.Current.Time)
					player.lastBeatProg = -1
				}

				if int64(float64(player.progressMsF-player.lastBeatStart)/player.lastBeatLength) > player.lastBeatProg {
					player.heartbeat.Play()

					wave := sprites.NewSpriteSingle(player.WaveTex, 0, bmath.NewVec2d(0, 0), bmath.NewVec2d(0, 0))
					wave.SetScale(0)
					bScale := player.cookieSize / float64(player.WaveTex.Height/2) * 0.7
					wave.AddTransform(animation.NewSingleTransform(animation.Fade, easing.OutQuad, player.progressMsF, player.progressMsF+1000, 0.5, 0))
					wave.AddTransform(animation.NewSingleTransform(animation.Scale, easing.OutQuad, player.progressMsF, player.progressMsF+1000, 1*bScale*player.hover, 1.4*bScale*player.hover))
					wave.AdjustTimesToTransformations()
					player.waves = append(player.waves, wave)

					norLength := math.Max(300, player.lastBeatLength)

					if player.bMap.Timings.Current.Kiai || player.lastBeatProg%4 == 0 {
						if !player.bMap.Timings.Current.Kiai || player.lastBeatProg%2 == 0 {
							player.leftPulse.AddTransform(animation.NewSingleTransform(animation.Fade, easing.OutQuad, player.progressMsF, player.progressMsF+norLength, 0.6*player.musicPlayer.GetLeftLevel(), 0))
						}

						if !player.bMap.Timings.Current.Kiai || player.lastBeatProg%2 == 1 {
							player.rightPulse.AddTransform(animation.NewSingleTransform(animation.Fade, easing.OutQuad, player.progressMsF, player.progressMsF+norLength, 0.6*player.musicPlayer.GetRightLevel(), 0))
						}
					}

					player.lastBeatProg++
				}

				player.beatProgress = float64(player.progressMsF-player.lastBeatStart)/player.lastBeatLength - float64(player.lastBeatProg)

			} else {
				player.beatProgress = 0
			}

			time.Sleep(time.Millisecond)
		}
	}()

	go func() {
		for {
			musicPlayer.Update()
			time.Sleep(15 * time.Millisecond)
		}
	}()
	player.profiler = utils.NewFPSCounter(60, false)
	player.musicPlayer = musicPlayer
	player.visualiser.SetTrack(player.musicPlayer)
	player.triangles.SetTrack(player.musicPlayer)

	player.bloomEffect = effects.NewBloomEffect(int(settings.Graphics.GetWidth()), int(settings.Graphics.GetHeight()))

	return player
}

func (pl *MainMenu) Show() {

}

func (pl *MainMenu) Draw(delta float64) {
	if pl.lastTime < 0 {
		pl.lastTime = qpc.GetNanoTime()
	}
	tim := qpc.GetNanoTime()
	timMs := float64(tim-pl.lastTime) / 1000000.0

	pl.profiler.PutSample(1000.0 / timMs)
	fps := pl.profiler.GetFPS()

	if pl.start {

		if fps > 58 && timMs > 18 {
			log.Println("Slow frame detected! Frame time:", timMs, "| Av. frame time:", 1000.0/fps)
		}

		pl.progressMs = int64(pl.progressMsF)
	}

	pl.lastTime = tim

	/*if len(pl.bMap.Queue) == 0 {
		pl.fadeOut -= timMs / (settings.Playfield.FadeOutTime * 1000)
		pl.fadeOut = math.Max(0.0, pl.fadeOut)
		pl.musicPlayer.SetVolumeRelative(pl.fadeOut)
	}*/

	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	bgAlpha := 1.0 //pl.dimGlider.GetValue()
	blurVal := 0.0

	cameras := pl.camera.GenRotated(settings.DIVIDES, -2*math.Pi/float64(settings.DIVIDES))

	if settings.Playfield.BlurEnable {
		blurVal = 1 //pl.blurGlider.GetValue()
		if settings.Playfield.UnblurToTheBeat {
			blurVal -= settings.Playfield.UnblurFill * (blurVal) * ( /*pl.Scl*/ 1.4 - 1.0) / (settings.Beat.BeatScale * 0.4)
		}
	}

	pl.background.Draw(pl.progressMs, pl.batch, blurVal*0.3, bgAlpha*0.5, cameras[0])

	pl.counter += timMs

	if pl.counter >= 1000.0/60 {
		pl.fpsC = pl.profiler.GetFPS()
		pl.fpsU = pl.profilerU.GetFPS()

		pl.vol = pl.musicPlayer.GetLevelCombined()
		pl.volAverage = pl.volAverage*0.9 + pl.vol*0.1

		//log.Println(pl.volAverage, pl.vol)

		pl.counter -= 1000.0 / 60
	}

	vprog := 1 - ((pl.vol - pl.volAverage) / 0.5)
	pV := math.Min(1.0, math.Max(0.0, 1.0-(vprog*0.5+pl.beatProgress*0.5)))

	ratio := math.Pow(0.5, timMs/16.6666666666667)

	pl.progress = pl.lastProgress*ratio + (pV)*(1-ratio)
	pl.lastProgress = pl.progress

	pl.batch.Begin()
	pl.batch.SetColor(1, 1, 1, 1)
	pl.batch.SetCamera(mgl32.Ortho(float32(-settings.Graphics.GetWidthF()/2), float32(settings.Graphics.GetWidthF()/2), float32(settings.Graphics.GetHeightF()/2), float32(-settings.Graphics.GetHeightF()/2), 1, -1))

	pl.triangles.Draw(pl.progressMsF, pl.batch)

	pl.visualiser.Draw(pl.progressMsF, pl.batch)

	pl.batch.Flush()

	pl.batch.SetColor(1, 1, 1, 1)

	pl.batch.SetCamera(mgl32.Ortho(float32(-settings.Graphics.GetWidthF()/2), float32(settings.Graphics.GetWidthF()/2), float32(settings.Graphics.GetHeightF()/2), -float32(settings.Graphics.GetHeightF()/2), -1, 1))
	//scl1 := (settings.Graphics.GetHeightF() / float64(pl.WaveTex.Height)) * (settings.Graphics.GetHeightF() / 600) * 0.6

	for i := 0; i < len(pl.waves); i++ {
		wave := pl.waves[i]

		wave.UpdateAndDraw(pl.progressMs, pl.batch)

		if wave.GetEndTime() < pl.progressMsF {
			pl.waves = pl.waves[1:]
			i--
		}
	}

	pl.batch.SetColor(1, 1, 1, 1)

	scl := (pl.cookieSize / 2048.0) * 0.7

	pl.LogoS1.SetScale((1.05 - easing.OutQuad(pl.progress)*0.05) * scl * pl.hover)
	pl.LogoS2.SetScale((1.05 + easing.OutQuad(pl.progress)*0.03) * scl * pl.hover)
	pl.visualiser.SetStartDistance(pl.cookieSize * 0.66 * pl.hover)

	alpha := 0.3
	if pl.bMap.Timings.Current.Kiai {
		alpha = 0.12
	}

	pl.LogoS2.SetAlpha(alpha)

	pl.LogoS1.UpdateAndDraw(pl.progressMs, pl.batch)
	pl.LogoS2.UpdateAndDraw(pl.progressMs, pl.batch)

	if pl.epiGlider.GetValue() > 0 {
		scl := (settings.Graphics.GetWidthF() / float64(pl.Epi.Width)) / 2 * 0.66
		pl.batch.SetScale(scl, scl)
		pl.batch.SetColor(1, 1, 1, pl.epiGlider.GetValue())
		pl.batch.DrawTexture(*pl.Epi)
	}

	pl.leftPulse.UpdateAndDraw(pl.progressMs, pl.batch)
	pl.rightPulse.UpdateAndDraw(pl.progressMs, pl.batch)

	pl.batch.End()

	if pl.start {
		settings.Objects.Colors.Update(timMs)
		settings.Objects.CustomSliderBorderColor.Update(timMs)
		settings.Cursor.Colors.Update(timMs)
	}

	colors1, _ := settings.Cursor.GetColors(settings.DIVIDES, 1 /*pl.Scl*/, 1, 1.0 /*pl.cursorGlider.GetValue()*/)

	scale2 := 1.0

	pl.camera.Update()

	if !settings.Cursor.ScaleToTheBeat {
		scale2 = 1
	}

	if settings.Playfield.BloomEnabled {
		pl.bloomEffect.SetThreshold(settings.Playfield.Bloom.Threshold)
		pl.bloomEffect.SetBlur(settings.Playfield.Bloom.Blur)
		pl.bloomEffect.SetPower(settings.Playfield.Bloom.Power + settings.Playfield.BloomBeatAddition*( /*pl.Scl*/ 1.4-1.0)/(settings.Beat.BeatScale*0.4))
		pl.bloomEffect.Begin()
	}

	pl.cursor.UpdateRenderer()

	gl.BlendFuncSeparate(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA, gl.ONE, gl.ONE_MINUS_SRC_ALPHA)
	gl.BlendEquation(gl.FUNC_ADD)
	pl.batch.SetAdditive(true)
	render.BeginCursorRender()

	pl.batch.SetCamera(pl.camera.GetProjectionView())
	pl.cursor.DrawM(scale2, pl.batch, colors1[0], colors1[0], 0)

	render.EndCursorRender()
	pl.batch.SetAdditive(false)

	if settings.Playfield.BloomEnabled {
		pl.bloomEffect.EndAndRender()
	}

	if settings.DEBUG || settings.FPS {
		pl.batch.Begin()
		pl.batch.SetColor(1, 1, 1, 1)
		pl.batch.SetScale(1, 1)
		pl.batch.SetCamera(pl.scamera.GetProjectionView())

		padDown := 4.0 * (settings.Graphics.GetHeightF() / 1080.0)
		shift := 16.0 * (settings.Graphics.GetHeightF() / 1080.0)
		size := 16.0 * (settings.Graphics.GetHeightF() / 1080.0)

		if settings.DEBUG {
			pl.font.Draw(pl.batch, 0, settings.Graphics.GetHeightF()-size*1.5, size*1.5, pl.mapFullName)
			pl.font.Draw(pl.batch, 0, padDown+shift*5, size, fmt.Sprintf("%0.0f FPS", pl.fpsC))
			pl.font.Draw(pl.batch, 0, padDown+shift*4, size, fmt.Sprintf("%0.2f ms", 1000/pl.fpsC))
			pl.font.Draw(pl.batch, 0, padDown+shift*3, size, fmt.Sprintf("%0.2f ms update", 1000/pl.fpsU))

			time := int(pl.musicPlayer.GetPosition())
			totalTime := int(pl.musicPlayer.GetLength())
			mapTime := int(pl.bMap.HitObjects[len(pl.bMap.HitObjects)-1].GetBasicData().EndTime / 1000)

			pl.font.Draw(pl.batch, 0, padDown+shift*2, size, fmt.Sprintf("%02d:%02d / %02d:%02d (%02d:%02d)", time/60, time%60, totalTime/60, totalTime%60, mapTime/60, mapTime%60))
		} else {
			pl.font.Draw(pl.batch, 0, padDown, size, fmt.Sprintf("%0.0f FPS", pl.fpsC))
		}

		pl.batch.End()
	}

}
