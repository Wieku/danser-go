package states

//region 无关0

import (
	"danser/animation"
	"danser/audio"
	"danser/beatmap"
	"danser/beatmap/objects"
	"danser/bmath"
	"danser/dance"
	"danser/hitjudge"
	"danser/render"
	"danser/render/effects"
	"danser/render/font"
	"danser/render/texture"
	"danser/replay"
	"danser/resultcache"
	"danser/score"
	"danser/settings"
	"danser/storyboard"
	"danser/utils"
	"fmt"
	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/wieku/glhf"
	"log"
	"math"
	"path/filepath"
	"time"
)

type Player struct {
	font           *font.Font
	bMap           *beatmap.BeatMap
	queue2         []objects.BaseObject
	processed      []objects.Renderable
	sliderRenderer *render.SliderRenderer
	blurEffect     *effects.BlurEffect
	bloomEffect    *effects.BloomEffect
	lastTime       int64
	progressMsF    float64
	progressMs     int64
	batch          *render.SpriteBatch
	controller     []dance.Controller
	//circles        []*objects.Circle
	//sliders        []*objects.Slider
	Background  *texture.TextureRegion
	Logo        *texture.TextureRegion
	BgScl       bmath.Vector2d
	Scl         float64
	SclA        float64
	CS          float64
	fxRotation  float64
	fadeOut     float64
	fadeIn      float64
	entry       float64
	start       bool
	mus         bool
	musicPlayer *audio.Music
	fxBatch     *render.FxBatch
	vao         *glhf.VertexSlice
	vaoD        []float32
	vaoDirty    bool
	rotation    float64
	profiler    *utils.FPSCounter
	profilerU   *utils.FPSCounter

	storyboard *storyboard.Storyboard

	camera         *bmath.Camera
	scamera        *bmath.Camera
	dimGlider      *animation.Glider
	blurGlider     *animation.Glider
	fxGlider       *animation.Glider
	cursorGlider   *animation.Glider
	counter        float64
	fpsC           float64
	fpsU           float64
	storyboardLoad float64
	mapFullName    string

	// 偏移位置参数
	fontsize		float64
	missfontsize	float64
	misssize		float64
	keysize 		float64
	modoffset		float64
	missoffsetX		float64
	missoffsetY		float64
	lineoffset		float64
	hitoffset		float64
	key1baseX		float64
	key2baseX		float64
	key3baseX		float64
	accbaseX		float64
	rankbaseX		float64
	ppbaseX			float64
	playerbaseX		float64
	keybaseY		float64
	fontbaseY		float64
	rankbaseY		float64
	hitbaseY		float64
}

//endregion

