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
	"os"
)

type Player struct {
	bMap *beatmap.BeatMap
	queue2 []objects.BaseObject
	processed []objects.BaseObject
	sliderRenderer *render.SliderRenderer
	blurEffect *render.BlurEffect
	bloomEffect *render.BloomEffect
	lastTime int64
	progressMsF float64
	progressMs int64
	batch *render.SpriteBatch
	controller dance.Controller
	circles []*objects.Circle
	sliders []*objects.Slider
	Background *glhf.Texture
	Logo *glhf.Texture
	BgScl bmath.Vector2d
	Scl float64
	SclA float64
	CS float64
	h, s, v float64
	fadeOut float64
	fadeIn float64
	entry float64
	start bool
	mus bool
	musicPlayer *audio.Music
	fxBatch *render.FxBatch
	vao *glhf.VertexSlice
	vaoD []float32
	vaoDirty bool
	rotation float64
	profiler *utils.FPSCounter

	camera *bmath.Camera
	dimGlider *animation.Glider
	blurGlider *animation.Glider
	fxGlider *animation.Glider
	cursorGlider *animation.Glider
}

func NewPlayer(beatMap *beatmap.BeatMap) *Player {
	player := &Player{}
	render.LoadTextures()
	player.batch = render.NewSpriteBatch()
	player.bMap = beatMap
	log.Println(beatMap.Name + " " + beatMap.Difficulty)
	player.CS = (1.0 - 0.7 * (beatMap.CircleSize - 5) / 5) / 2 * settings.Objects.CSMult
	render.CS = player.CS
	render.SetupSlider()

	log.Println(beatMap.Bg)
	var err error
	player.Background, err = utils.LoadTexture(settings.General.OsuSongsDir + string(os.PathSeparator) + beatMap.Dir + string(os.PathSeparator) + beatMap.Bg)
	player.Logo, err = utils.LoadTexture("assets/textures/logo-medium.png")
	log.Println(err)
	winscl := settings.Graphics.GetAspectRatio()

	if player.Background != nil {
		gl.ActiveTexture(gl.TEXTURE31)
		player.Background.Begin()
		player.blurEffect = render.NewBlurEffect(player.Background.Width(), player.Background.Height()/*int(settings.Graphics.GetHeight())*/)
		player.blurEffect.SetBlur(0.0, 0.0)
		imScl := float64(player.Background.Width())/float64(player.Background.Height())
		if imScl < winscl {
			player.BgScl = bmath.NewVec2d(1, winscl/imScl)
		} else {
			log.Println(winscl/imScl)
			player.BgScl = bmath.NewVec2d(imScl/winscl, 1)
		}
	}

	player.sliderRenderer = render.NewSliderRenderer()

	scl := (settings.Graphics.GetHeightF()*900.0/1080.0)/float64(384)*settings.Playfield.Scale

	osuAspect := 512.0/384.0
	screenAspect := settings.Graphics.GetWidthF()/settings.Graphics.GetHeightF()

	if osuAspect > screenAspect {
		scl = (settings.Graphics.GetWidthF()*900.0/1080.0)/float64(512)*settings.Playfield.Scale
	}

	player.camera = &bmath.Camera{}
	player.camera.SetViewport(int(settings.Graphics.GetWidth()), int(settings.Graphics.GetHeight()), true)
	player.camera.SetOrigin(bmath.NewVec2d(512.0 / 2, 384.0 / 2))
	player.camera.SetScale(bmath.NewVec2d(scl, scl))
	player.camera.Update()

	render.Camera = player.camera

	player.bMap.Reset()

	player.controller = dance.NewGenericController()
	player.controller.SetBeatMap(player.bMap)
	player.controller.InitCursors()

	player.lastTime = -1
	player.queue2 = make([]objects.BaseObject, len(player.bMap.Queue))
	copy(player.queue2, player.bMap.Queue)


	for _, o := range player.queue2 {
		if s, ok := o.(*objects.Slider); ok {
			s.InitCurve(player.sliderRenderer)
		}
	}
	player.start = false
	player.mus = false
	log.Println(beatMap.Audio)

	player.Scl = 1
	player.h, player.s, player.v = 0.0, 1.0, 1.0
	player.fadeOut = 1.0
	player.fadeIn = 0.0

	player.dimGlider = animation.NewGlider(0.0)
	player.blurGlider = animation.NewGlider(0.0)
	player.fxGlider = animation.NewGlider(0.0)
	player.cursorGlider = animation.NewGlider(0.0)

	tmS := float64(player.queue2[0].GetBasicData().StartTime)
	tmE := float64(player.queue2[len(player.queue2)-1].GetBasicData().EndTime)

	player.dimGlider.AddEvent(-1500, -1000, 1.0 - settings.Playfield.BackgroundInDim)
	player.blurGlider.AddEvent(-1500, -1000, settings.Playfield.BackgroundInBlur)
	player.fxGlider.AddEvent(-1500, -1000, 1.0 - settings.Playfield.SpectrumInDim)
	player.cursorGlider.AddEvent(-1500, -1000, 0.0)

	player.dimGlider.AddEvent(tmS - 750, tmS - 250, 1.0 - settings.Playfield.BackgroundDim)
	player.blurGlider.AddEvent(tmS - 750, tmS - 250, settings.Playfield.BackgroundBlur)
	player.fxGlider.AddEvent(tmS - 750, tmS - 250, 1.0 - settings.Playfield.SpectrumDim)
	player.cursorGlider.AddEvent(tmS - 750, tmS - 250, 1.0)

	fadeOut := settings.Playfield.FadeOutTime * 1000
	player.dimGlider.AddEvent(tmE, tmE + fadeOut, 0.0)
	player.fxGlider.AddEvent(tmE, tmE + fadeOut, 0.0)
	player.cursorGlider.AddEvent(tmE, tmE + fadeOut, 0.0)

	for _, p := range beatMap.Pauses {
		bd := p.GetBasicData()

		if bd.EndTime - bd.StartTime < 1000 {
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
		player.fxGlider.AddEvent(float64(bd.EndTime)-500, float64(bd.EndTime), 1.0 - settings.Playfield.SpectrumDim)
		player.cursorGlider.AddEvent(float64(bd.EndTime)-100, float64(bd.EndTime), 1.0)
	}

	musicPlayer := audio.NewMusic(settings.General.OsuSongsDir + string(os.PathSeparator) + beatMap.Dir + string(os.PathSeparator) + beatMap.Audio)

	go func() {
		player.entry = 1
		time.Sleep(time.Duration(settings.Playfield.LeadInTime*float64(time.Second)))

		start := -2000.0
		for i := 1; i <= 100; i++ {
			player.entry = float64(i) / 100
			start += 10
			player.dimGlider.Update(start)
			player.blurGlider.Update(start)
			player.fxGlider.Update(start)
			player.cursorGlider.Update(start)
			time.Sleep(10*time.Millisecond)
		}

		time.Sleep(time.Duration(settings.Playfield.LeadInHold*float64(time.Second)))

		for i := 1; i <= 100; i++ {
			player.fadeIn = float64(i) / 100
			start += 10
			player.dimGlider.Update(start)
			player.blurGlider.Update(start)
			player.fxGlider.Update(start)
			player.cursorGlider.Update(start)
			time.Sleep(10*time.Millisecond)
		}

		player.start = true
		musicPlayer.Play()
		musicPlayer.SetTempo(settings.SPEED)
	}()

	player.fxBatch = render.NewFxBatch()
	player.vao = player.fxBatch.CreateVao(2*3*(256+128))
	go func() {
		var last = musicPlayer.GetPosition()

		for {

			player.progressMsF = musicPlayer.GetPosition()*1000+float64(settings.Audio.Offset)

			player.bMap.Update(int64(player.progressMsF))
			player.controller.Update(int64(player.progressMsF), player.progressMsF-last)

			last = player.progressMsF

			if player.start && len(player.bMap.Queue) > 0 {
				player.dimGlider.Update(player.progressMsF)
				player.blurGlider.Update(player.progressMsF)
				player.fxGlider.Update(player.progressMsF)
				player.cursorGlider.Update(player.progressMsF)
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

			for i:=0; i < len(oldFFT); i++ {
				fft[i] = fft[i]*float32(math.Pow(float64(i+1), 0.33))
				oldFFT[i] =float32(math.Max(0.001, math.Max(math.Min(float64(fft[i]), float64(oldFFT[i]) + 0.05), float64(oldFFT[i]) - 0.025)))

				vI := bmath.NewVec2dRad(float64(i)/float64(len(oldFFT))*4*math.Pi, 0.005)
				vI2 := bmath.NewVec2dRad(float64(i)/float64(len(oldFFT))*4*math.Pi, 0.5)

				poH := bmath.NewVec2dRad(float64(i)/float64(len(oldFFT))*4*math.Pi, float64(oldFFT[i]))

				pLL := vI.Rotate(math.Pi/2).Add(vI2).Sub(poH.Scl(0.5))
				pLR := vI.Rotate(-math.Pi/2).Add(vI2).Sub(poH.Scl(0.5))
				pHL := vI.Rotate(math.Pi/2).Add(poH.Scl(0.5)).Add(vI2)
				pHR := vI.Rotate(-math.Pi/2).Add(poH.Scl(0.5)).Add(vI2)

				vertices[(i)*18], vertices[(i)*18+1], vertices[(i)*18+2] = pLL.X32(), pLL.Y32(), 0
				vertices[(i)*18+3], vertices[(i)*18+4], vertices[(i)*18+5] = pLR.X32(), pLR.Y32(), 0
				vertices[(i)*18+6], vertices[(i)*18+7], vertices[(i)*18+8] = pHR.X32(), pHR.Y32(), 0
				vertices[(i)*18+9], vertices[(i)*18+10], vertices[(i)*18+11] = pHR.X32(), pHR.Y32(), 0
				vertices[(i)*18+12], vertices[(i)*18+13], vertices[(i)*18+14] = pHL.X32(), pHL.Y32(), 0
				vertices[(i)*18+15], vertices[(i)*18+16], vertices[(i)*18+17] = pLL.X32(), pLL.Y32(), 0

			}

			player.vaoD = vertices
			player.vaoDirty = true

			time.Sleep(15*time.Millisecond)
		}
	}()
	player.profiler = utils.NewFPSCounter(60, true)
	player.musicPlayer = musicPlayer

	player.bloomEffect = render.NewBloomEffect(int(settings.Graphics.GetWidth()), int(settings.Graphics.GetHeight()))

	return player
}

func (pl *Player) Update() {
	if pl.lastTime < 0 {
		pl.lastTime = utils.GetNanoTime()
	}
	tim := utils.GetNanoTime()
	timMs := float64(tim-pl.lastTime)/1000000.0

	pl.profiler.PutSample(1000.0/timMs)
	fps := pl.profiler.GetFPS()

	if pl.start {

		if timMs > 5000.0/(fps) {
			log.Println("Slow frame detected! Frame time:", timMs, "| Av. frame time:", 1000.0/fps)
		}

		pl.progressMs = int64(pl.progressMsF)


		if pl.Scl < pl.SclA {
			pl.Scl += (pl.SclA-pl.Scl) * timMs/100
		} else if pl.Scl > pl.SclA {
			pl.Scl -= (pl.Scl-pl.SclA) * timMs/100
		}

	}

	pl.lastTime = tim

	if len(pl.queue2) > 0 {
		for i := 0; i < len(pl.queue2); i++ {
			if p := pl.queue2[i]; p.GetBasicData().StartTime-int64(pl.bMap.ARms) <= pl.progressMs {

				if s, ok := p.(*objects.Slider); ok {
					pl.sliders = append(pl.sliders, s)
				}
				if s, ok := p.(*objects.Circle); ok {
					pl.circles = append(pl.circles, s)
				}

				pl.queue2 = pl.queue2[1:]
				i--
			} else {
				break
			}
		}
	}

	pl.h += timMs/125
	if pl.h >=360.0 {
		pl.h -= 360.0
	}

	if len(pl.bMap.Queue) == 0 {
		pl.fadeOut -= timMs/7500
		pl.fadeOut = math.Max(0.0, pl.fadeOut)
		pl.musicPlayer.SetVolumeRelative(pl.fadeOut)
		pl.dimGlider.UpdateD(timMs)
		pl.blurGlider.UpdateD(timMs)
		pl.fxGlider.UpdateD(timMs)
		pl.cursorGlider.UpdateD(timMs)
	}

	render.CS = pl.CS
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	pl.batch.Begin()
	pl.batch.SetCamera(mgl32.Ortho( -1, 1 , 1, -1, 1, -1))
	bgAlpha := pl.dimGlider.GetValue()
	blurVal := 0.0

	if settings.Playfield.BlurEnable {
		blurVal = pl.blurGlider.GetValue()
		if settings.Playfield.UnblurToTheBeat {
			blurVal -= settings.Playfield.UnblurFill*(blurVal)*(pl.Scl-1.0)/(settings.Beat.BeatScale*0.4)
		}
	}

	if settings.Playfield.FlashToTheBeat {
		bgAlpha *= pl.Scl
	}

	pl.batch.SetColor(1, 1, 1, 1)
	pl.batch.ResetTransform()
	if pl.Background != nil {
		if settings.Playfield.BlurEnable {
			pl.blurEffect.SetBlur(blurVal, blurVal)
		}

		pl.batch.SetScale(1, 1)

		if settings.Playfield.BlurEnable {
			pl.blurEffect.Begin()
		} else {
			pl.batch.SetColor(1, 1, 1, bgAlpha)
			pl.batch.SetScale(pl.BgScl.X, pl.BgScl.Y)
		}

		pl.batch.DrawUnit(bmath.NewVec2d(0, 0), 31)

		pl.batch.SetColor(1, 1, 1, bgAlpha)

		if settings.Playfield.BlurEnable {
			texture := pl.blurEffect.EndAndProcess()
			pl.batch.SetScale(pl.BgScl.X, -pl.BgScl.Y)
			pl.batch.DrawUnscaled(bmath.NewVec2d(0, 0), texture)

		}

	}

	if pl.fxGlider.GetValue() > 0.0 {
		pl.batch.SetColor(1, 1, 1, pl.fxGlider.GetValue())
		pl.batch.SetCamera(mgl32.Ortho(float32(-settings.Graphics.GetWidthF()/2), float32(settings.Graphics.GetWidthF()/2), float32(settings.Graphics.GetHeightF()/2), float32(-settings.Graphics.GetHeightF()/2), 1, -1))
		scl := (settings.Graphics.GetWidthF()/float64(pl.Logo.Width()))/4
		pl.batch.SetScale(scl, scl)
		pl.batch.DrawTexture(bmath.NewVec2d(0, 0), pl.Logo)
		pl.batch.SetScale(scl*(1/pl.Scl), scl*(1/pl.Scl))
		pl.batch.SetColor(1, 1, 1, 0.25*pl.fxGlider.GetValue())
		pl.batch.DrawTexture(bmath.NewVec2d(0, 0), pl.Logo)
	}

	pl.batch.End()

	if pl.fxGlider.GetValue() > 0.0 {

		pl.fxBatch.Begin()
		pl.fxBatch.SetColor(1, 1, 1, 0.25*pl.Scl*pl.fxGlider.GetValue())
		pl.vao.Begin()

		if pl.vaoDirty {
			pl.vao.SetVertexData(pl.vaoD)
			pl.vaoDirty = false
		}

		base := mgl32.Ortho(-1920/2, 1920/2, 1080/2, -1080/2, -1, 1).Mul4(mgl32.Scale3D(600, 600, 0)).Mul4(mgl32.HomogRotate3DZ(float32(pl.h * math.Pi / 180.0)))

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
			pl.rotation += settings.Playfield.RotationSpeed/1000.0 * timMs
			for pl.rotation > 360.0 {
				pl.rotation -= 360.0
			}

			for pl.rotation < 0.0 {
				pl.rotation += 360.0
			}
		}
	}

	colors := settings.Objects.Colors.GetColors(settings.DIVIDES, pl.Scl, pl.fadeOut*pl.fadeIn)
	colors1 := settings.Cursor.GetColors(settings.DIVIDES, settings.TAG, pl.Scl, pl.cursorGlider.GetValue())
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

	cameras := pl.camera.GenRotated(settings.DIVIDES, -2*math.Pi/float64(settings.DIVIDES))

	if settings.Playfield.BloomEnabled {
		pl.bloomEffect.SetThreshold(settings.Playfield.Bloom.Threshold)
		pl.bloomEffect.SetBlur(settings.Playfield.Bloom.Blur)
		pl.bloomEffect.SetPower(settings.Playfield.Bloom.Power + settings.Playfield.BloomBeatAddition * (pl.Scl-1.0)/(settings.Beat.BeatScale*0.4))
		pl.bloomEffect.Begin()
	}

	if pl.start {

		pl.sliderRenderer.Begin()

		for j:=0; j < settings.DIVIDES; j++ {
			pl.sliderRenderer.SetCamera(cameras[j])
			ind := j-1
			if ind < 0 {
				ind = settings.DIVIDES - 1
			}

			for i := 0; i < len(pl.sliders); i++ {
				pl.sliderRenderer.SetScale(scale1)
				pl.sliders[i].Render(pl.progressMs, pl.bMap.ARms, colors2[j], colors2[ind], pl.sliderRenderer)
			}
		}

		pl.sliderRenderer.EndAndRender()

		if settings.DIVIDES >= settings.Objects.MandalaTexturesTrigger {
			gl.BlendFunc(gl.SRC_ALPHA, gl.ONE)
		} else {
			gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
		}


		pl.batch.Begin()
		pl.batch.SetScale(64*render.CS*scale1, 64*render.CS*scale1)
		objects.BeginSliderOverlay()
		for j:=0; j < settings.DIVIDES; j++ {

			pl.batch.SetCamera(cameras[j])

			for i := len(pl.sliders)-1; i >= 0 && len(pl.sliders) > 0 ; i-- {
				if i < len(pl.sliders) {
					res := pl.sliders[i].RenderOverlay(pl.progressMs, pl.bMap.ARms, colors[j], pl.batch)
					if res {
						pl.sliders = append(pl.sliders[:i], pl.sliders[(i+1):]...)
						i++
					}
				}
			}

		}

		objects.EndSliderOverlay()
		objects.BeginCircleRender()

		for j:=0; j < settings.DIVIDES; j++ {

			pl.batch.SetCamera(cameras[j])

			for i := len(pl.circles)-1; i >= 0 && len(pl.circles) > 0 ; i-- {
				if i < len(pl.circles) {
					res := pl.circles[i].Render(pl.progressMs, pl.bMap.ARms, colors[j], pl.batch)
					if res {
						pl.circles = append(pl.circles[:i], pl.circles[(i + 1):]...)
						i++
					}
				}
			}
		}

		objects.EndCircleRender()
		pl.batch.SetScale(1, 1)
		pl.batch.End()
	}

	for _, g := range pl.controller.GetCursors() {
		g.UpdateRenderer()
	}

	gl.BlendFuncSeparate(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA, gl.ONE, gl.ONE_MINUS_SRC_ALPHA)
	gl.BlendEquation(gl.FUNC_ADD)
	render.BeginCursorRender()
	for j:=0; j < settings.DIVIDES; j++ {

		pl.batch.SetCamera(cameras[j])

		for i, g := range pl.controller.GetCursors() {
			ind := j*len(pl.controller.GetCursors())+i-1
			if ind < 0 {
				ind = settings.DIVIDES*len(pl.controller.GetCursors()) - 1
			}

			g.DrawM(scale2, pl.batch, colors1[j*len(pl.controller.GetCursors())+i], colors1[ind])
		}

	}
	render.EndCursorRender()

	if settings.Playfield.BloomEnabled {
		pl.bloomEffect.EndAndRender()
	}

}