package objects

import (
	"fmt"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/wieku/danser-go/app/audio"
	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/app/skin"
	"github.com/wieku/danser-go/framework/bass"
	"github.com/wieku/danser-go/framework/graphics/batch"
	"github.com/wieku/danser-go/framework/graphics/sprite"
	"github.com/wieku/danser-go/framework/math/animation"
	"github.com/wieku/danser-go/framework/math/animation/easing"
	color2 "github.com/wieku/danser-go/framework/math/color"
	"github.com/wieku/danser-go/framework/math/math32"
	"github.com/wieku/danser-go/framework/math/mutils"
	"github.com/wieku/danser-go/framework/math/vector"
	"math"
	"math/rand"
	"strconv"
)

const rpms = 0.00795

var spinnerRed = color2.Color{R: 1, G: 0, B: 0, A: 1}
var spinnerBlue = color2.Color{R: 0.05, G: 0.5, B: 1.0, A: 1}

type Spinner struct {
	*HitObject

	Timings  *Timings
	diff     *difficulty.Difficulty
	sample   int
	rad      float32
	pos      vector.Vector2f
	fade     *animation.Glider
	lastTime float64
	rpm      float64

	spinnerbonus *bass.Sample
	loopSample   *bass.SampleChannel
	completion   float64

	newStyle     bool
	sprites      *sprite.Manager
	frontSprites *sprite.Manager

	glow     *sprite.Sprite
	bottom   *sprite.Sprite
	top      *sprite.Sprite
	middle2  *sprite.Sprite
	middle   *sprite.Sprite
	approach *sprite.Sprite

	clear *sprite.Sprite
	spin  *sprite.Sprite

	bonus      int
	bonusScale *animation.Glider
	bonusFade  *animation.Glider

	rpmBg *sprite.Sprite

	ScaledWidth  float64
	ScaledHeight float64
	background   *sprite.Sprite
	metre        *sprite.Sprite //nolint:misspell
}

func NewSpinner(data []string) *Spinner {
	spinner := &Spinner{
		HitObject: commonParse(data, 6),
	}

	spinner.EndTime, _ = strconv.ParseFloat(data[5], 64)

	sample, _ := strconv.ParseInt(data[4], 10, 64)

	spinner.sample = int(sample)

	return spinner
}

func NewDummySpinner(startTime, endTime float64) *Spinner {
	return &Spinner{
		HitObject: &HitObject{
			StartTime:   startTime,
			EndTime:     endTime,
			HitObjectID: -1,
		},
	}
}

func (spinner *Spinner) GetPosition() vector.Vector2f {
	return spinner.pos
}

func (spinner *Spinner) SetTiming(timings *Timings, _ int, _ bool) {
	spinner.Timings = timings
}

