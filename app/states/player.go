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
	"github.com/wieku/danser-go/app/input"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/app/states/components/common"
	"github.com/wieku/danser-go/app/states/components/containers"
	"github.com/wieku/danser-go/app/states/components/overlays"
	"github.com/wieku/danser-go/app/utils"
	"github.com/wieku/danser-go/framework/bass"
	"github.com/wieku/danser-go/framework/frame"
	batch2 "github.com/wieku/danser-go/framework/graphics/batch"
	"github.com/wieku/danser-go/framework/graphics/effects"
	"github.com/wieku/danser-go/framework/graphics/font"
	"github.com/wieku/danser-go/framework/graphics/texture"
	"github.com/wieku/danser-go/framework/math/animation"
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
)

const windowsOffset = 15

type Player struct {
	font        *font.Font
	bMap        *beatmap.BeatMap
	bloomEffect *effects.BloomEffect

	lastTime     int64
	lastMusicPos float64
	progressMsF  float64
	progressMs   int64

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
	counter         float64
	storyboardLoad  float64
	storyboardDrawn int
	mapFullName     string
	Epi             *texture.TextureRegion
	epiGlider       *animation.Glider
	overlay         overlays.Overlay
	blur            *effects.BlurEffect

	coin *common.DanserCoin

	hudGlider *animation.Glider

	volumeGlider *animation.Glider
	speedGlider  *animation.Glider
	pitchGlider  *animation.Glider

	startPoint    float64
	baseLimit     int
	updateLimiter *frame.Limiter

	objectsAlpha    *animation.Glider
	objectContainer *containers.HitObjectContainer

	MapEnd      float64
	RunningTime float64

	startOffset float64
	lateStart   bool
	mapEndL     float64
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

	player.musicPlayer = bass.NewTrack(filepath.Join(settings.General.OsuSongsDir, beatMap.Dir, beatMap.Audio))

	var err error
	player.Epi, err = utils.LoadTextureToAtlas(graphics.Atlas, "assets/textures/warning.png")

	if err != nil {
		log.Println(err)
	}

	if (settings.START > 0.01 || !math.IsInf(settings.END, 1)) && settings.PLAY {
		scrub := math.Max(0, settings.START*1000)
		end := settings.END*1000

		removed := false

		for i := 0; i < len(beatMap.HitObjects); i++ {
			o := beatMap.HitObjects[i]
			if o.GetStartTime() > scrub && end > o.GetEndTime() {
				continue
			}

			beatMap.HitObjects = append(beatMap.HitObjects[:i], beatMap.HitObjects[i+1:]...)
			i--

			removed = true
		}

		for i := 0; i < len(beatMap.HitObjects); i++ {
			beatMap.HitObjects[i].SetID(int64(i))
		}

		for i := 0; i < len(beatMap.Pauses); i++ {
			o := beatMap.Pauses[i]
			if o.GetStartTime() > scrub && end > o.GetEndTime() {
				continue
			}

			beatMap.Pauses = append(beatMap.Pauses[:i], beatMap.Pauses[i+1:]...)
			i--
		}

		if removed && settings.START > 0.01 {
			settings.START = 0
			settings.SKIP = true
		}
	}

	player.background = common.NewBackground()
	player.background.SetBeatmap(beatMap, settings.Playfield.Background.LoadStoryboards)

	player.camera = camera2.NewCamera()
	player.camera.SetOsuViewport(int(settings.Graphics.GetWidth()), int(settings.Graphics.GetHeight()), settings.Playfield.Scale, settings.Playfield.OsuShift)
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

	log.Println("Audio track:", beatMap.Audio)

	player.Scl = 1
	player.fadeOut = 1.0
	player.fadeIn = 0.0

	player.volumeGlider = animation.NewGlider(1)
	player.speedGlider = animation.NewGlider(settings.SPEED)
	player.pitchGlider = animation.NewGlider(settings.PITCH)
	player.hudGlider = animation.NewGlider(0)
	player.dimGlider = animation.NewGlider(0)
	player.blurGlider = animation.NewGlider(0)
	player.fxGlider = animation.NewGlider(0)
	player.cursorGlider = animation.NewGlider(0)
	player.epiGlider = animation.NewGlider(0)
	player.objectsAlpha = animation.NewGlider(1)

	if _, ok := player.overlay.(*overlays.ScoreOverlay); ok && player.controller.GetCursors()[0].Name == "" {
		player.cursorGlider.SetValue(1.0)
	}

	skipTime := 0.0
	if settings.SKIP {
		skipTime = beatMap.HitObjects[0].GetStartTime()
	}

	skipTime = math.Max(skipTime, settings.START*1000)

	beatmapStart := math.Max(beatMap.HitObjects[0].GetStartTime(), settings.START*1000)
	beatmapEnd := beatMap.HitObjects[len(beatMap.HitObjects)-1].GetEndTime() + float64(beatMap.Diff.Hit50)

	if !math.IsInf(settings.END, 1) {
		end := settings.END*1000
		beatmapEnd = math.Min(end, beatMap.HitObjects[len(beatMap.HitObjects)-1].GetEndTime()) + float64(beatMap.Diff.Hit50)
	}

	startOffset := 0.0

	if settings.SKIP || settings.START > 0.01 {
		startOffset = math.Max(0, skipTime-beatMap.Diff.Preempt)
		player.startPoint = startOffset

		for _, o := range beatMap.HitObjects {
			if o.GetStartTime() > player.startPoint {
				break
			}

			o.DisableAudioSubmission(true)
		}

		player.volumeGlider.SetValue(0.0)
		player.volumeGlider.AddEvent(skipTime-beatMap.Diff.Preempt, skipTime-beatMap.Diff.Preempt+difficulty.HitFadeIn, 1.0)

		if settings.START > 0.01 {
			player.objectsAlpha.SetValue(0.0)
			player.objectsAlpha.AddEvent(skipTime-beatMap.Diff.Preempt, skipTime-beatMap.Diff.Preempt+difficulty.HitFadeIn, 1.0)

			if player.overlay != nil {
				player.overlay.DisableAudioSubmission(true)
			}

			for i := -1000.0; i < player.startPoint-player.bMap.Diff.Preempt; i += 1.0 {
				player.controller.Update(i, 1)

				if player.overlay != nil {
					player.overlay.Update(i)
				}
			}

			if player.overlay != nil {
				player.overlay.DisableAudioSubmission(false)
			}
		}

		player.lateStart = true
	} else {
		startOffset = -beatMap.Diff.Preempt
	}

	startOffset += -settings.Playfield.LeadInHold * 1000

	player.dimGlider.AddEvent(startOffset-500, startOffset, 1.0-settings.Playfield.Background.Dim.Intro)
	player.blurGlider.AddEvent(startOffset-500, startOffset, settings.Playfield.Background.Blur.Values.Intro)
	player.fxGlider.AddEvent(startOffset-500, startOffset, 1.0-settings.Playfield.Logo.Dim.Intro)
	player.hudGlider.AddEvent(startOffset-500, startOffset, 1.0)

	if _, ok := player.overlay.(*overlays.ScoreOverlay); !ok {
		player.cursorGlider.AddEvent(startOffset-500, startOffset, 0.0)
	}

	player.dimGlider.AddEvent(beatmapStart-750, beatmapStart-250, 1.0-settings.Playfield.Background.Dim.Normal)
	player.blurGlider.AddEvent(beatmapStart-750, beatmapStart-250, settings.Playfield.Background.Blur.Values.Normal)
	player.fxGlider.AddEvent(beatmapStart-750, beatmapStart-250, 1.0-settings.Playfield.Logo.Dim.Normal)
	player.cursorGlider.AddEvent(beatmapStart-750, beatmapStart-250, 1.0)

	fadeOut := settings.Playfield.FadeOutTime * 1000

	if s, ok := player.overlay.(*overlays.ScoreOverlay); ok {
		if settings.Gameplay.ShowResultsScreen {
			beatmapEnd += 1000
			fadeOut = 250
		}

		s.SetBeatmapEnd(beatmapEnd+fadeOut)
	}

	if !math.IsInf(settings.END, 1) {
		for _, o := range beatMap.HitObjects {
			if o.GetEndTime() <= beatmapEnd {
				continue
			}

			o.DisableAudioSubmission(true)
		}

		if !settings.PLAY {
			player.objectsAlpha.AddEvent(beatmapEnd, beatmapEnd+fadeOut, 0)
		}
	}

	player.dimGlider.AddEvent(beatmapEnd, beatmapEnd+fadeOut, 0.0)
	player.fxGlider.AddEvent(beatmapEnd, beatmapEnd+fadeOut, 0.0)
	player.cursorGlider.AddEvent(beatmapEnd, beatmapEnd+fadeOut, 0.0)
	player.hudGlider.AddEvent(beatmapEnd, beatmapEnd+fadeOut, 0.0)

	player.mapEndL = beatmapEnd + fadeOut
	player.MapEnd = beatmapEnd + fadeOut

	if _, ok := player.overlay.(*overlays.ScoreOverlay); ok && settings.Gameplay.ShowResultsScreen {
		player.speedGlider.AddEvent(beatmapEnd+fadeOut, beatmapEnd+fadeOut, 1)
		player.pitchGlider.AddEvent(beatmapEnd+fadeOut, beatmapEnd+fadeOut, 1)

		player.MapEnd += (settings.Gameplay.ResultsScreenTime + 1) * 1000
		if player.MapEnd < player.musicPlayer.GetLength()*1000 {
			player.volumeGlider.AddEvent(player.MapEnd-settings.Gameplay.ResultsScreenTime*1000-500, player.MapEnd, 0.0)
		}
	} else {
		player.volumeGlider.AddEvent(beatmapEnd, beatmapEnd+fadeOut, 0.0)
	}

	player.MapEnd += 100

	if settings.Playfield.SeizureWarning.Enabled {
		am := math.Max(1000, settings.Playfield.SeizureWarning.Duration*1000)
		startOffset -= am
		player.epiGlider.AddEvent(startOffset, startOffset+500, 1.0)
		player.epiGlider.AddEvent(startOffset+am-500, startOffset+am, 0.0)
	}

	startOffset -= math.Max(settings.Playfield.LeadInTime*1000, 1000)

	player.startOffset = startOffset
	player.progressMsF = startOffset

	player.RunningTime = player.MapEnd - startOffset

	for _, p := range beatMap.Pauses {
		startTime := p.GetStartTime()
		endTime := p.GetEndTime()

		if endTime-startTime < 1000 || endTime < player.startPoint || startTime > player.MapEnd {
			continue
		}

		player.dimGlider.AddEvent(startTime, startTime+500*settings.SPEED, 1.0-settings.Playfield.Background.Dim.Breaks)
		player.blurGlider.AddEvent(startTime, startTime+500*settings.SPEED, settings.Playfield.Background.Blur.Values.Breaks)
		player.fxGlider.AddEvent(startTime, startTime+500*settings.SPEED, 1.0-settings.Playfield.Logo.Dim.Breaks)

		if !settings.Cursor.ShowCursorsOnBreaks {
			player.cursorGlider.AddEvent(startTime, startTime+100*settings.SPEED, 0.0)
		}

		player.dimGlider.AddEvent(endTime, endTime+500*settings.SPEED, 1.0-settings.Playfield.Background.Dim.Normal)
		player.blurGlider.AddEvent(endTime, endTime+500*settings.SPEED, settings.Playfield.Background.Blur.Values.Normal)
		player.fxGlider.AddEvent(endTime, endTime+500*settings.SPEED, 1.0-settings.Playfield.Logo.Dim.Normal)
		player.cursorGlider.AddEvent(endTime, endTime+100*settings.SPEED, 1.0)
	}

	player.background.SetTrack(player.musicPlayer)

	player.coin = common.NewDanserCoin()
	player.coin.SetMap(beatMap, player.musicPlayer)

	player.coin.SetScale(0.25 * math.Min(settings.Graphics.GetWidthF(), settings.Graphics.GetHeightF()))

	player.profiler = frame.NewCounter()

	player.bloomEffect = effects.NewBloomEffect(int(settings.Graphics.GetWidth()), int(settings.Graphics.GetHeight()))
	player.blur = effects.NewBlurEffect(int(settings.Graphics.GetWidth()), int(settings.Graphics.GetHeight()))

	player.background.Update(player.progressMsF, settings.Graphics.GetWidthF()/2, settings.Graphics.GetHeightF()/2)

	player.profilerU = frame.NewCounter()

	player.baseLimit = 2000

	player.updateLimiter = frame.NewLimiter(player.baseLimit)

	if settings.RECORD {
		return player
	}

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

		var lastTimeNano = qpc.GetNanoTime()

		for !input.Win.ShouldClose() {
			currentTimeNano := qpc.GetNanoTime()

			delta := float64(currentTimeNano-lastTimeNano) / 1000000.0

			player.profilerU.PutSample(delta)

			if player.musicPlayer.GetState() == bass.MUSIC_STOPPED {
				player.progressMsF += delta
			} else {
				platformOffset := 0.0
				if runtime.GOOS == "windows" {
					platformOffset = windowsOffset
				}

				musicPos := player.musicPlayer.GetPosition()*1000 + (platformOffset+float64(settings.Audio.Offset))*settings.SPEED

				if musicPos != player.lastMusicPos {
					player.progressMsF = musicPos
					player.lastMusicPos = musicPos
				} else {
					player.progressMsF += delta * settings.SPEED
				}
			}

			player.updateMain(delta)

			lastTimeNano = currentTimeNano

			player.updateLimiter.Sync()
		}

		player.musicPlayer.Stop()
		bass.StopLoops()
	}()

	return player
}

