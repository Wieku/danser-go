package states

import (
	"github.com/wieku/danser/beatmap"
	"github.com/wieku/danser/beatmap/objects"
	"github.com/wieku/danser/render"
	"time"
	"log"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/wieku/glhf"
	"math"
	"github.com/wieku/danser/audio"
	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/wieku/danser/utils"
	"github.com/wieku/danser/bmath"
	"github.com/wieku/danser/settings"
	"github.com/wieku/danser/dance"
	"github.com/wieku/danser/animation"
	"github.com/wieku/danser/render/effects"
	"github.com/wieku/danser/render/texture"
	"github.com/wieku/danser/render/font"
	"fmt"
	"path/filepath"
	"github.com/wieku/danser/render/batches"
	"github.com/wieku/danser/states/components"
)

type Player struct {
	font           *font.Font
	bMap           *beatmap.BeatMap
	queue2         []objects.BaseObject
	processed      []objects.Renderable
	sliderRenderer *render.SliderRenderer
	bloomEffect    *effects.BloomEffect
	lastTime       int64
	progressMsF    float64
	progressMs     int64
	batch          *batches.SpriteBatch
	controller     dance.Controller
	background     *components.Background
	Logo           *texture.TextureRegion
	BgScl          bmath.Vector2d
	Scl            float64
	SclA           float64
	CS             float64
	fxRotation     float64
	fadeOut        float64
	fadeIn         float64
	entry          float64
	start          bool
	mus            bool
	musicPlayer    *audio.Music
	fxBatch        *render.FxBatch
	vao            *glhf.VertexSlice
	vaoD           []float32
	vaoDirty       bool
	rotation       float64
	profiler       *utils.FPSCounter
	profilerU      *utils.FPSCounter

	camera         *bmath.Camera
	scamera        *bmath.Camera
	dimGlider      *animation.Glider
	blurGlider     *animation.Glider
	fxGlider       *animation.Glider
	cursorGlider   *animation.Glider
	playersGlider  *animation.Glider
	counter        float64
	fpsC           float64
	fpsU           float64
	storyboardLoad float64
	mapFullName    string
	Epi            *texture.TextureRegion
	epiGlider      *animation.Glider
	overlay 		components.Overlay
}

