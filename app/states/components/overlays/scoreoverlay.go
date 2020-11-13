package overlays

import (
	"fmt"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/wieku/danser-go/app/audio"
	"github.com/wieku/danser-go/app/beatmap/objects"
	"github.com/wieku/danser-go/app/bmath"
	camera2 "github.com/wieku/danser-go/app/bmath/camera"
	"github.com/wieku/danser-go/app/discord"
	"github.com/wieku/danser-go/app/graphics"
	"github.com/wieku/danser-go/app/graphics/font"
	"github.com/wieku/danser-go/app/input"
	"github.com/wieku/danser-go/app/rulesets/osu"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/app/skin"
	"github.com/wieku/danser-go/app/states/components/common"
	"github.com/wieku/danser-go/app/states/components/overlays/play"
	"github.com/wieku/danser-go/framework/bass"
	"github.com/wieku/danser-go/framework/graphics/batch"
	"github.com/wieku/danser-go/framework/graphics/shape"
	"github.com/wieku/danser-go/framework/graphics/sprite"
	"github.com/wieku/danser-go/framework/math/animation"
	"github.com/wieku/danser-go/framework/math/animation/easing"
	color2 "github.com/wieku/danser-go/framework/math/color"
	"github.com/wieku/danser-go/framework/math/vector"
	"math"
	"strconv"
	"strings"
)

type Overlay interface {
	Update(int64)
	DrawBeforeObjects(batch *batch.QuadBatch, colors []color2.Color, alpha float64)
	DrawNormal(batch *batch.QuadBatch, colors []color2.Color, alpha float64)
	DrawHUD(batch *batch.QuadBatch, colors []color2.Color, alpha float64)
	IsBroken(cursor *graphics.Cursor) bool
	NormalBeforeCursor() bool
}

type ScoreOverlay struct {
	font     *font.Font
	lastTime int64
	combo    int64
	newCombo int64
	maxCombo int64

	comboSlide     *animation.Glider
	newComboScale  *animation.Glider
	newComboScaleB *animation.Glider
	newComboFadeB  *animation.Glider

	oldScore    int64
	scoreGlider *animation.Glider
	ppGlider    *animation.Glider
	ruleset     *osu.OsuRuleSet
	cursor      *graphics.Cursor
	combobreak  *bass.Sample
	music       *bass.Track
	nextEnd     int64
	results     *play.HitResults

	keyStates   [4]bool
	keyCounters [4]int
	lastPresses [4]float64
	keyOverlay  *sprite.SpriteManager
	keys        []*sprite.Sprite

	ScaledWidth  float64
	ScaledHeight float64
	camera       *camera2.Camera
	scoreFont    *font.Font
	comboFont    *font.Font
	scoreEFont   *font.Font

	bgDim *animation.Glider

	hitErrorMeter *play.HitErrorMeter

	skip *sprite.Sprite

	healthBackground *sprite.Sprite
	healthBar        *sprite.Sprite
	displayHp        float64

	shapeRenderer *shape.Renderer

	boundaries *common.Boundaries
}