func (player *Player) Update(delta float64) bool {
	if player.musicPlayer.GetState() == bass.MUSIC_PLAYING {
		player.progressMsF += delta * player.musicPlayer.GetTempo()
	} else {
		player.progressMsF += delta
	}

	bass.GlobalTimeMs += delta

	if player.progressMsF > player.startPoint && settings.RECORD {
		player.musicPlayer.SetPositionF(player.progressMsF / 1000)
	}

	player.updateMain(delta)

	if player.progressMsF >= player.MapEnd {
		player.musicPlayer.Stop()
		bass.StopLoops()

		return true
	}

	return false
}

func (player *Player) GetTimeOffset() float64 {
	return player.progressMsF - player.startOffset
}

func (player *Player) updateMain(delta float64) {
	if player.progressMsF >= player.startPoint && !player.start {
		player.musicPlayer.Play()

		if ov, ok := player.overlay.(*overlays.ScoreOverlay); ok {
			ov.SetMusic(player.musicPlayer)
		}

		player.musicPlayer.SetPosition(player.startPoint / 1000)

		discord.SetDuration(int64((player.musicPlayer.GetLength() - player.musicPlayer.GetPosition()) * 1000 / settings.SPEED))

		if player.overlay == nil {
			discord.UpdateDance(settings.TAG, settings.DIVIDES)
		}

		player.start = true
	}

	player.speedGlider.Update(player.progressMsF)
	player.pitchGlider.Update(player.progressMsF)

	player.musicPlayer.SetTempo(player.speedGlider.GetValue())
	player.musicPlayer.SetPitch(player.pitchGlider.GetValue())

	if player.progressMsF >= player.startPoint-player.bMap.Diff.Preempt {
		if _, ok := player.controller.(*dance.GenericController); ok {
			player.bMap.Update(player.progressMsF)
		}

		player.objectContainer.Update(player.progressMsF)
	}

	if player.progressMsF >= player.startPoint/*-player.bMap.Diff.Preempt*/ || settings.PLAY {
		if player.progressMsF < player.mapEndL {
			player.controller.Update(player.progressMsF, delta)
		} else {
			if player.overlay != nil {
				player.overlay.DisableAudioSubmission(true)
			}
			player.controller.Update(player.bMap.HitObjects[len(player.bMap.HitObjects)-1].GetEndTime() + float64(player.bMap.Diff.Hit50) + 100, delta)
		}

		if player.lateStart {
			if player.overlay != nil {
				player.overlay.Update(player.progressMsF)
			}
		}
	}

	if player.overlay != nil && !player.lateStart {
		player.overlay.Update(player.progressMsF)
	}

	player.updateMusic(delta)

	player.coin.Update(player.progressMsF)
	player.coin.SetAlpha(float32(player.fxGlider.GetValue()))

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
	player.hudGlider.Update(player.progressMsF)
	player.volumeGlider.Update(player.progressMsF)
	player.objectsAlpha.Update(player.progressMsF)

	if player.musicPlayer.GetState() == bass.MUSIC_PLAYING {
		player.musicPlayer.SetVolumeRelative(player.volumeGlider.GetValue())
	}
}