func (spinner *Spinner) SetDifficulty(diff *difficulty.Difficulty) {
	spinner.diff = diff

	spinner.ScaledHeight = 768
	spinner.ScaledWidth = settings.Graphics.GetAspectRatio() * spinner.ScaledHeight

	spinner.fade = animation.NewGlider(0)
	spinner.fade.AddEvent(spinner.StartTime-diff.TimeFadeIn, spinner.StartTime, 1)
	spinner.fade.AddEvent(spinner.EndTime, spinner.EndTime+difficulty.HitFadeOut, 0)

	spinner.sprites = sprite.NewManager()
	spinner.frontSprites = sprite.NewManager()

	spinner.newStyle = skin.GetTexture("spinner-background") == nil

	if spinner.newStyle {
		spinner.glow = sprite.NewSpriteSingle(skin.GetTexture("spinner-glow"), 0.0, spinner.StartPosRaw.Copy64(), vector.Centre)
		spinner.glow.SetAdditive(true)
		spinner.bottom = sprite.NewSpriteSingle(skin.GetTexture("spinner-bottom"), 1.0, spinner.StartPosRaw.Copy64(), vector.Centre)
		spinner.top = sprite.NewSpriteSingle(skin.GetTexture("spinner-top"), 2.0, spinner.StartPosRaw.Copy64(), vector.Centre)
		spinner.middle2 = sprite.NewSpriteSingle(skin.GetTexture("spinner-middle2"), 3.0, spinner.StartPosRaw.Copy64(), vector.Centre)
		spinner.middle = sprite.NewSpriteSingle(skin.GetTexture("spinner-middle"), 4.0, spinner.StartPosRaw.Copy64(), vector.Centre)

		spinner.sprites.Add(spinner.glow)
		spinner.sprites.Add(spinner.bottom)
		spinner.sprites.Add(spinner.top)
		spinner.sprites.Add(spinner.middle2)
		spinner.sprites.Add(spinner.middle)

		spinner.glow.SetColor(spinnerBlue)
		spinner.glow.SetAlpha(0.0)
		spinner.middle.AddTransform(animation.NewColorTransform(animation.Color3, easing.Linear, spinner.StartTime, spinner.EndTime, color2.Color{R: 1, G: 1, B: 1, A: 1}, spinnerRed))
		spinner.middle.ResetValuesToTransforms()
	} else {
		spinner.background = sprite.NewSpriteSingle(skin.GetTexture("spinner-background"), 0.0, vector.NewVec2d(spinner.ScaledWidth/2, 46.5+350.4), vector.Centre)

		sMetre := skin.GetTexture("spinner-metre")

		metreHeight := 0.0
		if sMetre != nil {
			metreHeight = float64(sMetre.Height)
		}

		spinner.metre = sprite.NewSpriteSingle(sMetre, 2.0, vector.NewVec2d(spinner.ScaledWidth/2-512, 47.5+metreHeight), vector.BottomLeft) //nolint:misspell
		spinner.metre.SetCutOrigin(vector.BottomCentre)

		spinner.middle2 = sprite.NewSpriteSingle(skin.GetTexture("spinner-circle"), 1.0, spinner.StartPosRaw.Copy64(), vector.Centre)

		spinner.sprites.Add(spinner.middle2)
	}

	if !diff.CheckModActive(difficulty.Hidden) {
		spinner.approach = sprite.NewSpriteSingle(skin.GetTexture("spinner-approachcircle"), 5.0, spinner.StartPosRaw.Copy64(), vector.Centre)
		spinner.sprites.Add(spinner.approach)
		spinner.approach.AddTransform(animation.NewSingleTransform(animation.Scale, easing.Linear, spinner.StartTime, spinner.EndTime, 1.9, 0.1))
		spinner.approach.ResetValuesToTransforms()
	}

	spinner.UpdateCompletion(0.0)

	spinner.clear = sprite.NewSpriteSingle(skin.GetTexture("spinner-clear"), 10.0, vector.NewVec2d(spinner.ScaledWidth/2 /*46.5+240*/, 256-16-8), vector.Centre)
	spinner.clear.SetAlpha(0.0)

	spinner.frontSprites.Add(spinner.clear)

	spinner.spin = sprite.NewSpriteSingle(skin.GetTexture("spinner-spin"), 10.0, vector.NewVec2d(spinner.ScaledWidth/2 /*46.5+536*/, 608-12.8-16), vector.Centre)

	spinner.frontSprites.Add(spinner.spin)

	spinner.spinnerbonus = audio.LoadSample("spinnerbonus")
	spinner.bonusFade = animation.NewGlider(0.0)
	spinner.bonusScale = animation.NewGlider(0.0)

	spinner.rpmBg = sprite.NewSpriteSingle(skin.GetTexture("spinner-rpm"), 0.0, vector.NewVec2d(spinner.ScaledWidth/2-139, spinner.ScaledHeight-56), vector.TopLeft)

	skin.GetFont("score")
}

func (spinner *Spinner) Update(time float64) bool {
	spinner.fade.Update(time)

	if time >= spinner.StartTime && time <= spinner.EndTime {
		if (!settings.PLAY && !settings.KNOCKOUT) || settings.PLAYERS > 1 {
			rRPMS := rpms * mutils.Clamp(float32(time-spinner.StartTime)/500, 0.0, 1.0)

			spinner.rad = rRPMS * float32(time-spinner.StartTime) * 2 * math32.Pi

			spinner.rpm = float64(rRPMS) * 1000 * 60

			spinner.SetRotation(float64(spinner.rad))

			spinner.UpdateCompletion((time - spinner.StartTime) / (spinner.EndTime - spinner.StartTime))

			if spinner.lastTime < spinner.StartTime {
				spinner.StartSpinSample()
			}
		}
	}

	spinner.pos = spinner.StartPosRaw

	if spinner.lastTime < spinner.EndTime && time >= spinner.EndTime {
		if (!settings.PLAY && !settings.KNOCKOUT) || settings.PLAYERS > 1 {
			spinner.StopSpinSample()
			spinner.Clear()
			spinner.Hit(time, true)
		}
	}

	spinner.sprites.Update(time)

	spinner.frontSprites.Update(time)

	spinner.bonusFade.Update(time)
	spinner.bonusScale.Update(time)

	spinner.lastTime = time

	return true
}