func NewScoreOverlay(ruleset *osu.OsuRuleSet, cursor *graphics.Cursor) *ScoreOverlay {
	overlay := new(ScoreOverlay)
	overlay.results = play.NewHitResults(ruleset.GetBeatMap().Diff)
	overlay.ruleset = ruleset
	overlay.cursor = cursor
	overlay.font = font.GetFont("Exo 2 Bold")

	overlay.comboSlide = animation.NewGlider(0)
	overlay.comboSlide.SetEasing(easing.OutQuad)

	overlay.newComboScale = animation.NewGlider(1)
	overlay.newComboScaleB = animation.NewGlider(1)
	overlay.newComboFadeB = animation.NewGlider(1)

	overlay.scoreGlider = animation.NewGlider(0)
	overlay.scoreGlider.SetEasing(easing.OutQuad)

	overlay.ppGlider = animation.NewGlider(0)
	overlay.ppGlider.SetEasing(easing.OutQuint)

	overlay.bgDim = animation.NewGlider(1)

	overlay.combobreak = audio.LoadSample("combobreak")

	for _, p := range ruleset.GetBeatMap().Pauses {
		bd := p.GetBasicData()

		if bd.EndTime-bd.StartTime < 1000 {
			continue
		}

		overlay.comboSlide.AddEvent(float64(bd.StartTime), float64(bd.StartTime)+500, -1)
		overlay.comboSlide.AddEvent(float64(bd.EndTime-500), float64(bd.EndTime), 0)

		overlay.bgDim.AddEvent(float64(bd.StartTime), float64(bd.StartTime)+500, 0)
		overlay.bgDim.AddEvent(float64(bd.EndTime-500), float64(bd.EndTime), 1)
	}

	discord.UpdatePlay(cursor.Name)

	overlay.scoreEFont = skin.GetFont("scoreentry")
	overlay.scoreFont = skin.GetFont("score")
	overlay.comboFont = skin.GetFont("combo")

	ruleset.SetListener(func(cursor *graphics.Cursor, time int64, number int64, position vector.Vector2d, result osu.HitResult, comboResult osu.ComboResult, pp float64, score1 int64) {

		if result&(osu.BaseHitsM) > 0 {
			overlay.results.AddResult(time, result, position)
		}

		_, hC := ruleset.GetBeatMap().HitObjects[number].(*objects.Circle)
		allowCircle := hC && (result&osu.BaseHits > 0)
		_, sl := ruleset.GetBeatMap().HitObjects[number].(*objects.Slider)
		allowSlider := sl && result == osu.SliderStart

		if allowCircle || allowSlider {
			timeDiff := float64(time) - float64(ruleset.GetBeatMap().HitObjects[number].GetBasicData().StartTime)

			overlay.hitErrorMeter.Add(float64(time), timeDiff)
		}

		if comboResult == osu.ComboResults.Increase {
			overlay.newComboScaleB.Reset()
			overlay.newComboScaleB.AddEventS(float64(time), float64(time+300), 1.7, 1.1)

			overlay.newComboFadeB.Reset()
			overlay.newComboFadeB.AddEventS(float64(time), float64(time+300), 0.6, 0.0)

			overlay.animate(time)

			overlay.combo = overlay.newCombo
			overlay.newCombo++
			overlay.nextEnd = time + 300
		} else if comboResult == osu.ComboResults.Reset {
			if overlay.newCombo > 20 {
				overlay.combobreak.Play()
			}
			overlay.newCombo = 0
		}

		_, _, score, _ := overlay.ruleset.GetResults(overlay.cursor)

		overlay.scoreGlider.Reset()
		overlay.scoreGlider.AddEvent(float64(time), float64(time+500), float64(score))
		overlay.ppGlider.Reset()
		overlay.ppGlider.AddEvent(float64(time), float64(time+500), pp)

		overlay.oldScore = score
	})

	overlay.ScaledHeight = 768
	overlay.ScaledWidth = settings.Graphics.GetAspectRatio() * overlay.ScaledHeight

	overlay.camera = camera2.NewCamera()
	overlay.camera.SetViewportF(0, int(overlay.ScaledHeight), int(overlay.ScaledWidth), 0)
	overlay.camera.Update()

	overlay.keyOverlay = sprite.NewSpriteManager()

	keyBg := sprite.NewSpriteSingle(skin.GetTexture("inputoverlay-background"), 0, vector.NewVec2d(overlay.ScaledWidth, overlay.ScaledHeight/2-64), bmath.Origin.TopLeft)
	keyBg.SetScaleV(vector.NewVec2d(1.05, 1))
	keyBg.ShowForever(true)
	keyBg.SetRotation(math.Pi / 2)

	overlay.keyOverlay.Add(keyBg)

	for i := 0; i < 4; i++ {
		posY := overlay.ScaledHeight/2 - 64 + (30.4+float64(i)*47.2)*settings.Gameplay.KeyOverlay.Scale

		key := sprite.NewSpriteSingle(skin.GetTexture("inputoverlay-key"), 1, vector.NewVec2d(overlay.ScaledWidth-24*settings.Gameplay.KeyOverlay.Scale, posY), bmath.Origin.Centre)
		key.ShowForever(true)

		overlay.keys = append(overlay.keys, key)
		overlay.keyOverlay.Add(key)
	}

	overlay.hitErrorMeter = play.NewHitErrorMeter(overlay.ScaledWidth, overlay.ScaledHeight, ruleset.GetBeatMap().Diff)

	start := overlay.ruleset.GetBeatMap().HitObjects[0].GetBasicData().StartTime - 2000

	if start > 2000 {
		skipFrames := skin.GetFrames("play-skip", true)
		overlay.skip = sprite.NewAnimation(skipFrames, skin.GetInfo().GetFrameTime(len(skipFrames)), true, 0.0, vector.NewVec2d(overlay.ScaledWidth, overlay.ScaledHeight), bmath.Origin.BottomRight)
		overlay.skip.SetAlpha(0.0)
		overlay.skip.AddTransform(animation.NewSingleTransform(animation.Fade, easing.OutQuad, 0, 500, 0.0, 0.6))
		overlay.skip.AddTransform(animation.NewSingleTransform(animation.Fade, easing.OutQuad, float64(start), float64(start+300), 0.6, 0.0))
	}

	overlay.healthBackground = sprite.NewSpriteSingle(skin.GetTexture("scorebar-bg"), 0, vector.NewVec2d(0, 0), bmath.Origin.TopLeft)

	pos := vector.NewVec2d(4.8, 16)
	if skin.GetTexture("scorebar-marker") != nil {
		pos = vector.NewVec2d(12, 12.5)
	}

	barTextures := skin.GetFrames("scorebar-colour", true)

	overlay.healthBar = sprite.NewAnimation(barTextures, skin.GetInfo().GetFrameTime(len(barTextures)), true, 0.0, pos, bmath.Origin.TopLeft)
	overlay.healthBar.SetCutOrigin(bmath.Origin.CentreLeft)

	overlay.shapeRenderer = shape.NewRenderer()

	overlay.boundaries = common.NewBoundaries()

	return overlay
}