func NewPlayer(beatMap *beatmap.BeatMap) *Player {
	player := new(Player)
	render.LoadTextures()
	render.SetupSlider()
	player.batch = batches.NewSpriteBatch()
	player.sliderRenderer = render.NewSliderRenderer()
	player.font = font.GetFont("Roboto Bold")

	player.bMap = beatMap
	player.mapFullName = fmt.Sprintf("%s - %s [%s]", beatMap.Artist, beatMap.Name, beatMap.Difficulty)
	log.Println("Playing:", player.mapFullName)

	player.CS = 32 * (1.0 - 0.7*(beatMap.CircleSize-5)/5) * settings.Objects.CSMult
	render.CS = player.CS

	var err error

	player.Logo, err = utils.LoadTextureToAtlas(render.Atlas, "assets/textures/logo-medium.png")
	player.Epi, err = utils.LoadTextureToAtlas(render.Atlas, "assets/textures/warning.png")

	if err != nil {
		log.Println(err)
	}

	player.background = components.NewBackground(beatMap)

	player.camera = bmath.NewCamera()
	player.camera.SetOsuViewport(int(settings.Graphics.GetWidth()), int(settings.Graphics.GetHeight()), settings.Playfield.Scale)
	player.camera.Update()

	player.scamera = bmath.NewCamera()
	player.scamera.SetViewport(int(settings.Graphics.GetWidth()), int(settings.Graphics.GetHeight()), false)
	player.scamera.SetOrigin(bmath.NewVec2d(settings.Graphics.GetWidthF()/2, settings.Graphics.GetHeightF()/2))
	player.scamera.Update()

	render.Camera = player.camera

	player.bMap.Reset()
	if settings.PLAY {
		player.controller = dance.NewPlayerController()

		player.controller.SetBeatMap(player.bMap)
		player.controller.InitCursors()
		player.overlay = components.NewScoreOverlay(player.controller.(*dance.PlayerController).GetRuleset(), player.controller.GetCursors()[0])
	} else if settings.KNOCKOUT != "" {
		controller := dance.NewReplayController()
		player.controller = controller
		player.controller.SetBeatMap(player.bMap)
		player.controller.InitCursors()
		player.overlay = components.NewKnockoutOverlay(controller.(*dance.ReplayController))
	} else {
		player.controller = dance.NewGenericController()
		player.controller.SetBeatMap(player.bMap)
		player.controller.InitCursors()
	}

	player.lastTime = -1
	player.queue2 = make([]objects.BaseObject, len(player.bMap.Queue))
	copy(player.queue2, player.bMap.Queue)

	log.Println("Music:", beatMap.Audio)

	player.Scl = 1
	player.fxRotation = 0.0
	player.fadeOut = 1.0
	player.fadeIn = 0.0

	player.dimGlider = animation.NewGlider(0.0)
	player.blurGlider = animation.NewGlider(0.0)
	player.fxGlider = animation.NewGlider(0.0)
	player.cursorGlider = animation.NewGlider(0.0)
	player.playersGlider = animation.NewGlider(0.0)

	tmS := float64(player.queue2[0].GetBasicData().StartTime)
	tmE := float64(player.queue2[len(player.queue2)-1].GetBasicData().EndTime)

	player.dimGlider.AddEvent(-1500, -1000, 1.0-settings.Playfield.BackgroundInDim)
	player.blurGlider.AddEvent(-1500, -1000, settings.Playfield.BackgroundInBlur)
	player.fxGlider.AddEvent(-1500, -1000, 1.0-settings.Playfield.SpectrumInDim)
	player.cursorGlider.AddEvent(-1500, -1000, 0.0)
	player.playersGlider.AddEvent(-1500, -1000, 1.0)

	player.dimGlider.AddEvent(tmS-750, tmS-250, 1.0-settings.Playfield.BackgroundDim)
	player.blurGlider.AddEvent(tmS-750, tmS-250, settings.Playfield.BackgroundBlur)
	player.fxGlider.AddEvent(tmS-750, tmS-250, 1.0-settings.Playfield.SpectrumDim)
	player.cursorGlider.AddEvent(tmS-750, tmS-250, 1.0)

	fadeOut := settings.Playfield.FadeOutTime * 1000
	player.dimGlider.AddEvent(tmE, tmE+fadeOut, 0.0)
	player.fxGlider.AddEvent(tmE, tmE+fadeOut, 0.0)
	player.cursorGlider.AddEvent(tmE, tmE+fadeOut, 0.0)
	player.playersGlider.AddEvent(tmE, tmE+fadeOut, 0.0)

	player.epiGlider = animation.NewGlider(0)
	player.epiGlider.AddEvent(0, 500, 1.0)
	player.epiGlider.AddEvent(2500, 3000, 0.0)

	for _, p := range beatMap.Pauses {
		bd := p.GetBasicData()

		if bd.EndTime-bd.StartTime < 1000 {
			continue
		}

		player.dimGlider.AddEvent(float64(bd.StartTime), float64(bd.StartTime)+500, 1.0-settings.Playfield.BackgroundDimBreaks)
		player.blurGlider.AddEvent(float64(bd.StartTime), float64(bd.StartTime)+500, settings.Playfield.BackgroundBlurBreaks)
		player.fxGlider.AddEvent(float64(bd.StartTime), float64(bd.StartTime)+500, 1.0-settings.Playfield.SpectrumDimBreaks)
		if !settings.Cursor.ShowCursorsOnBreaks {
			player.cursorGlider.AddEvent(float64(bd.StartTime), float64(bd.StartTime)+100, 0.0)
		}

		player.dimGlider.AddEvent(float64(bd.EndTime)-500, float64(bd.EndTime), 1.0-settings.Playfield.BackgroundDim)
		player.blurGlider.AddEvent(float64(bd.EndTime)-500, float64(bd.EndTime), settings.Playfield.BackgroundBlur)
		player.fxGlider.AddEvent(float64(bd.EndTime)-500, float64(bd.EndTime), 1.0-settings.Playfield.SpectrumDim)
		player.cursorGlider.AddEvent(float64(bd.EndTime)-100, float64(bd.EndTime), 1.0)
	}

	musicPlayer := audio.NewMusic(filepath.Join(settings.General.OsuSongsDir, beatMap.Dir, beatMap.Audio))

	go func() {
		player.entry = 1
		time.Sleep(time.Duration(settings.Playfield.LeadInTime * float64(time.Second)))

		for i := 0; i <= 300; i++ {
			player.epiGlider.Update(float64(i * 10))
			time.Sleep(10 * time.Millisecond)
		}

		start := -2000.0
		for i := 1; i <= 100; i++ {
			player.entry = float64(i) / 100
			start += 10
			player.dimGlider.Update(start)
			player.blurGlider.Update(start)
			player.fxGlider.Update(start)
			player.cursorGlider.Update(start)
			player.playersGlider.Update(start)
			time.Sleep(10 * time.Millisecond)
		}

		time.Sleep(time.Duration(settings.Playfield.LeadInHold * float64(time.Second)))

		for i := 1; i <= 100; i++ {
			player.fadeIn = float64(i) / 100
			start += 10
			player.dimGlider.Update(start)
			player.blurGlider.Update(start)
			player.fxGlider.Update(start)
			player.cursorGlider.Update(start)
			player.playersGlider.Update(start)
			time.Sleep(10 * time.Millisecond)
		}

		player.start = true
		musicPlayer.Play()
		musicPlayer.SetTempo(settings.SPEED)
		musicPlayer.SetPitch(settings.PITCH)
		//musicPlayer.SetPosition(295)
	}()

	player.fxBatch = render.NewFxBatch()
	player.vao = player.fxBatch.CreateVao(2 * 3 * (256 + 128))
	player.profilerU = utils.NewFPSCounter(60, false)
	go func() {
		var last = musicPlayer.GetPosition()
		var lastT = utils.GetNanoTime()
		for {

			currtime := utils.GetNanoTime()

			player.profilerU.PutSample(1000.0 / (float64(currtime-lastT) / 1000000.0))

			lastT = currtime

			player.progressMsF = musicPlayer.GetPosition()*1000 + float64(settings.Audio.Offset)

			player.bMap.Update(int64(player.progressMsF))
			player.controller.Update(int64(player.progressMsF), player.progressMsF-last)
			if player.overlay != nil {
				player.overlay.Update(int64(player.progressMsF))
			}

			player.background.Update(int64(player.progressMsF))

			last = player.progressMsF

			if player.start && len(player.bMap.Queue) > 0 {
				player.dimGlider.Update(player.progressMsF)
				player.blurGlider.Update(player.progressMsF)
				player.fxGlider.Update(player.progressMsF)
				player.cursorGlider.Update(player.progressMsF)
				player.playersGlider.Update(player.progressMsF)
			}

			time.Sleep(time.Millisecond)
		}
	}()

	go func() {
		vertices := make([]float32, (256+128)*3*3*2)
		oldFFT := make([]float32, 256+128)
		for {

			musicPlayer.Update()
			player.SclA = math.Min(1.4*settings.Beat.BeatScale, math.Max(math.Sin(musicPlayer.GetBeat()*math.Pi/2)*0.4*settings.Beat.BeatScale+1.0, 1.0))

			fft := musicPlayer.GetFFT()

			for i := 0; i < len(oldFFT); i++ {
				fft[i] = fft[i] * float32(math.Pow(float64(i+1), 0.33))
				oldFFT[i] = float32(math.Max(0.001, math.Max(math.Min(float64(fft[i]), float64(oldFFT[i])+0.05), float64(oldFFT[i])-0.025)))

				vI := bmath.NewVec2dRad(float64(i)/float64(len(oldFFT))*4*math.Pi, 0.005)
				vI2 := bmath.NewVec2dRad(float64(i)/float64(len(oldFFT))*4*math.Pi, 0.5)

				poH := bmath.NewVec2dRad(float64(i)/float64(len(oldFFT))*4*math.Pi, float64(oldFFT[i]))

				pLL := vI.Rotate(math.Pi / 2).Add(vI2).Sub(poH.Scl(0.5))
				pLR := vI.Rotate(-math.Pi / 2).Add(vI2).Sub(poH.Scl(0.5))
				pHL := vI.Rotate(math.Pi / 2).Add(poH.Scl(0.5)).Add(vI2)
				pHR := vI.Rotate(-math.Pi / 2).Add(poH.Scl(0.5)).Add(vI2)

				vertices[(i)*18], vertices[(i)*18+1], vertices[(i)*18+2] = pLL.X32(), pLL.Y32(), 0
				vertices[(i)*18+3], vertices[(i)*18+4], vertices[(i)*18+5] = pLR.X32(), pLR.Y32(), 0
				vertices[(i)*18+6], vertices[(i)*18+7], vertices[(i)*18+8] = pHR.X32(), pHR.Y32(), 0
				vertices[(i)*18+9], vertices[(i)*18+10], vertices[(i)*18+11] = pHR.X32(), pHR.Y32(), 0
				vertices[(i)*18+12], vertices[(i)*18+13], vertices[(i)*18+14] = pHL.X32(), pHL.Y32(), 0
				vertices[(i)*18+15], vertices[(i)*18+16], vertices[(i)*18+17] = pLL.X32(), pLL.Y32(), 0

			}

			player.vaoD = vertices
			player.vaoDirty = true

			time.Sleep(15 * time.Millisecond)
		}
	}()
	player.profiler = utils.NewFPSCounter(60, false)
	player.musicPlayer = musicPlayer

	player.bloomEffect = effects.NewBloomEffect(int(settings.Graphics.GetWidth()), int(settings.Graphics.GetHeight()))

	return player
}