func (spinner *Spinner) Draw(time float64, color color2.Color, batch *batch.QuadBatch) bool {
	batch.SetTranslation(vector.NewVec2d(0, 0))

	shiftX := -float32(settings.Playfield.ShiftX*spinner.ScaledHeight) / 480

	// Objects are not aware of their backing camera so we need to apply scaling and shifting here as it only applies to spinners
	shiftY := float32(settings.Playfield.ShiftY*spinner.ScaledHeight) / 480
	if !settings.Playfield.OsuShift {
		shiftY = 8 * float32(spinner.ScaledHeight) / 480
	}

	overScale := (float32(1.0/settings.Playfield.Scale) - 1) / 2
	overdrawX := overScale * float32(spinner.ScaledWidth)
	overdrawY := overScale * float32(spinner.ScaledHeight)

	scaledOrtho := mgl32.Ortho(-overdrawX+shiftX, float32(spinner.ScaledWidth)+overdrawX+shiftX, float32(spinner.ScaledHeight)+overdrawY+shiftY, -overdrawY+shiftY, -1, 1)

	alpha := spinner.fade.GetValue() * float64(color.A)

	batch.SetColor(1, 1, 1, alpha)

	scale := batch.GetScale()
	batch.SetScale(1, 1)

	if !spinner.newStyle {
		oldCamera := batch.Projection

		batch.SetCamera(scaledOrtho)

		spinner.background.Draw(time, batch)

		batch.SetCamera(oldCamera)
	}

	batch.SetScale(384.0/480*0.78, 384.0/480*0.78)

	spinner.sprites.Draw(time, batch)

	if !spinner.newStyle {
		batch.ResetTransform()
		oldCamera := batch.Projection

		batch.SetCamera(scaledOrtho)

		spinner.metre.Draw(time, batch)

		batch.SetCamera(oldCamera)
		batch.SetScale(384.0/480*0.78, 384.0/480*0.78)
	}

	scoreFont := skin.GetFont("score")

	if spinner.bonusFade.GetValue() > 0.01 {
		batch.SetColor(1.0, 1.0, 1.0, spinner.bonusFade.GetValue()*alpha)

		scoreFont.DrawOrigin(batch, 256, 192+80, vector.Centre, spinner.bonusScale.GetValue()*scoreFont.GetSize()*0.8, false, strconv.Itoa(spinner.bonus))
	}

	batch.ResetTransform()
	batch.SetColor(1.0, 1.0, 1.0, alpha)

	oldCamera := batch.Projection

	batch.SetCamera(scaledOrtho)

	spinner.frontSprites.Draw(time, batch)

	batch.SetCamera(mgl32.Ortho(shiftX, float32(spinner.ScaledWidth)+shiftX, float32(spinner.ScaledHeight), 0, -1, 1))

	spinner.rpmBg.Draw(time, batch)

	rpmTxt := fmt.Sprintf("%d", int(spinner.rpm))
	scoreFont.DrawOrigin(batch, spinner.ScaledWidth/2+139, spinner.ScaledHeight-56, vector.TopRight, scoreFont.GetSize(), false, rpmTxt)

	batch.SetCamera(oldCamera)
	batch.ResetTransform()
	batch.SetScale(scale.X, scale.Y)

	if time >= spinner.EndTime && spinner.fade.GetValue() <= 0.01 {
		return true
	}

	return false
}

func (spinner *Spinner) DrawApproach(_ float64, _ color2.Color, _ *batch.QuadBatch) {}

func (spinner *Spinner) Hit(_ float64, isHit bool) {
	if !isHit || spinner.audioSubmissionDisabled {
		return
	}

	point := spinner.Timings.GetPointAt(spinner.EndTime)

	index := spinner.BasicHitSound.CustomIndex
	if index == 0 {
		index = point.SampleIndex
	}

	sampleSet := spinner.BasicHitSound.SampleSet
	if sampleSet == 0 {
		sampleSet = point.SampleSet
	}

	audio.PlaySample(sampleSet, spinner.BasicHitSound.AdditionSet, spinner.sample, index, point.SampleVolume, spinner.HitObjectID, spinner.StartPosRaw.X64())
}

