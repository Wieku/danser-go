package states

import (
	"danser/beatmap"
	"danser/beatmap/objects"
	"danser/render"
	"time"
	"log"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/faiface/glhf"
	"math"
	"danser/audio"
	"github.com/go-gl/gl/v3.3-core/gl"
	"danser/utils"
	"danser/bmath"
	"danser/settings"

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
	BgScl bmath.Vector2d
	Cam mgl32.Mat4
	Scl float64
	SclA float64
	CS float64
	h, s, v float64
	fadeOut float64
	start bool
	mus bool
	musicPlayer *audio.Music
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
	log.Println(err)
	winscl := 1920.0/1080.0
	imScl := float64(player.Background.Width())/float64(player.Background.Height())
	if imScl > winscl {
		player.BgScl = bmath.NewVec2d(1, winscl/imScl)
	} else {
		player.BgScl = bmath.NewVec2d(imScl/winscl, 1)
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
	//file, err := os.Open(beatMap.Audio)
	//if err != nil {
	//	log.Println(err)
	//} else {



		/*sd, format, err2 := mp3.Decode(file)
		if err2 == nil {
			speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))

			fun := beep.StreamerFunc(func(samples [][2]float64) (n int, ok bool) {
				sd.Stream(samples)
				av := 0.0
				for i := range samples {
					av += (math.Abs(samples[i][0]) + math.Abs(samples[i][1]))/2
				}
				av /= float64(len(samples))
				player.SclA = math.Min(1.1, math.Max(math.Log10(av*100)/math.Sqrt(2)-0.05, 0.9))

				return len(samples), true
			})

			player.effect = &effects.Volume{ fun, 10, -1.5, false}*/

			//go func(){
			//	time.Sleep(5*time.Second)
				/*player.start = true
				speaker.Play(player.effect)*/
				//log.Println("START", time.Now().UnixNano())
			//}()
		//}
	//}



	player.cursor = render.NewCursor()

	scl = float32(800)/float32(384)*3/4
	log.Println(scl)
	player.Cam = mgl32.Ortho( -1920/2, 1920.0/2 , 1080.0/2, -1080/2, 1, -1)

	mat = mgl32.Scale3D(scl, scl, 1)

	//player.Cam = player.Cam.Mul4(mgl32.Translate3D((1920.0-512.0*scl)/2, (1080.0-384.0*scl)/2, 0).Mul4(mat))
	player.Scl = 1
	player.h, player.s, player.v = 0.0, 1.0, 1.0
	player.fadeOut = 1.0

	musicPlayer := audio.NewMusic(beatMap.Audio)

	go func() {
		time.Sleep(5*time.Second)
		musicPlayer.Play()
	}()

	go func() {
		var last = musicPlayer.GetPosition()

		player.start = true

		for {

			player.progressMsF = musicPlayer.GetPosition()*1000

			player.bMap.Update(int64(player.progressMsF), player.cursor)
			player.cursor.Update(player.progressMsF - last)

			last = player.progressMsF

			time.Sleep(time.Millisecond)
		}
	}()
	player.musicPlayer = musicPlayer
	return player
}

func (pl *Player) Update() {


	if pl.lastTime < 0 {
		pl.lastTime = time.Now().UnixNano()
	}
	tim := time.Now().UnixNano()

	timMs := float64(tim-pl.lastTime)/1000000.0

	//log.Println(1000/timMs)

	if pl.start {

		pl.progressMs = int64(pl.progressMsF)


		if pl.Scl < pl.SclA {
			pl.Scl += (pl.SclA-pl.Scl) * timMs/50
		} else if pl.Scl > pl.SclA {
			pl.Scl -= (pl.Scl-pl.SclA) * timMs/50
		}
		pl.Scl = 1
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

	//pl.bMap.Update(pl.progressMs, pl.cursor)

	pl.h += timMs/125
	if pl.h >=360.0 {
		pl.h -= 360.0
	}
	h1 := pl.h+90.0
	if h1 >=360.0 {
		h1 -= 360.0
	}

	if len(pl.bMap.Queue) == 0 {
		pl.fadeOut -= timMs/7500
		if pl.fadeOut < 0.0 {
			pl.fadeOut = 0.0
		}
		pl.musicPlayer.SetVolume(0.1*pl.fadeOut)
	}


	colors := render.GetColors(pl.h, 360.0/float64(settings.DIVIDES), settings.DIVIDES, pl.fadeOut)

	//log.Println(pl.Scl*3)
	render.CS = pl.CS
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	pl.batch.Begin()
	pl.batch.SetCamera(mgl32.Ortho( -1, 1 , 1, -1, 1, -1))
	pl.batch.SetColor(1, 1, 1, 0.05*pl.Scl*pl.fadeOut)
	pl.batch.ResetTransform()
	pl.batch.SetScale(pl.BgScl.X, pl.BgScl.Y)
	pl.batch.DrawUnscaled(bmath.NewVec2d(0, 0), pl.Background)
	pl.batch.End()

	if settings.DIVIDES > 2 {
		pl.sliderRenderer.Begin()
	}

	for j:=0; j < settings.DIVIDES; j++ {

		vc := bmath.NewVec2d(0, 1).Rotate(float64(j)*2*math.Pi/float64(settings.DIVIDES))
		lookAt := mgl32.LookAtV(mgl32.Vec3{0,0, 0}, mgl32.Vec3{0,0, -1}, mgl32.Vec3{float32(vc.X), float32(vc.Y), 0})
		pl.sliderRenderer.SetCamera(pl.Cam.Mul4(lookAt).Mul4(mgl32.Translate3D(-512.0*scl/2, -384.0*scl/2, 0)).Mul4(mat))

		if settings.DIVIDES <= 2 {
			pl.sliderRenderer.Begin()
		}

		pl.sliderRenderer.SetColor(colors[j])

		for i := 0; i < len(pl.sliders); i++ {
			pl.sliderRenderer.SetScale(pl.Scl)
			pl.sliders[i].Render(pl.progressMs, pl.bMap.ARms)
		}

		if settings.DIVIDES <= 2 {
			pl.sliderRenderer.EndAndRender()

			pl.batch.SetCamera(pl.Cam.Mul4(lookAt).Mul4(mgl32.Translate3D(-512.0*scl/2, -384.0*scl/2, 0)).Mul4(mat))
			pl.batch.SetScale(pl.Scl * 64*render.CS, pl.Scl *64*render.CS)
			pl.batch.Begin()
			for i := 0; i < len(pl.sliders); i++ {
				res := pl.sliders[i].RenderOverlay(pl.progressMs, pl.bMap.ARms, colors[j], pl.batch)
				if res {
					pl.sliders = append(pl.sliders[:i], pl.sliders[(i+1):]...)
					i--
				}
			}
			pl.batch.End()
			pl.batch.SetScale(1, 1)
		}
	}

	if settings.DIVIDES > 2 {
		pl.sliderRenderer.EndAndRender()

		for j:=0; j < settings.DIVIDES; j++ {

			vc := bmath.NewVec2d(0, 1).Rotate(float64(j)*2*math.Pi/float64(settings.DIVIDES))
			lookAt := mgl32.LookAtV(mgl32.Vec3{0,0, 0}, mgl32.Vec3{0,0, -1}, mgl32.Vec3{float32(vc.X), float32(vc.Y), 0})
			pl.batch.SetCamera(pl.Cam.Mul4(lookAt).Mul4(mgl32.Translate3D(-512.0*scl/2, -384.0*scl/2, 0)).Mul4(mat))

			pl.batch.SetScale(pl.Scl * 64*render.CS, pl.Scl *64*render.CS)
			pl.batch.Begin()
			for i := 0; i < len(pl.sliders); i++ {
				res := pl.sliders[i].RenderOverlay(pl.progressMs, pl.bMap.ARms, colors[j], pl.batch)
				if res {
					pl.sliders = append(pl.sliders[:i], pl.sliders[(i+1):]...)
					i--
				}
			}
			pl.batch.End()
			pl.batch.SetScale(1, 1)

		}
	}

	pl.batch.Begin()
	for j:=0; j < settings.DIVIDES; j++ {

		pl.batch.SetScale(64*render.CS*pl.Scl, 64*render.CS*pl.Scl)

		vc := bmath.NewVec2d(0, 1).Rotate(float64(j)*2*math.Pi/float64(settings.DIVIDES))
		lookAt := mgl32.LookAtV(mgl32.Vec3{0,0, 0}, mgl32.Vec3{0,0, -1}, mgl32.Vec3{float32(vc.X), float32(vc.Y), 0})
		pl.batch.SetCamera(pl.Cam.Mul4(lookAt).Mul4(mgl32.Translate3D(-512.0*scl/2, -384.0*scl/2, 0)).Mul4(mat))

		for i := len(pl.circles)-1; i >= 0 && len(pl.circles) > 0 ; i-- {
			res := pl.circles[i].Render(pl.progressMs, pl.bMap.ARms, colors[j], pl.batch)
			if res {
				pl.circles = append(pl.circles[:i], pl.circles[(i + 1):]...)
				i++
			}
		}
	}
	pl.batch.End()

	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE)
	gl.BlendEquation(gl.FUNC_ADD)
	for j:=0; j < settings.DIVIDES; j++ {


		vc := bmath.NewVec2d(0, 1).Rotate(float64(j)*2*math.Pi/float64(settings.DIVIDES))
		lookAt := mgl32.LookAtV(mgl32.Vec3{0,0, 0}, mgl32.Vec3{0,0, -1}, mgl32.Vec3{float32(vc.X), float32(vc.Y), 0})
		pl.batch.SetCamera(pl.Cam.Mul4(lookAt).Mul4(mgl32.Translate3D(-512.0*scl/2, -384.0*scl/2, 0)).Mul4(mat))
		pl.cursor.Draw(pl.Scl, pl.batch, colors[j])

	}
}