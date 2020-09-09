package states

import (
	"fmt"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/wieku/danser-go/app/audio"
	"github.com/wieku/danser-go/app/beatmap"
	"github.com/wieku/danser-go/app/beatmap/objects"
	"github.com/wieku/danser-go/app/bmath"
	"github.com/wieku/danser-go/app/dance"
	"github.com/wieku/danser-go/app/discord"
	"github.com/wieku/danser-go/app/graphics/sliderrenderer"
	"github.com/wieku/danser-go/app/render"
	"github.com/wieku/danser-go/app/render/batches"
	"github.com/wieku/danser-go/app/render/effects"
	"github.com/wieku/danser-go/app/render/font"
	"github.com/wieku/danser-go/app/render/gui/drawables"
	"github.com/wieku/danser-go/app/render/sprites"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/app/states/components"
	"github.com/wieku/danser-go/app/utils"
	"github.com/wieku/danser-go/framework/bass"
	"github.com/wieku/danser-go/framework/frame"
	"github.com/wieku/danser-go/framework/graphics/texture"
	"github.com/wieku/danser-go/framework/math/easing"
	"github.com/wieku/danser-go/framework/math/glider"
	"github.com/wieku/danser-go/framework/qpc"
	"log"
	"math"
	"path/filepath"
	"time"
)

type Player struct {
	font        *font.Font
	bMap        *beatmap.BeatMap
	queue2      []objects.BaseObject
	processed   []objects.Renderable
	bloomEffect *effects.BloomEffect
	lastTime    int64
	progressMsF float64
	progressMs  int64
	batch       *batches.SpriteBatch
	controller  dance.Controller
	background  *components.Background
	BgScl       bmath.Vector2d
	Scl         float64
	SclA        float64
	CS          float64
	fadeOut     float64
	fadeIn      float64
	entry       float64
	start       bool
	mus         bool
	musicPlayer *bass.Track
	rotation    float64
	profiler    *frame.Counter
	profilerU   *frame.Counter

	camera         *bmath.Camera
	camera1        *bmath.Camera
	scamera        *bmath.Camera
	dimGlider      *glider.Glider
	blurGlider     *glider.Glider
	fxGlider       *glider.Glider
	cursorGlider   *glider.Glider
	playersGlider  *glider.Glider
	unfold         *glider.Glider
	counter        float64
	fpsC           float64
	fpsU           float64
	storyboardLoad float64
	mapFullName    string
	Epi            *texture.TextureRegion
	epiGlider      *glider.Glider
	overlay        components.Overlay
	velocity       float64
	hGlider        *glider.Glider
	vGlider        *glider.Glider
	oGlider        *glider.Glider
	flashGlider    *glider.Glider
	danserGlider   *glider.Glider
	resnadGlider   *glider.Glider
	blur           *effects.BlurEffect
	lastFromQueue  objects.BaseObject
	x              float64
	y              float64

	currentBeatVal float64
	lastBeatLength float64
	lastBeatStart  float64
	beatProgress   float64
	lastBeatProg   int64

	progress, lastProgress float64
	LogoS1                 *sprites.Sprite
	LogoS2                 *sprites.Sprite

	vol        float64
	volAverage float64
	cookieSize float64
	visualiser *drawables.Visualiser
}

type hsv struct {
	h, s, v float64
}

var hsvarray = []hsv{
	hsv{3, 0.88, 0.79},
	hsv{295, 0.85, 0.82},
	hsv{251, 0.80, 0.73},
	hsv{209, 0.85, 0.82},
	hsv{165, 0.85, 0.78},
}

var hsvindex = 0
var hsvDir = 1