func (pl *Player) Show() {

}

func (pl *Player) Draw(delta float64) {
	if pl.lastTime < 0 {
		pl.lastTime = utils.GetNanoTime()
	}
	tim := utils.GetNanoTime()
	timMs := float64(tim-pl.lastTime) / 1000000.0

	pl.profiler.PutSample(1000.0 / timMs)
	fps := pl.profiler.GetFPS()

	if pl.start {

		if fps > 58 && timMs > 18 {
			log.Println("Slow frame detected! Frame time:", timMs, "| Av. frame time:", 1000.0/fps)
		}

		pl.progressMs = int64(pl.progressMsF)

		if pl.Scl < pl.SclA {
			pl.Scl += (pl.SclA - pl.Scl) * timMs / 100
		} else if pl.Scl > pl.SclA {
			pl.Scl -= (pl.Scl - pl.SclA) * timMs / 100
		}

	}

	pl.lastTime = tim

	if len(pl.queue2) > 0 {
		for i := 0; i < len(pl.queue2); i++ {
			if p := pl.queue2[i]; p.GetBasicData().StartTime-15000 <= pl.progressMs {
				if s, ok := p.(*objects.Slider); ok {
					s.InitCurve(pl.sliderRenderer)
				}

				if p := pl.queue2[i]; p.GetBasicData().StartTime-int64(pl.bMap.Preempt) <= pl.progressMs {

					pl.processed = append(pl.processed, p.(objects.Renderable))

					pl.queue2 = pl.queue2[1:]
					i--
				}
			} else {
				break
			}
		}
	}

	pl.fxRotation += timMs / 125
	if pl.fxRotation >= 360.0 {
		pl.fxRotation -= 360.0
	}

	if len(pl.bMap.Queue) == 0 {
		pl.fadeOut -= timMs / (settings.Playfield.FadeOutTime * 1000)
		pl.fadeOut = math.Max(0.0, pl.fadeOut)
		pl.musicPlayer.SetVolumeRelative(pl.fadeOut)
		pl.dimGlider.UpdateD(timMs)
		pl.blurGlider.UpdateD(timMs)
		pl.fxGlider.UpdateD(timMs)
		pl.cursorGlider.UpdateD(timMs)
		pl.playersGlider.UpdateD(timMs)
	}

	render.CS = pl.CS
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	bgAlpha := pl.dimGlider.GetValue()
	blurVal := 0.0

	cameras := pl.camera.GenRotated(settings.DIVIDES, -2*math.Pi/float64(settings.DIVIDES))

	if settings.Playfield.BlurEnable {
		blurVal = pl.blurGlider.GetValue()
		if settings.Playfield.UnblurToTheBeat {
			blurVal -= settings.Playfield.UnblurFill * (blurVal) * (pl.Scl - 1.0) / (settings.Beat.BeatScale * 0.4)
		}
	}

	if settings.Playfield.FlashToTheBeat {
		bgAlpha *= pl.Scl
	}

	pl.background.Draw(pl.progressMs, pl.batch, blurVal, bgAlpha, cameras[0])

	if pl.fxGlider.GetValue() > 0.0 || pl.epiGlider.GetValue() > 0 {
		pl.batch.Begin()
		pl.batch.SetColor(1, 1, 1, pl.fxGlider.GetValue())
		pl.batch.SetCamera(mgl32.Ortho(float32(-settings.Graphics.GetWidthF()/2), float32(settings.Graphics.GetWidthF()/2), float32(settings.Graphics.GetHeightF()/2), float32(-settings.Graphics.GetHeightF()/2), 1, -1))
		if pl.fxGlider.GetValue() > 0.0 {
			scl := (settings.Graphics.GetWidthF() / float64(pl.Logo.Width)) / 4
			pl.batch.SetScale(scl, scl)
			pl.batch.DrawTexture(*pl.Logo)
			pl.batch.SetScale(scl*(1/pl.Scl), scl*(1/pl.Scl))
			pl.batch.SetColor(1, 1, 1, 0.25*pl.fxGlider.GetValue())
			pl.batch.DrawTexture(*pl.Logo)
		}

		if pl.epiGlider.GetValue() > 0 {
			scl := (settings.Graphics.GetWidthF() / float64(pl.Epi.Width)) / 2 * 0.66
			pl.batch.SetScale(scl, scl)
			pl.batch.SetColor(1, 1, 1, pl.epiGlider.GetValue())
			pl.batch.DrawTexture(*pl.Epi)
		}

		pl.batch.End()
	}

	pl.counter += timMs

	if pl.counter >= 1000.0/60 {
		pl.fpsC = pl.profiler.GetFPS()
		pl.fpsU = pl.profilerU.GetFPS()
		pl.counter -= 1000.0 / 60
		if pl.background.GetStoryboard() != nil {
			pl.storyboardLoad = pl.background.GetStoryboard().GetLoad()
		}
	}

	if pl.fxGlider.GetValue() > 0.0 {

		pl.fxBatch.Begin()
		pl.batch.SetCamera(mgl32.Ortho(-1, 1, 1, -1, 1, -1))
		pl.fxBatch.SetColor(1, 1, 1, 0.25*pl.Scl*pl.fxGlider.GetValue())
		pl.vao.Begin()

		if pl.vaoDirty {
			pl.vao.SetVertexData(pl.vaoD)
			pl.vaoDirty = false
		}

		base := mgl32.Ortho(-1920/2, 1920/2, 1080/2, -1080/2, -1, 1).Mul4(mgl32.Scale3D(600, 600, 0)).Mul4(mgl32.HomogRotate3DZ(float32(pl.fxRotation * math.Pi / 180.0)))

		pl.fxBatch.SetTransform(base)
		pl.vao.Draw()

		pl.fxBatch.SetTransform(base.Mul4(mgl32.HomogRotate3DZ(math.Pi)))
		pl.vao.Draw()

		pl.vao.End()
		pl.fxBatch.End()
	}

	if pl.start {
		settings.Objects.Colors.Update(timMs)
		settings.Objects.CustomSliderBorderColor.Update(timMs)
		settings.Cursor.Colors.Update(timMs)
		if settings.Playfield.RotationEnabled {
			pl.rotation += settings.Playfield.RotationSpeed / 1000.0 * timMs
			for pl.rotation > 360.0 {
				pl.rotation -= 360.0
			}

			for pl.rotation < 0.0 {
				pl.rotation += 360.0
			}
		}
	}

	colors := settings.Objects.Colors.GetColors(settings.DIVIDES, pl.Scl, pl.fadeOut*pl.fadeIn)
	colors1 := settings.Cursor.GetColors(settings.DIVIDES, /*settings.TAG*/ len(pl.controller.GetCursors()), pl.Scl, pl.cursorGlider.GetValue())
	colors2 := colors

	if settings.Objects.EnableCustomSliderBorderColor {
		colors2 = settings.Objects.CustomSliderBorderColor.GetColors(settings.DIVIDES, pl.Scl, pl.fadeOut*pl.fadeIn)
	}

	scale1 := pl.Scl
	scale2 := pl.Scl
	rotationRad := (pl.rotation + settings.Playfield.BaseRotation) * math.Pi / 180.0

	pl.camera.SetRotation(-rotationRad)
	pl.camera.Update()

	if !settings.Objects.ScaleToTheBeat {
		scale1 = 1
	}

	if !settings.Cursor.ScaleToTheBeat {
		scale2 = 1
	}

	if settings.Playfield.BloomEnabled {
		pl.bloomEffect.SetThreshold(settings.Playfield.Bloom.Threshold)
		pl.bloomEffect.SetBlur(settings.Playfield.Bloom.Blur)
		pl.bloomEffect.SetPower(settings.Playfield.Bloom.Power + settings.Playfield.BloomBeatAddition*(pl.Scl-1.0)/(settings.Beat.BeatScale*0.4))
		pl.bloomEffect.Begin()
	}

	if pl.start {

		if settings.Objects.SliderMerge {
			pl.sliderRenderer.Begin()

			for j := 0; j < settings.DIVIDES; j++ {
				pl.sliderRenderer.SetCamera(cameras[j])
				ind := j - 1
				if ind < 0 {
					ind = settings.DIVIDES - 1
				}

				for i := len(pl.processed) - 1; i >= 0; i-- {
					if s, ok := pl.processed[i].(*objects.Slider); ok {
						pl.sliderRenderer.SetScale(scale1)
						s.DrawBody(pl.progressMs, colors2[j], colors2[ind], pl.sliderRenderer)
					}
				}
			}

			pl.sliderRenderer.EndAndRender()
		}

		pl.batch.Begin()

		if settings.DIVIDES >= settings.Objects.MandalaTexturesTrigger {
			pl.batch.SetAdditive(true)
		} else {
			pl.batch.SetAdditive(false)
		}

		pl.batch.SetScale(render.CS*scale1, render.CS*scale1)

		for j := 0; j < settings.DIVIDES; j++ {
			if !settings.Objects.SliderMerge {
				pl.sliderRenderer.SetCamera(cameras[j])
			}
			pl.batch.SetCamera(cameras[j])
			ind := j - 1
			if ind < 0 {
				ind = settings.DIVIDES - 1
			}

			for i := len(pl.processed) - 1; i >= 0 && len(pl.processed) > 0; i-- {
				if i < len(pl.processed) {
					if !settings.Objects.SliderMerge {
						if s, ok := pl.processed[i].(*objects.Slider); ok {
							pl.batch.Flush()
							pl.sliderRenderer.Begin()
							pl.sliderRenderer.SetScale(scale1)
							s.DrawBody(pl.progressMs, colors2[j], colors2[ind], pl.sliderRenderer)
							pl.sliderRenderer.EndAndRender()
						}
					}
					res := pl.processed[i].Draw(pl.progressMs, colors[j], pl.batch)
					if res {
						pl.processed = append(pl.processed[:i], pl.processed[(i + 1):]...)
						i++
					}
				}
			}
		}

		if settings.DIVIDES < settings.Objects.MandalaTexturesTrigger && settings.Objects.DrawApproachCircles {
			pl.batch.Flush()

			for j := 0; j < settings.DIVIDES; j++ {

				pl.batch.SetCamera(cameras[j])

				for i := len(pl.processed) - 1; i >= 0 && len(pl.processed) > 0; i-- {
					pl.processed[i].DrawApproach(pl.progressMs, colors[j], pl.batch)
				}
			}
		}

		pl.batch.SetScale(1, 1)
		pl.batch.End()
	}

	for _, g := range pl.controller.GetCursors() {
		g.UpdateRenderer()
	}

	gl.BlendFuncSeparate(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA, gl.ONE, gl.ONE_MINUS_SRC_ALPHA)
	gl.BlendEquation(gl.FUNC_ADD)
	pl.batch.SetAdditive(true)
	render.BeginCursorRender()
	for j := 0; j < settings.DIVIDES; j++ {

		pl.batch.SetCamera(cameras[j])

		for i, g := range pl.controller.GetCursors() {
			if pl.overlay != nil && pl.overlay.IsBroken(g) {
				continue
			}

			ind := j*len(pl.controller.GetCursors()) + i - 1
			if ind < 0 {
				ind = settings.DIVIDES*len(pl.controller.GetCursors()) - 1
			}

			g.DrawM(scale2, pl.batch, colors1[j*len(pl.controller.GetCursors())+i], colors1[ind])
		}

	}
	render.EndCursorRender()
	pl.batch.SetAdditive(false)

	if pl.overlay != nil {
		pl.batch.Begin()
		pl.batch.SetScale(1, 1)
		pl.overlay.DrawNormal(pl.batch, colors1, pl.playersGlider.GetValue())
		pl.batch.SetCamera(pl.scamera.GetProjectionView())

		pl.overlay.DrawHUD(pl.batch, colors1, pl.playersGlider.GetValue())

		pl.batch.End()
	}

	if settings.Playfield.BloomEnabled {
		pl.bloomEffect.EndAndRender()
	}

	if settings.DEBUG || settings.FPS {
		pl.batch.Begin()
		pl.batch.SetColor(1, 1, 1, 1)
		pl.batch.SetScale(1, 1)
		pl.batch.SetCamera(pl.scamera.GetProjectionView())

		padDown := 4.0*(settings.Graphics.GetHeightF()/1080.0)
		shift := 16.0*(settings.Graphics.GetHeightF()/1080.0)
		size := 16.0*(settings.Graphics.GetHeightF()/1080.0)

		if settings.DEBUG {
			pl.font.Draw(pl.batch, 0, settings.Graphics.GetHeightF()-size*1.5, size*1.5, pl.mapFullName)
			pl.font.Draw(pl.batch, 0, padDown+shift*5, size, fmt.Sprintf("%0.0f FPS", pl.fpsC))
			pl.font.Draw(pl.batch, 0, padDown+shift*4, size, fmt.Sprintf("%0.2f ms", 1000/pl.fpsC))
			pl.font.Draw(pl.batch, 0, padDown+shift*3, size, fmt.Sprintf("%0.2f ms update", 1000/pl.fpsU))

			time := int(pl.musicPlayer.GetPosition())
			totalTime := int(pl.musicPlayer.GetLength())
			mapTime := int(pl.bMap.HitObjects[len(pl.bMap.HitObjects)-1].GetBasicData().EndTime / 1000)

			pl.font.Draw(pl.batch, 0, padDown+shift*2, size, fmt.Sprintf("%02d:%02d / %02d:%02d (%02d:%02d)", time/60, time%60, totalTime/60, totalTime%60, mapTime/60, mapTime%60))
			pl.font.Draw(pl.batch, 0, padDown+shift, size, fmt.Sprintf("%d(*%d) hitobjects, %d total", len(pl.processed), settings.DIVIDES, len(pl.bMap.HitObjects)))

			if storyboard := pl.background.GetStoryboard(); storyboard != nil {
				pl.font.Draw(pl.batch, 0, padDown, size, fmt.Sprintf("%d storyboard sprites (%0.2fx load), %d in queue (%d total)", storyboard.GetProcessedSprites(), pl.storyboardLoad, storyboard.GetQueueSprites(), storyboard.GetTotalSprites()))
			} else {
				pl.font.Draw(pl.batch, 0, padDown, size, "No storyboard")
			}
		} else {
			pl.font.Draw(pl.batch, 0, padDown, size, fmt.Sprintf("%0.0f FPS", pl.fpsC))
		}

		pl.batch.End()
	}

}
