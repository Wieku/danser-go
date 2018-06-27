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
	cursors []*render.Cursor
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
	player.Background, err = utils.LoadTexture(beatMap.Bg)
	player.Logo, err = utils.LoadTexture("assets/textures/logo.png")
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
	player.cursors = make([]*render.Cursor, settings.TAG)
	for i := range player.cursors {
		player.cursors[i] = render.NewCursor()
	}

	player.bMap.SetCursors(player.cursors)
	player.bMap.Reset()
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

	musicPlayer := audio.NewMusic(beatMap.Audio)

	go func() {
		player.entry = 1
		time.Sleep(time.Duration(settings.Playfield.LeadInTime*float64(time.Second)))

		/*for i := 1; i <= 100; i++ {
			player.entry = float64(i) / 100
			time.Sleep(10*time.Millisecond)
		}

		time.Sleep(2*time.Second)*/

		for i := 1; i <= 100; i++ {
			player.fadeIn = float64(i) / 100
			time.Sleep(10*time.Millisecond)
		}
		player.start = true
		musicPlayer.Play()
		musicPlayer.SetTempo(settings.SPEED)
	}()

	player.fxBatch = render.NewFxBatch()
	player.vao = player.fxBatch.CreateVao(3*(256+128))
	go func() {
		var last = musicPlayer.GetPosition()

		for {

			player.progressMsF = musicPlayer.GetPosition()*1000

			player.bMap.Update(int64(player.progressMsF), nil/*player.cursor*/)
			for _, g := range player.cursors {
				g.Update(player.progressMsF - last)
			}
			//player.cursor.Update(player.progressMsF - last)

			last = player.progressMsF

			time.Sleep(time.Millisecond)
		}
	}()

	go func() {
		vertices := make([]float32, (256+128)*3*3)
		oldFFT := make([]float32, 256+128)
		for {

			musicPlayer.Update()
			player.SclA = math.Min(1.4*settings.Beat.BeatScale, math.Max(math.Sin(musicPlayer.GetBeat()*math.Pi/2)*0.4*settings.Beat.BeatScale+1.0, 1.0))

			fft := musicPlayer.GetFFT()

			//last := fft[0]

			for i:=0; i < len(oldFFT); i++ {
				fft[i] = float32(1.0 - math.Abs(math.Max(-50.0,20*math.Log10(float64(fft[i]))/50)))//*10.0
				oldFFT[i] =float32(math.Max(0.001, math.Max(math.Min(float64(fft[i]) /** 3*/, float64(oldFFT[i]) + 0.02), float64(oldFFT[i]) - 0.0075)))
				angl := 2*float64(i)/float64(len(oldFFT))*math.Pi
				angl1 := 2*(float64(i)/float64(len(oldFFT))-0.01)*math.Pi
				angl2 := 2*(float64(i)/float64(len(oldFFT))+0.01)*math.Pi
				x, y := float32(math.Cos(angl)), float32(math.Sin(angl))
				x1, y1 := float32(math.Cos(angl1)), float32(math.Sin(angl1))
				x2, y2 := float32(math.Cos(angl2)), float32(math.Sin(angl2))

				vertices[(i)*9], vertices[(i)*9+1], vertices[(i)*9+2] = x1*0.01, y1*0.01, 0
				vertices[(i)*9+3], vertices[(i)*9+4], vertices[(i)*9+5] = x2*0.01, y2*0.01, 0
				vertices[(i)*9+6], vertices[(i)*9+7], vertices[(i)*9+8] = x*oldFFT[i], y*oldFFT[i], 0
				/*vertices[(i)*9], vertices[(i)*9+1], vertices[(i)*9+2] = -1*//*+last*//*, 2*float32(i)/float32(len(oldFFT))-1 + 0.2/float32(len(oldFFT)), 0
				vertices[(i)*9+3], vertices[(i)*9+4], vertices[(i)*9+5] = -1, 2*float32(i)/float32(len(oldFFT))-1 - 0.2/float32(len(oldFFT)), 0
				vertices[(i)*9+6], vertices[(i)*9+7], vertices[(i)*9+8] = -1+oldFFT[i]*3, 2*float32(i)/float32(len(oldFFT))-1, 0*/
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
		if p := pl.queue2[0]; p.GetBasicData().StartTime-int64(pl.bMap.ARms) <= pl.progressMs {

			if s, ok := p.(*objects.Slider); ok {
				pl.sliders = append(pl.sliders, s)
			}
			if s, ok := p.(*objects.Circle); ok {
				pl.circles = append(pl.circles, s)
			}

			pl.queue2 = pl.queue2[1:]
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
	}

	render.CS = pl.CS
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	pl.batch.Begin()
	pl.batch.SetCamera(mgl32.Ortho( -1, 1 , 1, -1, 1, -1))
	bgAlpha := ((1.0-settings.Playfield.BackgroundDim)+((settings.Playfield.BackgroundDim - settings.Playfield.BackgroundInDim)*(1-pl.fadeIn)))*pl.fadeOut*pl.entry
	blurVal := 0.0

	if settings.Playfield.BlurEnable {
		blurVal = (settings.Playfield.BackgroundBlur - (settings.Playfield.BackgroundBlur-settings.Playfield.BackgroundInBlur)*(1-pl.fadeIn))*pl.fadeOut
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

	pl.batch.End()

	/*pl.fxBatch.Begin()
	pl.fxBatch.SetColor(1, 1, 1, 0.12*pl.Scl*pl.fadeOut)
	pl.vao.Begin()

	if pl.vaoDirty {
		pl.vao.SetVertexData(pl.vaoD)

		pl.vaoDirty = false
	}

	base := mgl32.Ortho( -1920/2, 1920/2 , 1080/2, -1080/2, -1, 1).Mul4(mgl32.Scale3D(600, 600, 0)).Mul4(mgl32.HomogRotate3DZ(float32(pl.h*math.Pi/180.0)))

	pl.fxBatch.SetTransform(base)
	pl.vao.Draw()

	pl.fxBatch.SetTransform(base.Mul4(mgl32.HomogRotate3DZ(math.Pi)))
	pl.vao.Draw()

	pl.vao.End()
	pl.fxBatch.End()*/

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
	colors1 := settings.Cursor.Colors.GetColors(settings.DIVIDES*len(pl.cursors), pl.Scl, pl.fadeOut*pl.fadeIn)
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

			for i := 0; i < len(pl.sliders); i++ {
				pl.sliderRenderer.SetScale(scale1)
				pl.sliders[i].Render(pl.progressMs, pl.bMap.ARms, colors2[j], pl.sliderRenderer)
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

	for _, g := range pl.cursors {
		g.UpdateRenderer()
	}

	gl.BlendFuncSeparate(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA, gl.ONE, gl.ONE_MINUS_SRC_ALPHA)
	gl.BlendEquation(gl.FUNC_ADD)
	render.BeginCursorRender()
	for j:=0; j < settings.DIVIDES; j++ {

		pl.batch.SetCamera(cameras[j])

		for i, g := range pl.cursors {
			ind := j*len(pl.cursors)+i-1
			if ind < 0 {
				ind = settings.DIVIDES*len(pl.cursors) - 1
			}

			g.DrawM(scale2, pl.batch, colors1[j*len(pl.cursors)+i], colors1[ind])
		}

	}
	render.EndCursorRender()

	if settings.Playfield.BloomEnabled {
		pl.bloomEffect.EndAndRender()
	}

}