func NewPlayer(beatMap *beatmap.BeatMap) *Player {
	//region 无关1
	player := new(Player)
	render.LoadTextures()
	render.SetupSlider()
	player.batch = render.NewSpriteBatch()
	player.sliderRenderer = render.NewSliderRenderer()
	player.font = font.GetFont("Roboto Bold")

	player.bMap = beatMap

	player.mapFullName = fmt.Sprintf("%s - %s [%s]", beatMap.Artist, beatMap.Name, beatMap.Difficulty)
	log.Println("Playing:", player.mapFullName)

	player.CS = (1.0 - 0.7*(beatMap.CircleSize-5)/5) / 2 * settings.Objects.CSMult
	render.CS = player.CS

	var err error
	player.Background, err = utils.LoadTextureToAtlas(render.Atlas, filepath.Join(settings.General.OsuSongsDir, beatMap.Dir, beatMap.Bg))
	if err != nil {
		log.Println(err)
	}

	//if settings.Playfield.StoryboardEnabled {
	//	player.storyboard = storyboard.NewStoryboard(player.bMap)
	//
	//	if player.storyboard == nil {
	//		log.Println("Storyboard not found!")
	//	}
	//}

	//player.Logo, err = utils.LoadTextureToAtlas(render.Atlas, "assets/textures/logo-medium.png")

	if err != nil {
		log.Println(err)
	}

	winscl := settings.Graphics.GetAspectRatio()

	player.blurEffect = effects.NewBlurEffect(int(settings.Graphics.GetWidth()), int(settings.Graphics.GetHeight()))

	if player.Background != nil {
		imScl := float64(player.Background.Width) / float64(player.Background.Height)

		condition := imScl < winscl
		if player.storyboard != nil && !player.storyboard.IsWideScreen() {
			condition = !condition
		}

		if condition {
			player.BgScl = bmath.NewVec2d(1, winscl/imScl)
		} else {
			player.BgScl = bmath.NewVec2d(imScl/winscl, 1)
		}
	}

	scl := (settings.Graphics.GetHeightF() * 900.0 / 1080.0) / float64(384) * settings.Playfield.Scale

	osuAspect := 512.0 / 384.0
	screenAspect := settings.Graphics.GetWidthF() / settings.Graphics.GetHeightF()

	if osuAspect > screenAspect {
		scl = (settings.Graphics.GetWidthF() * 900.0 / 1080.0) / float64(512) * settings.Playfield.Scale
	}

	player.camera = bmath.NewCamera()
	player.camera.SetViewport(int(settings.Graphics.GetWidth()), int(settings.Graphics.GetHeight()), true)
	player.camera.SetOrigin(bmath.NewVec2d(512.0/2, 384.0/2))
	player.camera.SetScale(bmath.NewVec2d(scl, scl))
	player.camera.Update()

	player.scamera = bmath.NewCamera()
	player.scamera.SetViewport(int(settings.Graphics.GetWidth()), int(settings.Graphics.GetHeight()), false)
	player.scamera.SetOrigin(bmath.NewVec2d(settings.Graphics.GetWidthF()/2, settings.Graphics.GetHeightF()/2))
	player.scamera.Update()

	render.Camera = player.camera

	player.bMap.Reset()

	// 控制初始化
	player.controller = make([]dance.Controller, settings.General.Players)
	for k := 0; k < settings.General.Players; k++ {
		player.controller[k] = dance.NewReplayController()
		player.controller[k].SetBeatMap(player.bMap)
		player.controller[k].InitCursors()
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

	tmS := float64(player.queue2[0].GetBasicData().StartTime)
	tmE := float64(player.queue2[len(player.queue2)-1].GetBasicData().EndTime)

	player.dimGlider.AddEvent(-1500, -1000, 1.0-settings.Playfield.BackgroundInDim)
	player.blurGlider.AddEvent(-1500, -1000, settings.Playfield.BackgroundInBlur)
	player.fxGlider.AddEvent(-1500, -1000, 1.0-settings.Playfield.SpectrumInDim)
	player.cursorGlider.AddEvent(-1500, -1000, 0.0)

	player.dimGlider.AddEvent(tmS-750, tmS-250, 1.0-settings.Playfield.BackgroundDim)
	player.blurGlider.AddEvent(tmS-750, tmS-250, settings.Playfield.BackgroundBlur)
	player.fxGlider.AddEvent(tmS-750, tmS-250, 1.0-settings.Playfield.SpectrumDim)
	player.cursorGlider.AddEvent(tmS-750, tmS-250, 1.0)

	fadeOut := settings.Playfield.FadeOutTime * 1000
	player.dimGlider.AddEvent(tmE, tmE+fadeOut, 0.0)
	player.fxGlider.AddEvent(tmE, tmE+fadeOut, 0.0)
	player.cursorGlider.AddEvent(tmE, tmE+fadeOut, 0.0)

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

	//endregion

	//region 计算大小偏移位置常量

	player.fontsize = 1.75 * settings.General.BaseSize
	player.missfontsize = settings.General.MissMult * player.fontsize
	player.misssize = 1.5 * settings.General.MissMult * settings.General.BaseSize
	player.keysize = 1.25 * settings.General.BaseSize
	player.modoffset =  1.25 * settings.General.BaseSize
	player.missoffsetX =  2 * settings.General.MissMult * settings.General.BaseSize
	player.missoffsetY =  0.6 * settings.General.MissMult * settings.General.BaseSize
	player.lineoffset = 2.25 * settings.General.BaseSize
	player.hitoffset = 1.75 * settings.General.BaseSize
	player.key1baseX = settings.General.BaseX
	player.key2baseX = player.key1baseX + 2 * player.keysize
	player.key3baseX = player.key2baseX + 2 * player.keysize
	player.accbaseX = player.key3baseX + 2 * settings.General.BaseSize
	player.rankbaseX = player.accbaseX + 8.375 * settings.General.BaseSize
	player.ppbaseX = player.rankbaseX + 1.625 * settings.General.BaseSize
	player.playerbaseX = player.ppbaseX + 8.75 * settings.General.BaseSize
	player.keybaseY = settings.General.BaseY
	player.fontbaseY = settings.General.BaseY - 0.75 * settings.General.BaseSize
	player.rankbaseY = settings.General.BaseY - 0.25 * settings.General.BaseSize
	player.hitbaseY = settings.General.BaseY - 0.25 * settings.General.BaseSize

	//endregion

	//region replay处理

	// 读取replay
	replays, err := replay.GetOsrFiles()
	if err != nil {
		panic(err)
	}
	// 解析每个replay的判定
	t := time.Now()
	if settings.General.ReadResultCache {
		log.Println("本次选择读取缓存replay结果")
		for k := 0; k < settings.General.Players; k++ {
			t1 := time.Now()
			log.Println("读取第", k+1, "个replay缓存")
			result, totalresult := resultcache.ReadResult(k+1)
			player.controller[k].SetHitResult(result)
			player.controller[k].SetTotalResult(totalresult)
			// 设置计算数组、初始化acc、rank和pp
			player.controller[k].SetAcc(100.0)
			player.controller[k].SetRank(*render.RankX)
			player.controller[k].SetPP(0.0)
			// 设置初始显示
			player.controller[k].SetIsShow(true)
			log.Println("读取第", k+1, "个replay缓存完成，耗时", time.Now().Sub(t1), "，总耗时", time.Now().Sub(t))
		}
	}else {
		log.Println("本次选择读取解析replay")
		for k := 0; k < settings.General.Players; k++ {
			t1 := time.Now()
			log.Println("解析第", k+1, "个replay")
			result, totalresult := hitjudge.ParseHits(settings.General.OsuSongsDir+beatMap.Dir+"/"+beatMap.File, replays[k])
			player.controller[k].SetHitResult(result)
			player.controller[k].SetTotalResult(totalresult)
			// 设置计算数组、初始化acc、rank和pp
			player.controller[k].SetAcc(100.0)
			player.controller[k].SetRank(*render.RankX)
			player.controller[k].SetPP(0.0)
			// 设置初始显示
			player.controller[k].SetIsShow(true)
			// 保存结果缓存
			if settings.General.SaveResultCache {
				resultcache.SaveResult(result, totalresult, k+1)
				log.Println("已保存第", k+1, "个replay的结果缓存")
			}
			log.Println("解析第", k+1, "个replay完成，耗时", time.Now().Sub(t1), "，总耗时", time.Now().Sub(t))
		}
	}

	//endregion

	//region 音乐？

	go func() {
		player.entry = 1
		time.Sleep(time.Duration(settings.Playfield.LeadInTime * float64(time.Second)))

		start := -2000.0
		for i := 1; i <= 100; i++ {
			player.entry = float64(i) / 100
			start += 10
			player.dimGlider.Update(start)
			player.blurGlider.Update(start)
			player.fxGlider.Update(start)
			player.cursorGlider.Update(start)
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
			time.Sleep(10 * time.Millisecond)
		}

		player.start = true
		musicPlayer.Play()
		musicPlayer.SetTempo(settings.SPEED)
		musicPlayer.SetPitch(settings.PITCH)
	}()

	player.fxBatch = render.NewFxBatch()
	player.vao = player.fxBatch.CreateVao(2 * 3 * (256 + 128))
	player.profilerU = utils.NewFPSCounter(60, false)

	//endregion

	//region 重写更新时间和坐标函数

	for k := 0; k < settings.General.Players; k++ {
		go func(k int) {
			// 获取replay信息
			r := replay.ExtractReplay(replays[k])
			index := 3

			// 设置player名
			player.controller[k].SetPlayername(r.Username)

			// 判断mod
			mods := r.Mods
			player.controller[k].SetIsHD(mods&8 > 0)
			player.controller[k].SetIsHR(mods&16 > 0)
			player.controller[k].SetIsEZ(mods&2 > 0)

			// 开始时间
			r1 := *r.ReplayData[1]
			r2 := *r.ReplayData[2]
			start := r1.Time + r2.Time

			var last= musicPlayer.GetPosition()
			for {
				// 获取第index个replay数据
				rdata := *r.ReplayData[index]
				offset := rdata.Time
				posX := rdata.MosueX
				posY := rdata.MouseY
				PressKey := *rdata.KeyPressed

				// 如果offset=-12345，结束
				if offset == -12345 {
					time.Sleep(1000 * time.Second)
				}

				if index == 3 {
					offset += start
				}

				progressMsF := musicPlayer.GetPosition()*1000 + float64(settings.Audio.Offset)

				//真实的offset
				true_offset := progressMsF - last

				// 如果真实offset大于等于读到的offset，更新
				if true_offset >= float64(offset) {
					// 如果是HR，上下翻转
					if player.controller[k].GetIsHR(){
						player.controller[k].Update(int64(progressMsF), true_offset, bmath.NewVec2d(float64(posX), float64(384 - posY)))
					}else {
						player.controller[k].Update(int64(progressMsF), true_offset, bmath.NewVec2d(float64(posX), float64(posY)))
					}

					player.controller[k].SetPresskey(PressKey)

					// 修正last
					last += float64(offset)

					index++
				}

				time.Sleep(time.Millisecond)
			}
		}(k)
	}

	//endregion

	//region 独立绘图

	go func() {
		for {
			player.progressMsF = musicPlayer.GetPosition()*1000 + float64(settings.Audio.Offset)
			player.bMap.Update(int64(player.progressMsF))
			if player.start && len(player.bMap.Queue) > 0 {
				player.dimGlider.Update(player.progressMsF)
				player.blurGlider.Update(player.progressMsF)
				player.fxGlider.Update(player.progressMsF)
				player.cursorGlider.Update(player.progressMsF)
			}
			time.Sleep(time.Millisecond)
		}
	}()

	//endregion

	//region 无关2

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

	//endregion
}

func (pl *Player) Show() {

}

func (pl *Player) Draw(delta float64) {

	//region 无关3

	if pl.lastTime < 0 {
		pl.lastTime = utils.GetNanoTime()
	}
	tim := utils.GetNanoTime()
	timMs := float64(tim-pl.lastTime) / 1000000.0

	pl.profiler.PutSample(1000.0 / timMs)
	//fps := pl.profiler.GetFPS()

	if pl.start {

		//if fps > 58 && timMs > 17 {
		//	log.Println("Slow frame detected! Frame time:", timMs, "| Av. frame time:", 1000.0/fps)
		//}

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

				if p := pl.queue2[i]; p.GetBasicData().StartTime-int64(pl.bMap.ARms) <= pl.progressMs {

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

	// 结束标志
	if len(pl.bMap.Queue) == 0 {
		pl.fadeOut -= timMs / (settings.Playfield.FadeOutTime * 1000)
		pl.fadeOut = math.Max(0.0, pl.fadeOut)
		pl.musicPlayer.SetVolumeRelative(pl.fadeOut)
		pl.dimGlider.UpdateD(timMs)
		pl.blurGlider.UpdateD(timMs)
		//pl.fxGlider.UpdateD(timMs)
		pl.cursorGlider.UpdateD(timMs)
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

	//if settings.Playfield.FlashToTheBeat {
	//	bgAlpha *= pl.Scl
	//}

	pl.batch.Begin()

	pl.batch.SetColor(1, 1, 1, 1)
	pl.batch.ResetTransform()
	pl.batch.SetAdditive(false)
	if pl.Background != nil || pl.storyboard != nil {
		if settings.Playfield.BlurEnable {
			pl.blurEffect.SetBlur(blurVal, blurVal)
			pl.blurEffect.Begin()
		}

		if pl.Background != nil && (pl.storyboard == nil || !pl.storyboard.BGFileUsed()) {
			pl.batch.SetCamera(mgl32.Ortho(-1, 1, -1, 1, 1, -1))
			pl.batch.SetScale(pl.BgScl.X, -pl.BgScl.Y)
			if !settings.Playfield.BlurEnable {
				pl.batch.SetColor(1, 1, 1, bgAlpha)
			}
			pl.batch.DrawUnit(*pl.Background)
		}

		//if pl.storyboard != nil {
		//	pl.batch.SetScale(1, 1)
		//	if !settings.Playfield.BlurEnable {
		//		pl.batch.SetColor(bgAlpha, bgAlpha, bgAlpha, 1)
		//	}
		//	pl.batch.SetCamera(cameras[0])
		//	pl.storyboard.Draw(pl.progressMs, pl.batch)
		//	pl.batch.Flush()
		//}

		if settings.Playfield.BlurEnable {
			pl.batch.End()

			texture := pl.blurEffect.EndAndProcess()
			pl.batch.Begin()
			pl.batch.SetColor(1, 1, 1, bgAlpha)
			pl.batch.SetCamera(mgl32.Ortho(-1, 1, -1, 1, 1, -1))
			pl.batch.DrawUnscaled(texture.GetRegion())
		}

	}

	pl.batch.Flush()

	//if pl.fxGlider.GetValue() > 0.0 {
	//	pl.batch.SetColor(1, 1, 1, pl.fxGlider.GetValue())
	//	pl.batch.SetCamera(mgl32.Ortho(float32(-settings.Graphics.GetWidthF()/2), float32(settings.Graphics.GetWidthF()/2), float32(settings.Graphics.GetHeightF()/2), float32(-settings.Graphics.GetHeightF()/2), 1, -1))
	//	scl := (settings.Graphics.GetWidthF() / float64(pl.Logo.Width)) / 4
	//	pl.batch.SetScale(scl, scl)
	//	pl.batch.DrawTexture(*pl.Logo)
	//	pl.batch.SetScale(scl*(1/pl.Scl), scl*(1/pl.Scl))
	//	pl.batch.SetColor(1, 1, 1, 0.25*pl.fxGlider.GetValue())
	//	pl.batch.DrawTexture(*pl.Logo)
	//}
	//
	//pl.batch.End()

	pl.counter += timMs

	if pl.counter >= 1000.0/60 {
		pl.fpsC = pl.profiler.GetFPS()
		pl.fpsU = pl.profilerU.GetFPS()
		pl.counter -= 1000.0 / 60
		//if pl.storyboard != nil {
		//	pl.storyboardLoad = pl.storyboard.GetLoad()
		//}
	}

	//if pl.fxGlider.GetValue() > 0.0 {
	//
	//	pl.fxBatch.Begin()
	//	pl.batch.SetCamera(mgl32.Ortho(-1, 1, 1, -1, 1, -1))
	//	pl.fxBatch.SetColor(1, 1, 1, 0.25*pl.Scl*pl.fxGlider.GetValue())
	//	pl.vao.Begin()
	//
	//	if pl.vaoDirty {
	//		pl.vao.SetVertexData(pl.vaoD)
	//		pl.vaoDirty = false
	//	}
	//
	//	base := mgl32.Ortho(-1920/2, 1920/2, 1080/2, -1080/2, -1, 1).Mul4(mgl32.Scale3D(600, 600, 0)).Mul4(mgl32.HomogRotate3DZ(float32(pl.fxRotation * math.Pi / 180.0)))
	//
	//	pl.fxBatch.SetTransform(base)
	//	pl.vao.Draw()
	//
	//	pl.fxBatch.SetTransform(base.Mul4(mgl32.HomogRotate3DZ(math.Pi)))
	//	pl.vao.Draw()
	//
	//	pl.vao.End()
	//	pl.fxBatch.End()
	//}

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

	colors1 := settings.Cursor.GetColors(settings.General.Players + 1, settings.TAG, pl.Scl, pl.cursorGlider.GetValue())

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

	//endregion

	//region 渲染按键
	pl.batch.Begin()
	pl.batch.SetCamera(pl.scamera.GetProjectionView())
	for k := 0; k < settings.General.Players; k++ {
		colornum := (settings.General.CursorColorSkipNum*k*len(pl.controller[k].GetCursors())) % settings.General.Players
		namecolor := colors1[colornum]
		if settings.General.EnableBreakandQuit && (!pl.controller[k].GetIsShow()) {
			namecolor[3] = float32(math.Max(0.0, float64(namecolor[3]) - (pl.progressMsF - pl.controller[k].GetDishowTime()) / settings.General.PlayerFadeTime))
		}
		playerkey := pl.controller[k].GetPresskey()
		if playerkey.Key1 {
			pl.batch.SetTranslation(bmath.NewVec2d(pl.key1baseX, pl.keybaseY - pl.lineoffset * float64(k)))
			pl.batch.SetScale(pl.keysize, pl.keysize)
			pl.batch.SetColor(float64(namecolor[0]), float64(namecolor[1]), float64(namecolor[2]), float64(namecolor[3]))
			pl.batch.DrawUnit(*render.PressKey)
		} else {
			pl.batch.SetTranslation(bmath.NewVec2d(pl.key1baseX, pl.keybaseY - pl.lineoffset * float64(k)))
			pl.batch.SetScale(pl.keysize, pl.keysize)
			pl.batch.SetColor(1, 1, 1, 0)
			pl.batch.DrawUnit(*render.PressKey)
		}
		if playerkey.Key2 {
			pl.batch.SetTranslation(bmath.NewVec2d(pl.key2baseX, pl.keybaseY - pl.lineoffset * float64(k)))
			pl.batch.SetScale(pl.keysize, pl.keysize)
			pl.batch.SetColor(float64(namecolor[0]), float64(namecolor[1]), float64(namecolor[2]), float64(namecolor[3]))
			pl.batch.DrawUnit(*render.PressKey)
		} else {
			pl.batch.SetTranslation(bmath.NewVec2d(pl.key2baseX, pl.keybaseY - pl.lineoffset * float64(k)))
			pl.batch.SetScale(pl.keysize, pl.keysize)
			pl.batch.SetColor(1, 1, 1, 0)
			pl.batch.DrawUnit(*render.PressKey)
		}
		if playerkey.LeftClick && !playerkey.Key1 {
			pl.batch.SetTranslation(bmath.NewVec2d(pl.key3baseX, pl.keybaseY - pl.lineoffset * float64(k)))
			pl.batch.SetScale(pl.keysize, pl.keysize)
			pl.batch.SetColor(float64(namecolor[0]), float64(namecolor[1]), float64(namecolor[2]), float64(namecolor[3]))
			pl.batch.DrawUnit(*render.PressKey)
		} else {
			pl.batch.SetTranslation(bmath.NewVec2d(pl.key3baseX, pl.keybaseY - pl.lineoffset * float64(k)))
			pl.batch.SetScale(pl.keysize, pl.keysize)
			pl.batch.SetColor(1, 1, 1, 0)
			pl.batch.DrawUnit(*render.PressKey)
		}
	}
	pl.batch.End()

	//endregion

	//region 渲染文字

	// 文字的公用X轴
	var lastPos []float64
	lastPos = make([]float64, settings.General.Players)
	for k := 0; k < settings.General.Players; k++ {
		lastPos[k] = 0.0
	}
	// 渲染player名
	pl.batch.Begin()
	pl.batch.SetCamera(pl.scamera.GetProjectionView())
	for k := 0; k < settings.General.Players; k++ {
		pl.batch.SetAdditive(true)
		colornum := (settings.General.CursorColorSkipNum*k*len(pl.controller[k].GetCursors())) % settings.General.Players
		namecolor := colors1[colornum]
		if settings.General.EnableBreakandQuit && (!pl.controller[k].GetIsShow()) {
			namecolor[3] = float32(math.Max(0.0, float64(namecolor[3]) - (pl.progressMsF - pl.controller[k].GetDishowTime()) / settings.General.PlayerFadeTime))
		}
		// 渲染player名
		pl.batch.SetColor(float64(namecolor[0]), float64(namecolor[1]), float64(namecolor[2]), float64(namecolor[3]))
		lastPos[k] = pl.font.DrawAndGetLastPosition(pl.batch, pl.playerbaseX, pl.fontbaseY - pl.lineoffset * float64(k), pl.fontsize, pl.controller[k].GetPlayname())
		// 渲染mod
		mods := "+"
		if pl.controller[k].GetIsHR(){
			mods += "HR"
		}
		if pl.controller[k].GetIsHD(){
			mods += "HD"
		}
		if pl.controller[k].GetIsEZ(){
			mods += "EZ"
		}
		if mods != "+" {
			pl.batch.SetColor(1, 1, 1, float64(namecolor[3]))
			lastPos[k] = pl.font.DrawAndGetLastPosition(pl.batch, lastPos[k]+pl.modoffset, pl.fontbaseY-pl.lineoffset*float64(k), pl.fontsize, mods)
		}
	}
	pl.batch.End()

	//endregion

	//region 渲染300、100、50、miss、acc、rank、pp

	// 断连文字的公用X轴
	var lastmissPos []float64
	lastmissPos = make([]float64, settings.General.Players)

	pl.batch.Begin()
	pl.batch.SetCamera(pl.scamera.GetProjectionView())
	for k := 0; k < settings.General.Players; k++ {
		colornum := (settings.General.CursorColorSkipNum*k*len(pl.controller[k].GetCursors())) % settings.General.Players
		namecolor := colors1[colornum]
		// 如果设置不显示，开始降低透明度
		if settings.General.EnableBreakandQuit && (!pl.controller[k].GetIsShow()) {
			namecolor[3] = float32(math.Max(0.0, float64(namecolor[3]) - (pl.progressMsF - pl.controller[k].GetDishowTime()) / settings.General.PlayerFadeTime))
			// 显示断连者名字
			pl.batch.SetColor(float64(namecolor[0]), float64(namecolor[1]), float64(namecolor[2]), float64(namecolor[3]))
			lastmissPos[k] = pl.font.DrawAndGetLastPosition(pl.batch, bmath.GetX(pl.controller[k].GetDishowPos()), bmath.GetY(pl.controller[k].GetDishowPos()), pl.missfontsize, pl.controller[k].GetPlayname())
			// 显示miss
			pl.batch.SetTranslation(bmath.NewVec2d(lastmissPos[k] + pl.missoffsetX, bmath.GetY(pl.controller[k].GetDishowPos()) + pl.missoffsetY))
			pl.batch.SetColor(1, 1, 1, float64(namecolor[3]))
			pl.batch.SetScale(2.75 * pl.misssize, pl.misssize)
			pl.batch.DrawUnit(*render.Hit0)
		}
		// 如果现在时间大于第一个result的时间，渲染这个result，并在渲染一定时间后弹出
		if len(pl.controller[k].GetHitResult()) != 0 {
			if pl.progressMs > pl.controller[k].GetHitResult()[0].JudgeTime {
				judge := *render.Hit300
				pl.batch.SetColor(1, 1, 1, float64(namecolor[3]))
				switch pl.controller[k].GetHitResult()[0].Result {
				case hitjudge.Hit300:
					pl.batch.SetColor(1, 1, 1, 0)
					break
				case hitjudge.Hit100:
					judge = *render.Hit100
					break
				case hitjudge.Hit50:
					judge = *render.Hit50
					break
				case hitjudge.HitMiss:
					judge = *render.Hit0
					break
				}
				if pl.controller[k].GetHitResult()[0].IsBreak {
					// 断连后设置不显示
					if pl.controller[k].GetIsShow() {
						pl.controller[k].SetIsShow(false)
						// 保存消失时间、消失位置
						pl.controller[k].SetDishowTime(pl.progressMsF)
						pl.controller[k].SetDishowPos(pl.controller[k].GetHitResult()[0].JudgePos)
					}
				}
				pl.batch.SetTranslation(bmath.NewVec2d(lastPos[k] + pl.hitoffset, pl.hitbaseY - pl.lineoffset * float64(k)))
				pl.batch.SetScale(2.75 * settings.General.BaseSize, settings.General.BaseSize)
				pl.batch.DrawUnit(judge)
				// 渲染时间结束，弹出
				if pl.progressMs > pl.controller[k].GetHitResult()[0].JudgeTime + settings.General.HitFadeTime {
					// 设置acc、rank和pp
					pl.controller[k].SetAcc(pl.controller[k].GetTotalResult()[0].Acc)
					pl.controller[k].SetPP(pl.controller[k].GetTotalResult()[0].PP.Total)
					switch pl.controller[k].GetTotalResult()[0].Rank {
					case score.SS:
						pl.controller[k].SetRank(*render.RankX)
						break
					case score.S:
						pl.controller[k].SetRank(*render.RankS)
						break
					case score.A:
						pl.controller[k].SetRank(*render.RankA)
						break
					case score.B:
						pl.controller[k].SetRank(*render.RankB)
						break
					case score.C:
						pl.controller[k].SetRank(*render.RankC)
						break
					case score.D:
						pl.controller[k].SetRank(*render.RankD)
						break
					}
					// 弹出
					pl.controller[k].SetHitResult(pl.controller[k].GetHitResult()[1:])
					pl.controller[k].SetTotalResult(pl.controller[k].GetTotalResult()[1:])
				}
			}
		}
		// 渲染acc
		pl.batch.SetColor(1, 1, 1, float64(namecolor[3]))
		pl.font.Draw(pl.batch, pl.accbaseX, pl.fontbaseY - pl.lineoffset * float64(k), pl.fontsize, fmt.Sprintf("%.2f", pl.controller[k].GetAcc()) + "%")
		// 渲染rank
		pl.batch.SetTranslation(bmath.NewVec2d(pl.rankbaseX, pl.rankbaseY - pl.lineoffset * float64(k)))
		pl.batch.SetColor(1, 1, 1, float64(namecolor[3]))
		pl.batch.SetScale(settings.General.BaseSize, settings.General.BaseSize)
		pl.batch.DrawUnit(pl.controller[k].GetRank())
		// 渲染pp
		pl.batch.SetColor(1, 1, 1, float64(namecolor[3]))
		pl.font.Draw(pl.batch, pl.ppbaseX, pl.fontbaseY - pl.lineoffset * float64(k), pl.fontsize, fmt.Sprintf("%.2f", pl.controller[k].GetPP()) + " pp")
	}
	pl.batch.End()

	//endregion

	//region 多个光标渲染

	for k := 0; k < settings.General.Players; k++ {
		if !(settings.General.EnableBreakandQuit && (!pl.controller[k].GetIsShow())) {
			for _, g := range pl.controller[k].GetCursors() {
				g.UpdateRenderer()
			}
			gl.BlendFuncSeparate(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA, gl.ONE, gl.ONE_MINUS_SRC_ALPHA)
			gl.BlendEquation(gl.FUNC_ADD)
			pl.batch.SetAdditive(true)
			render.BeginCursorRender()
			for j := 0; j < settings.DIVIDES; j++ {
				pl.batch.SetCamera(cameras[j])
				for i, g := range pl.controller[k].GetCursors() {
					ind := k*len(pl.controller[k].GetCursors()) + i - 1
					if ind < 0 {
						ind = settings.DIVIDES*len(pl.controller[k].GetCursors()) - 1
					}
					colornum := (settings.General.CursorColorSkipNum*k*len(pl.controller[k].GetCursors())) % settings.General.Players
					g.DrawM(scale2, pl.batch, colors1[colornum], colors1[ind])
				}
			}
			render.EndCursorRender()
		}
	}

	//endregion

	//region 无关4

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
						s.DrawBody(pl.progressMs, pl.bMap.ARms, colors1[settings.General.CursorColorNum-1], colors1[settings.General.CursorColorNum-1], pl.sliderRenderer)
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

		pl.batch.SetScale(64*render.CS*scale1, 64*render.CS*scale1)

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
							s.DrawBody(pl.progressMs, pl.bMap.ARms, colors1[settings.General.CursorColorNum-1], colors1[settings.General.CursorColorNum-1], pl.sliderRenderer)
							pl.sliderRenderer.EndAndRender()
						}
					}
					res := pl.processed[i].Draw(pl.progressMs, pl.bMap.ARms, colors1[settings.General.CursorColorNum-1], pl.batch)
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
					pl.processed[i].DrawApproach(pl.progressMs, pl.bMap.ARms, colors1[settings.General.CursorColorNum-1], pl.batch)
				}
			}
		}

		pl.batch.SetScale(1, 1)
		pl.batch.End()
	}

	pl.batch.SetAdditive(false)
	if settings.Playfield.BloomEnabled {
		pl.bloomEffect.EndAndRender()
	}

	//if settings.DEBUG || settings.FPS {
	//	pl.batch.Begin()
	//	pl.batch.SetColor(1, 1, 1, 1)
	//	pl.batch.SetCamera(pl.scamera.GetProjectionView())
	//
	//	padDown := 4.0
	//	shift := 16.0
	//
	//	if settings.DEBUG {
	//		pl.font.Draw(pl.batch, 0, settings.Graphics.GetHeightF()-24, 24, pl.mapFullName)
	//		pl.font.Draw(pl.batch, 0, padDown+shift*5, 16, fmt.Sprintf("%0.0f FPS", pl.fpsC))
	//		pl.font.Draw(pl.batch, 0, padDown+shift*4, 16, fmt.Sprintf("%0.2f ms", 1000/pl.fpsC))
	//		pl.font.Draw(pl.batch, 0, padDown+shift*3, 16, fmt.Sprintf("%0.2f ms update", 1000/pl.fpsU))
	//
	//		time := int(pl.musicPlayer.GetPosition())
	//		totalTime := int(pl.musicPlayer.GetLength())
	//		mapTime := int(pl.bMap.HitObjects[len(pl.bMap.HitObjects)-1].GetBasicData().EndTime / 1000)
	//
	//		pl.font.Draw(pl.batch, 0, padDown+shift*2, 16, fmt.Sprintf("%02d:%02d / %02d:%02d (%02d:%02d)", time/60, time%60, totalTime/60, totalTime%60, mapTime/60, mapTime%60))
	//		pl.font.Draw(pl.batch, 0, padDown+shift, 16, fmt.Sprintf("%d(*%d) hitobjects, %d total", len(pl.processed), settings.DIVIDES, len(pl.bMap.HitObjects)))
	//
	//		if pl.storyboard != nil {
	//			pl.font.Draw(pl.batch, 0, padDown, 16, fmt.Sprintf("%d storyboard sprites (%0.2fx load), %d in queue (%d total)", pl.storyboard.GetProcessedSprites(), pl.storyboardLoad, pl.storyboard.GetQueueSprites(), pl.storyboard.GetTotalSprites()))
	//		} else {
	//			pl.font.Draw(pl.batch, 0, padDown, 16, "No storyboard")
	//		}
	//	} else {
	//		pl.font.Draw(pl.batch, 0, padDown, 16, fmt.Sprintf("%0.0f FPS", pl.fpsC))
	//	}
	//
	//	pl.batch.End()
	//}

	//endregion

}