func (player *Player) updateMusic(delta float64) {
	player.musicPlayer.Update()

	target := bmath.ClampF64(player.musicPlayer.GetBoost()*(settings.Audio.BeatScale-1.0)+1.0, 1.0, settings.Audio.BeatScale)

	if settings.Audio.BeatUseTimingPoints {
		player.Scl = 1 + player.coin.Progress*(settings.Audio.BeatScale-1.0)
	} else if player.Scl < target {
		player.Scl += (target - player.Scl) * 0.3 * delta / 16.66667
	} else if player.Scl > target {
		player.Scl -= (player.Scl - target) * 0.15 * delta / 16.66667
	}
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

	if fps > 58 && timMs > 18 && !settings.RECORD {
		log.Println(fmt.Sprintf("Slow frame detected! Frame time: %.3fms | Av. frame time: %.3fms", timMs, 1000.0/fps))
	}

	player.progressMs = int64(player.progressMsF)

	player.profiler.PutSample(timMs)
	player.lastTime = tim

	bgAlpha := player.dimGlider.GetValue()

	cameras := player.camera.GenRotated(settings.DIVIDES, -2*math.Pi/float64(settings.DIVIDES))
	cameras1 := player.camera1.GenRotated(settings.DIVIDES, -2*math.Pi/float64(settings.DIVIDES))

	if settings.Playfield.Background.FlashToTheBeat {
		bgAlpha = bmath.ClampF64(bgAlpha*player.Scl, 0, 1)
	}

	player.background.Draw(player.progressMsF, player.batch, player.blurGlider.GetValue(), bgAlpha, cameras1[0])

	if player.start {
		settings.Cursor.Colors.Update(timMs)
	}

	cursorColors := settings.Cursor.GetColors(settings.DIVIDES, len(player.controller.GetCursors()), player.Scl, player.cursorGlider.GetValue())

	if player.overlay != nil {
		player.batch.Begin()
		player.batch.ResetTransform()
		player.batch.SetScale(1, 1)

		player.batch.SetCamera(cameras[0])

		player.overlay.DrawBeforeObjects(player.batch, cursorColors, player.hudGlider.GetValue())

		player.batch.End()
		player.batch.ResetTransform()
		player.batch.SetColor(1, 1, 1, 1)
	}

	if player.epiGlider.GetValue() > 0.01 {
		player.batch.Begin()
		player.batch.ResetTransform()
		player.batch.SetColor(1, 1, 1, player.epiGlider.GetValue())
		player.batch.SetCamera(mgl32.Ortho(float32(-settings.Graphics.GetWidthF()/2), float32(settings.Graphics.GetWidthF()/2), float32(settings.Graphics.GetHeightF()/2), float32(-settings.Graphics.GetHeightF()/2), 1, -1))

		scl := scaling.Fit.Apply(player.Epi.Width, player.Epi.Height, float32(settings.Graphics.GetWidthF()), float32(settings.Graphics.GetHeightF()))
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

		player.coin.DrawVisualiser(settings.Playfield.Logo.DrawSpectrum)
		player.coin.Draw(player.progressMsF, player.batch)

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

	player.objectContainer.Draw(player.batch, cameras, player.progressMsF, float32(player.Scl), float32(player.objectsAlpha.GetValue()))

	if player.overlay != nil {
		player.batch.Begin()
		player.batch.SetScale(1, 1)

		player.batch.SetCamera(cameras[0])

		player.overlay.DrawNormal(player.batch, cursorColors, player.hudGlider.GetValue())

		player.batch.End()
	}

	player.background.DrawOverlay(player.progressMsF, player.batch, bgAlpha, cameras1[0])

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

		player.batch.SetCamera(player.scamera.GetProjectionView())

		player.overlay.DrawHUD(player.batch, cursorColors, player.hudGlider.GetValue())

		player.batch.End()
	}

	if settings.Playfield.Bloom.Enabled {
		player.bloomEffect.EndAndRender()
	}

	player.drawDebug()
}