func (spinner *Spinner) SetRotation(f float64) {
	spinner.rad = float32(f)

	if spinner.newStyle {
		spinner.bottom.SetRotation(f / 3)
		spinner.top.SetRotation(f * 0.5)
		spinner.middle2.SetRotation(f)
	} else if spinner.middle2 != nil {
		spinner.middle2.SetRotation(f)
	}
}

func (spinner *Spinner) SetRPM(rpm float64) {
	spinner.rpm = rpm
}

func (spinner *Spinner) UpdateCompletion(completion float64) {
	if completion > 0 && spinner.completion == 0 {
		spinner.spin.AddTransform(animation.NewSingleTransform(animation.Fade, easing.Linear, spinner.lastTime, spinner.lastTime+300, 1.0, 0.0))
	}

	spinner.completion = completion

	if skin.GetInfo().SpinnerFrequencyModulate && spinner.loopSample != nil {
		bass.SetRate(spinner.loopSample, min(100000, 20000+(40000*completion)))
	}

	scale := 0.8 + min(1.0, completion)*0.2

	if spinner.newStyle {
		spinner.glow.SetScale(scale)
		spinner.bottom.SetScale(scale)
		spinner.top.SetScale(scale)
		spinner.middle2.SetScale(scale)
		spinner.middle.SetScale(scale)

		spinner.glow.SetAlpha(min(1.0, float32(completion)))
	} else if spinner.metre != nil {
		bars := int(min(0.99, completion) * 10)

		if skin.GetInfo().SpinnerNoBlink || rand.Float64() < math.Mod(completion*10, 1) {
			bars++
		}

		spinner.metre.SetCutY(1.0-float64(bars)/10, 0)
	}
}

func (spinner *Spinner) StartSpinSample() {
	if spinner.audioSubmissionDisabled {
		return
	}

	if spinner.loopSample == nil {
		sample := audio.LoadSample("spinnerspin")
		if sample != nil {
			spinner.loopSample = sample.PlayLoop()
		}
	} else {
		bass.PlaySample(spinner.loopSample)
	}

	if skin.GetInfo().SpinnerFrequencyModulate && spinner.loopSample != nil {
		bass.SetRate(spinner.loopSample, min(100000, 20000+(40000*spinner.completion)))
	}
}

func (spinner *Spinner) PauseSpinSample() {
	if spinner.audioSubmissionDisabled {
		return
	}

	if spinner.loopSample != nil {
		bass.PauseSample(spinner.loopSample)
	}
}

func (spinner *Spinner) StopSpinSample() {
	if spinner.audioSubmissionDisabled {
		return
	}

	if spinner.loopSample != nil {
		bass.StopSample(spinner.loopSample)
		spinner.loopSample = nil
	}
}

func (spinner *Spinner) Clear() {
	spinner.clear.AddTransform(animation.NewSingleTransform(animation.Scale, easing.OutBack, spinner.lastTime, spinner.lastTime+spinner.diff.TimeFadeIn, 2.0, 1.0))
	spinner.clear.AddTransform(animation.NewSingleTransform(animation.Fade, easing.OutQuad, spinner.lastTime, spinner.lastTime+spinner.diff.TimeFadeIn, 0.0, 1.0))
}

func (spinner *Spinner) Bonus(bonusValue int) {
	if spinner.glow != nil {
		spinner.glow.AddTransform(animation.NewColorTransform(animation.Color3, easing.OutQuad, spinner.lastTime, spinner.lastTime+difficulty.HitFadeOut, color2.Color{R: 1, G: 1, B: 1, A: 1}, spinnerBlue))
	}

	if spinner.spinnerbonus != nil && !spinner.audioSubmissionDisabled {
		spinner.spinnerbonus.Play()
	}

	spinner.bonusFade.Reset()
	spinner.bonusFade.AddEventSEase(spinner.lastTime, spinner.lastTime+800, 1.0, 0.0, easing.OutQuad)

	spinner.bonusScale.Reset()
	spinner.bonusScale.AddEventSEase(spinner.lastTime, spinner.lastTime+800, 2.0, 1.28, easing.OutQuad)

	spinner.bonus += bonusValue
}

func (spinner *Spinner) GetType() Type {
	return SPINNER
}