func NewPlayer(beatMap *beatmap.BeatMap) *Player {
	player := new(Player)
	render.LoadTextures()
	player.batch = batches.NewSpriteBatch()
	player.font = font.GetFont("Exo 2 Bold")

	discord.SetMap(beatMap)

	player.bMap = beatMap
	player.mapFullName = fmt.Sprintf("%s - %s [%s]", beatMap.Artist, beatMap.Name, beatMap.Difficulty)
	log.Println("Playing:", player.mapFullName)

	player.CS = beatMap.Diff.CircleRadius * settings.Objects.CSMult

	var err error

	LogoT, err := utils.LoadTextureToAtlas(render.Atlas, "assets/textures/coinbig.png")
	player.LogoS1 = sprites.NewSpriteSingle(LogoT, 0, bmath.NewVec2d(0, 0), bmath.NewVec2d(0, 0))
	player.LogoS2 = sprites.NewSpriteSingle(LogoT, 0, bmath.NewVec2d(0, 0), bmath.NewVec2d(0, 0))

	if settings.Graphics.GetWidthF() > settings.Graphics.GetHeightF() {
		player.cookieSize = 0.5 * settings.Graphics.GetHeightF()
	} else {
		player.cookieSize = 0.5 * settings.Graphics.GetWidthF()
	}

	player.Epi, err = utils.LoadTextureToAtlas(render.Atlas, "assets/textures/warning.png")

	if err != nil {
		log.Println(err)
	}

	player.background = components.NewBackground(beatMap, 0.1, true)

	player.camera = bmath.NewCamera()
	player.camera.SetOsuViewport(int(settings.Graphics.GetWidth()), int(settings.Graphics.GetHeight()), settings.Playfield.Scale, settings.Playfield.OsuShift)
	//player.camera.SetOrigin(bmath.NewVec2d(256, 192.0-5))
	player.camera.Update()

	player.camera1 = bmath.NewCamera()
	player.camera1.SetOsuViewport(int(settings.Graphics.GetWidth()), int(settings.Graphics.GetHeight()), settings.Playfield.Scale, false)
	player.camera1.Update()

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
	} else if settings.KNOCKOUT {
		controller := dance.NewReplayController()
		player.controller = controller
		player.controller.SetBeatMap(player.bMap)
		player.controller.InitCursors()

		if settings.PLAYERS == 1 {
			player.overlay = components.NewScoreOverlay(player.controller.(*dance.ReplayController).GetRuleset(), player.controller.GetCursors()[0])
		} else {
			player.overlay = components.NewKnockoutOverlay(controller.(*dance.ReplayController))
		}
	} else {
		player.controller = dance.NewGenericController()
		player.controller.SetBeatMap(player.bMap)
		player.controller.InitCursors()
	}

	player.lastTime = -1
	player.queue2 = make([]objects.BaseObject, len(player.bMap.Queue))

	copy(player.queue2, player.bMap.Queue)

	log.Println("Track:", beatMap.Audio)

	player.Scl = 1
	player.fadeOut = 1.0
	player.fadeIn = 0.0

	player.dimGlider = glider.NewGlider(0.0)
	player.blurGlider = glider.NewGlider(0.0)
	player.fxGlider = glider.NewGlider(0.0)
	if _, ok := player.overlay.(*components.ScoreOverlay); !ok {
		player.cursorGlider = glider.NewGlider(0.0)
	} else {
		player.cursorGlider = glider.NewGlider(1.0)
	}
	player.playersGlider = glider.NewGlider(0.0)

	tmS := float64(player.queue2[0].GetBasicData().StartTime)
	tmE := float64(player.queue2[len(player.queue2)-1].GetBasicData().EndTime)

	player.dimGlider.AddEvent(-1500, -1000, 1.0-settings.Playfield.BackgroundInDim)
	player.blurGlider.AddEvent(-1500, -1000, settings.Playfield.BackgroundInBlur)
	player.fxGlider.AddEvent(-1500, -1000, 1.0-settings.Playfield.SpectrumInDim)
	if _, ok := player.overlay.(*components.ScoreOverlay); !ok {
		player.cursorGlider.AddEvent(-1500, -1000, 0.0)
	}
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

	player.epiGlider = glider.NewGlider(0)
	player.epiGlider.AddEvent(0, 500, 1.0)
	player.epiGlider.AddEvent(4500, 5000, 0.0)

	player.hGlider = glider.NewGlider(1)
	player.vGlider = glider.NewGlider(1)
	player.oGlider = glider.NewGlider(-36)
	player.flashGlider = glider.NewGlider(1)
	player.danserGlider = glider.NewGlider(0.0)
	player.danserGlider.AddEventS(151562, 151856, 0.0, 1.0)
	//player.danserGlider.AddEventS(187000, 189000, 0.0, 1.0)
	player.resnadGlider = glider.NewGlider(0.0)
	player.resnadGlider.AddEventS(221432, 226542, 0.0, 1.0)

	player.unfold = glider.NewGlider(1)

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

	musicPlayer := bass.NewTrack(filepath.Join(settings.General.OsuSongsDir, beatMap.Dir, beatMap.Audio))
	player.background.SetTrack(musicPlayer)
	player.visualiser = drawables.NewVisualiser(player.cookieSize*0.66, player.cookieSize*2, bmath.NewVec2d(0, 0))
	player.visualiser.SetTrack(musicPlayer)

	audio.AddListener(func(sampleSet int, hitsoundIndex, index int, volume float64, objNum int64) {
		//_, isSlider := player.bMap.HitObjects[objNum].(*objects.Slider)
		startTime := player.progressMsF
		endTime := player.progressMsF + 500
		/*if isSlider{
			startTime = float64(s.GetBasicData().StartTime)
			endTime = float64(s.GetBasicData().EndTime)
		}*/
		if index == 2 && hitsoundIndex == 3 {
			if hsvDir < 0 {
				hsvindex--
			} else {
				hsvindex++
			}
			if hsvindex > 4 {
				hsvindex = 0
			} else if hsvindex < 0 {
				hsvindex = 4
			}
			col := hsvarray[hsvindex]
			/*settings.Cursor.Colors.BaseColor.Hue = col.h
			settings.Objects.Colors.BaseColor.Hue = col.h
			settings.Cursor.Colors.BaseColor.Value = col.v
			settings.Objects.Colors.BaseColor.Value = col.v*/
			//player.hGlider.Reset()
			//player.hGlider.AddEvent(startTime, endTime, col.h)
			//player.vGlider.Reset()
			//player.vGlider.AddEvent(startTime, endTime, col.v)
			player.flashGlider.Reset()
			player.flashGlider.AddEventS(startTime, endTime, 0.0, col.s)
			//println("hat", sampleSet, index)
		} // else if hitsoundIndex == 1 && player.bMap.HitObjects[objNum].GetBasicData().NewCombo {
		//	//prev := hsvarray[hsvindex]
		//	/*if hsvDir < 0 {
		//		hsvindex--
		//	} else {
		//		hsvindex++
		//	}*/
		//	hsvindex++
		//	if hsvindex > 4 {
		//		hsvindex = 0
		//	} else if hsvindex < 0 {
		//		hsvindex = 4
		//	}
		//
		//	hsvDir *= -1.0
		//
		//	col := hsvarray[hsvindex]
		//	/*settings.Cursor.Colors.BaseColor.Hue = col.h
		//	settings.Objects.Colors.BaseColor.Hue = col.h
		//	settings.Cursor.Colors.BaseColor.Value = col.v
		//	settings.Objects.Colors.BaseColor.Value = col.v*/
		//	//println("clap", sampleSet, index)
		//
		//	player.hGlider.Reset()
		//	player.hGlider.AddEvent(startTime, endTime, col.h)
		//	player.vGlider.Reset()
		//	player.vGlider.AddEvent(startTime, endTime, col.v)
		//
		//	player.oGlider.Reset()
		//	if hsvDir > 0 {
		//		player.oGlider.AddEvent(startTime, endTime, 0.0)
		//	} else {
		//		player.oGlider.AddEvent(startTime, endTime, -36)
		//	}
		//
		//	/*player.flashGlider.Reset()
		//	player.flashGlider.AddEvent(startTime, endTime, col.s)*/
		//}
	})

	player.background.Update(0, settings.Graphics.GetWidthF()/2, settings.Graphics.GetHeightF()/2)

	go func() {
		player.entry = 1
		time.Sleep(time.Duration(settings.Playfield.LeadInTime * float64(time.Second)))

		if settings.Playfield.ShowWarning {
			for i := 0; i <= 500; i++ {
				player.epiGlider.Update(float64(i * 10))
				time.Sleep(10 * time.Millisecond)
			}
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
			//player.background.Update(int64(start), 0, 0)
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
			//player.background.Update(int64(start), 0, 0)
			time.Sleep(10 * time.Millisecond)
		}

		musicPlayer.Play()
		musicPlayer.SetTempo(settings.SPEED)
		musicPlayer.SetPitch(settings.PITCH)
		if ov, ok := player.overlay.(*components.ScoreOverlay); ok {
			ov.SetMusic(musicPlayer)
		}
		//musicPlayer.SetPosition(1 * 30)
		discord.SetDuration(int64(musicPlayer.GetLength() * 1000 / settings.SPEED))
		if player.overlay == nil {
			discord.UpdateDance(settings.TAG, settings.DIVIDES)
		}
		player.start = true
	}()

	player.profilerU = frame.NewCounter()
	limiter := frame.NewLimiter(5000)
	go func() {
		var last = musicPlayer.GetPosition()
		var lastT = qpc.GetNanoTime()
		firstT := lastT
		//lastPos := bmath.NewVec2f(0, 0)
		//rotation := 0.0
		//dstRotation := 0.0
		for {
			_ = firstT
			//player.background.Update((lastT-firstT)/1000000, 0, 0)
			currtime := qpc.GetNanoTime()

			player.profilerU.PutSample(float64(currtime-lastT) / 1000000.0)
			if player.start {

				if musicPlayer.GetState() == bass.MUSIC_STOPPED {
					player.progressMsF += float64(currtime-lastT) / 1000000.0
				} else {
					player.progressMsF = musicPlayer.GetPosition()*1000 + float64(settings.Audio.Offset)
				}

				if _, ok := player.controller.(*dance.GenericController); ok {
					player.bMap.Update(int64(player.progressMsF))
				}

				player.controller.Update(int64(player.progressMsF), player.progressMsF-last)
				if player.overlay != nil {
					player.overlay.Update(int64(player.progressMsF))
				}

				//p := player.controller.GetCursors()[0].Position
				//player.camera.SetOrigin(p.Copy64())
				//if p.Dst(lastPos) > 5 {
				//	rot := float64(lastPos.AngleRV(p))
				//	//player.camera.SetRotation(rot-math.Pi)
				//	dstRotation = rot
				//	lastPos = p
				//}
				//rotation += (dstRotation-rotation) * (player.progressMsF-last) / 300
				////player.camera.SetRotation(rotation)
				//player.camera.Update()

				/*if player.lastFromQueue != nil && len(player.bMap.Queue) > 0 {


				}

				if len(player.bMap.Queue) > 0 && player.bMap.Queue[0].GetBasicData().StartTime > int64(player.progressMsF) {
					player.lastFromQueue = player.bMap.Queue[0]
					prog := int64(player.progressMsF)
				}

				if player.lastFromQueue.GetBasicData().*/

				bTime := player.bMap.Timings.Current.BaseBpm

				if bTime != player.lastBeatLength {
					player.lastBeatLength = bTime
					player.lastBeatStart = float64(player.bMap.Timings.Current.Time)
					player.lastBeatProg = -1
				}

				if int64(float64(player.progressMsF-player.lastBeatStart)/player.lastBeatLength) > player.lastBeatProg {
					player.lastBeatProg++
				}

				player.beatProgress = float64(player.progressMsF-player.lastBeatStart)/player.lastBeatLength - float64(player.lastBeatProg)
				player.visualiser.Update(player.progressMsF)

				//crsr := player.controller.GetCursors()[0].Position

				player.background.Update(int64(player.progressMsF) /*crsr.X*settings.Graphics.GetHeightF()/384, crsr.Y*settings.Graphics.GetHeightF()/384*/, 0, 0)

				last = player.progressMsF

				if player.start && len(player.bMap.Queue) > 0 {
					player.dimGlider.Update(player.progressMsF)
					player.blurGlider.Update(player.progressMsF)
					player.fxGlider.Update(player.progressMsF)
					player.cursorGlider.Update(player.progressMsF)
					player.playersGlider.Update(player.progressMsF)
					player.unfold.Update(player.progressMsF)
					player.hGlider.Update(player.progressMsF)
					player.vGlider.Update(player.progressMsF)
					player.flashGlider.Update(player.progressMsF)
					player.danserGlider.Update(player.progressMsF)
					player.resnadGlider.Update(player.progressMsF)
					player.oGlider.Update(player.progressMsF)
					/*settings.Objects.Colors.BaseColor.Hue = player.hGlider.GetValue()
					settings.Cursor.Colors.BaseColor.Hue = player.hGlider.GetValue()

					settings.Objects.Colors.BaseColor.Value = player.vGlider.GetValue()
					settings.Cursor.Colors.BaseColor.Value = player.vGlider.GetValue()

					settings.Objects.Colors.BaseColor.Saturation = player.flashGlider.GetValue()
					settings.Cursor.Colors.BaseColor.Saturation = player.flashGlider.GetValue()

					settings.Objects.Colors.HueOffset = player.oGlider.GetValue()
					settings.Cursor.Colors.HueOffset = player.oGlider.GetValue()*/
				}

			}
			lastT = currtime
			limiter.Sync()
		}
	}()

	go func() {
		for {

			musicPlayer.Update()
			//log.Println(musicPlayer.GetBeat())
			player.SclA = bmath.ClampF64(musicPlayer.GetBeat()*0.666*(settings.Audio.BeatScale-1.0)+1.0, 1.0, settings.Audio.BeatScale) //math.Min(1.4*settings.Audio.BeatScale, math.Max(math.Sin(musicPlayer.GetBeat()*math.Pi/2)*0.4*settings.Audio.BeatScale+1.0, 1.0))

			fft := musicPlayer.GetFFT()

			bars := 40
			boost := 0.0

			for i := 0; i < bars; i++ {
				boost += 2 * float64(fft[i]) * float64(bars-i) / float64(bars)
			}

			//oldVelocity := player.velocity

			player.velocity = math.Max(player.velocity, boost*1.5)

			player.velocity *= 1.0 - 0.05

			//player.Scl = 1 + player.progress*(settings.Audio.BeatScale-1.0)

			//log.Println(player.velocity)

			time.Sleep(15 * time.Millisecond)
		}
	}()
	player.profiler = frame.NewCounter()
	player.musicPlayer = musicPlayer

	player.bloomEffect = effects.NewBloomEffect(int(settings.Graphics.GetWidth()), int(settings.Graphics.GetHeight()))
	player.blur = effects.NewBlurEffect(int(settings.Graphics.GetWidth()), int(settings.Graphics.GetHeight()))

	return player
}

