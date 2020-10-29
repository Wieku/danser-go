package states

import (
	"fmt"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/wieku/danser-go/app/beatmap"
	"github.com/wieku/danser-go/app/beatmap/objects"
	"github.com/wieku/danser-go/app/bmath"
	"github.com/wieku/danser-go/app/bmath/difficulty"
	"github.com/wieku/danser-go/app/dance"
	"github.com/wieku/danser-go/app/discord"
	"github.com/wieku/danser-go/app/graphics"
	"github.com/wieku/danser-go/app/graphics/font"
	"github.com/wieku/danser-go/app/graphics/gui/drawables"
	"github.com/wieku/danser-go/app/graphics/sliderrenderer"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/app/skin"
	"github.com/wieku/danser-go/app/states/components"
	"github.com/wieku/danser-go/app/utils"
	"github.com/wieku/danser-go/framework/bass"
	"github.com/wieku/danser-go/framework/frame"
	"github.com/wieku/danser-go/framework/graphics/effects"
	"github.com/wieku/danser-go/framework/graphics/sprite"
	"github.com/wieku/danser-go/framework/graphics/texture"
	"github.com/wieku/danser-go/framework/math/animation"
	"github.com/wieku/danser-go/framework/math/animation/easing"
	"github.com/wieku/danser-go/framework/math/scaling"
	"github.com/wieku/danser-go/framework/math/vector"
	"github.com/wieku/danser-go/framework/qpc"
	"github.com/wieku/danser-go/framework/statistic"
	"log"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"time"
)

type Player struct {
	font        *font.Font
	bMap        *beatmap.BeatMap
	queue2      []objects.BaseObject
	processed   []*RenderableProxy
	bloomEffect *effects.BloomEffect
	lastTime    int64
	progressMsF float64
	progressMs  int64
	batch       *sprite.SpriteBatch
	controller  dance.Controller
	background  *components.Background
	BgScl       vector.Vector2d
	Scl         float64
	SclA        float64
	fadeOut     float64
	fadeIn      float64
	start       bool
	musicPlayer *bass.Track
	profiler    *frame.Counter
	profilerU   *frame.Counter

	camera          *bmath.Camera
	camera1         *bmath.Camera
	scamera         *bmath.Camera
	dimGlider       *animation.Glider
	blurGlider      *animation.Glider
	fxGlider        *animation.Glider
	cursorGlider    *animation.Glider
	playersGlider   *animation.Glider
	unfold          *animation.Glider
	counter         float64
	storyboardLoad  float64
	storyboardDrawn int
	mapFullName     string
	Epi             *texture.TextureRegion
	epiGlider       *animation.Glider
	overlay         components.Overlay
	velocity        float64
	blur            *effects.BlurEffect

	currentBeatVal float64
	lastBeatLength float64
	lastBeatStart  float64
	beatProgress   float64
	lastBeatProg   int64

	progress, lastProgress float64
	LogoS1                 *sprite.Sprite
	LogoS2                 *sprite.Sprite

	vol        float64
	volAverage float64
	cookieSize float64
	visualiser *drawables.Visualiser

	followpoints *sprite.SpriteManager
	hudGlider    *animation.Glider

	volumeGlider *animation.Glider
	startPoint   float64
}

type RenderableProxy struct {
	renderable   objects.Renderable
	IsSliderBody bool
	depth        int64
	endTime      int64
}

