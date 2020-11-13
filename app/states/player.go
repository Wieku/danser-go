package states

import (
	"fmt"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/wieku/danser-go/app/beatmap"
	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/bmath"
	camera2 "github.com/wieku/danser-go/app/bmath/camera"
	"github.com/wieku/danser-go/app/dance"
	"github.com/wieku/danser-go/app/discord"
	"github.com/wieku/danser-go/app/graphics"
	"github.com/wieku/danser-go/app/graphics/font"
	"github.com/wieku/danser-go/app/graphics/gui/drawables"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/app/states/components/common"
	"github.com/wieku/danser-go/app/states/components/containers"
	"github.com/wieku/danser-go/app/states/components/overlays"
	"github.com/wieku/danser-go/app/utils"
	"github.com/wieku/danser-go/framework/bass"
	"github.com/wieku/danser-go/framework/frame"
	batch2 "github.com/wieku/danser-go/framework/graphics/batch"
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
	"strconv"
	"time"
)

type Player struct {
	font        *font.Font
	bMap        *beatmap.BeatMap
	bloomEffect *effects.BloomEffect
	lastTime    int64
	progressMsF float64
	progressMs  int64
	batch       *batch2.QuadBatch
	controller  dance.Controller
	background  *common.Background
	BgScl       vector.Vector2d
	Scl         float64
	SclA        float64
	fadeOut     float64
	fadeIn      float64
	start       bool
	musicPlayer *bass.Track
	profiler    *frame.Counter
	profilerU   *frame.Counter

	camera          *camera2.Camera
	camera1         *camera2.Camera
	scamera         *camera2.Camera
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
	overlay         overlays.Overlay
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

	hudGlider *animation.Glider

	volumeGlider  *animation.Glider
	startPoint    float64
	baseLimit     int
	updateLimiter *frame.Limiter

	objectContainer *containers.HitObjectContainer
}

