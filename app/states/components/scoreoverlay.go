package components

import (
	"fmt"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/wieku/danser-go/app/audio"
	"github.com/wieku/danser-go/app/beatmap/objects"
	"github.com/wieku/danser-go/app/bmath"
	"github.com/wieku/danser-go/app/bmath/difficulty"
	"github.com/wieku/danser-go/app/discord"
	"github.com/wieku/danser-go/app/graphics"
	"github.com/wieku/danser-go/app/graphics/font"
	"github.com/wieku/danser-go/app/input"
	"github.com/wieku/danser-go/app/rulesets/osu"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/app/skin"
	"github.com/wieku/danser-go/framework/bass"
	"github.com/wieku/danser-go/framework/graphics/sprite"
	"github.com/wieku/danser-go/framework/math/animation"
	"github.com/wieku/danser-go/framework/math/animation/easing"
	"github.com/wieku/danser-go/framework/math/vector"
	"math"
	"math/rand"
	"strconv"
	"strings"
)

const errorScale = 2.0

var colors = []bmath.Color{{0.2, 0.8, 1, 1}, {0.44, 0.98, 0.18, 1}, {0.85, 0.68, 0.27, 1}}

type Overlay interface {
	Update(int64)
	DrawBeforeObjects(batch *sprite.SpriteBatch, colors []mgl32.Vec4, alpha float64)
	DrawNormal(batch *sprite.SpriteBatch, colors []mgl32.Vec4, alpha float64)
	DrawHUD(batch *sprite.SpriteBatch, colors []mgl32.Vec4, alpha float64)
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
	results     *sprite.SpriteManager

	keyStates   [4]bool
	keyCounters [4]int
	lastPresses [4]float64
	keyOverlay  *sprite.SpriteManager
	keys        []*sprite.Sprite

	ScaledWidth  float64
	ScaledHeight float64
	camera       *bmath.Camera
	scoreFont    *font.Font
	comboFont    *font.Font
	scoreEFont   *font.Font

	bgDim *animation.Glider

	errorDisplay     *sprite.SpriteManager
	errorCurrent     float64
	triangle         *sprite.Sprite
	errorDisplayFade *animation.Glider
}