func NewPlayer(beatMap *beatmap.BeatMap) *Player {
	player := new(Player)
	graphics.LoadTextures()
	player.batch = sprite.NewSpriteBatch()
	player.font = font.GetFont("Exo 2 Bold")

	discord.SetMap(beatMap.Artist, beatMap.Name, beatMap.Difficulty)

	player.bMap = beatMap
	player.mapFullName = fmt.Sprintf("%s - %s [%s]", beatMap.Artist, beatMap.Name, beatMap.Difficulty)
	log.Println("Playing:", player.mapFullName)

	var err error

	LogoT, err := utils.LoadTextureToAtlas(graphics.Atlas, "assets/textures/coinbig.png")
	player.LogoS1 = sprite.NewSpriteSingle(LogoT, 0, vector.NewVec2d(0, 0), vector.NewVec2d(0, 0))
	player.LogoS2 = sprite.NewSpriteSingle(LogoT, 0, vector.NewVec2d(0, 0), vector.NewVec2d(0, 0))

	if settings.Graphics.GetWidthF() > settings.Graphics.GetHeightF() {
		player.cookieSize = 0.5 * settings.Graphics.GetHeightF()
	} else {
		player.cookieSize = 0.5 * settings.Graphics.GetWidthF()
	}

	player.Epi, err = utils.LoadTextureToAtlas(graphics.Atlas, "assets/textures/warning.png")

	if err != nil {
		log.Println(err)
	}

	player.background = components.NewBackground(beatMap)

	player.camera = bmath.NewCamera()
	player.camera.SetOsuViewport(int(settings.Graphics.GetWidth()), int(settings.Graphics.GetHeight()), settings.Playfield.Scale, settings.Playfield.OsuShift)
	//player.camera.SetOrigin(bmath.NewVec2d(256, 192.0-5))
	player.camera.Update()

	player.camera1 = bmath.NewCamera()

	sbScale := 1.0
	if settings.Playfield.ScaleStoryboardWithPlayfield {
		sbScale = settings.Playfield.Scale
	}

	player.camera1.SetOsuViewport(int(settings.Graphics.GetWidth()), int(settings.Graphics.GetHeight()), sbScale, false)
	player.camera1.Update()

	player.scamera = bmath.NewCamera()
	player.scamera.SetViewport(int(settings.Graphics.GetWidth()), int(settings.Graphics.GetHeight()), false)
	player.scamera.SetOrigin(vector.NewVec2d(settings.Graphics.GetWidthF()/2, settings.Graphics.GetHeightF()/2))
	player.scamera.Update()

	graphics.Camera = player.camera

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

	player.followpoints = sprite.NewSpriteManager()

	prempt := 800.0
	postmt := 240.0
	lineDist := 32.0

	for i := 1; i < len(player.queue2); i++ {
		_, ok1 := player.queue2[i-1].(*objects.Spinner)
		_, ok2 := player.queue2[i].(*objects.Spinner)
		if ok1 || ok2 || player.queue2[i].GetBasicData().NewCombo {
			continue
		}

		prevTime := float64(player.queue2[i-1].GetBasicData().EndTime)
		prevPos := player.queue2[i-1].GetBasicData().EndPos.Copy64()

		nextTime := float64(player.queue2[i].GetBasicData().StartTime)
		nextPos := player.queue2[i].GetBasicData().StartPos.Copy64()

		vec := nextPos.Sub(prevPos)
		duration := nextTime - prevTime
		distance := vec.Len()
		rotation := vec.AngleR()

		for prog := lineDist * 1.5; prog < distance-lineDist; prog += lineDist {
			t := prog / distance

			tStart := prevTime + t*duration - prempt
			tEnd := prevTime + t*duration

			pos := prevPos.Add(vec.Scl(t))

			textures := skin.GetFrames("followpoint", true)

			sprite := sprite.NewAnimation(textures, 1000.0/float64(len(textures)), true, -float64(i), pos, bmath.Origin.Centre)
			sprite.SetRotation(rotation)
			sprite.ShowForever(false)

			sprite.AddTransform(animation.NewSingleTransform(animation.Fade, easing.Linear, tStart, tStart+postmt, 0, 1))
			sprite.AddTransform(animation.NewSingleTransform(animation.Fade, easing.Linear, tEnd, tEnd+postmt, 1, 0))
			sprite.AdjustTimesToTransformations()
			sprite.SetAlpha(0)
			player.followpoints.Add(sprite)
		}
	}

	log.Println("Track:", beatMap.Audio)

	player.Scl = 1
	player.fadeOut = 1.0
	player.fadeIn = 0.0

	player.volumeGlider = animation.NewGlider(1.0)
	player.hudGlider = animation.NewGlider(1.0)
	player.dimGlider = animation.NewGlider(0.0)
	player.blurGlider = animation.NewGlider(0.0)
	player.fxGlider = animation.NewGlider(0.0)
	if _, ok := player.overlay.(*components.ScoreOverlay); !ok {
		player.cursorGlider = animation.NewGlider(0.0)
	} else {
		player.cursorGlider = animation.NewGlider(1.0)
	}
	player.playersGlider = animation.NewGlider(0.0)

	skipTime := 0.0
	if settings.SKIP {
		skipTime = float64(player.queue2[0].GetBasicData().StartTime)
	}

	skipTime = math.Max(skipTime, settings.SCRUB*1000)

	tmS := math.Max(float64(player.queue2[0].GetBasicData().StartTime), settings.SCRUB*1000)
	tmE := float64(player.queue2[len(player.queue2)-1].GetBasicData().EndTime)

	startOffset := 0.0

	if settings.SKIP || settings.SCRUB > 0.01 {
		startOffset = skipTime
		player.startPoint = skipTime - beatMap.Diff.Preempt
		player.volumeGlider.SetValue(0.0)
		player.volumeGlider.AddEvent(skipTime-beatMap.Diff.Preempt, skipTime-beatMap.Diff.Preempt+difficulty.HitFadeIn, 1.0)
	}

	startOffset += -settings.Playfield.LeadInHold*1000 - beatMap.Diff.Preempt

	player.dimGlider.AddEvent(startOffset-500, startOffset, 1.0-settings.Playfield.Background.Dim.Intro)
	player.blurGlider.AddEvent(startOffset-500, startOffset, settings.Playfield.Background.Blur.Values.Intro)
	player.fxGlider.AddEvent(startOffset-500, startOffset, 1.0-settings.Playfield.Logo.Dim.Intro)
	if _, ok := player.overlay.(*components.ScoreOverlay); !ok {
		player.cursorGlider.AddEvent(startOffset-500, startOffset, 0.0)
	}
	player.playersGlider.AddEvent(startOffset-500, startOffset, 1.0)

	player.dimGlider.AddEvent(tmS-750, tmS-250, 1.0-settings.Playfield.Background.Dim.Normal)
	player.blurGlider.AddEvent(tmS-750, tmS-250, settings.Playfield.Background.Blur.Values.Normal)
	player.fxGlider.AddEvent(tmS-750, tmS-250, 1.0-settings.Playfield.Logo.Dim.Normal)
	player.cursorGlider.AddEvent(tmS-750, tmS-250, 1.0)

	fadeOut := settings.Playfield.FadeOutTime * 1000
	player.dimGlider.AddEvent(tmE, tmE+fadeOut, 0.0)
	player.fxGlider.AddEvent(tmE, tmE+fadeOut, 0.0)
	player.cursorGlider.AddEvent(tmE, tmE+fadeOut, 0.0)
	player.playersGlider.AddEvent(tmE, tmE+fadeOut, 0.0)

	player.hudGlider.AddEvent(tmE, tmE+fadeOut, 0.0)

	player.volumeGlider.AddEvent(tmE, tmE+settings.Playfield.FadeOutTime*1000, 0.0)

	player.epiGlider = animation.NewGlider(0)

	if settings.Playfield.SeizureWarning.Enabled {
		am := math.Max(1000, settings.Playfield.SeizureWarning.Duration*1000)
		startOffset -= am
		player.epiGlider.AddEvent(startOffset, startOffset+500, 1.0)
		player.epiGlider.AddEvent(startOffset+am-500, startOffset+am, 0.0)
	}

	startOffset -= settings.Playfield.LeadInTime * 1000

	player.progressMsF = startOffset

	player.unfold = animation.NewGlider(1)

	for _, p := range beatMap.Pauses {
		bd := p.GetBasicData()

		if bd.EndTime-bd.StartTime < 1000 {
			continue
		}

		//player.hudGlider.AddEvent(float64(bd.StartTime), float64(bd.StartTime)+500, 0.0)
		player.dimGlider.AddEvent(float64(bd.StartTime), float64(bd.StartTime)+500, 1.0-settings.Playfield.Background.Dim.Breaks)
		player.blurGlider.AddEvent(float64(bd.StartTime), float64(bd.StartTime)+500, settings.Playfield.Background.Blur.Values.Breaks)
		player.fxGlider.AddEvent(float64(bd.StartTime), float64(bd.StartTime)+500, 1.0-settings.Playfield.Logo.Dim.Breaks)

		if !settings.Cursor.ShowCursorsOnBreaks {
			player.cursorGlider.AddEvent(float64(bd.StartTime), float64(bd.StartTime)+100, 0.0)
		}

		//player.hudGlider.AddEvent(float64(bd.EndTime)-500, float64(bd.EndTime), 1.0)
		player.dimGlider.AddEvent(float64(bd.EndTime)-500, float64(bd.EndTime), 1.0-settings.Playfield.Background.Dim.Normal)
		player.blurGlider.AddEvent(float64(bd.EndTime)-500, float64(bd.EndTime), settings.Playfield.Background.Blur.Values.Normal)
		player.fxGlider.AddEvent(float64(bd.EndTime)-500, float64(bd.EndTime), 1.0-settings.Playfield.Logo.Dim.Normal)
		player.cursorGlider.AddEvent(float64(bd.EndTime)-100, float64(bd.EndTime), 1.0)
	}

	musicPlayer := bass.NewTrack(filepath.Join(settings.General.OsuSongsDir, beatMap.Dir, beatMap.Audio))
	player.background.SetTrack(musicPlayer)
	player.visualiser = drawables.NewVisualiser(player.cookieSize*0.66, player.cookieSize*2, vector.NewVec2d(0, 0))
	player.visualiser.SetTrack(musicPlayer)

	player.profiler = frame.NewCounter()
	player.musicPlayer = musicPlayer

	player.bloomEffect = effects.NewBloomEffect(int(settings.Graphics.GetWidth()), int(settings.Graphics.GetHeight()))
	player.blur = effects.NewBlurEffect(int(settings.Graphics.GetWidth()), int(settings.Graphics.GetHeight()))

	player.background.Update(0, settings.Graphics.GetWidthF()/2, settings.Graphics.GetHeightF()/2)

	player.profilerU = frame.NewCounter()
	limiter := frame.NewLimiter(10000)

	go func() {
		defer func() {
			if err := recover(); err != nil {
				log.Println("panic:", err)

				for _, s := range utils.GetPanicStackTrace() {
					log.Println(s)
				}

				os.Exit(1)
			}
		}()

		runtime.LockOSThread()

		var lastT = qpc.GetNanoTime()

		for {
			currtime := qpc.GetNanoTime()

			player.profilerU.PutSample(float64(currtime-lastT) / 1000000.0)

			if musicPlayer.GetState() == bass.MUSIC_STOPPED {
				player.progressMsF += float64(currtime-lastT) / 1000000.0
			} else {
				player.progressMsF = musicPlayer.GetPosition()*1000 + float64(settings.Audio.Offset)
			}

			if player.progressMsF >= player.startPoint && !player.start {
				musicPlayer.Play()
				musicPlayer.SetTempo(settings.SPEED)
				musicPlayer.SetPitch(settings.PITCH)

				if ov, ok := player.overlay.(*components.ScoreOverlay); ok {
					ov.SetMusic(musicPlayer)
				}

				musicPlayer.SetPosition(player.startPoint / 1000)

				discord.SetDuration(int64((musicPlayer.GetLength() - musicPlayer.GetPosition()) * 1000 / settings.SPEED))

				if player.overlay == nil {
					discord.UpdateDance(settings.TAG, settings.DIVIDES)
				}

				player.start = true
			}

			if player.progressMsF >= player.startPoint-player.bMap.Diff.Preempt {
				if _, ok := player.controller.(*dance.GenericController); ok {
					player.bMap.Update(int64(player.progressMsF))
				}
			}

			if player.progressMsF >= player.startPoint-player.bMap.Diff.Preempt || settings.PLAY {
				player.controller.Update(int64(player.progressMsF), float64(currtime-lastT)/1000000)
			}

			if player.overlay != nil {
				player.overlay.Update(int64(player.progressMsF))
			}

			bTime := player.bMap.Timings.Current.BaseBpm

			if bTime != player.lastBeatLength {
				player.lastBeatLength = bTime
				player.lastBeatStart = float64(player.bMap.Timings.Current.Time)
				player.lastBeatProg = int64((player.progressMsF-player.lastBeatStart)/player.lastBeatLength) - 1
			}

			if int64(float64(player.progressMsF-player.lastBeatStart)/player.lastBeatLength) > player.lastBeatProg {
				player.lastBeatProg++
			}

			player.beatProgress = float64(player.progressMsF-player.lastBeatStart)/player.lastBeatLength - float64(player.lastBeatProg)
			player.visualiser.Update(player.progressMsF)

			var offset vector.Vector2d

			for _, c := range player.controller.GetCursors() {
				offset = offset.Add(player.camera.Project(c.Position.Copy64()).Mult(vector.NewVec2d(2/settings.Graphics.GetWidthF(), -2/settings.Graphics.GetHeightF())))
			}

			offset = offset.Scl(1 / float64(len(player.controller.GetCursors())))

			player.background.Update(player.progressMsF, offset.X*player.cursorGlider.GetValue(), offset.Y*player.cursorGlider.GetValue())

			player.followpoints.Update(int64(player.progressMsF))

			player.epiGlider.Update(player.progressMsF)
			player.dimGlider.Update(player.progressMsF)
			player.blurGlider.Update(player.progressMsF)
			player.fxGlider.Update(player.progressMsF)
			player.cursorGlider.Update(player.progressMsF)
			player.playersGlider.Update(player.progressMsF)
			player.hudGlider.Update(player.progressMsF)
			player.unfold.Update(player.progressMsF)

			player.volumeGlider.Update(player.progressMsF)
			player.musicPlayer.SetVolumeRelative(player.volumeGlider.GetValue())

			lastT = currtime
			limiter.Sync()
		}
	}()

	go func() {
		for {

			musicPlayer.Update()
			//log.Println(musicPlayer.GetBeat())
			//player.Scl = bmath.ClampF64(musicPlayer.GetBeat()*0.666*(settings.Audio.BeatScale-1.0)+1.0, 1.0, settings.Audio.BeatScale) //math.Min(1.4*settings.Audio.BeatScale, math.Max(math.Sin(musicPlayer.GetBeat()*math.Pi/2)*0.4*settings.Audio.BeatScale+1.0, 1.0))

			fft := musicPlayer.GetFFT()

			bars := 40
			boost := 0.0

			for i := 0; i < bars; i++ {
				boost += 2 * float64(fft[i]) * float64(bars-i) / float64(bars)
			}

			//oldVelocity := player.velocity

			player.velocity = math.Min(math.Max(player.velocity, math.Max(0, (boost*1.5/6)-0.5)/0.5), 1)

			player.velocity *= 1.0 - 0.05

			player.Scl = 1 + player.progress*(settings.Audio.BeatScale-1.0)

			//log.Println(player.velocity)

			time.Sleep(15 * time.Millisecond)
		}
	}()

	return player
}

