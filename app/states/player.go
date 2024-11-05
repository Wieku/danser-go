package states

import (
	"fmt"
	"github.com/dustin/go-humanize"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/wieku/danser-go/app/audio"
	"github.com/wieku/danser-go/app/beatmap"
	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/bmath"
	camera2 "github.com/wieku/danser-go/app/bmath/camera"
	"github.com/wieku/danser-go/app/dance"
	"github.com/wieku/danser-go/app/discord"
	"github.com/wieku/danser-go/app/graphics"
	"github.com/wieku/danser-go/app/input"
	"github.com/wieku/danser-go/app/osuapi"
	"github.com/wieku/danser-go/app/rulesets/osu"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/app/states/components/common"
	"github.com/wieku/danser-go/app/states/components/containers"
	"github.com/wieku/danser-go/app/states/components/overlays"
	"github.com/wieku/danser-go/app/utils"
	"github.com/wieku/danser-go/framework/bass"
	"github.com/wieku/danser-go/framework/frame"
	"github.com/wieku/danser-go/framework/goroutines"
	batch2 "github.com/wieku/danser-go/framework/graphics/batch"
	"github.com/wieku/danser-go/framework/graphics/effects"
	"github.com/wieku/danser-go/framework/graphics/font"
	"github.com/wieku/danser-go/framework/graphics/shape"
	"github.com/wieku/danser-go/framework/graphics/texture"
	"github.com/wieku/danser-go/framework/math/animation"
	"github.com/wieku/danser-go/framework/math/animation/easing"
	color2 "github.com/wieku/danser-go/framework/math/color"
	"github.com/wieku/danser-go/framework/math/mutils"
	"github.com/wieku/danser-go/framework/math/scaling"
	"github.com/wieku/danser-go/framework/math/vector"
	"github.com/wieku/danser-go/framework/profiler"
	"github.com/wieku/danser-go/framework/qpc"
	"log"
	"math"
	"math/rand"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const windowsOffset = 15

type Player struct {
	font        *font.Font
	bMap        *beatmap.BeatMap
	bloomEffect *effects.BloomEffect

	lastTime        int64
	lastMusicPos    float64
	lastProgressMsF float64
	progressMsF     float64
	rawPositionF    float64
	progressMs      int64

	batch       *batch2.QuadBatch
	controller  dance.Controller
	background  *common.Background
	BgScl       vector.Vector2d
	Scl         float64
	SclA        float64
	fadeOut     float64
	fadeIn      float64
	start       bool
	musicPlayer bass.ITrack
	profiler    *frame.Counter
	profilerU   *frame.Counter

	onlineOffset float64

	mainCamera   *camera2.Camera
	objectCamera *camera2.Camera
	bgCamera     *camera2.Camera
	uiCamera     *camera2.Camera

	dimGlider       *bmath.DimGlider
	blurGlider      *bmath.DimGlider
	fxGlider        *bmath.DimGlider
	cursorGlider    *animation.Glider
	counter         float64
	storyboardDrawn int
	mapFullName     string
	Epi             *texture.TextureRegion
	epiGlider       *animation.Glider
	overlay         overlays.Overlay
	blur            *effects.BlurEffect

	coin *common.DanserCoin

	hudGlider *animation.Glider

	volumeGlider    *animation.Glider
	speedGlider     *animation.Glider
	pitchGlider     *animation.Glider
	frequencyGlider *animation.Glider

	startPoint  float64
	startPointE float64

	baseLimit     int
	updateLimiter *frame.Limiter

	objectsAlpha    *animation.Glider
	objectContainer *containers.HitObjectContainer

	MapEnd      float64
	RunningTime float64

	startOffset float64
	lateStart   bool
	mapEndL     float64

	ScaledWidth  float64
	ScaledHeight float64

	nightcore *common.NightcoreProcessor

	realTime         float64
	objectsAlphaFail *animation.Glider
	failOX           *animation.Glider
	failOY           *animation.Glider
	failRotation     *animation.Glider

	failing bool
	failAt  float64
	failed  bool

	mProfiler *frame.Counter
	mStats1   *runtime.MemStats
	mStats2   *runtime.MemStats
	mBuffer   []byte
	memTicker *time.Ticker
	ftGraph   *shape.SteppingGraph
}

func NewPlayer(beatMap *beatmap.BeatMap) *Player {
	player := new(Player)
	player.mProfiler = frame.NewCounter()
	player.mStats1 = new(runtime.MemStats)
	player.mStats2 = new(runtime.MemStats)
	player.mBuffer = make([]byte, 0, 256)

	player.ftGraph = shape.NewSteppingGraph(1920-600-20, 1080-300-40, 600, 250, 6, 16.667, "ms")
	player.ftGraph.SetLabel(0, "BG Tasks", color2.NewIA(0x00ff34ff))
	player.ftGraph.SetLabel(1, "Input", color2.NewIA(0xffff00ff))
	player.ftGraph.SetLabel(2, "Draw", color2.NewIA(0xbf6500ff))
	player.ftGraph.SetLabel(3, "SwapBuffers", color2.NewIA(0xff0000ff))
	player.ftGraph.SetLabel(4, "Sleep", color2.NewIA(0x5555bbff))
	player.ftGraph.SetLabel(5, "Overhead", color2.NewIA(0xffffffff))

	graphics.LoadTextures()

	if settings.Graphics.Experimental.UsePersistentBuffers {
		player.batch = batch2.NewQuadBatchPersistent()
	} else {
		player.batch = batch2.NewQuadBatch()
	}

	player.font = font.GetFont("Quicksand Bold")

	discord.SetMap(beatMap.Artist, beatMap.Name, beatMap.Difficulty)

	player.bMap = beatMap
	player.mapFullName = fmt.Sprintf("%s - %s [%s]", beatMap.Artist, beatMap.Name, beatMap.Difficulty)
	log.Println("Playing:", player.mapFullName)

	var track *bass.TrackBass
	if fPath, err2 := beatMap.GetAudioFile(); err2 == nil {
		track = bass.NewTrack(fPath)
	}

	if track == nil {
		log.Println("Failed to create music stream, creating a dummy stream...")

		player.musicPlayer = bass.NewTrackVirtual(beatMap.HitObjects[len(beatMap.HitObjects)-1].GetEndTime()/1000 + 1)
	} else {
		log.Println("Audio track:", beatMap.Audio)

		player.musicPlayer = track
	}

	var err error
	player.Epi, err = utils.LoadTextureToAtlas(graphics.Atlas, "assets/textures/warning.png")

	if err != nil {
		log.Println(err)
	}

	settings.START = min(settings.START, (beatMap.HitObjects[len(beatMap.HitObjects)-1].GetStartTime()-1)/1000) // cap start to start time of the last HitObject - 1ms

	if (settings.START > 0.01 || !math.IsInf(settings.END, 1)) && (settings.PLAY || !settings.KNOCKOUT) {
		scrub := max(0, settings.START*1000)
		end := settings.END * 1000

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

	player.background = common.NewBackground(true)
	player.background.SetBeatmap(beatMap, true, true)

	player.mainCamera = camera2.NewCamera()
	player.mainCamera.SetOsuViewport(int(settings.Graphics.GetWidth()), int(settings.Graphics.GetHeight()), settings.Playfield.Scale, true, settings.Playfield.OsuShift)
	player.mainCamera.Update()

	player.objectCamera = camera2.NewCamera()
	player.objectCamera.SetOsuViewport(int(settings.Graphics.GetWidth()), int(settings.Graphics.GetHeight()), settings.Playfield.Scale, true, settings.Playfield.OsuShift)
	player.objectCamera.Update()

	player.bgCamera = camera2.NewCamera()

	sbScale := 1.0
	if settings.Playfield.ScaleStoryboardWithPlayfield {
		sbScale = settings.Playfield.Scale
	}

	player.bgCamera.SetOsuViewport(int(settings.Graphics.GetWidth()), int(settings.Graphics.GetHeight()), sbScale, !settings.Playfield.OsuShift && settings.Playfield.MoveStoryboardWithPlayfield, false)
	player.bgCamera.Update()

	player.ScaledHeight = 1080.0
	player.ScaledWidth = player.ScaledHeight * settings.Graphics.GetAspectRatio()

	player.uiCamera = camera2.NewCamera()
	player.uiCamera.SetViewport(int(player.ScaledWidth), int(player.ScaledHeight), true)
	player.uiCamera.SetViewportF(0, int(player.ScaledHeight), int(player.ScaledWidth), 0)
	player.uiCamera.Update()

	graphics.Camera = player.mainCamera

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

	player.Scl = 1
	player.fadeOut = 1.0
	player.fadeIn = 0.0

	player.volumeGlider = animation.NewGlider(1)
	player.speedGlider = animation.NewGlider(1)
	player.pitchGlider = animation.NewGlider(1)
	player.frequencyGlider = animation.NewGlider(1)

	player.hudGlider = animation.NewGlider(0)
	player.hudGlider.SetEasing(easing.OutQuad)

	player.dimGlider = bmath.NewDimGlider(0)
	player.dimGlider.SetEasing(easing.OutQuad)

	player.blurGlider = bmath.NewDimGlider(0)
	player.blurGlider.SetEasing(easing.OutQuad)

	player.fxGlider = bmath.NewDimGlider(0)
	player.cursorGlider = animation.NewGlider(0)
	player.epiGlider = animation.NewGlider(0)
	player.objectsAlpha = animation.NewGlider(1)

	player.objectsAlphaFail = animation.NewGlider(1)
	player.failOX = animation.NewGlider(0)
	player.failOY = animation.NewGlider(0)
	player.failRotation = animation.NewGlider(0)

	player.trySetupFail()

	preempt := min(1800, beatMap.Diff.Preempt)

	skipTime := 0.0
	if settings.SKIP {
		skipTime = beatMap.HitObjects[0].GetStartTime()
	}

	skipTime = max(skipTime, settings.START*1000) - preempt

	beatmapStart := max(beatMap.HitObjects[0].GetStartTime(), settings.START*1000) - preempt
	beatmapEnd := beatMap.HitObjects[len(beatMap.HitObjects)-1].GetEndTime() + float64(beatMap.Diff.Hit50)

	if !math.IsInf(settings.END, 1) {
		end := settings.END * 1000
		beatmapEnd = min(end, beatMap.HitObjects[len(beatMap.HitObjects)-1].GetEndTime()) + float64(beatMap.Diff.Hit50)
	}

	startOffset := 0.0

	if max(0, skipTime) > 0.01 {
		startOffset = skipTime
		player.startPoint = max(0, startOffset)

		for _, o := range beatMap.HitObjects {
			if o.GetStartTime() > player.startPoint {
				break
			}

			o.DisableAudioSubmission(true)
		}

		player.volumeGlider.SetValue(0.0)
		player.volumeGlider.AddEvent(skipTime, skipTime+beatMap.Diff.TimeFadeIn, 1.0)

		player.objectsAlpha.SetValue(0.0)
		player.objectsAlpha.AddEvent(skipTime, skipTime+beatMap.Diff.TimeFadeIn, 1.0)

		if player.overlay != nil {
			player.overlay.DisableAudioSubmission(true)
		}

		for i := -1000.0; i < startOffset; i += 1.0 {
			player.controller.Update(i, 1)

			if player.overlay != nil {
				player.overlay.Update(i)
			}
		}

		if player.overlay != nil {
			player.overlay.DisableAudioSubmission(false)
		}

		player.lateStart = true
	} else {
		startOffset = -preempt
	}

	player.startPointE = startOffset

	startOffset += -settings.Playfield.LeadInHold * 1000

	player.dimGlider.AddEvent(startOffset-500, startOffset, bmath.Intro)
	player.blurGlider.AddEvent(startOffset-500, startOffset, bmath.Intro)
	player.fxGlider.AddEvent(startOffset-500, startOffset, bmath.Intro)
	player.hudGlider.AddEvent(startOffset-500, startOffset, 1.0)

	if _, ok := player.overlay.(*overlays.ScoreOverlay); ok {
		player.cursorGlider.AddEvent(startOffset-750, startOffset-250, 1.0)
	} else {
		player.cursorGlider.AddEvent(beatmapStart-750, beatmapStart-250, 1.0)
	}

	player.dimGlider.AddEvent(beatmapStart, beatmapStart+1000, bmath.Normal)
	player.blurGlider.AddEvent(beatmapStart, beatmapStart+1000, bmath.Normal)
	player.fxGlider.AddEvent(beatmapStart, beatmapStart+1000, bmath.Normal)

	fadeOut := settings.Playfield.FadeOutTime * 1000

	if s, ok := player.overlay.(*overlays.ScoreOverlay); ok {
		if settings.Gameplay.ShowResultsScreen {
			beatmapEnd += 1000
			fadeOut = 250
		}

		s.SetBeatmapEnd(beatmapEnd + fadeOut)
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

	player.dimGlider.AddEventV(beatmapEnd, beatmapEnd+fadeOut, 0.0, bmath.Absolute)
	player.fxGlider.AddEventV(beatmapEnd, beatmapEnd+fadeOut, 0.0, bmath.Absolute)
	player.cursorGlider.AddEvent(beatmapEnd, beatmapEnd+fadeOut, 0.0)
	player.hudGlider.AddEvent(beatmapEnd, beatmapEnd+fadeOut, 0.0)

	player.mapEndL = beatmapEnd + fadeOut
	player.MapEnd = beatmapEnd + fadeOut

	if _, ok := player.overlay.(*overlays.ScoreOverlay); ok && settings.Gameplay.ShowResultsScreen {
		player.speedGlider.AddEvent(beatmapEnd+fadeOut, beatmapEnd+fadeOut, 0)
		player.pitchGlider.AddEvent(beatmapEnd+fadeOut, beatmapEnd+fadeOut, 0)

		player.MapEnd += (settings.Gameplay.ResultsScreenTime + 1) * 1000
		if player.MapEnd < player.musicPlayer.GetLength()*1000 {
			player.volumeGlider.AddEvent(player.MapEnd-settings.Gameplay.ResultsScreenTime*1000-500, player.MapEnd, 0.0)
		}
	} else {
		player.volumeGlider.AddEvent(beatmapEnd, beatmapEnd+fadeOut, 0.0)
	}

	player.MapEnd += 100

	// See https://github.com/Wieku/danser-go/issues/121
	player.musicPlayer.AddSilence(max(0, player.MapEnd/1000-player.musicPlayer.GetLength()))

	if settings.Playfield.SeizureWarning.Enabled {
		am := max(1000, settings.Playfield.SeizureWarning.Duration*1000)
		startOffset -= am
		player.epiGlider.AddEvent(startOffset, startOffset+500, 1.0)
		player.epiGlider.AddEvent(startOffset+am-500, startOffset+am, 0.0)
	}

	startOffset -= max(settings.Playfield.LeadInTime*1000, 1000)

	player.startOffset = startOffset
	player.progressMsF = startOffset
	player.rawPositionF = startOffset

	player.RunningTime = player.MapEnd - startOffset

	for _, p := range beatMap.Pauses {
		startTime := p.GetStartTime()
		endTime := p.GetEndTime()

		speed := settings.SPEED * player.bMap.Diff.GetSpeed()

		if endTime-startTime < 1000*speed || endTime < player.startPoint || startTime > player.MapEnd {
			continue
		}

		player.dimGlider.AddEvent(startTime, startTime+1000*speed, bmath.Break)
		player.blurGlider.AddEvent(startTime, startTime+1000*speed, bmath.Break)
		player.fxGlider.AddEvent(startTime, startTime+1000*speed, bmath.Break)

		if !settings.Cursor.ShowCursorsOnBreaks {
			player.cursorGlider.AddEvent(startTime, startTime+100*speed, 0.0)
		}

		player.dimGlider.AddEvent(endTime, endTime+1000*speed, bmath.Normal)
		player.blurGlider.AddEvent(endTime, endTime+1000*speed, bmath.Normal)
		player.fxGlider.AddEvent(endTime, endTime+1000*speed, bmath.Normal)
		player.cursorGlider.AddEvent(endTime, endTime+1000*speed, 1.0)
	}

	player.background.SetTrack(player.musicPlayer)

	player.coin = common.NewDanserCoin()
	player.coin.SetMap(beatMap, player.musicPlayer)

	player.coin.SetScale(0.25 * min(settings.Graphics.GetWidthF(), settings.Graphics.GetHeightF()))

	player.profiler = frame.NewCounter()

	player.bloomEffect = effects.NewBloomEffect(int(settings.Graphics.GetWidth()), int(settings.Graphics.GetHeight()))
	player.blur = effects.NewBlurEffect(int(settings.Graphics.GetWidth()), int(settings.Graphics.GetHeight()))

	player.background.Update(player.progressMsF, settings.Graphics.GetWidthF()/2, settings.Graphics.GetHeightF()/2)

	player.profilerU = frame.NewCounter()

	player.baseLimit = 1000

	player.updateLimiter = frame.NewLimiter(player.baseLimit)

	if player.bMap.Diff.CheckModActive(difficulty.Nightcore) {
		player.nightcore = common.NewNightcoreProcessor()
		player.nightcore.SetMap(player.bMap, player.musicPlayer)
	}

	if settings.Audio.OnlineOffset { // Try to load online offset
		onlineBeatmap, err2 := osuapi.LookupBeatmap(beatMap.MD5)
		if err2 != nil {
			log.Println("Failed to load online offset:", err.Error())
		} else if onlineBeatmap != nil {
			player.onlineOffset = onlineBeatmap.Beatmapset.Offset
			log.Println(fmt.Sprintf("Online offset loaded: %.0fms", player.onlineOffset))
		}
	}

	if settings.RECORD {
		return player
	}

	goroutines.RunOS(func() {
		var lastTimeNano = qpc.GetNanoTime()

		for !input.Win.ShouldClose() {
			currentTimeNano := qpc.GetNanoTime()

			delta := float64(currentTimeNano-lastTimeNano) / 1000000.0

			player.profilerU.PutSample(delta)

			musicState := player.musicPlayer.GetState()

			speed := 1.0

			if musicState == bass.MusicStopped {
				if player.rawPositionF < player.startPointE || player.start {
					player.rawPositionF += delta
				} else {
					speed = settings.SPEED * player.bMap.Diff.GetSpeed()
					player.rawPositionF += delta * speed
				}
			} else {
				musicPos := player.musicPlayer.GetPosition() * 1000
				speed = player.musicPlayer.GetSpeed()

				if musicPos != player.lastMusicPos || musicState == bass.MusicPaused {
					player.rawPositionF = musicPos
					player.lastMusicPos = musicPos
				} else if musicPos > 1 {
					// In DirectSound mode with VistaTruePos set to FALSE music is reported at 10ms intervals so we need to *interpolate* it
					// Wait at least 1ms because before interpolating because there's a 60ish ms delay before music in playing state starts reporting time and we don't want to jump back in time
					player.rawPositionF += delta * speed
				}
			}

			platformOffset := 0.0
			if runtime.GOOS == "windows" { // For some reason WASAPI reports time with 15ms delay, so we need to correct it
				platformOffset = windowsOffset
			}

			oldOffset := 0.0
			if player.bMap.Version < 5 {
				oldOffset = 24
			}

			player.progressMsF = player.rawPositionF + (platformOffset+float64(settings.Audio.Offset))*speed - oldOffset - float64(settings.LOCALOFFSET) - player.onlineOffset

			player.updateMain(delta)

			lastTimeNano = currentTimeNano

			player.updateLimiter.Sync()
		}

		player.musicPlayer.Stop()
		bass.StopLoops()
	})

	return player
}

func (player *Player) trySetupFail() {
	if sO, ok := player.overlay.(*overlays.ScoreOverlay); ok {
		var ruleset *osu.OsuRuleSet

		if rC, ok1 := player.controller.(*dance.ReplayController); ok1 {
			ruleset = rC.GetRuleset()
		} else if rP, ok2 := player.controller.(*dance.PlayerController); ok2 {
			ruleset = rP.GetRuleset()
		}

		if ruleset != nil {
			ruleset.SetFailListener(func(cursor *graphics.Cursor) {
				if !settings.RECORD {
					audio.PlayFailSound()
				}

				log.Println("Player failed!")

				sO.Fail(true)

				player.frequencyGlider.AddEvent(player.realTime, player.realTime+2400, 0.0)
				player.objectsAlphaFail.AddEvent(player.realTime, player.realTime+2400, 0.0)

				player.failOX.AddEvent(player.realTime, player.realTime+2400, camera2.OsuWidth*(rand.Float64()-0.5)/2)
				player.failOY.AddEvent(player.realTime, player.realTime+2400, -camera2.OsuHeight*(1+rand.Float64()*0.2))

				rotBase := rand.Float64()

				player.failRotation.AddEvent(player.realTime, player.realTime+2400, math.Copysign((math.Abs(rotBase)*0.5+0.5)/6*math.Pi, rotBase))

				player.failing = true
				player.failAt = player.realTime + 2400

				player.dimGlider.Reset()
				player.blurGlider.Reset()
				player.hudGlider.Reset()
				player.fxGlider.Reset()
				player.cursorGlider.Reset()
				player.objectsAlpha.Reset()
			})
		}

		if player.controller.GetCursors()[0].IsPlayer && !player.controller.GetCursors()[0].IsAutoplay {
			player.cursorGlider.SetValue(1.0)
		}
	}
}

func (player *Player) Update(delta float64) bool {
	speed := 1.0

	if player.musicPlayer.GetState() == bass.MusicPlaying {
		speed = player.musicPlayer.GetSpeed()
	} else if !(player.progressMsF < player.startPointE || player.start) {
		speed = settings.SPEED * player.bMap.Diff.GetSpeed()
	}

	player.rawPositionF += delta * speed

	oldOffset := 0.0
	if player.bMap.Version < 5 {
		oldOffset = 24
	}

	player.progressMsF = player.rawPositionF - oldOffset - float64(settings.LOCALOFFSET) - player.onlineOffset

	player.updateMain(delta)

	if player.progressMsF >= player.MapEnd {
		player.musicPlayer.Stop()
		bass.StopLoops()

		return true
	}

	return false
}

func (player *Player) GetTime() float64 {
	return player.progressMsF
}

func (player *Player) GetTimeOffset() float64 {
	return player.progressMsF - player.startOffset
}

func (player *Player) updateMain(delta float64) {
	player.realTime += delta

	if player.rawPositionF >= player.startPoint && !player.start {
		player.musicPlayer.Play()

		if player.overlay != nil {
			player.overlay.SetMusic(player.musicPlayer)
		}

		player.musicPlayer.SetPosition(player.startPoint / 1000)

		discord.SetDuration(int64((player.mapEndL-player.musicPlayer.GetPosition()*1000)/(settings.SPEED*player.bMap.Diff.GetSpeed()) + (player.MapEnd - player.mapEndL)))

		if player.overlay == nil {
			discord.UpdateDance(settings.TAG, settings.DIVIDES)
		}

		player.start = true
	}

	player.speedGlider.Update(player.progressMsF)
	player.pitchGlider.Update(player.progressMsF)
	player.frequencyGlider.Update(player.realTime)

	player.objectsAlphaFail.Update(player.realTime)
	player.failOX.Update(player.realTime)
	player.failOY.Update(player.realTime)
	player.failRotation.Update(player.realTime)

	player.objectCamera.SetOrigin(vector.NewVec2d(player.failOX.GetValue(), player.failOY.GetValue()))
	player.objectCamera.SetRotation(player.failRotation.GetValue())
	player.objectCamera.Update()

	if player.failing && player.realTime >= player.failAt {
		if !player.failed {
			player.musicPlayer.Pause()
			player.MapEnd = player.progressMsF
		}

		player.failed = true
	}

	speedAdjust := mutils.Lerp(1, settings.SPEED, player.speedGlider.GetValue())
	freqAdjust := 1.0

	speedVal := mutils.Lerp(1, player.bMap.Diff.GetSpeed(), player.speedGlider.GetValue())
	if player.bMap.Diff.AdjustsPitch() {
		freqAdjust = speedVal
	} else {
		speedAdjust *= speedVal
	}

	player.musicPlayer.SetTempo(speedAdjust)
	player.musicPlayer.SetPitch(mutils.Lerp(1, settings.PITCH, player.pitchGlider.GetValue()))
	player.musicPlayer.SetRelativeFrequency(freqAdjust * player.frequencyGlider.GetValue())

	if player.progressMsF >= player.startPointE {
		if _, ok := player.controller.(*dance.GenericController); ok {
			player.bMap.Update(player.progressMsF)
		}

		player.objectContainer.Update(player.progressMsF)
	}

	if player.progressMsF >= player.startPointE || settings.PLAY {
		if player.progressMsF < player.mapEndL {
			player.controller.Update(player.progressMsF, delta)

			if player.nightcore != nil {
				player.nightcore.Update(player.progressMsF)
			}
		} else if settings.Gameplay.ShowResultsScreen {
			if player.overlay != nil {
				player.overlay.DisableAudioSubmission(true)
			}
			player.controller.Update(player.bMap.HitObjects[len(player.bMap.HitObjects)-1].GetEndTime()+float64(player.bMap.Diff.Hit50)+100, delta)
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
		offset = offset.Add(player.mainCamera.Project(c.Position.Copy64()).Mult(vector.NewVec2d(2/settings.Graphics.GetWidthF(), -2/settings.Graphics.GetHeightF())))
	}

	offset = offset.Scl(1 / float64(len(player.controller.GetCursors())))

	player.background.Update(player.progressMsF, offset.X*player.cursorGlider.GetValue(), offset.Y*player.cursorGlider.GetValue())

	bgDim := settings.Playfield.Background.Dim
	blurDim := settings.Playfield.Background.Blur.Values
	fxDim := settings.Playfield.Logo.Dim

	player.dimGlider.Update(player.progressMsF, 1-bgDim.Intro, 1-bgDim.Normal, 1-bgDim.Breaks)
	player.blurGlider.Update(player.progressMsF, blurDim.Intro, blurDim.Normal, blurDim.Breaks)
	player.fxGlider.Update(player.progressMsF, 1-fxDim.Intro, 1-fxDim.Normal, 1-fxDim.Breaks)

	player.epiGlider.Update(player.progressMsF)
	player.cursorGlider.Update(player.progressMsF)
	player.hudGlider.Update(player.progressMsF)
	player.volumeGlider.Update(player.progressMsF)
	player.objectsAlpha.Update(player.progressMsF)

	if player.musicPlayer.GetState() == bass.MusicPlaying {
		player.musicPlayer.SetVolumeRelative(player.volumeGlider.GetValue())
	}
}

func (player *Player) updateMusic(delta float64) {
	player.musicPlayer.Update()

	target := mutils.Clamp(player.musicPlayer.GetBoost()*(settings.Audio.BeatScale-1.0)+1.0, 1.0, settings.Audio.BeatScale)

	if settings.Audio.BeatUseTimingPoints {
		player.Scl = 1 + player.coin.Beat*(settings.Audio.BeatScale-1.0)
	} else if player.Scl < target {
		player.Scl += (target - player.Scl) * 0.3 * delta / 16.66667
	} else if player.Scl > target {
		player.Scl -= (player.Scl - target) * 0.15 * delta / 16.66667
	}
}

func (player *Player) DrawMain(float64) {
	profiler.StartGroup("Player.DrawMain", profiler.PDraw)

	if player.lastTime <= 0 {
		player.lastTime = qpc.GetNanoTime()
	}

	tim := qpc.GetNanoTime()
	timMs := float64(tim-player.lastTime) / 1000000.0

	fps := player.profiler.GetFPS()

	player.updateLimiter.FPS = mutils.Clamp(int(fps*1.2), player.baseLimit, 10000)

	if player.background.GetStoryboard() != nil {
		player.background.GetStoryboard().SetFPS(mutils.Clamp(int(fps*1.2), player.baseLimit, 10000))
	}

	if fps > 58 && timMs > 18 && !settings.RECORD {
		log.Println(fmt.Sprintf("Slow frame detected! Frame time: %.3fms | Av. frame time: %.3fms", timMs, 1000.0/fps))
	}

	player.progressMs = int64(player.progressMsF)

	player.profiler.PutSample(timMs)
	player.lastTime = tim

	objectCameras := player.objectCamera.GenRotated(settings.DIVIDES, -2*math.Pi/float64(settings.DIVIDES))
	cursorCameras := player.mainCamera.GenRotated(settings.DIVIDES, -2*math.Pi/float64(settings.DIVIDES))

	bgAlpha := player.dimGlider.GetValue()
	if settings.Playfield.Background.FlashToTheBeat {
		bgAlpha = mutils.Clamp(bgAlpha*player.Scl, 0, 1)
	}

	player.background.Draw(player.progressMsF, player.batch, player.blurGlider.GetValue(), bgAlpha, player.bgCamera.GetProjectionView())

	if player.progressMsF > 0 {
		timeDiff := player.progressMsF - player.lastProgressMsF
		settings.Cursor.Colors.Update(timeDiff)
		player.lastProgressMsF = player.progressMsF
	}

	cursorColors := settings.Cursor.GetColors(settings.DIVIDES, len(player.controller.GetCursors()), player.Scl, player.cursorGlider.GetValue())

	if player.overlay != nil {
		player.drawOverlayPart(player.overlay.DrawBackground, cursorColors, cursorCameras[0], 1)
	}

	player.drawEpilepsyWarning()

	player.counter += timMs

	if player.counter >= 1000.0/60 {
		player.counter -= 1000.0 / 60
		if player.background.GetStoryboard() != nil {
			player.storyboardDrawn = player.background.GetStoryboard().GetRenderedSprites()
		}
	}

	player.drawCoin()

	scale2 := player.Scl
	if !settings.Cursor.ScaleToTheBeat {
		scale2 = 1
	}

	bloomEnabled := settings.Playfield.Bloom.Enabled

	if bloomEnabled {
		player.bloomEffect.SetThreshold(settings.Playfield.Bloom.Threshold)
		player.bloomEffect.SetBlur(settings.Playfield.Bloom.Blur)

		bPower := settings.Playfield.Bloom.Power
		if settings.Playfield.Bloom.BloomToTheBeat {
			bPower += settings.Playfield.Bloom.BloomBeatAddition * (player.Scl - 1.0) / (settings.Audio.BeatScale * 0.4)
		}

		player.bloomEffect.SetPower(bPower)
		player.bloomEffect.Begin()
	}

	if player.overlay != nil {
		player.drawOverlayPart(player.overlay.DrawBeforeObjects, cursorColors, objectCameras[0], player.objectsAlphaFail.GetValue())
	}

	player.objectContainer.Draw(player.batch, player.mainCamera.GetProjectionView(), objectCameras, player.progressMsF, float32(player.Scl), float32(player.objectsAlpha.GetValue()*player.objectsAlphaFail.GetValue()))

	if player.overlay != nil {
		player.drawOverlayPart(player.overlay.DrawNormal, cursorColors, objectCameras[0], 1)
	}

	player.background.DrawOverlay(player.progressMsF, player.batch, bgAlpha, player.bgCamera.GetProjectionView())

	if player.overlay != nil && player.overlay.ShouldDrawHUDBeforeCursor() {
		player.drawOverlayPart(player.overlay.DrawHUD, cursorColors, player.uiCamera.GetProjectionView(), 1)
	}

	if settings.Playfield.DrawCursors {
		for _, g := range player.controller.GetCursors() {
			g.UpdateRenderer()
		}

		player.batch.SetAdditive(false)

		graphics.BeginCursorRender()

		for j := 0; j < settings.DIVIDES; j++ {
			player.batch.SetCamera(cursorCameras[j])

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

	if player.overlay != nil && !player.overlay.ShouldDrawHUDBeforeCursor() {
		player.drawOverlayPart(player.overlay.DrawHUD, cursorColors, player.uiCamera.GetProjectionView(), 1)
	}

	if bloomEnabled {
		player.bloomEffect.EndAndRender()
	}

	profiler.EndGroup()
}

func (player *Player) Draw(d float64) {
	profiler.StartGroup("Player.Draw", profiler.PDraw)

	player.DrawMain(d)
	player.drawDebug()

	profiler.EndGroup()
}

func (player *Player) drawEpilepsyWarning() {
	if player.epiGlider.GetValue() < 0.01 {
		return
	}

	player.batch.Begin()
	player.batch.ResetTransform()
	player.batch.SetColor(1, 1, 1, player.epiGlider.GetValue())
	player.batch.SetCamera(mgl32.Ortho(float32(-settings.Graphics.GetWidthF()/2), float32(settings.Graphics.GetWidthF()/2), float32(settings.Graphics.GetHeightF()/2), float32(-settings.Graphics.GetHeightF()/2), 1, -1))

	scl := scaling.Fit.Apply(player.Epi.Width, player.Epi.Height, float32(settings.Graphics.GetWidthF()), float32(settings.Graphics.GetHeightF()))
	scl = scl.Scl(0.5).Scl(0.66)
	player.batch.SetScale(scl.X64(), scl.Y64())
	player.batch.DrawUnit(*player.Epi)

	player.batch.ResetTransform()
	player.batch.End()
	player.batch.SetColor(1, 1, 1, 1)
}

func (player *Player) drawCoin() {
	if !settings.Playfield.Logo.Enabled || player.fxGlider.GetValue() < 0.01 {
		return
	}

	player.batch.Begin()
	player.batch.ResetTransform()
	player.batch.SetColor(1, 1, 1, player.fxGlider.GetValue())
	player.batch.SetCamera(mgl32.Ortho(float32(-settings.Graphics.GetWidthF()/2), float32(settings.Graphics.GetWidthF()/2), float32(settings.Graphics.GetHeightF()/2), float32(-settings.Graphics.GetHeightF()/2), 1, -1))

	player.coin.DrawVisualiser(settings.Playfield.Logo.DrawSpectrum)
	player.coin.Draw(player.progressMsF, player.batch)

	player.batch.ResetTransform()
	player.batch.End()
}

func (player *Player) drawOverlayPart(drawFunc func(*batch2.QuadBatch, []color2.Color, float64), cursorColors []color2.Color, camera mgl32.Mat4, alpha float64) {
	player.batch.Begin()
	player.batch.ResetTransform()
	player.batch.SetColor(1, 1, 1, 1)

	player.batch.SetCamera(camera)

	drawFunc(player.batch, cursorColors, player.hudGlider.GetValue()*alpha)

	player.batch.End()
	player.batch.ResetTransform()
	player.batch.SetColor(1, 1, 1, 1)
}

func (player *Player) drawDebug() {
	profiler.StartGroup("Player.DrawDebug", profiler.PDraw)
	if settings.DEBUG && player.memTicker == nil {
		player.memTicker = time.NewTicker(100 * time.Millisecond)

		goroutines.RunOS(func() {
			prevT := time.Now()
			runtime.ReadMemStats(player.mStats2)

			for t := range player.memTicker.C {
				diff := t.Sub(prevT)
				prevT = t

				player.mStats1, player.mStats2 = player.mStats2, player.mStats1
				runtime.ReadMemStats(player.mStats2)

				mDelta := float64(int64(player.mStats2.Alloc)-int64(player.mStats1.Alloc)) * (1000 / float64(diff.Milliseconds()))
				if mDelta > 0 {
					player.mProfiler.PutSample(mDelta)
				}

				if !settings.DEBUG && player.memTicker != nil {
					player.memTicker.Stop()
					player.memTicker = nil
				}
			}
		})
	}

	var profRes *profiler.PNode

	if settings.DEBUG || settings.PerfGraph {
		profRes = profiler.GetLastProfileResult()
	}

	if settings.DEBUG || settings.Graphics.ShowFPS || settings.PerfGraph {
		padDown := 4.0
		size := 16.0

		drawShadowed := func(right bool, pos float64, format string, a ...any) {
			pX := 0.0
			origin := vector.BottomLeft

			if right {
				pX = player.ScaledWidth
				origin = vector.BottomRight
			}

			pY := player.ScaledHeight - (size+padDown)*pos - padDown

			player.mBuffer = player.mBuffer[:0]
			player.mBuffer = fmt.Appendf(player.mBuffer, format, a...)

			player.batch.SetColor(0, 0, 0, 1)
			player.font.DrawOrigin(player.batch, pX+size*0.1, pY+size*0.1, origin, size, true, string(player.mBuffer))

			player.batch.SetColor(1, 1, 1, 1)
			player.font.DrawOrigin(player.batch, pX, pY, origin, size, true, string(player.mBuffer))
		}

		player.batch.Begin()
		player.batch.ResetTransform()
		player.batch.SetColor(1, 1, 1, 1)
		player.batch.SetCamera(player.uiCamera.GetProjectionView())

		if settings.DEBUG {
			player.batch.SetColor(0, 0, 0, 1)
			player.font.DrawOrigin(player.batch, size*1.5*0.1, padDown+size*1.5*0.1, vector.TopLeft, size*1.5, false, player.mapFullName)

			player.batch.SetColor(1, 1, 1, 1)
			player.font.DrawOrigin(player.batch, 0, padDown, vector.TopLeft, size*1.5, false, player.mapFullName)

			pos := 3.0

			player.font.DrawBg(true)
			player.font.SetBgBorderSize(padDown / 2)
			player.font.SetBgColor(color2.NewLA(0, 0.8))

			drawWithBackground := func(format string, a ...any) {
				player.mBuffer = player.mBuffer[:0]
				player.mBuffer = fmt.Appendf(player.mBuffer, format, a...)

				player.font.DrawOrigin(player.batch, 0, (size+padDown)*pos, vector.CentreLeft, size, true, string(player.mBuffer))
				pos++
			}

			drawWithBackground("VSync: %t", settings.Graphics.VSync)
			drawWithBackground("Blur: %t", settings.Playfield.Background.Blur.Enabled)
			drawWithBackground("Bloom: %t", settings.Playfield.Bloom.Enabled)

			msaa := "OFF"
			if settings.Graphics.MSAA > 0 {
				msaa = strconv.Itoa(int(settings.Graphics.MSAA)) + "x"
			}

			drawWithBackground("MSAA: %s", msaa)

			drawWithBackground("FBO Binds: %d", profiler.GetPreviousStat(profiler.FBOBinds))
			drawWithBackground("VAO Binds: %d", profiler.GetPreviousStat(profiler.VAOBinds))
			drawWithBackground("VBO Binds: %d", profiler.GetPreviousStat(profiler.VBOBinds))
			drawWithBackground("Vertex Upload: %.2fk", float64(profiler.GetPreviousStat(profiler.VertexUpload))/1000)
			drawWithBackground("Vertices Drawn: %.2fk", float64(profiler.GetPreviousStat(profiler.VerticesDrawn))/1000)
			drawWithBackground("Draw Calls: %d", profiler.GetPreviousStat(profiler.DrawCalls))
			drawWithBackground("Sprites Drawn: %d", profiler.GetPreviousStat(profiler.SpritesDrawn))

			if storyboard := player.background.GetStoryboard(); storyboard != nil {
				drawWithBackground("SB sprites: %d", player.storyboardDrawn)
			}

			pos++
			drawWithBackground("Memory:")

			drawWithBackground("Allocated: %s", humanize.Bytes(player.mStats2.Alloc))
			drawWithBackground("Allocs/s: %s", humanize.Bytes(uint64(player.mProfiler.GetAverage())))
			drawWithBackground("System: %s", humanize.Bytes(player.mStats2.Sys))
			drawWithBackground("GC Runs: %d", player.mStats2.NumGC)
			drawWithBackground("GC Time: %.3fms", float64(player.mStats2.PauseTotalNs)/1000000)

			if profRes != nil && settings.CallGraph {
				pos++
				player.traverseGraph(0, profRes, drawWithBackground)
			}

			player.font.DrawBg(false)

			player.batch.ResetTransform()

			currentTime := int(player.musicPlayer.GetPosition())
			totalTime := int(player.musicPlayer.GetLength())
			mapTime := int(player.bMap.HitObjects[len(player.bMap.HitObjects)-1].GetEndTime() / 1000)

			drawShadowed(false, 2, "%02d:%02d / %02d:%02d (%02d:%02d)", currentTime/60, currentTime%60, totalTime/60, totalTime%60, mapTime/60, mapTime%60)
			drawShadowed(false, 1, "%d(*%d) hitobjects, %d total", player.objectContainer.GetNumProcessed(), settings.DIVIDES, len(player.bMap.HitObjects))

			if storyboard := player.background.GetStoryboard(); storyboard != nil {
				drawShadowed(false, 0, "%d storyboard sprites, %d in queue (%d total)", player.background.GetStoryboard().GetProcessedSprites(), storyboard.GetQueueSprites(), storyboard.GetTotalSprites())
			} else {
				drawShadowed(false, 0, "No storyboard")
			}
		}

		if settings.DEBUG || settings.Graphics.ShowFPS || settings.PerfGraph {
			if settings.PerfGraph {
				if profRes != nil && input.Win.GetKey(glfw.KeyLeftShift) != glfw.Press {
					root := profRes.TimeTotal
					sched := profRes.Nodes[0].TimeTotal
					input := profRes.Nodes[1].TimeTotal
					draw := profRes.Nodes[2].TimeTotal
					swp := profRes.Nodes[3].TimeTotal
					slp := profRes.Nodes[4].TimeTotal
					root += -input - draw - swp - slp - sched

					player.ftGraph.Advance(sched, input, draw, swp, slp, root)
				}

				player.batch.Flush()

				player.ftGraph.SetCamera(player.uiCamera.GetProjectionView())
				player.ftGraph.SetMaxValue(settings.Debug.PerfGraph.MaxRange)
				player.ftGraph.Draw()
			}

			fpsC := player.profiler.GetFPS()
			fpsU := player.profilerU.GetFPS()

			sbThread := player.background.GetStoryboard() != nil && player.background.GetStoryboard().HasVisuals()

			drawFPS := fmt.Sprintf("%0.0ffps (%0.2fms)", fpsC, 1000/fpsC)
			updateFPS := fmt.Sprintf("%0.0ffps (%0.2fms)", fpsU, 1000/fpsU)
			sbFPS := ""

			off := 0.0

			if sbThread {
				off = 1.0

				fpsS := player.background.GetStoryboard().GetFPS()
				sbFPS = fmt.Sprintf("%0.0ffps (%0.2fms)", fpsS, 1000/fpsS)
			}

			shift := strconv.Itoa(max(len(drawFPS), max(len(updateFPS), len(sbFPS))))

			drawShadowed(true, 1+off, "Draw: %"+shift+"s", drawFPS)
			drawShadowed(true, 0+off, "Update: %"+shift+"s", updateFPS)

			if sbThread {
				drawShadowed(true, 0, "Storyboard: %"+shift+"s", sbFPS)
			}
		}

		player.batch.End()
	}

	profiler.EndGroup()
}

func (player *Player) traverseGraph(ident int, node *profiler.PNode, draw func(format string, a ...any)) {
	draw(strings.Repeat("  ", ident)+"%s: %.5fms", node.GetFullName(), node.TimeTotal)
	for _, sNode := range node.Nodes {
		player.traverseGraph(ident+1, sNode, draw)
	}
}

func (player *Player) Show() {}

func (player *Player) Hide() {}

func (player *Player) Dispose() {}