func NewPlayer(beatMap *beatmap.BeatMap) *Player {
	player := new(Player)

	graphics.LoadTextures()

	player.batch = batch2.NewQuadBatch()
	player.font = font.GetFont("Exo 2 Bold")

	discord.SetMap(beatMap.Artist, beatMap.Name, beatMap.Difficulty)

	player.bMap = beatMap
	player.mapFullName = fmt.Sprintf("%s - %s [%s]", beatMap.Artist, beatMap.Name, beatMap.Difficulty)
	log.Println("Playing:", player.mapFullName)

	var err error

	LogoT, err := utils.LoadTextureToAtlas(graphics.Atlas, "assets/textures/coinbig.png")
	if err != nil {
		panic(err)
	}

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

	player.background = common.NewBackground()
	player.background.SetBeatmap(beatMap, settings.Playfield.Background.LoadStoryboards)

	player.camera = camera2.NewCamera()
	player.camera.SetOsuViewport(int(settings.Graphics.GetWidth()), int(settings.Graphics.GetHeight()), settings.Playfield.Scale, settings.Playfield.OsuShift)
	//player.camera.SetOrigin(bmath.NewVec2d(256, 192.0-5))
	player.camera.Update()

	player.camera1 = camera2.NewCamera()

	sbScale := 1.0
	if settings.Playfield.ScaleStoryboardWithPlayfield {
		sbScale = settings.Playfield.Scale
	}

	player.camera1.SetOsuViewport(int(settings.Graphics.GetWidth()), int(settings.Graphics.GetHeight()), sbScale, false)
	player.camera1.Update()

	player.scamera = camera2.NewCamera()
	player.scamera.SetViewport(int(settings.Graphics.GetWidth()), int(settings.Graphics.GetHeight()), false)
	player.scamera.SetOrigin(vector.NewVec2d(settings.Graphics.GetWidthF()/2, settings.Graphics.GetHeightF()/2))
	player.scamera.Update()

	graphics.Camera = player.camera

	player.bMap.Reset()
	if settings.PLAY {
		player.controller = dance.NewPlayerController()

		player.controller.SetBeatMap(player.bMap)
		player.controller.InitCursors()
		player.overlay = overlays.NewScoreOverlay(player.controller.(*dance.PlayerController).GetRuleset(), player.controller.GetCursors()[0])
	} else if settings.KNOCKOUT {
		controller := dance.NewReplayController()
		player.controller = controller
		player.controller.SetBeatMap(player.bMap)
		player.controller.InitCursors()

		if settings.PLAYERS == 1 {
			player.overlay = overlays.NewScoreOverlay(player.controller.(*dance.ReplayController).GetRuleset(), player.controller.GetCursors()[0])
		} else {
			player.overlay = overlays.NewKnockoutOverlay(controller.(*dance.ReplayController))
		}
	} else {
		player.controller = dance.NewGenericController()
		player.controller.SetBeatMap(player.bMap)
		player.controller.InitCursors()
	}

	player.lastTime = -1

	player.objectContainer = containers.NewHitObjectContainer(beatMap)

	log.Println("Track:", beatMap.Audio)

	player.Scl = 1
	player.fadeOut = 1.0
	player.fadeIn = 0.0

	player.volumeGlider = animation.NewGlider(1.0)
	player.hudGlider = animation.NewGlider(1.0)
	player.dimGlider = animation.NewGlider(0.0)
	player.blurGlider = animation.NewGlider(0.0)
	player.fxGlider = animation.NewGlider(0.0)
	if _, ok := player.overlay.(*overlays.ScoreOverlay); !ok {
		player.cursorGlider = animation.NewGlider(0.0)
	} else {
		player.cursorGlider = animation.NewGlider(1.0)
	}
	player.playersGlider = animation.NewGlider(0.0)

	skipTime := 0.0
	if settings.SKIP {
		skipTime = float64(beatMap.HitObjects[0].GetBasicData().StartTime)
	}

	skipTime = math.Max(skipTime, settings.SCRUB*1000)

	tmS := math.Max(float64(beatMap.HitObjects[0].GetBasicData().StartTime), settings.SCRUB*1000)
	tmE := float64(beatMap.HitObjects[len(beatMap.HitObjects)-1].GetBasicData().EndTime)

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
	if _, ok := player.overlay.(*overlays.ScoreOverlay); !ok {
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

	player.baseLimit = 2000

	player.updateLimiter = frame.NewLimiter(player.baseLimit)

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

				if ov, ok := player.overlay.(*overlays.ScoreOverlay); ok {
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

				player.objectContainer.Update(player.progressMsF)
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

			player.updateLimiter.Sync()
		}
	}()

	go func() {
		for {
			musicPlayer.Update()

			target := bmath.ClampF64(musicPlayer.GetBoost()*(settings.Audio.BeatScale-1.0)+1.0, 1.0, settings.Audio.BeatScale) //math.Min(1.4*settings.Audio.BeatScale, math.Max(math.Sin(musicPlayer.GetBeat()*math.Pi/2)*0.4*settings.Audio.BeatScale+1.0, 1.0))

			ratio1 := 15 / 16.6666666666667

			player.vol = player.musicPlayer.GetLevelCombined()
			player.volAverage = player.volAverage*0.9 + player.vol*0.1

			vprog := 1 - ((player.vol - player.volAverage) / 0.5)
			pV := math.Min(1.0, math.Max(0.0, 1.0-(vprog*0.5+player.beatProgress*0.5)))

			ratio := math.Pow(0.5, ratio1)

			player.progress = player.lastProgress*ratio + (pV)*(1-ratio)
			player.lastProgress = player.progress

			if settings.Audio.BeatUseTimingPoints {
				player.Scl = 1 + player.progress*(settings.Audio.BeatScale-1.0)
			} else {
				if player.Scl < target {
					player.Scl += (target - player.Scl) * 30 / 100
				} else if player.Scl > target {
					player.Scl -= (player.Scl - target) * 15 / 100
				}
			}

			time.Sleep(15 * time.Millisecond)
		}
	}()

	return player
}

func (player *Player) Show() {

}

