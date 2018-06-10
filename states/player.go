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

var scl float32 = 0.0
var mat mgl32.Mat4

type Player struct {
	bMap *beatmap.BeatMap
	queue2 []objects.BaseObject
	processed []objects.BaseObject
	sliderRenderer *render.SliderRenderer
	lastTime int64
	progressMsF float64
	progressMs int64
	batch *render.SpriteBatch
	cursor *render.Cursor
	circles []*objects.Circle
	sliders []*objects.Slider
	Background *glhf.Texture
	Logo *glhf.Texture
	BgScl bmath.Vector2d
	Cam mgl32.Mat4
	Scl float64
	SclA float64
	CS float64
	h, s, v float64
	fadeOut float64
	fadeIn float64
	start bool
	mus bool
	musicPlayer *audio.Music
	fxBatch *render.FxBatch
	vao *glhf.VertexSlice
	vaoD []float32
	vaoDirty bool
}

func NewPlayer(beatMap *beatmap.BeatMap) *Player {
	player := &Player{}
	render.LoadTextures()
	player.batch = render.NewSpriteBatch()
	player.bMap = beatMap
	log.Println(beatMap.Name + " " + beatMap.Difficulty)
	player.CS = (1.0 - 0.7 * (beatMap.CircleSize - 5) / 5) / 2
	render.CS = player.CS
	render.SetupSlider()

	log.Println(beatMap.Bg)
	var err error
	player.Background, err = utils.LoadTexture(beatMap.Bg)
	player.Logo, err = utils.LoadTexture("assets/textures/logo.png")
	log.Println(err)
	winscl := settings.Graphics.GetAspectRatio()

	if player.Background != nil {
		imScl := float64(player.Background.Width())/float64(player.Background.Height())
		if imScl > winscl {
			player.BgScl = bmath.NewVec2d(1, winscl/imScl)
		} else {
			player.BgScl = bmath.NewVec2d(imScl/winscl, 1)
		}
	}

	player.sliderRenderer = render.NewSliderRenderer()
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

	player.cursor = render.NewCursor()

	scl = float32(settings.Graphics.GetHeightF()*800.0/1080.0)/float32(384)*float32(settings.Playfield.Scale)
	log.Println(scl)
	player.Cam = mgl32.Ortho( -float32(settings.Graphics.GetWidthF())/2, float32(settings.Graphics.GetWidthF())/2 , float32(settings.Graphics.GetHeightF())/2, -float32(settings.Graphics.GetHeightF())/2, 1, -1)

	mat = mgl32.Scale3D(scl, scl, 1)

	player.Scl = 1
	player.h, player.s, player.v = 0.0, 1.0, 1.0
	player.fadeOut = 1.0
	player.fadeIn = 0.0

	musicPlayer := audio.NewMusic(beatMap.Audio)

	go func() {
		time.Sleep(2*time.Second)

		for i := 1; i <= 100; i++ {
			player.fadeIn = float64(i) / 100
			time.Sleep(10*time.Millisecond)
		}
		time.Sleep(500*time.Millisecond)
		player.start = true
		musicPlayer.Play()
	}()

	player.fxBatch = render.NewFxBatch()
	player.vao = player.fxBatch.CreateVao(3*(256+128))
	go func() {
		var last = musicPlayer.GetPosition()

		for {

			player.progressMsF = musicPlayer.GetPosition()*1000

			player.bMap.Update(int64(player.progressMsF), player.cursor)
			player.cursor.Update(player.progressMsF - last)

			last = player.progressMsF

			time.Sleep(time.Millisecond)
		}
	}()

	go func() {
		vertices := make([]float32, (256+128)*3*3)
		oldFFT := make([]float32, 256+128)
		for {

			musicPlayer.Update()
			player.SclA = math.Min(1.4*settings.Beat.BeatScale, math.Max(musicPlayer.GetBeat()*settings.Beat.BeatScale+1.0, 1.0))

			fft := musicPlayer.GetFFT()

			//last := fft[0]

			for i:=0; i < len(oldFFT); i++ {
				fft[i] = float32(math.Log10(float64(fft[i])*40))
				oldFFT[i] = float32(math.Max(0.001, math.Max(math.Min(float64(fft[i]) /** 3*/, float64(oldFFT[i]) + 0.02), float64(oldFFT[i]) - 0.015)))
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

			time.Sleep(40*time.Millisecond)
		}
	}()

	player.musicPlayer = musicPlayer
	return player
}

func (pl *Player) Update() {


	if pl.lastTime < 0 {
		pl.lastTime = utils.GetNanoTime()
	}
	tim := utils.GetNanoTime()
	timMs := float64(tim-pl.lastTime)/1000000.0

	fps := 1000.0/timMs

	if fps < 100 {
		log.Println(fps)
	}

	if pl.start {

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


	//colors := render.GetColors(pl.h, 360.0/float64(settings.DIVIDES), settings.DIVIDES, pl.fadeOut*pl.fadeIn)
	/*offst := 0.0
	if settings.Objects.Colors.FlashToTheBeat {
		offst = settings.Objects.Colors.FlashAmplitude * ((pl.Scl-1.0)/(0.4*settings.Beat.BeatScale))
	}*/


	render.CS = pl.CS
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	pl.batch.Begin()
	pl.batch.SetCamera(mgl32.Ortho( -1, 1 , 1, -1, 1, -1))
	pl.batch.SetColor(1, 1, 1, (0.05+(0.95*(1-pl.fadeIn)))*pl.Scl*pl.fadeOut)
	pl.batch.ResetTransform()
	if pl.Background != nil {
		pl.batch.SetScale(pl.BgScl.X, pl.BgScl.Y)
		pl.batch.DrawUnscaled(bmath.NewVec2d(0, 0), pl.Background)
	}
	/*pl.batch.SetColor(1, 1, 1, 1)
	pl.batch.DrawUnscaled(bmath.NewVec2d(0, 0), render.SliderGradient)*/
	//pl.batch.SetCamera(mgl32.Ortho( -1920/2, 1920/2 , 1080/2, -1080/2, -1, 1))
	//pl.batch.SetScale(0.5, 0.5)
	//pl.batch.SetColor(1, 1, 1, 1-pl.fadeIn)
	//pl.batch.DrawTexture(bmath.NewVec2d(0, 0), pl.Logo)
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
		settings.Cursor.Colors.Update(timMs)
	}

	colors := settings.Objects.Colors.GetColors(settings.DIVIDES, pl.Scl, pl.fadeOut*pl.fadeIn)
	colors1 := settings.Cursor.Colors.GetColors(settings.DIVIDES, pl.Scl, pl.fadeOut*pl.fadeIn)
	scale1 := pl.Scl
	scale2 := pl.Scl

	if !settings.Objects.ScaleToTheBeat {
		scale1 = 1
	}
	
	if !settings.Cursor.ScaleToTheBeat {
		scale2 = 1
	}
	
	if pl.start {

		pl.sliderRenderer.Begin()

		for j:=0; j < settings.DIVIDES; j++ {

			vc := bmath.NewVec2d(0, 1).Rotate(float64(j)*2*math.Pi/float64(settings.DIVIDES))
			lookAt := mgl32.LookAtV(mgl32.Vec3{0,0, 0}, mgl32.Vec3{0,0, -1}, mgl32.Vec3{float32(vc.X), float32(vc.Y), 0})
			pl.sliderRenderer.SetCamera(pl.Cam.Mul4(lookAt).Mul4(mgl32.Translate3D(-512.0*scl/2, -384.0*scl/2, 0)).Mul4(mat))

			pl.sliderRenderer.SetColor(colors[j])

			for i := 0; i < len(pl.sliders); i++ {
				pl.sliderRenderer.SetScale(scale1)
				pl.sliders[i].Render(pl.progressMs, pl.bMap.ARms)
			}

		}

		pl.sliderRenderer.EndAndRender()

		for j:=0; j < settings.DIVIDES; j++ {

			vc := bmath.NewVec2d(0, 1).Rotate(float64(j)*2*math.Pi/float64(settings.DIVIDES))
			lookAt := mgl32.LookAtV(mgl32.Vec3{0,0, 0}, mgl32.Vec3{0,0, -1}, mgl32.Vec3{float32(vc.X), float32(vc.Y), 0})
			pl.batch.SetCamera(pl.Cam.Mul4(lookAt).Mul4(mgl32.Translate3D(-512.0*scl/2, -384.0*scl/2, 0)).Mul4(mat))

			pl.batch.SetScale(scale1 * 64*render.CS, scale1 *64*render.CS)
			pl.batch.Begin()
			for i := len(pl.sliders)-1; i >= 0 && len(pl.sliders) > 0 ; i-- {
				if i < len(pl.sliders) {
					res := pl.sliders[i].RenderOverlay(pl.progressMs, pl.bMap.ARms, colors[j], pl.batch)
					if res {
						pl.sliders = append(pl.sliders[:i], pl.sliders[(i+1):]...)
						i++
					}
				}
			}
			pl.batch.End()
			pl.batch.SetScale(1, 1)

		}

		pl.batch.Begin()
		for j:=0; j < settings.DIVIDES; j++ {

			pl.batch.SetScale(64*render.CS*scale1, 64*render.CS*scale1)

			vc := bmath.NewVec2d(0, 1).Rotate(float64(j)*2*math.Pi/float64(settings.DIVIDES))
			lookAt := mgl32.LookAtV(mgl32.Vec3{0,0, 0}, mgl32.Vec3{0,0, -1}, mgl32.Vec3{float32(vc.X), float32(vc.Y), 0})
			pl.batch.SetCamera(pl.Cam.Mul4(lookAt).Mul4(mgl32.Translate3D(-512.0*scl/2, -384.0*scl/2, 0)).Mul4(mat))

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
		pl.batch.End()

		//gl.BlendFunc(gl.SRC_ALPHA, gl.ONE)

	}

	gl.BlendFuncSeparate(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA, gl.ONE, gl.ONE_MINUS_SRC_ALPHA)
	gl.BlendEquation(gl.FUNC_ADD)
	for j:=0; j < settings.DIVIDES; j++ {

		vc := bmath.NewVec2d(0, 1).Rotate(float64(j)*2*math.Pi/float64(settings.DIVIDES))
		lookAt := mgl32.LookAtV(mgl32.Vec3{0,0, 0}, mgl32.Vec3{0,0, -1}, mgl32.Vec3{float32(vc.X), float32(vc.Y), 0})
		pl.batch.SetCamera(pl.Cam.Mul4(lookAt).Mul4(mgl32.Translate3D(-512.0*scl/2, -384.0*scl/2, 0)).Mul4(mat))
		ind := j-1
		if ind < 0 {
			ind = settings.DIVIDES - 1
		}
		pl.cursor.DrawM(scale2, pl.batch, colors1[j], colors1[ind])

	}

}