func (pl *Player) Show() {

}

func (pl *Player) AddProxy(proxy *RenderableProxy) {
	n := sort.Search(len(pl.processed), func(j int) bool {
		return proxy.depth < pl.processed[j].depth
	})

	pl.processed = append(pl.processed, nil) //allocate bigger array in case when len=cap
	copy(pl.processed[n+1:], pl.processed[n:])

	pl.processed[n] = proxy
}

func (pl *Player) Draw(float64) {
	if pl.lastTime <= 0 {
		pl.lastTime = qpc.GetNanoTime()
	}

	tim := qpc.GetNanoTime()
	timMs := float64(tim-pl.lastTime) / 1000000.0

	fps := pl.profiler.GetFPS()

	if fps > 58 && timMs > 18 {
		log.Println(fmt.Sprintf("Slow frame detected! Frame time: %.3fms | Av. frame time: %.3fms", timMs, 1000.0/fps))
	}

	pl.progressMs = int64(pl.progressMsF)

	//if pl.Scl < pl.SclA {
	//	pl.Scl += (pl.SclA - pl.Scl) * timMs / 100
	//} else if pl.Scl > pl.SclA {
	//	pl.Scl -= (pl.Scl - pl.SclA) * timMs / 100
	//}

	pl.profiler.PutSample(timMs)
	pl.lastTime = tim

	if len(pl.queue2) > 0 {
		for i := 0; i < len(pl.queue2); i++ {
			if p := pl.queue2[i]; p.GetBasicData().StartTime-15000 <= pl.progressMs {
				if p := pl.queue2[i]; p.GetBasicData().StartTime-int64(pl.bMap.Diff.Preempt) <= pl.progressMs {

					if _, ok := p.(*objects.Spinner); ok {
						pl.AddProxy(&RenderableProxy{
							renderable:   p.(objects.Renderable),
							IsSliderBody: false,
							depth:        math.MaxInt64,
							endTime:      p.GetBasicData().EndTime + difficulty.HitFadeOut,
						})
					} else {
						pl.AddProxy(&RenderableProxy{
							renderable:   p.(objects.Renderable),
							IsSliderBody: false,
							depth:        p.GetBasicData().StartTime,
							endTime:      p.GetBasicData().EndTime + difficulty.HitFadeOut,
						})
					}

					if _, ok := p.(*objects.Slider); ok {
						pl.AddProxy(&RenderableProxy{
							renderable:   p.(objects.Renderable),
							IsSliderBody: true,
							depth:        p.GetBasicData().EndTime + 10,
							endTime:      p.GetBasicData().EndTime + difficulty.HitFadeOut,
						})
					}

					pl.queue2 = pl.queue2[1:]
					i--
				}
			} else {
				break
			}
		}
	}

	bgAlpha := pl.dimGlider.GetValue()

	cameras := pl.camera.GenRotated(settings.DIVIDES, -2*math.Pi/float64(settings.DIVIDES) /**pl.unfold.GetValue()*/)
	cameras1 := pl.camera1.GenRotated(settings.DIVIDES, -2*math.Pi/float64(settings.DIVIDES) /**pl.unfold.GetValue()*/)

	if settings.Playfield.Background.FlashToTheBeat {
		bgAlpha *= pl.Scl
	}

	pl.background.Draw(pl.progressMs, pl.batch, pl.blurGlider.GetValue(), bgAlpha, cameras1[0])

	if pl.start {
		settings.Objects.Colors.Color.Update(timMs)
		settings.Objects.Colors.Sliders.Border.Color.Update(timMs)
		settings.Objects.Colors.Sliders.Body.Color.Update(timMs)
		settings.Cursor.Colors.Update(timMs)
	}

	objectColors := settings.Objects.Colors.Color.GetColors(settings.DIVIDES, pl.Scl, pl.fadeOut*pl.fadeIn)
	cursorColors := settings.Cursor.GetColors(settings.DIVIDES, len(pl.controller.GetCursors()), pl.Scl, pl.cursorGlider.GetValue())
	borderColors := objectColors
	bodyColors := objectColors

	if !settings.Objects.Colors.Sliders.Border.UseHitCircleColor {
		borderColors = settings.Objects.Colors.Sliders.Border.Color.GetColors(settings.DIVIDES, pl.Scl, pl.fadeOut*pl.fadeIn)
	}

	if !settings.Objects.Colors.Sliders.Body.UseHitCircleColor {
		bodyColors = settings.Objects.Colors.Sliders.Body.Color.GetColors(settings.DIVIDES, pl.Scl, pl.fadeOut*pl.fadeIn)
	}

	if pl.overlay != nil {
		pl.batch.Begin()
		pl.batch.ResetTransform()
		pl.batch.SetScale(1, 1)

		pl.batch.SetCamera(cameras[0])

		pl.overlay.DrawBeforeObjects(pl.batch, cursorColors, pl.playersGlider.GetValue()*pl.hudGlider.GetValue())

		pl.batch.End()
		pl.batch.ResetTransform()
		pl.batch.SetColor(1, 1, 1, 1)
	}

	if pl.epiGlider.GetValue() > 0.01 {
		pl.batch.Begin()
		pl.batch.ResetTransform()
		pl.batch.SetColor(1, 1, 1, pl.epiGlider.GetValue())
		pl.batch.SetCamera(mgl32.Ortho(float32(-settings.Graphics.GetWidthF()/2), float32(settings.Graphics.GetWidthF()/2), float32(settings.Graphics.GetHeightF()/2), float32(-settings.Graphics.GetHeightF()/2), 1, -1))

		scl := scaling.Fit.Apply(float32(pl.Epi.Width), float32(pl.Epi.Height), float32(settings.Graphics.GetWidthF()), float32(settings.Graphics.GetHeightF()))
		scl = scl.Scl(0.5).Scl(0.66)
		pl.batch.SetScale(scl.X64(), scl.Y64())
		pl.batch.DrawUnit(*pl.Epi)

		//pl.batch.SetScale(1.0, -1.0)
		//s := "Support me on ko-fi.com/wiekus"
		//width := pl.font.GetWidth(settings.Graphics.GetHeightF()/40, s)
		//pl.font.Draw(pl.batch, -width/2, (0.77)*(settings.Graphics.GetHeightF()/2), settings.Graphics.GetHeightF()/40, s)

		pl.batch.ResetTransform()
		pl.batch.End()
		pl.batch.SetColor(1, 1, 1, 1)
	}

	pl.counter += timMs

	if pl.counter >= 1000.0/60 {

		pl.vol = pl.musicPlayer.GetLevelCombined()
		pl.volAverage = pl.volAverage*0.9 + pl.vol*0.1

		pl.counter -= 1000.0 / 60
		if pl.background.GetStoryboard() != nil {
			pl.storyboardLoad = pl.background.GetStoryboard().GetLoad()
			pl.storyboardDrawn = pl.background.GetStoryboard().GetRenderedSprites()
		}
	}

	vprog := 1 - ((pl.vol - pl.volAverage) / 0.5)
	pV := math.Min(1.0, math.Max(0.0, 1.0-(vprog*0.5+pl.beatProgress*0.5)))

	ratio := math.Pow(0.5, timMs/16.6666666666667)

	pl.progress = pl.lastProgress*ratio + (pV)*(1-ratio)
	pl.lastProgress = pl.progress

	if pl.fxGlider.GetValue() > 0.01 {
		pl.batch.Begin()
		pl.batch.ResetTransform()
		pl.batch.SetColor(1, 1, 1, pl.fxGlider.GetValue())
		pl.batch.SetCamera(mgl32.Ortho(float32(-settings.Graphics.GetWidthF()/2), float32(settings.Graphics.GetWidthF()/2), float32(settings.Graphics.GetHeightF()/2), float32(-settings.Graphics.GetHeightF()/2), 1, -1))

		innerCircleScale := 1.05 - easing.OutQuad(pl.progress)*0.05
		outerCircleScale := 1.05 + easing.OutQuad(pl.progress)*0.03

		if settings.Playfield.Logo.DrawSpectrum {
			pl.visualiser.SetStartDistance(pl.cookieSize * 0.5 * innerCircleScale)
			pl.visualiser.Draw(pl.progressMsF, pl.batch)
		}

		pl.batch.SetColor(1, 1, 1, pl.fxGlider.GetValue())

		scl := (pl.cookieSize / 2048.0) * 1.05

		pl.LogoS1.SetScale(innerCircleScale * scl)
		pl.LogoS2.SetScale(outerCircleScale * scl)

		alpha := 0.3
		if pl.bMap.Timings.Current.Kiai {
			alpha = 0.12
		}

		pl.LogoS2.SetAlpha(float32(alpha * (1 - easing.OutQuad(pl.progress))))

		pl.LogoS1.UpdateAndDraw(pl.progressMs, pl.batch)
		pl.LogoS2.UpdateAndDraw(pl.progressMs, pl.batch)
		pl.batch.ResetTransform()
		pl.batch.End()
	}

	scale1 := pl.Scl
	scale2 := pl.Scl

	if !settings.Objects.ScaleToTheBeat {
		scale1 = 1
	}

	if !settings.Cursor.ScaleToTheBeat {
		scale2 = 1
	}

	if settings.Playfield.Bloom.Enabled {
		pl.bloomEffect.SetThreshold(settings.Playfield.Bloom.Threshold)
		pl.bloomEffect.SetBlur(settings.Playfield.Bloom.Blur)
		pl.bloomEffect.SetPower(settings.Playfield.Bloom.Power + settings.Playfield.Bloom.BloomBeatAddition*(pl.Scl-1.0)/(settings.Audio.BeatScale*0.4))
		pl.bloomEffect.Begin()
	}

	if settings.Playfield.DrawObjects {
		pl.batch.Begin()
		pl.batch.ResetTransform()
		pl.batch.SetColor(1, 1, 1, 1)
		pl.batch.SetScale(scale1*pl.bMap.Diff.CircleRadius/64, scale1*pl.bMap.Diff.CircleRadius/64)

		if settings.DIVIDES < settings.Objects.Colors.MandalaTexturesTrigger && settings.Objects.DrawFollowPoints {
			for j := 0; j < settings.DIVIDES; j++ {
				pl.batch.SetCamera(cameras[j])
				pl.followpoints.Draw(pl.progressMs, pl.batch)
			}
		}

		pl.batch.Flush()
		pl.batch.SetScale(1, 1)

		for i := len(pl.processed) - 1; i >= 0; i-- {
			if s, ok := pl.processed[i].renderable.(*objects.Slider); ok && pl.processed[i].IsSliderBody {
				s.DrawBodyBase(pl.progressMs, cameras[0])
			}
		}

		if settings.Objects.Sliders.SliderMerge {
			enabled := false

			for j := 0; j < settings.DIVIDES; j++ {
				ind := j - 1
				if ind < 0 {
					ind = settings.DIVIDES - 1
				}

				for i := len(pl.processed) - 1; i >= 0; i-- {
					if s, ok := pl.processed[i].renderable.(*objects.Slider); ok && pl.processed[i].IsSliderBody {
						if !enabled {
							enabled = true
							sliderrenderer.BeginRendererMerge()
						}

						s.DrawBody(pl.progressMs, bodyColors[j], borderColors[j], borderColors[ind], cameras[j], float32(scale1))
					}
				}
			}
			if enabled {
				sliderrenderer.EndRendererMerge()
			}
		}

		if settings.DIVIDES >= settings.Objects.Colors.MandalaTexturesTrigger {
			pl.batch.SetAdditive(true)
		} else {
			pl.batch.SetAdditive(false)
		}

		pl.batch.SetScale(scale1*pl.bMap.Diff.CircleRadius/64, scale1*pl.bMap.Diff.CircleRadius/64)

		for j := 0; j < settings.DIVIDES; j++ {
			pl.batch.SetCamera(cameras[j])
			ind := j - 1
			if ind < 0 {
				ind = settings.DIVIDES - 1
			}

			pl.batch.Flush()

			enabled := false

			for i := len(pl.processed) - 1; i >= 0; i-- {
				proxy := pl.processed[i]

				if !proxy.IsSliderBody {
					if enabled && !settings.Objects.Sliders.SliderMerge {
						enabled = false

						sliderrenderer.EndRenderer()
					}

					_, sp := pl.processed[i].renderable.(*objects.Spinner)
					if !sp || j == 0 {
						proxy.renderable.Draw(pl.progressMs, objectColors[j], pl.batch)
					}
				} else if !settings.Objects.Sliders.SliderMerge {
					if !enabled {
						enabled = true

						pl.batch.Flush()

						sliderrenderer.BeginRenderer()
					}

					proxy.renderable.(*objects.Slider).DrawBody(pl.progressMs, bodyColors[j], borderColors[j], borderColors[ind], cameras[j], float32(scale1))
				}

				if proxy.endTime <= pl.progressMs {
					pl.processed = append(pl.processed[:i], pl.processed[(i+1):]...)
				}
			}

			if enabled {
				sliderrenderer.EndRenderer()
			}
		}

		if settings.DIVIDES < settings.Objects.Colors.MandalaTexturesTrigger && settings.Objects.DrawApproachCircles {
			pl.batch.Flush()

			for j := 0; j < settings.DIVIDES; j++ {
				pl.batch.SetCamera(cameras[j])

				for i := len(pl.processed) - 1; i >= 0; i-- {
					if s := pl.processed[i]; !s.IsSliderBody {
						s.renderable.DrawApproach(pl.progressMs, objectColors[j], pl.batch)
					}
				}
			}
		}

		pl.batch.SetScale(1, 1)
		pl.batch.End()
	}

	if pl.overlay != nil && pl.overlay.NormalBeforeCursor() {
		pl.batch.Begin()
		pl.batch.SetScale(1, 1)

		pl.batch.SetCamera(cameras[0])

		pl.overlay.DrawNormal(pl.batch, cursorColors, pl.playersGlider.GetValue()*pl.hudGlider.GetValue())

		pl.batch.End()
	}

	pl.background.DrawOverlay(pl.progressMs, pl.batch, bgAlpha, cameras1[0])

	if settings.Playfield.DrawCursors {
		for _, g := range pl.controller.GetCursors() {
			g.UpdateRenderer()
		}

		pl.batch.SetAdditive(false)

		graphics.BeginCursorRender()

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

				col1 := cursorColors[baseIndex]
				col2 := cursorColors[ind]

				g.DrawM(scale2, pl.batch, col1, col2)
			}
		}

		graphics.EndCursorRender()
	}

	pl.batch.SetAdditive(false)

	if pl.overlay != nil {
		pl.batch.Begin()
		pl.batch.SetScale(1, 1)

		if !pl.overlay.NormalBeforeCursor() {
			pl.overlay.DrawNormal(pl.batch, cursorColors, pl.playersGlider.GetValue()*pl.hudGlider.GetValue())
		}

		pl.batch.SetCamera(pl.scamera.GetProjectionView())

		pl.overlay.DrawHUD(pl.batch, cursorColors, pl.playersGlider.GetValue()*pl.hudGlider.GetValue())

		pl.batch.End()
	}

	if settings.Playfield.Bloom.Enabled {
		pl.bloomEffect.EndAndRender()
	}

	if settings.DEBUG || settings.Graphics.ShowFPS {
		pl.batch.Begin()
		pl.batch.SetColor(1, 1, 1, 1)
		pl.batch.SetScale(1, 1)
		pl.batch.SetCamera(pl.scamera.GetProjectionView())

		padDown := 4.0 * (settings.Graphics.GetHeightF() / 1080.0)
		//shift := 16.0 * (settings.Graphics.GetHeightF() / 1080.0)
		size := 16.0 * (settings.Graphics.GetHeightF() / 1080.0)

		if settings.DEBUG {
			drawShadowed := func(pos float64, text string) {
				pl.batch.SetColor(0, 0, 0, 0.5)
				pl.font.DrawMonospaced(pl.batch, 0-size*0.05, (size+padDown)*pos+padDown/2-size*0.05, size, text)

				pl.batch.SetColor(1, 1, 1, 1)
				pl.font.DrawMonospaced(pl.batch, 0, (size+padDown)*pos+padDown/2, size, text)
			}

			pl.batch.SetColor(0, 0, 0, 1)
			pl.font.Draw(pl.batch, 0-size*1.5*0.1, settings.Graphics.GetHeightF()-size*1.5-size*1.5*0.1, size*1.5, pl.mapFullName)

			pl.batch.SetColor(1, 1, 1, 1)
			pl.font.Draw(pl.batch, 0, settings.Graphics.GetHeightF()-size*1.5, size*1.5, pl.mapFullName)

			type tx struct {
				pos  float64
				text string
			}

			var queue []tx

			drawWithBackground := func(pos float64, text string) {
				pl.batch.SetColor(0, 0, 0, 0.8)

				width := pl.font.GetWidthMonospaced(size, text)

				pl.batch.SetTranslation(vector.NewVec2d(width/2, settings.Graphics.GetHeightF()-(size+padDown)*(pos-0.5)))
				pl.batch.SetSubScale(width/2, (size+padDown)/2)
				pl.batch.DrawUnit(graphics.Pixel.GetRegion())

				queue = append(queue, tx{pos, text})
			}

			drawWithBackground(3, fmt.Sprintf("VSync: %t", settings.Graphics.VSync))
			drawWithBackground(4, fmt.Sprintf("Blur: %t", settings.Playfield.Background.Blur.Enabled))
			drawWithBackground(5, fmt.Sprintf("Bloom: %t", settings.Playfield.Bloom.Enabled))

			drawWithBackground(6, fmt.Sprintf("FBO Binds: %d", statistic.GetPrevious(statistic.FBOBinds)))
			drawWithBackground(7, fmt.Sprintf("VAO Binds: %d", statistic.GetPrevious(statistic.VAOBinds)))
			drawWithBackground(8, fmt.Sprintf("VBO Binds: %d", statistic.GetPrevious(statistic.VBOBinds)))
			drawWithBackground(9, fmt.Sprintf("Vertex Upload: %.2fk", float64(statistic.GetPrevious(statistic.VertexUpload))/1000))
			drawWithBackground(10, fmt.Sprintf("Vertices Drawn: %.2fk", float64(statistic.GetPrevious(statistic.VerticesDrawn))/1000))
			drawWithBackground(11, fmt.Sprintf("Draw Calls: %d", statistic.GetPrevious(statistic.DrawCalls)))
			drawWithBackground(12, fmt.Sprintf("Sprites Drawn: %d", statistic.GetPrevious(statistic.SpritesDrawn)))

			if storyboard := pl.background.GetStoryboard(); storyboard != nil {
				drawWithBackground(13, fmt.Sprintf("SB sprites: %d", pl.storyboardDrawn))
				drawWithBackground(14, fmt.Sprintf("SB load: %.2f", pl.storyboardLoad))
			}

			for _, t := range queue {
				pl.batch.SetColor(1, 1, 1, 1)
				pl.font.DrawMonospaced(pl.batch, 0, settings.Graphics.GetHeightF()-(size+padDown)*t.pos+padDown/2, size, t.text)
			}

			currentTime := int(pl.musicPlayer.GetPosition())
			totalTime := int(pl.musicPlayer.GetLength())
			mapTime := int(pl.bMap.HitObjects[len(pl.bMap.HitObjects)-1].GetBasicData().EndTime / 1000)

			drawShadowed(2, fmt.Sprintf("%02d:%02d / %02d:%02d (%02d:%02d)", currentTime/60, currentTime%60, totalTime/60, totalTime%60, mapTime/60, mapTime%60))
			drawShadowed(1, fmt.Sprintf("%d(*%d) hitobjects, %d total", len(pl.processed), settings.DIVIDES, len(pl.bMap.HitObjects)))

			if storyboard := pl.background.GetStoryboard(); storyboard != nil {
				drawShadowed(0, fmt.Sprintf("%d storyboard sprites, %d in queue (%d total)", pl.background.GetStoryboard().GetProcessedSprites(), storyboard.GetQueueSprites(), storyboard.GetTotalSprites()))
			} else {
				drawShadowed(0, "No storyboard")
			}
		}

		if settings.DEBUG || settings.Graphics.ShowFPS {
			drawShadowed := func(pos float64, text string) {
				pl.batch.SetColor(0, 0, 0, 0.5)
				pl.font.DrawMonospaced(pl.batch, settings.Graphics.GetWidthF()-pl.font.GetWidthMonospaced(size, text)-size*0.05, (size+padDown)*pos+padDown/2-size*0.05, size, text)

				pl.batch.SetColor(1, 1, 1, 1)
				pl.font.DrawMonospaced(pl.batch, settings.Graphics.GetWidthF()-pl.font.GetWidthMonospaced(size, text), (size+padDown)*pos+padDown/2, size, text)
			}

			fpsC := pl.profiler.GetFPS()
			fpsU := pl.profilerU.GetFPS()

			off := 0.0
			if pl.background.GetStoryboard() != nil {
				off = 1.0
			}

			drawFPS := fmt.Sprintf("%0.0ffps (%0.2fms)", fpsC, 1000/fpsC)
			updateFPS := fmt.Sprintf("%0.0ffps (%0.2fms)", fpsU, 1000/fpsU)
			sbFPS := ""

			if pl.background.GetStoryboard() != nil {
				fpsS := pl.background.GetStoryboard().GetFPS()
				sbFPS = fmt.Sprintf("%0.0ffps (%0.2fms)", fpsS, 1000/fpsS)
			}

			shift := strconv.Itoa(bmath.MaxI(len(drawFPS), bmath.MaxI(len(updateFPS), len(sbFPS))))

			drawShadowed(1+off, fmt.Sprintf("Draw: %"+shift+"s", drawFPS))
			drawShadowed(0+off, fmt.Sprintf("Update: %"+shift+"s", updateFPS))

			if pl.background.GetStoryboard() != nil {
				drawShadowed(0, fmt.Sprintf("Storyboard: %"+shift+"s", sbFPS))
			}
		}

		pl.batch.End()
	}
}