func (player *Player) Draw(float64) {
	if player.lastTime <= 0 {
		player.lastTime = qpc.GetNanoTime()
	}

	tim := qpc.GetNanoTime()
	timMs := float64(tim-player.lastTime) / 1000000.0

	fps := player.profiler.GetFPS()

	player.updateLimiter.FPS = bmath.ClampI(int(fps*1.2), player.baseLimit, 10000)

	if player.background.GetStoryboard() != nil {
		player.background.GetStoryboard().SetFPS(bmath.ClampI(int(fps*1.2), player.baseLimit, 10000))
	}

	if fps > 58 && timMs > 18 {
		log.Println(fmt.Sprintf("Slow frame detected! Frame time: %.3fms | Av. frame time: %.3fms", timMs, 1000.0/fps))
	}

	player.progressMs = int64(player.progressMsF)

	player.profiler.PutSample(timMs)
	player.lastTime = tim

	bgAlpha := player.dimGlider.GetValue()

	cameras := player.camera.GenRotated(settings.DIVIDES, -2*math.Pi/float64(settings.DIVIDES) /**player.unfold.GetValue()*/)
	cameras1 := player.camera1.GenRotated(settings.DIVIDES, -2*math.Pi/float64(settings.DIVIDES) /**player.unfold.GetValue()*/)

	if settings.Playfield.Background.FlashToTheBeat {
		bgAlpha *= player.Scl
	}

	player.background.Draw(player.progressMs, player.batch, player.blurGlider.GetValue(), bgAlpha, cameras1[0])

	if player.start {
		settings.Cursor.Colors.Update(timMs)
	}

	cursorColors := settings.Cursor.GetColors(settings.DIVIDES, len(player.controller.GetCursors()), player.Scl, player.cursorGlider.GetValue())

	if player.overlay != nil {
		player.batch.Begin()
		player.batch.ResetTransform()
		player.batch.SetScale(1, 1)

		player.batch.SetCamera(cameras[0])

		player.overlay.DrawBeforeObjects(player.batch, cursorColors, player.playersGlider.GetValue()*player.hudGlider.GetValue())

		player.batch.End()
		player.batch.ResetTransform()
		player.batch.SetColor(1, 1, 1, 1)
	}

	if player.epiGlider.GetValue() > 0.01 {
		player.batch.Begin()
		player.batch.ResetTransform()
		player.batch.SetColor(1, 1, 1, player.epiGlider.GetValue())
		player.batch.SetCamera(mgl32.Ortho(float32(-settings.Graphics.GetWidthF()/2), float32(settings.Graphics.GetWidthF()/2), float32(settings.Graphics.GetHeightF()/2), float32(-settings.Graphics.GetHeightF()/2), 1, -1))

		scl := scaling.Fit.Apply(float32(player.Epi.Width), float32(player.Epi.Height), float32(settings.Graphics.GetWidthF()), float32(settings.Graphics.GetHeightF()))
		scl = scl.Scl(0.5).Scl(0.66)
		player.batch.SetScale(scl.X64(), scl.Y64())
		player.batch.DrawUnit(*player.Epi)

		//player.batch.SetScale(1.0, -1.0)
		//s := "Support me on ko-fi.com/wiekus"
		//width := player.font.GetWidth(settings.Graphics.GetHeightF()/40, s)
		//player.font.Draw(player.batch, -width/2, (0.77)*(settings.Graphics.GetHeightF()/2), settings.Graphics.GetHeightF()/40, s)

		player.batch.ResetTransform()
		player.batch.End()
		player.batch.SetColor(1, 1, 1, 1)
	}

	player.counter += timMs

	if player.counter >= 1000.0/60 {
		player.counter -= 1000.0 / 60
		if player.background.GetStoryboard() != nil {
			player.storyboardLoad = player.background.GetStoryboard().GetLoad()
			player.storyboardDrawn = player.background.GetStoryboard().GetRenderedSprites()
		}
	}

	if player.fxGlider.GetValue() > 0.01 {
		player.batch.Begin()
		player.batch.ResetTransform()
		player.batch.SetColor(1, 1, 1, player.fxGlider.GetValue())
		player.batch.SetCamera(mgl32.Ortho(float32(-settings.Graphics.GetWidthF()/2), float32(settings.Graphics.GetWidthF()/2), float32(settings.Graphics.GetHeightF()/2), float32(-settings.Graphics.GetHeightF()/2), 1, -1))

		innerCircleScale := 1.05 - easing.OutQuad(player.progress)*0.05
		outerCircleScale := 1.05 + easing.OutQuad(player.progress)*0.03

		if settings.Playfield.Logo.DrawSpectrum {
			player.visualiser.SetStartDistance(player.cookieSize * 0.5 * innerCircleScale)
			player.visualiser.Draw(player.progressMsF, player.batch)
		}

		player.batch.SetColor(1, 1, 1, player.fxGlider.GetValue())

		scl := (player.cookieSize / 2048.0) * 1.05

		player.LogoS1.SetScale(innerCircleScale * scl)
		player.LogoS2.SetScale(outerCircleScale * scl)

		alpha := 0.3
		if player.bMap.Timings.Current.Kiai {
			alpha = 0.12
		}

		player.LogoS2.SetAlpha(float32(alpha * (1 - easing.OutQuad(player.progress))))

		player.LogoS1.UpdateAndDraw(player.progressMs, player.batch)
		player.LogoS2.UpdateAndDraw(player.progressMs, player.batch)
		player.batch.ResetTransform()
		player.batch.End()
	}

	scale2 := player.Scl

	if !settings.Cursor.ScaleToTheBeat {
		scale2 = 1
	}

	if settings.Playfield.Bloom.Enabled {
		player.bloomEffect.SetThreshold(settings.Playfield.Bloom.Threshold)
		player.bloomEffect.SetBlur(settings.Playfield.Bloom.Blur)
		player.bloomEffect.SetPower(settings.Playfield.Bloom.Power + settings.Playfield.Bloom.BloomBeatAddition*(player.Scl-1.0)/(settings.Audio.BeatScale*0.4))
		player.bloomEffect.Begin()
	}

	player.objectContainer.Draw(player.batch, cameras, player.progressMsF, float32(player.Scl), 1.0)

	if player.overlay != nil && player.overlay.NormalBeforeCursor() {
		player.batch.Begin()
		player.batch.SetScale(1, 1)

		player.batch.SetCamera(cameras[0])

		player.overlay.DrawNormal(player.batch, cursorColors, player.playersGlider.GetValue()*player.hudGlider.GetValue())

		player.batch.End()
	}

	player.background.DrawOverlay(player.progressMs, player.batch, bgAlpha, cameras1[0])

	if settings.Playfield.DrawCursors {
		for _, g := range player.controller.GetCursors() {
			g.UpdateRenderer()
		}

		player.batch.SetAdditive(false)

		graphics.BeginCursorRender()

		for j := 0; j < settings.DIVIDES; j++ {
			player.batch.SetCamera(cameras[j])

			for i, g := range player.controller.GetCursors() {
				if player.overlay != nil && player.overlay.IsBroken(g) {
					continue
				}

				baseIndex := j*len(player.controller.GetCursors()) + i

				ind := baseIndex - 1
				if ind < 0 {
					ind = settings.DIVIDES*len(player.controller.GetCursors()) - 1
				}

				col1 := cursorColors[baseIndex]
				col2 := cursorColors[ind]

				g.DrawM(scale2, player.batch, col1, col2)
			}
		}

		graphics.EndCursorRender()
	}

	player.batch.SetAdditive(false)

	if player.overlay != nil {
		player.batch.Begin()
		player.batch.SetScale(1, 1)

		if !player.overlay.NormalBeforeCursor() {
			player.overlay.DrawNormal(player.batch, cursorColors, player.playersGlider.GetValue()*player.hudGlider.GetValue())
		}

		player.batch.SetCamera(player.scamera.GetProjectionView())

		player.overlay.DrawHUD(player.batch, cursorColors, player.playersGlider.GetValue()*player.hudGlider.GetValue())

		player.batch.End()
	}

	if settings.Playfield.Bloom.Enabled {
		player.bloomEffect.EndAndRender()
	}

	if settings.DEBUG || settings.Graphics.ShowFPS {
		player.batch.Begin()
		player.batch.SetColor(1, 1, 1, 1)
		player.batch.SetScale(1, 1)
		player.batch.SetCamera(player.scamera.GetProjectionView())

		padDown := 4.0 * (settings.Graphics.GetHeightF() / 1080.0)
		//shift := 16.0 * (settings.Graphics.GetHeightF() / 1080.0)
		size := 16.0 * (settings.Graphics.GetHeightF() / 1080.0)

		if settings.DEBUG {
			drawShadowed := func(pos float64, text string) {
				player.batch.SetColor(0, 0, 0, 0.5)
				player.font.DrawMonospaced(player.batch, 0-size*0.05, (size+padDown)*pos+padDown/2-size*0.05, size, text)

				player.batch.SetColor(1, 1, 1, 1)
				player.font.DrawMonospaced(player.batch, 0, (size+padDown)*pos+padDown/2, size, text)
			}

			player.batch.SetColor(0, 0, 0, 1)
			player.font.Draw(player.batch, 0-size*1.5*0.1, settings.Graphics.GetHeightF()-size*1.5-size*1.5*0.1, size*1.5, player.mapFullName)

			player.batch.SetColor(1, 1, 1, 1)
			player.font.Draw(player.batch, 0, settings.Graphics.GetHeightF()-size*1.5, size*1.5, player.mapFullName)

			type tx struct {
				pos  float64
				text string
			}

			var queue []tx

			drawWithBackground := func(pos float64, text string) {
				player.batch.SetColor(0, 0, 0, 0.8)

				width := player.font.GetWidthMonospaced(size, text)

				player.batch.SetTranslation(vector.NewVec2d(width/2, settings.Graphics.GetHeightF()-(size+padDown)*(pos-0.5)))
				player.batch.SetSubScale(width/2, (size+padDown)/2)
				player.batch.DrawUnit(graphics.Pixel.GetRegion())

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

			if storyboard := player.background.GetStoryboard(); storyboard != nil {
				drawWithBackground(13, fmt.Sprintf("SB sprites: %d", player.storyboardDrawn))
				drawWithBackground(14, fmt.Sprintf("SB load: %.2f", player.storyboardLoad))
			}

			for _, t := range queue {
				player.batch.SetColor(1, 1, 1, 1)
				player.font.DrawMonospaced(player.batch, 0, settings.Graphics.GetHeightF()-(size+padDown)*t.pos+padDown/2, size, t.text)
			}

			currentTime := int(player.musicPlayer.GetPosition())
			totalTime := int(player.musicPlayer.GetLength())
			mapTime := int(player.bMap.HitObjects[len(player.bMap.HitObjects)-1].GetBasicData().EndTime / 1000)

			drawShadowed(2, fmt.Sprintf("%02d:%02d / %02d:%02d (%02d:%02d)", currentTime/60, currentTime%60, totalTime/60, totalTime%60, mapTime/60, mapTime%60))
			drawShadowed(1, fmt.Sprintf("%d(*%d) hitobjects, %d total" /*len(player.processed)*/, 0, settings.DIVIDES, len(player.bMap.HitObjects)))

			if storyboard := player.background.GetStoryboard(); storyboard != nil {
				drawShadowed(0, fmt.Sprintf("%d storyboard sprites, %d in queue (%d total)", player.background.GetStoryboard().GetProcessedSprites(), storyboard.GetQueueSprites(), storyboard.GetTotalSprites()))
			} else {
				drawShadowed(0, "No storyboard")
			}
		}

		if settings.DEBUG || settings.Graphics.ShowFPS {
			drawShadowed := func(pos float64, text string) {
				player.batch.SetColor(0, 0, 0, 0.5)
				player.font.DrawMonospaced(player.batch, settings.Graphics.GetWidthF()-player.font.GetWidthMonospaced(size, text)-size*0.05, (size+padDown)*pos+padDown/2-size*0.05, size, text)

				player.batch.SetColor(1, 1, 1, 1)
				player.font.DrawMonospaced(player.batch, settings.Graphics.GetWidthF()-player.font.GetWidthMonospaced(size, text), (size+padDown)*pos+padDown/2, size, text)
			}

			fpsC := player.profiler.GetFPS()
			fpsU := player.profilerU.GetFPS()

			off := 0.0
			if player.background.GetStoryboard() != nil {
				off = 1.0
			}

			drawFPS := fmt.Sprintf("%0.0ffps (%0.2fms)", fpsC, 1000/fpsC)
			updateFPS := fmt.Sprintf("%0.0ffps (%0.2fms)", fpsU, 1000/fpsU)
			sbFPS := ""

			if player.background.GetStoryboard() != nil {
				fpsS := player.background.GetStoryboard().GetFPS()
				sbFPS = fmt.Sprintf("%0.0ffps (%0.2fms)", fpsS, 1000/fpsS)
			}

			shift := strconv.Itoa(bmath.MaxI(len(drawFPS), bmath.MaxI(len(updateFPS), len(sbFPS))))

			drawShadowed(1+off, fmt.Sprintf("Draw: %"+shift+"s", drawFPS))
			drawShadowed(0+off, fmt.Sprintf("Update: %"+shift+"s", updateFPS))

			if player.background.GetStoryboard() != nil {
				drawShadowed(0, fmt.Sprintf("Storyboard: %"+shift+"s", sbFPS))
			}
		}

		player.batch.End()
	}
}

func (player *Player) Hide() {

}

func (player *Player) Dispose() {

}