func (overlay *ScoreOverlay) animate(time int64) {
	overlay.newComboScale.Reset()
	overlay.newComboScale.AddEventSEase(float64(time), float64(time+50), 1.0, 1.2, easing.InQuad)
	overlay.newComboScale.AddEventSEase(float64(time+50), float64(time+100), 1.2, 1.0, easing.OutQuad)
}

func (overlay *ScoreOverlay) Update(time int64) {

	if input.Win.GetKey(glfw.KeySpace) == glfw.Press {
		if overlay.music != nil && overlay.music.GetState() == bass.MUSIC_PLAYING {
			start := overlay.ruleset.GetBeatMap().HitObjects[0].GetBasicData().StartTime
			if start-time > 4000 {
				overlay.music.SetPosition((float64(start) - 2000) / 1000)
			}
		}
	}

	for sTime := overlay.lastTime + 1; sTime <= time; sTime++ {
		overlay.newComboScale.Update(float64(sTime))
		overlay.newComboScaleB.Update(float64(sTime))
		overlay.newComboFadeB.Update(float64(sTime))
		overlay.scoreGlider.Update(float64(sTime))
		overlay.ppGlider.Update(float64(sTime))
		overlay.comboSlide.Update(float64(sTime))

		if sTime%17 == 0 {
			if overlay.combo > overlay.newCombo && overlay.newCombo == 0 {
				overlay.combo--
			}
		}
	}

	if overlay.combo != overlay.newCombo && overlay.nextEnd < time+140 {
		overlay.animate(time)
		overlay.combo = overlay.newCombo
		overlay.nextEnd = math.MaxInt64
	}

	currentHp := overlay.ruleset.GetHP(overlay.cursor)

	if overlay.displayHp < currentHp {
		overlay.displayHp = math.Min(1.0, overlay.displayHp+math.Abs(currentHp-overlay.displayHp)/4*float64(time-overlay.lastTime)/16.667)
	} else if overlay.displayHp > currentHp {
		overlay.displayHp = math.Max(0.0, overlay.displayHp-math.Abs(overlay.displayHp-currentHp)/6*float64(time-overlay.lastTime)/16.667)
	}

	overlay.healthBar.SetCutX(1.0 - overlay.displayHp)

	overlay.results.Update(float64(time))

	overlay.hitErrorMeter.Update(float64(time))

	currentStates := [4]bool{overlay.cursor.LeftKey, overlay.cursor.RightKey, overlay.cursor.LeftMouse && !overlay.cursor.LeftKey, overlay.cursor.RightMouse && !overlay.cursor.RightKey}

	for i, state := range currentStates {
		color := color2.Color{R: 1.0, G: 222.0 / 255, B: 0, A: 0}
		if i > 1 {
			color = color2.Color{R: 248.0 / 255, G: 0, B: 158.0 / 255, A: 0}
		}

		if !overlay.keyStates[i] && state {
			key := overlay.keys[i]
			key.ClearTransformationsOfType(animation.Scale)
			key.AddTransform(animation.NewSingleTransform(animation.Scale, easing.OutQuad, float64(time), float64(time+100), 1.0, 0.8))

			key.AddTransform(animation.NewColorTransform(animation.Color3, easing.OutQuad, float64(time), float64(time+100), color2.Color{R: 1, G: 1, B: 1, A: 1}, color))

			overlay.lastPresses[i] = float64(time + 100)
			overlay.keyCounters[i]++
		}

		if overlay.keyStates[i] && !state {
			key := overlay.keys[i]
			key.ClearTransformationsOfType(animation.Scale)
			key.AddTransform(animation.NewSingleTransform(animation.Scale, easing.OutQuad, math.Max(float64(time), overlay.lastPresses[i]), float64(time+100), key.GetScale().Y, 1.0))

			key.AddTransform(animation.NewColorTransform(animation.Color3, easing.OutQuad, float64(time), float64(time+100), color, color2.Color{R: 1, G: 1, B: 1, A: 1}))
		}

		overlay.keyStates[i] = state
	}

	overlay.keyOverlay.Update(time)
	overlay.bgDim.Update(float64(time))
	overlay.healthBackground.Update(time)
	overlay.healthBar.Update(time)

	if overlay.skip != nil {
		overlay.skip.Update(time)
	}

	overlay.lastTime = time
}