func NewScoreOverlay(ruleset *osu.OsuRuleSet, cursor *graphics.Cursor) *ScoreOverlay {
	overlay := new(ScoreOverlay)
	overlay.results = sprite.NewSpriteManager()
	overlay.ruleset = ruleset
	overlay.cursor = cursor
	overlay.font = font.GetFont("Exo 2 Bold")

	overlay.comboSlide = animation.NewGlider(0)
	overlay.comboSlide.SetEasing(easing.OutQuad)

	overlay.newComboScale = animation.NewGlider(1)
	overlay.newComboScaleB = animation.NewGlider(1)
	overlay.newComboFadeB = animation.NewGlider(1)

	overlay.scoreGlider = animation.NewGlider(0)
	overlay.scoreGlider.SetEasing(easing.OutQuint)

	overlay.ppGlider = animation.NewGlider(0)
	overlay.ppGlider.SetEasing(easing.OutQuint)

	overlay.bgDim = animation.NewGlider(1)

	overlay.combobreak = audio.LoadSample("assets/default-skin/combobreak")

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

	skin.GetFrames("hit0", true)
	skin.GetFrames("hit50", true)
	skin.GetFrames("hit100", true)

	ruleset.SetListener(func(cursor *graphics.Cursor, time int64, number int64, position vector.Vector2d, result osu.HitResult, comboResult osu.ComboResult, pp float64, score1 int64) {

		if result == osu.HitResults.Hit100 || result == osu.HitResults.Hit50 || result == osu.HitResults.Miss {
			overlay.AddResult(time, result, position)
		}

		_, hC := ruleset.GetBeatMap().HitObjects[number].(*objects.Circle)
		allowCircle := hC && (result == osu.HitResults.Hit300 || result == osu.HitResults.Hit100 || result == osu.HitResults.Hit50)
		_, sl := ruleset.GetBeatMap().HitObjects[number].(*objects.Slider)
		allowSlider := sl && result == osu.HitResults.SliderStart

		if allowCircle || allowSlider {
			timeDiff := float64(time) - float64(ruleset.GetBeatMap().HitObjects[number].GetBasicData().StartTime)
			timeDiffA := int64(math.Abs(timeDiff))

			pixel := graphics.Pixel.GetRegion()
			middle := sprite.NewSpriteSingle(&pixel, 3.0, vector.NewVec2d(overlay.ScaledWidth/2+timeDiff*errorScale, overlay.ScaledHeight-10*errorScale), bmath.Origin.Centre)
			middle.SetScaleV(vector.NewVec2d(1.5*errorScale, 20*errorScale))
			middle.SetAdditive(true)

			var col bmath.Color
			switch {
			case timeDiffA < ruleset.GetBeatMap().Diff.Hit300:
				col = colors[0]
			case timeDiffA < ruleset.GetBeatMap().Diff.Hit100:
				col = colors[1]
			case timeDiffA < ruleset.GetBeatMap().Diff.Hit50:
				col = colors[2]
			}

			middle.SetColor(col)

			middle.AddTransform(animation.NewSingleTransform(animation.Fade, easing.Linear, float64(time), float64(time+10000), 0.4, 0.0))
			middle.AdjustTimesToTransformations()

			overlay.errorDisplay.Add(middle)

			overlay.errorCurrent = overlay.errorCurrent*0.8 + timeDiff*0.2

			overlay.triangle.ClearTransformations()
			overlay.triangle.AddTransform(animation.NewSingleTransform(animation.MoveX, easing.OutQuad, float64(time), float64(time+800), overlay.triangle.GetPosition().X, overlay.ScaledWidth/2+overlay.errorCurrent))

			overlay.errorDisplayFade.Reset()
			overlay.errorDisplayFade.SetValue(1.0)
			overlay.errorDisplayFade.AddEventSEase(float64(time+4000), float64(time+5000), 1.0, 0.0, easing.InQuad)
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

	overlay.camera = bmath.NewCamera()
	overlay.camera.SetViewportF(0, int(overlay.ScaledHeight), int(overlay.ScaledWidth), 0)
	overlay.camera.Update()

	overlay.keyOverlay = sprite.NewSpriteManager()

	keyBg := sprite.NewSpriteSingle(skin.GetTexture("inputoverlay-background"), 0, vector.NewVec2d(overlay.ScaledWidth, overlay.ScaledHeight/2-40), bmath.Origin.TopLeft)
	keyBg.SetScaleV(vector.NewVec2d(1.05, 1))
	keyBg.ShowForever(true)
	keyBg.SetRotation(math.Pi / 2)

	overlay.keyOverlay.Add(keyBg)

	for i := 0; i < 4; i++ {
		posY := overlay.ScaledHeight/2 - 40 + 30 + float64(i)*47.5

		key := sprite.NewSpriteSingle(skin.GetTexture("inputoverlay-key"), 1, vector.NewVec2d(overlay.ScaledWidth-24, posY), bmath.Origin.Centre)
		key.ShowForever(true)

		overlay.keys = append(overlay.keys, key)
		overlay.keyOverlay.Add(key)
	}

	overlay.errorDisplayFade = animation.NewGlider(0.0)

	overlay.errorDisplay = sprite.NewSpriteManager()

	sum := ruleset.GetBeatMap().Diff.Hit50

	pixel := graphics.Pixel.GetRegion()
	bg := sprite.NewSpriteSingle(&pixel, 0.0, vector.NewVec2d(overlay.ScaledWidth/2, overlay.ScaledHeight-10*errorScale), bmath.Origin.Centre)
	bg.SetScaleV(vector.NewVec2d(float64(sum)*2*errorScale, 20*errorScale))
	bg.SetColor(bmath.Color{0, 0, 0, 1})
	bg.SetAlpha(0.8)
	overlay.errorDisplay.Add(bg)

	vals := []float64{float64(ruleset.GetBeatMap().Diff.Hit300), float64(ruleset.GetBeatMap().Diff.Hit100), float64(ruleset.GetBeatMap().Diff.Hit50)}

	for i, v := range vals {
		pos := 0.0
		width := v

		if i > 0 {
			pos = vals[i-1]
			width -= vals[i-1]
		}

		left := sprite.NewSpriteSingle(&pixel, 1.0, vector.NewVec2d(overlay.ScaledWidth/2-pos*errorScale, overlay.ScaledHeight-10*errorScale), bmath.Origin.CentreRight)
		left.SetScaleV(vector.NewVec2d(width*errorScale, 4*errorScale))
		left.SetColor(colors[i])
		left.SetAlpha(0.8)

		overlay.errorDisplay.Add(left)

		right := sprite.NewSpriteSingle(&pixel, 1.0, vector.NewVec2d(overlay.ScaledWidth/2+pos*errorScale, overlay.ScaledHeight-10*errorScale), bmath.Origin.CentreLeft)
		right.SetScaleV(vector.NewVec2d(width*errorScale, 4*errorScale))
		right.SetColor(colors[i])
		right.SetAlpha(0.8)

		overlay.errorDisplay.Add(right)
	}

	middle := sprite.NewSpriteSingle(&pixel, 2.0, vector.NewVec2d(overlay.ScaledWidth/2, overlay.ScaledHeight-10*errorScale), bmath.Origin.Centre)
	middle.SetScaleV(vector.NewVec2d(2*errorScale, 20*errorScale))
	middle.SetAlpha(0.8)

	overlay.errorDisplay.Add(middle)

	overlay.triangle = sprite.NewSpriteSingle(graphics.TriangleSmall, 2.0, vector.NewVec2d(overlay.ScaledWidth/2, overlay.ScaledHeight-12*errorScale), bmath.Origin.BottomCentre)
	overlay.triangle.SetScaleV(vector.NewVec2d(errorScale/8, errorScale/8))
	overlay.triangle.SetAlpha(0.8)

	overlay.errorDisplay.Add(overlay.triangle)

	return overlay
}

func (overlay *ScoreOverlay) animate(time int64) {
	overlay.newComboScale.Reset()
	overlay.newComboScale.AddEventSEase(float64(time), float64(time+50), 1.0, 1.2, easing.InQuad)
	overlay.newComboScale.AddEventSEase(float64(time+50), float64(time+100), 1.2, 1.0, easing.OutQuad)
}

func (overlay *ScoreOverlay) AddResult(time int64, result osu.HitResult, position vector.Vector2d) {
	var tex string

	switch result {
	case osu.HitResults.Hit100:
		tex = "hit100"
	case osu.HitResults.Hit50:
		tex = "hit50"
	case osu.HitResults.Miss:
		tex = "hit0"
	}

	if tex == "" {
		return
	}

	frames := skin.GetFrames(tex, true)

	sprite := sprite.NewAnimation(frames, 1000.0/60, false, -float64(time), position, bmath.Origin.Centre)

	fadeIn := float64(time + difficulty.ResultFadeIn)
	postEmpt := float64(time + difficulty.PostEmpt)
	fadeOut := postEmpt + float64(difficulty.ResultFadeOut)

	sprite.AddTransformUnordered(animation.NewSingleTransform(animation.Fade, easing.Linear, float64(time), fadeIn, 0.0, 1.0))
	sprite.AddTransformUnordered(animation.NewSingleTransform(animation.Fade, easing.Linear, postEmpt, fadeOut, 1.0, 0.0))

	if len(frames) == 1 {
		sprite.AddTransformUnordered(animation.NewSingleTransform(animation.Scale, easing.Linear, float64(time), float64(time+difficulty.ResultFadeIn*0.8), 0.6, 1.1))
		sprite.AddTransformUnordered(animation.NewSingleTransform(animation.Scale, easing.Linear, fadeIn, float64(time+difficulty.ResultFadeIn*1.2), 1.1, 0.9))
		sprite.AddTransformUnordered(animation.NewSingleTransform(animation.Scale, easing.Linear, float64(time+difficulty.ResultFadeIn*1.2), float64(time+difficulty.ResultFadeIn*1.4), 0.9, 1.0))

		if result == osu.HitResults.Miss {
			rotation := rand.Float64()*0.3 - 0.15

			sprite.AddTransformUnordered(animation.NewSingleTransform(animation.Rotate, easing.Linear, float64(time), fadeIn, 0.0, rotation))
			sprite.AddTransformUnordered(animation.NewSingleTransform(animation.Rotate, easing.Linear, fadeIn, fadeOut, rotation, rotation*2))

			sprite.AddTransformUnordered(animation.NewSingleTransform(animation.MoveY, easing.Linear, float64(time), fadeOut, position.Y-5, position.Y+40))
		}
	}

	sprite.SortTransformations()
	sprite.AdjustTimesToTransformations()
	sprite.ResetValuesToTransforms()

	overlay.results.Add(sprite)
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

	overlay.results.Update(time)
	overlay.errorDisplayFade.Update(float64(time))
	overlay.errorDisplay.Update(time)

	currentStates := [4]bool{overlay.cursor.LeftButton, overlay.cursor.RightButton, false, false}

	for i, state := range currentStates {
		color := bmath.Color{R: 1.0, G: 222.0 / 255, B: 0, A: 0}
		if i > 1 {
			color = bmath.Color{R: 248.0 / 255, G: 0, B: 258.0 / 255, A: 0}
		}

		if !overlay.keyStates[i] && state {
			key := overlay.keys[i]
			key.ClearTransformationsOfType(animation.Scale)
			key.AddTransform(animation.NewSingleTransform(animation.Scale, easing.OutQuad, float64(time), float64(time+100), 1.0, 0.8))

			key.AddTransform(animation.NewColorTransform(animation.Color3, easing.OutQuad, float64(time), float64(time+100), bmath.Color{R: 1, G: 1, B: 1, A: 1}, color))

			overlay.lastPresses[i] = float64(time + 100)
			overlay.keyCounters[i]++
		}

		if overlay.keyStates[i] && !state {
			key := overlay.keys[i]
			key.ClearTransformationsOfType(animation.Scale)
			key.AddTransform(animation.NewSingleTransform(animation.Scale, easing.OutQuad, math.Max(float64(time), overlay.lastPresses[i]), float64(time+100), key.GetScale().Y, 1.0))

			key.AddTransform(animation.NewColorTransform(animation.Color3, easing.OutQuad, float64(time), float64(time+100), color, bmath.Color{R: 1, G: 1, B: 1, A: 1}))
		}

		overlay.keyStates[i] = state
	}

	overlay.keyOverlay.Update(time)
	overlay.bgDim.Update(float64(time))

	overlay.lastTime = time
}

func (overlay *ScoreOverlay) SetMusic(music *bass.Track) {
	overlay.music = music
}

func (overlay *ScoreOverlay) DrawBeforeObjects(batch *sprite.SpriteBatch, colors []mgl32.Vec4, alpha float64) {
	alpha *= overlay.bgDim.GetValue()
	cs := overlay.ruleset.GetBeatMap().Diff.CircleRadius
	sizeX := 512 + (cs+0.3)*2
	sizeY := 384 + (cs+0.3)*2

	batch.SetScale(sizeX/2, sizeY/2)
	batch.SetColor(0, 0, 0, 0.8*alpha)
	batch.SetTranslation(vector.NewVec2d(256, 192)) //bg
	batch.DrawUnit(graphics.Pixel.GetRegion())

	batch.SetColor(1, 1, 1, alpha)
	batch.SetScale(sizeX/2, 0.3)
	batch.SetTranslation(vector.NewVec2d(256, -cs)) //top line
	batch.DrawUnit(graphics.Pixel.GetRegion())

	batch.SetTranslation(vector.NewVec2d(256, 384+cs)) //bottom line
	batch.DrawUnit(graphics.Pixel.GetRegion())

	batch.SetScale(0.3, sizeY/2)
	batch.SetTranslation(vector.NewVec2d(-cs, 192)) //left line
	batch.DrawUnit(graphics.Pixel.GetRegion())
	batch.SetTranslation(vector.NewVec2d(512+cs, 192)) //right line
	batch.DrawUnit(graphics.Pixel.GetRegion())
	batch.SetScale(1, 1)
}

func (overlay *ScoreOverlay) DrawNormal(batch *sprite.SpriteBatch, colors []mgl32.Vec4, alpha float64) {
	scale := overlay.ruleset.GetBeatMap().Diff.CircleRadius / 64
	batch.SetScale(scale, scale)

	overlay.results.Draw(overlay.lastTime, batch)

	prev := batch.Projection
	batch.SetCamera(overlay.camera.GetProjectionView())
	batch.SetScale(1, 1)
	batch.SetColor(1, 1, 1, overlay.errorDisplayFade.GetValue())

	overlay.errorDisplay.Draw(overlay.lastTime, batch)

	batch.SetCamera(prev)
}

func (overlay *ScoreOverlay) DrawHUD(batch *sprite.SpriteBatch, colors []mgl32.Vec4, alpha float64) {
	prev := batch.Projection
	batch.SetCamera(overlay.camera.GetProjectionView())
	batch.SetScale(1, 1)

	fntSize := overlay.scoreFont.GetSize()
	cmbSize := overlay.comboFont.GetSize()

	//region Combo rendering

	batch.SetColor(1, 1, 1, overlay.newComboFadeB.GetValue()*alpha)

	shiftL := overlay.comboSlide.GetValue() * overlay.comboFont.GetWidth(cmbSize*overlay.newComboScale.GetValue(), fmt.Sprintf("%dx", overlay.combo))

	overlay.comboFont.Draw(batch, shiftL, overlay.ScaledHeight-cmbSize*overlay.newComboScaleB.GetValue()/2, cmbSize*overlay.newComboScaleB.GetValue(), fmt.Sprintf("%dx", overlay.newCombo))
	batch.SetColor(1, 1, 1, alpha)
	overlay.comboFont.Draw(batch, shiftL, overlay.ScaledHeight-cmbSize*overlay.newComboScale.GetValue()/2, cmbSize*overlay.newComboScale.GetValue(), fmt.Sprintf("%dx", overlay.combo))

	//endregion

	//region Score+progress+accuracy

	scoreText := fmt.Sprintf("%08d", int64(overlay.scoreGlider.GetValue()))
	overlay.scoreFont.DrawMonospaced(batch, overlay.ScaledWidth-overlay.scoreFont.GetWidthMonospaced(fntSize, scoreText), fntSize/2, fntSize, scoreText)

	acc, _, _, _ := overlay.ruleset.GetResults(overlay.cursor)
	accText := fmt.Sprintf("%0.2f%%", acc)
	overlay.scoreFont.Draw(batch, overlay.ScaledWidth-overlay.scoreFont.GetWidth(fntSize*0.45, accText), 9+fntSize+fntSize*0.45/2, fntSize*0.45, accText)

	if _, _, _, grade := overlay.ruleset.GetResults(overlay.cursor); grade != osu.NONE {
		gText := strings.ToLower(strings.ReplaceAll(osu.GradesText[grade], "SS", "X"))

		text := skin.GetTexture("ranking-" + gText + "-small")

		batch.SetTranslation(vector.NewVec2d(overlay.ScaledWidth-overlay.scoreFont.GetWidth(fntSize*0.45, "100.00%")-float64(text.Width)/2, 9+fntSize+fntSize*0.45/2))
		batch.DrawTexture(*text)
	}

	hObjects := overlay.ruleset.GetBeatMap().HitObjects

	startTime := float64(hObjects[0].GetBasicData().StartTime)
	endTime := float64(hObjects[len(hObjects)-1].GetBasicData().EndTime)
	musicPos := 0.0
	if overlay.music != nil {
		musicPos = overlay.music.GetPosition() * 1000
	}

	progress := math.Min(1.0, math.Max(0.0, (musicPos-startTime)/(endTime-startTime)))
	//log.Println(progress)
	batch.SetColor(0.2, 0.6, 0.2, alpha*0.8)

	batch.SetSubScale(100*progress, 3)
	batch.SetTranslation(vector.NewVec2d(overlay.ScaledWidth-5-200+progress*100, fntSize+4))
	batch.DrawUnit(graphics.Pixel.GetRegion())

	batch.SetColor(1, 1, 1, alpha)
	batch.SetSubScale(1, 1)

	batch.SetScale(1, -1)
	overlay.font.DrawMonospaced(batch, 0, 150, fntSize*0.7, fmt.Sprintf("%0.2fpp", overlay.ppGlider.GetValue()))
	batch.SetScale(1, 1)

	//endregion

	batch.ResetTransform()
	batch.SetColor(1, 1, 1, alpha)

	overlay.keyOverlay.Draw(overlay.lastTime, batch)

	for i := 0; i < 4; i++ {
		posX := overlay.ScaledWidth - 24
		posY := overlay.ScaledHeight/2 - 40 + 30 + float64(i)*47.5
		scale := overlay.keys[i].GetScale().Y

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

			batch.SetColor(1, 1, 1, alpha)
			batch.SetScale(1, -1)
			overlay.font.DrawMonospaced(batch, posX-texLen/2, posY+scale*14/3, scale*14, text)
		} else {
			siz := scale * overlay.scoreEFont.GetSize()
			batch.SetScale(1, 1)
			overlay.scoreEFont.DrawCentered(batch, posX, posY, siz, text)
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