func (pl *Player) Show() {

}

func (pl *Player) Draw(delta float64) {

	if pl.lastTime < 0 {
		pl.lastTime = qpc.GetNanoTime()
	}
	tim := qpc.GetNanoTime()
	timMs := float64(tim-pl.lastTime) / 1000000.0

	fps := pl.profiler.GetFPS()

	if pl.start {
		if fps > 58 && timMs > 18 {
			log.Println(fmt.Sprintf("Slow frame detected! Frame time: %.3fms | Av. frame time: %.3fms", timMs, 1000.0/fps))
		}

		pl.progressMs = int64(pl.progressMsF)

		if pl.Scl < pl.SclA {
			pl.Scl += (pl.SclA - pl.Scl) * timMs / 100
		} else if pl.Scl > pl.SclA {
			pl.Scl -= (pl.Scl - pl.SclA) * timMs / 100
		}

	}
	pl.profiler.PutSample(timMs)
	pl.lastTime = tim

	if len(pl.queue2) > 0 {
		for i := 0; i < len(pl.queue2); i++ {
			if p := pl.queue2[i]; p.GetBasicData().StartTime-15000 <= pl.progressMs {
				if p := pl.queue2[i]; p.GetBasicData().StartTime-int64(pl.bMap.Diff.Preempt) <= pl.progressMs {

					pl.processed = append(pl.processed, p.(objects.Renderable))

					pl.queue2 = pl.queue2[1:]
					i--
				}
			} else {
				break
			}
		}
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

	bgAlpha := pl.dimGlider.GetValue()
	blurVal := 0.0

	cameras := pl.camera.GenRotated(settings.DIVIDES, -2*math.Pi/float64(settings.DIVIDES) /**pl.unfold.GetValue()*/)
	cameras1 := pl.camera1.GenRotated(settings.DIVIDES, -2*math.Pi/float64(settings.DIVIDES) /**pl.unfold.GetValue()*/)

	if settings.Playfield.BlurEnable {
		blurVal = pl.blurGlider.GetValue()
		if settings.Playfield.UnblurToTheBeat {
			blurVal -= settings.Playfield.UnblurFill * (blurVal) * (pl.Scl - 1.0) / (settings.Audio.BeatScale * 0.4)
		}
	}

	if settings.Playfield.FlashToTheBeat {
		bgAlpha *= pl.Scl
	}

	pl.background.Draw(pl.progressMs, pl.batch, blurVal, bgAlpha, cameras1[0])

	if pl.fxGlider.GetValue() > 0.01 || pl.epiGlider.GetValue() > 0.01 {
		pl.batch.Begin()
		pl.batch.SetColor(1, 1, 1, pl.fxGlider.GetValue())
		pl.batch.SetCamera(mgl32.Ortho(float32(-settings.Graphics.GetWidthF()/2), float32(settings.Graphics.GetWidthF()/2), float32(settings.Graphics.GetHeightF()/2), float32(-settings.Graphics.GetHeightF()/2), 1, -1))

		if pl.epiGlider.GetValue() > 0.01 {
			scl := (settings.Graphics.GetWidthF() / float64(pl.Epi.Width)) / 2 * 0.66
			pl.batch.SetScale(scl, scl)
			pl.batch.SetColor(1, 1, 1, pl.epiGlider.GetValue())
			pl.batch.DrawTexture(*pl.Epi)
			pl.batch.SetScale(1.0, -1.0)
			s := "Support me on ko-fi.com/wiekus"
			width := pl.font.GetWidth(settings.Graphics.GetHeightF()/40, s)
			pl.font.Draw(pl.batch, -width/2, (0.77)*(settings.Graphics.GetHeightF()/2), settings.Graphics.GetHeightF()/40, s)
		}

		pl.batch.End()
	}

	pl.counter += timMs

	if pl.counter >= 1000.0/60 {
		pl.fpsC = pl.profiler.GetFPS()
		pl.fpsU = pl.profilerU.GetFPS()

		pl.vol = pl.musicPlayer.GetLevelCombined()
		pl.volAverage = pl.volAverage*0.9 + pl.vol*0.1

		pl.counter -= 1000.0 / 60
		if pl.background.GetStoryboard() != nil {
			pl.storyboardLoad = pl.background.GetStoryboard().GetLoad()
		}
	}

	vprog := 1 - ((pl.vol - pl.volAverage) / 0.5)
	pV := math.Min(1.0, math.Max(0.0, 1.0-(vprog*0.5+pl.beatProgress*0.5)))

	ratio := math.Pow(0.5, timMs/16.6666666666667)

	pl.progress = pl.lastProgress*ratio + (pV)*(1-ratio)
	pl.lastProgress = pl.progress

	if pl.fxGlider.GetValue() > 0.01 {
		pl.batch.Begin()
		pl.visualiser.SetStartDistance(pl.cookieSize * 0.5)
		pl.visualiser.Draw(pl.progressMsF, pl.batch)

		pl.batch.SetColor(1, 1, 1, pl.Scl*pl.fxGlider.GetValue())

		scl := (pl.cookieSize / 2048.0) * 1.05

		pl.LogoS1.SetScale((1.05 - easing.OutQuad(pl.progress)*0.05) * scl)
		pl.LogoS2.SetScale((1.05 + easing.OutQuad(pl.progress)*0.03) * scl)

		alpha := 0.3
		if pl.bMap.Timings.Current.Kiai {
			alpha = 0.12
		}

		pl.LogoS2.SetAlpha(alpha * (1 - easing.OutQuad(pl.progress)))

		pl.LogoS1.UpdateAndDraw(pl.progressMs, pl.batch)
		pl.LogoS2.UpdateAndDraw(pl.progressMs, pl.batch)
		pl.batch.End()
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
	colors1, hshifts := settings.Cursor.GetColors(settings.DIVIDES /*settings.TAG*/, len(pl.controller.GetCursors()), pl.Scl, pl.cursorGlider.GetValue())
	colors2 := colors

	if settings.Objects.EnableCustomSliderBorderColor {
		colors2 = settings.Objects.CustomSliderBorderColor.GetColors(settings.DIVIDES, pl.Scl, pl.fadeOut*pl.fadeIn)
	}

	scale1 := pl.Scl
	scale2 := pl.Scl
	if settings.Playfield.RotationEnabled {
		rotationRad := (pl.rotation + settings.Playfield.BaseRotation) * math.Pi / 180.0
		pl.camera.SetRotation(-rotationRad)
		pl.camera.Update()
	}

	if !settings.Objects.ScaleToTheBeat {
		scale1 = 1
	}

	if !settings.Cursor.ScaleToTheBeat {
		scale2 = 1
	}

	if pl.overlay != nil {
		pl.batch.Begin()
		pl.batch.SetScale(1, 1)

		pl.batch.SetCamera(cameras[0])

		pl.overlay.DrawBeforeObjects(pl.batch, colors1, pl.playersGlider.GetValue()*0.8)

		pl.batch.End()
	}

	if settings.Playfield.BloomEnabled {
		pl.bloomEffect.SetThreshold(settings.Playfield.Bloom.Threshold)
		pl.bloomEffect.SetBlur(settings.Playfield.Bloom.Blur)
		pl.bloomEffect.SetPower(settings.Playfield.Bloom.Power + settings.Playfield.BloomBeatAddition*(pl.Scl-1.0)/(settings.Audio.BeatScale*0.4))
		pl.bloomEffect.Begin()
	}

	if pl.start && settings.Playfield.DrawObjects {

		for i := len(pl.processed) - 1; i >= 0; i-- {
			if s, ok := pl.processed[i].(*objects.Slider); ok {
				s.DrawBodyBase(pl.progressMs, cameras[0])
			}
		}

		if settings.Objects.SliderMerge {
			enabled := false

			for j := 0; j < settings.DIVIDES; j++ {
				ind := j - 1
				if ind < 0 {
					ind = settings.DIVIDES - 1
				}

				for i := len(pl.processed) - 1; i >= 0; i-- {
					if s, ok := pl.processed[i].(*objects.Slider); ok {
						if !enabled {
							enabled = true
							sliderrenderer.BeginRendererMerge()
						}

						s.DrawBody(pl.progressMs, colors2[j], colors2[ind], cameras[j], float32(scale1))
					}
				}
			}
			if enabled {
				sliderrenderer.EndRendererMerge()
			}
		}

		pl.batch.Begin()

		if settings.DIVIDES >= settings.Objects.MandalaTexturesTrigger {
			pl.batch.SetAdditive(true)
		} else {
			pl.batch.SetAdditive(false)
		}

		pl.batch.SetScale(scale1, scale1)

		for j := 0; j < settings.DIVIDES; j++ {
			pl.batch.SetCamera(cameras[j])
			ind := j - 1
			if ind < 0 {
				ind = settings.DIVIDES - 1
			}

			pl.batch.Flush()

			if !settings.Objects.SliderMerge {
				enabled := false

				for i := len(pl.processed) - 1; i >= 0 && len(pl.processed) > 0; i-- {
					if i < len(pl.processed) {
						if s, ok := pl.processed[i].(*objects.Slider); ok {
							if !enabled {
								enabled = true
								sliderrenderer.BeginRenderer()
							}

							s.DrawBody(pl.progressMs, colors2[j], colors2[ind], cameras[j], float32(scale1))
						}
					}
				}

				if enabled {
					sliderrenderer.EndRenderer()
				}
			}

			for i := len(pl.processed) - 1; i >= 0 && len(pl.processed) > 0; i-- {
				if i < len(pl.processed) {
					res := pl.processed[i].Draw(pl.progressMs, colors[j], pl.batch)
					if res {
						pl.processed = append(pl.processed[:i], pl.processed[(i+1):]...)
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

	if pl.overlay != nil && pl.overlay.NormalBeforeCursor() {
		pl.batch.Begin()
		pl.batch.SetScale(1, 1)

		//for j := 0; j < settings.DIVIDES; j++ {

		pl.batch.SetCamera(cameras[0])

		pl.overlay.DrawNormal(pl.batch, colors1, pl.playersGlider.GetValue()*0.8)
		//}

		pl.batch.End()
	}

	pl.background.DrawOverlay(pl.progressMs, pl.batch, bgAlpha, cameras1[0])

	for _, g := range pl.controller.GetCursors() {
		g.UpdateRenderer()
	}

	pl.batch.SetAdditive(false)
	render.BeginCursorRender()
	for j := 0; j < settings.DIVIDES; j++ {

		pl.batch.SetCamera(cameras[j])

		for i, g := range pl.controller.GetCursors() {
			if pl.overlay != nil && pl.overlay.IsBroken(g) {
				continue
			}

			baseIndex := j*len(pl.controller.GetCursors()) + i
			ind := baseIndex - 1
			if ind < 0 {
				ind = settings.DIVIDES*len(pl.controller.GetCursors()) - 1
			}

			col1 := colors1[baseIndex]
			col2 := colors1[ind]

			/*if i == 0 {
				col1[3] *= float32(pl.danserGlider.GetValue())
				col2[3] *= float32(pl.danserGlider.GetValue())
			}*/ /* else if i == 1 {
				col1[3] *= float32(pl.resnadGlider.GetValue())
				col2[3] *= float32(pl.resnadGlider.GetValue())
			}*/

			g.DrawM(scale2, pl.batch, col1, col2, hshifts[baseIndex])
		}

	}
	render.EndCursorRender()
	pl.batch.SetAdditive(false)

	if pl.overlay != nil {
		pl.batch.Begin()
		pl.batch.SetScale(1, 1)
		if !pl.overlay.NormalBeforeCursor() {
			pl.overlay.DrawNormal(pl.batch, colors1, pl.playersGlider.GetValue())
		}

		pl.batch.SetCamera(pl.scamera.GetProjectionView())

		pl.overlay.DrawHUD(pl.batch, colors1, pl.playersGlider.GetValue())

		pl.batch.End()
	}

	if settings.Playfield.BloomEnabled {
		pl.bloomEffect.EndAndRender()
	}

	if settings.DEBUG || settings.Graphics.ShowFPS {
		pl.batch.Begin()
		pl.batch.SetColor(1, 1, 1, 1)
		pl.batch.SetScale(1, 1)
		pl.batch.SetCamera(pl.scamera.GetProjectionView())

		padDown := 4.0 * (settings.Graphics.GetHeightF() / 1080.0)
		shift := 16.0 * (settings.Graphics.GetHeightF() / 1080.0)
		size := 16.0 * (settings.Graphics.GetHeightF() / 1080.0)

		if settings.DEBUG {
			pl.font.Draw(pl.batch, 0, settings.Graphics.GetHeightF()-size*1.5, size*1.5, pl.mapFullName)

			time := int(pl.musicPlayer.GetPosition())
			totalTime := int(pl.musicPlayer.GetLength())
			mapTime := int(pl.bMap.HitObjects[len(pl.bMap.HitObjects)-1].GetBasicData().EndTime / 1000)

			translate := func(b bool) string {
				if b {
					return "on"
				}
				return "off"
			}

			pl.font.Draw(pl.batch, 0, padDown+shift*4, size, fmt.Sprintf("Blur %s", translate(settings.Playfield.BlurEnable)))
			pl.font.Draw(pl.batch, 0, padDown+shift*3, size, fmt.Sprintf("Bloom %s", translate(settings.Playfield.BloomEnabled)))
			pl.font.Draw(pl.batch, 0, padDown+shift*2, size, fmt.Sprintf("%02d:%02d / %02d:%02d (%02d:%02d)", time/60, time%60, totalTime/60, totalTime%60, mapTime/60, mapTime%60))
			pl.font.Draw(pl.batch, 0, padDown+shift, size, fmt.Sprintf("%d(*%d) hitobjects, %d total", len(pl.processed), settings.DIVIDES, len(pl.bMap.HitObjects)))

			if storyboard := pl.background.GetStoryboard(); storyboard != nil {
				pl.font.Draw(pl.batch, 0, padDown, size, fmt.Sprintf("%d storyboard sprites (%0.2fx load), %d in queue (%d total)", storyboard.GetProcessedSprites(), pl.storyboardLoad, storyboard.GetQueueSprites(), storyboard.GetTotalSprites()))
			} else {
				pl.font.Draw(pl.batch, 0, padDown, size, "No storyboard")
			}
			vSync := "VSync " + translate(settings.Graphics.VSync)
			pl.font.Draw(pl.batch, settings.Graphics.GetWidthF()-pl.font.GetWidth(size, vSync), padDown+shift*2, size, vSync)

		}

		if settings.DEBUG || settings.Graphics.ShowFPS {
			fpsText := fmt.Sprintf("%0.0ffps (%0.2fms)", pl.fpsC, 1000/pl.fpsC)
			tpsText := fmt.Sprintf("%0.0ftps (%0.2fms)", pl.fpsU, 1000/pl.fpsU)
			pl.font.Draw(pl.batch, settings.Graphics.GetWidthF()-pl.font.GetWidth(size, fpsText), padDown+shift, size, fpsText)
			pl.font.Draw(pl.batch, settings.Graphics.GetWidthF()-pl.font.GetWidth(size, tpsText), padDown, size, tpsText)
		}

		pl.batch.End()
	}

}