func (overlay *ScoreOverlay) SetMusic(music *bass.Track) {
	overlay.music = music
}

func (overlay *ScoreOverlay) DrawBeforeObjects(batch *batch.QuadBatch, colors []color2.Color, alpha float64) {
	overlay.boundaries.Draw(batch.Projection, float32(overlay.ruleset.GetBeatMap().Diff.CircleRadius), float32(alpha*overlay.bgDim.GetValue()))
}

func (overlay *ScoreOverlay) DrawNormal(batch *batch.QuadBatch, colors []color2.Color, alpha float64) {
	scale := overlay.ruleset.GetBeatMap().Diff.CircleRadius / 64
	batch.SetScale(scale, scale)

	overlay.results.Draw(batch, 1.0)

	prev := batch.Projection
	batch.SetCamera(overlay.camera.GetProjectionView())

	overlay.hitErrorMeter.Draw(batch, alpha)

	batch.SetScale(1, 1)
	batch.SetColor(1, 1, 1, alpha)

	if overlay.skip != nil {
		overlay.skip.Draw(overlay.lastTime, batch)
	}

	batch.SetCamera(prev)
}

func (overlay *ScoreOverlay) DrawHUD(batch *batch.QuadBatch, colors []color2.Color, alpha float64) {
	prev := batch.Projection
	batch.SetCamera(overlay.camera.GetProjectionView())
	batch.ResetTransform()
	batch.SetColor(1, 1, 1, alpha)

	hObjects := overlay.ruleset.GetBeatMap().HitObjects

	startTime := float32(hObjects[0].GetBasicData().StartTime)
	endTime := float32(hObjects[len(hObjects)-1].GetBasicData().EndTime)
	musicPos := float32(0.0)
	if overlay.music != nil {
		musicPos = float32(overlay.music.GetPosition()) * 1000
	}

	progress := bmath.ClampF32((musicPos-startTime)/(endTime-startTime), 0.0, 1.0)
	if musicPos < startTime {
		progress = bmath.ClampF32(-1.0+musicPos/startTime, -1.0, 0.0)
	}

	if scoreAlpha := settings.Gameplay.Score.Opacity; scoreAlpha > 0.001 && settings.Gameplay.Score.Show && settings.Gameplay.ProgressBar == "Pie" {
		scoreScale := settings.Gameplay.Score.Scale
		fntSize := overlay.scoreFont.GetSize() * scoreScale * 0.96

		rightOffset := -9.6
		accOffset := overlay.ScaledWidth - overlay.scoreFont.GetWidthMonospaced(fntSize*0.6, "99.99%") - 38.4 + rightOffset
		vAccOffset := 4.8

		overlay.shapeRenderer.SetCamera(overlay.camera.GetProjectionView())

		if progress < 0.0 {
			overlay.shapeRenderer.SetColor(0.4, 0.8, 0.4, alpha*0.6*scoreAlpha)
		} else {
			overlay.shapeRenderer.SetColor(1, 1, 1, 0.6*alpha*scoreAlpha)
		}

		overlay.shapeRenderer.Begin()
		overlay.shapeRenderer.DrawCircleProgressS(vector.NewVec2f(float32(accOffset), float32(fntSize+vAccOffset+fntSize*0.6/2)), 16, 40, float32(progress))
		overlay.shapeRenderer.End()
	}

	overlay.healthBackground.Draw(overlay.lastTime, batch)
	overlay.healthBar.Draw(overlay.lastTime, batch)

	//region Combo rendering

	if comboAlpha := settings.Gameplay.ComboCounter.Opacity; comboAlpha > 0.001 && settings.Gameplay.ComboCounter.Show {
		cmbSize := overlay.comboFont.GetSize() * settings.Gameplay.ComboCounter.Scale

		batch.SetColor(1, 1, 1, overlay.newComboFadeB.GetValue()*alpha*comboAlpha)

		shiftL := overlay.comboSlide.GetValue() * overlay.comboFont.GetWidth(cmbSize*overlay.newComboScale.GetValue(), fmt.Sprintf("%dx", overlay.combo))

		overlay.comboFont.Draw(batch, shiftL, overlay.ScaledHeight-cmbSize*overlay.newComboScaleB.GetValue()/2, cmbSize*overlay.newComboScaleB.GetValue(), fmt.Sprintf("%dx", overlay.newCombo))
		batch.SetColor(1, 1, 1, alpha*comboAlpha)
		overlay.comboFont.Draw(batch, shiftL, overlay.ScaledHeight-cmbSize*overlay.newComboScale.GetValue()/2, cmbSize*overlay.newComboScale.GetValue(), fmt.Sprintf("%dx", overlay.combo))
	}

	//endregion

	//region Score+progress+accuracy

	if scoreAlpha := settings.Gameplay.Score.Opacity; scoreAlpha > 0.001 && settings.Gameplay.Score.Show {
		batch.ResetTransform()

		scoreScale := settings.Gameplay.Score.Scale
		fntSize := overlay.scoreFont.GetSize() * scoreScale * 0.96
		rightOffset := -9.6
		accOffset := overlay.ScaledWidth - overlay.scoreFont.GetWidthMonospaced(fntSize*0.6, "99.99%") - 38.4 + rightOffset
		vAccOffset := 4.8

		if settings.Gameplay.ProgressBar == "Pie" {
			text := skin.GetTextureSource("circularmetre", skin.LOCAL)

			batch.SetColor(1, 1, 1, alpha*scoreAlpha)
			batch.SetTranslation(vector.NewVec2d(accOffset, fntSize+vAccOffset+fntSize*0.6/2))
			batch.DrawTexture(*text)

			accOffset -= 44.8
		} else if progress > 0.0 {
			batch.SetColor(0.2, 0.6, 0.2, alpha*0.8*scoreAlpha)

			batch.SetSubScale(272*float64(progress)*scoreScale/2, 2.5*scoreScale)
			batch.SetTranslation(vector.NewVec2d(overlay.ScaledWidth+(-12-272+float64(progress)*272/2)*scoreScale, fntSize))
			batch.DrawUnit(graphics.Pixel.GetRegion())
		}

		batch.ResetTransform()
		batch.SetColor(1, 1, 1, alpha*scoreAlpha)

		scoreText := fmt.Sprintf("%08d", int64(math.Round(overlay.scoreGlider.GetValue())))
		overlay.scoreFont.DrawMonospaced(batch, overlay.ScaledWidth+rightOffset-overlay.scoreFont.GetWidthMonospaced(fntSize, scoreText)+skin.GetInfo().ScoreOverlap, fntSize/2, fntSize, scoreText)

		acc, _, _, _ := overlay.ruleset.GetResults(overlay.cursor)

		var accText string
		if acc == 100 {
			accText = fmt.Sprintf("%5.1f%%", acc)
		} else {
			accText = fmt.Sprintf("%5.2f%%", acc)
		}

		overlay.scoreFont.DrawMonospaced(batch, overlay.ScaledWidth+rightOffset-overlay.scoreFont.GetWidthMonospaced(fntSize*0.6, accText)+skin.GetInfo().ScoreOverlap*0.6, fntSize+vAccOffset+fntSize*0.6/2, fntSize*0.6, accText)

		if _, _, _, grade := overlay.ruleset.GetResults(overlay.cursor); grade != osu.NONE {
			gText := strings.ToLower(strings.ReplaceAll(osu.GradesText[grade], "SS", "X"))

			text := skin.GetTexture("ranking-" + gText + "-small")

			aspect := float64(text.Width) / float64(text.Height)

			batch.SetTranslation(vector.NewVec2d(accOffset, fntSize+vAccOffset+fntSize*0.6/2))
			batch.SetSubScale(fntSize*aspect*0.6/2, fntSize*0.6/2)
			batch.DrawUnit(*text)
		}
	}

	//endregion

	//region pp

	batch.SetColor(1, 1, 1, alpha)
	batch.SetScale(1, -1)
	batch.SetSubScale(1, 1)

	overlay.font.DrawMonospaced(batch, 0, 150, 40, fmt.Sprintf("%0.2fpp", overlay.ppGlider.GetValue()))

	batch.SetScale(1, 1)

	//endregion

	batch.ResetTransform()

	if keyAlpha := settings.Gameplay.KeyOverlay.Opacity; keyAlpha > 0.001 && settings.Gameplay.KeyOverlay.Show {
		keyScale := settings.Gameplay.KeyOverlay.Scale

		batch.SetColor(1, 1, 1, alpha*keyAlpha)
		batch.SetScale(keyScale, keyScale)

		overlay.keyOverlay.Draw(overlay.lastTime, batch)

		col := skin.GetInfo().InputOverlayText
		batch.SetColor(float64(col.R), float64(col.G), float64(col.B), alpha*keyAlpha)

		for i := 0; i < 4; i++ {
			posX := overlay.ScaledWidth - 24*keyScale
			posY := overlay.ScaledHeight/2 - 64 + (30.4+float64(i)*47.2)*keyScale
			scale := overlay.keys[i].GetScale().Y * keyScale

			text := strconv.Itoa(overlay.keyCounters[i])

			if overlay.keyCounters[i] == 0 {
				text = "K"
				if i > 1 {
					text = "M"
				}

				text += strconv.Itoa(i%2 + 1)
			}

			if overlay.keyCounters[i] == 0 || overlay.scoreEFont == nil {
				texLen := overlay.font.GetWidthMonospaced(scale*14, text)

				batch.SetScale(1, -1)
				overlay.font.DrawMonospaced(batch, posX-texLen/2, posY+scale*14/3, scale*14, text)
			} else {
				siz := scale * overlay.scoreEFont.GetSize()
				batch.SetScale(1, 1)
				overlay.scoreEFont.DrawCentered(batch, posX, posY, siz, text)
			}
		}
	}

	batch.ResetTransform()
	batch.SetColor(1, 1, 1, alpha)

	batch.SetCamera(prev)
}

func (overlay *ScoreOverlay) IsBroken(cursor *graphics.Cursor) bool {
	return false
}

func (overlay *ScoreOverlay) NormalBeforeCursor() bool {
	return true
}