func (player *Player) drawDebug() {
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

			msaa := "OFF"
			if settings.Graphics.MSAA > 0 {
				msaa = strconv.Itoa(int(settings.Graphics.MSAA)) + "x"
			}

			drawWithBackground(6, fmt.Sprintf("MSAA: %s", msaa))

			drawWithBackground(7, fmt.Sprintf("FBO Binds: %d", statistic.GetPrevious(statistic.FBOBinds)))
			drawWithBackground(8, fmt.Sprintf("VAO Binds: %d", statistic.GetPrevious(statistic.VAOBinds)))
			drawWithBackground(9, fmt.Sprintf("VBO Binds: %d", statistic.GetPrevious(statistic.VBOBinds)))
			drawWithBackground(10, fmt.Sprintf("Vertex Upload: %.2fk", float64(statistic.GetPrevious(statistic.VertexUpload))/1000))
			drawWithBackground(11, fmt.Sprintf("Vertices Drawn: %.2fk", float64(statistic.GetPrevious(statistic.VerticesDrawn))/1000))
			drawWithBackground(12, fmt.Sprintf("Draw Calls: %d", statistic.GetPrevious(statistic.DrawCalls)))
			drawWithBackground(13, fmt.Sprintf("Sprites Drawn: %d", statistic.GetPrevious(statistic.SpritesDrawn)))

			if storyboard := player.background.GetStoryboard(); storyboard != nil {
				drawWithBackground(14, fmt.Sprintf("SB sprites: %d", player.storyboardDrawn))
				drawWithBackground(15, fmt.Sprintf("SB load: %.2f", player.storyboardLoad))
			}

			for _, t := range queue {
				player.batch.SetColor(1, 1, 1, 1)
				player.font.DrawMonospaced(player.batch, 0, settings.Graphics.GetHeightF()-(size+padDown)*t.pos+padDown/2, size, t.text)
			}

			currentTime := int(player.musicPlayer.GetPosition())
			totalTime := int(player.musicPlayer.GetLength())
			mapTime := int(player.bMap.HitObjects[len(player.bMap.HitObjects)-1].GetEndTime() / 1000)

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

func (player *Player) Show() {}

func (player *Player) Hide() {}

func (player *Player) Dispose() {}
