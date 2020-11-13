package objects

import (
	"fmt"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/wieku/danser-go/app/audio"
	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/bmath"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/app/skin"
	"github.com/wieku/danser-go/framework/bass"
	"github.com/wieku/danser-go/framework/graphics/batch"
	"github.com/wieku/danser-go/framework/graphics/sprite"
	"github.com/wieku/danser-go/framework/math/animation"
	"github.com/wieku/danser-go/framework/math/animation/easing"
	color2 "github.com/wieku/danser-go/framework/math/color"
	"github.com/wieku/danser-go/framework/math/math32"
	"github.com/wieku/danser-go/framework/math/vector"
	"math"
	"math/rand"
	"strconv"
)

const rpms = 0.00795

var spinnerRed = color2.Color{R: 1, G: 0, B: 0, A: 1}
var spinnerBlue = color2.Color{R: 0.05, G: 0.5, B: 1.0, A: 1}

type Spinner struct {
	objData  *basicData
	Timings  *Timings
	sample   int
	rad      float32
	pos      vector.Vector2f
	fade     *animation.Glider
	lastTime int64
	rpm      float64

	spinnerbonus *bass.Sample
	loopSample   bass.SubSample
	completion   float64

	newStyle     bool
	sprites      *sprite.SpriteManager
	frontSprites *sprite.SpriteManager

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
	metre        *sprite.Sprite
}

func NewSpinner(data []string) *Spinner {
	spinner := &Spinner{}
	spinner.objData = commonParse(data)
	spinner.objData.parseExtras(data, 6)

	spinner.objData.EndTime, _ = strconv.ParseInt(data[5], 10, 64)

	sample, _ := strconv.ParseInt(data[4], 10, 64)

	spinner.sample = int(sample)

	spinner.objData.EndPos = spinner.objData.StartPos
	return spinner
}

func (spinner *Spinner) GetBasicData() *basicData {
	return spinner.objData
}

func (spinner *Spinner) GetPosition() vector.Vector2f {
	return spinner.pos
}

func (spinner *Spinner) SetTiming(timings *Timings) {
	spinner.Timings = timings
}

func (spinner *Spinner) UpdateStacking() {}

func (spinner *Spinner) SetDifficulty(diff *difficulty.Difficulty) {
	spinner.ScaledHeight = 768
	spinner.ScaledWidth = settings.Graphics.GetAspectRatio() * spinner.ScaledHeight

	spinner.fade = animation.NewGlider(0)
	spinner.fade.AddEvent(float64(spinner.objData.StartTime)-difficulty.HitFadeIn, float64(spinner.objData.StartTime), 1)
	spinner.fade.AddEvent(float64(spinner.objData.EndTime), float64(spinner.objData.EndTime)+difficulty.HitFadeOut, 0)

	spinner.sprites = sprite.NewSpriteManager()
	spinner.frontSprites = sprite.NewSpriteManager()

	spinner.newStyle = skin.GetTexture("spinner-background") == nil

	if spinner.newStyle {
		spinner.glow = sprite.NewSpriteSingle(skin.GetTexture("spinner-glow"), 0.0, spinner.objData.StartPos.Copy64(), bmath.Origin.Centre)
		spinner.glow.SetAdditive(true)
		spinner.bottom = sprite.NewSpriteSingle(skin.GetTexture("spinner-bottom"), 1.0, spinner.objData.StartPos.Copy64(), bmath.Origin.Centre)
		spinner.top = sprite.NewSpriteSingle(skin.GetTexture("spinner-top"), 2.0, spinner.objData.StartPos.Copy64(), bmath.Origin.Centre)
		spinner.middle2 = sprite.NewSpriteSingle(skin.GetTexture("spinner-middle2"), 3.0, spinner.objData.StartPos.Copy64(), bmath.Origin.Centre)
		spinner.middle = sprite.NewSpriteSingle(skin.GetTexture("spinner-middle"), 4.0, spinner.objData.StartPos.Copy64(), bmath.Origin.Centre)

		spinner.sprites.Add(spinner.glow)
		spinner.sprites.Add(spinner.bottom)
		spinner.sprites.Add(spinner.top)
		spinner.sprites.Add(spinner.middle2)
		spinner.sprites.Add(spinner.middle)

		spinner.glow.SetColor(spinnerBlue)
		spinner.glow.SetAlpha(0.0)
		spinner.middle.AddTransform(animation.NewColorTransform(animation.Color3, easing.Linear, float64(spinner.objData.StartTime), float64(spinner.objData.EndTime), color2.Color{R: 1, G: 1, B: 1, A: 1}, spinnerRed))
		spinner.middle.ResetValuesToTransforms()

	} else {
		spinner.background = sprite.NewSpriteSingle(skin.GetTexture("spinner-background"), 0.0, vector.NewVec2d(spinner.ScaledWidth/2, 46.5+350.4), bmath.Origin.Centre)
		spinner.metre = sprite.NewSpriteSingle(skin.GetTexture("spinner-metre"), 1.0, vector.NewVec2d(spinner.ScaledWidth/2-512, 46.5), bmath.Origin.TopLeft)
		spinner.metre.SetCutOrigin(bmath.Origin.BottomCentre)

		spinner.middle2 = sprite.NewSpriteSingle(skin.GetTexture("spinner-circle"), 2.0, spinner.objData.StartPos.Copy64(), bmath.Origin.Centre)

		spinner.sprites.Add(spinner.middle2)
	}

	spinner.approach = sprite.NewSpriteSingle(skin.GetTexture("spinner-approachcircle"), 5.0, spinner.objData.StartPos.Copy64(), bmath.Origin.Centre)
	spinner.sprites.Add(spinner.approach)
	spinner.approach.AddTransform(animation.NewSingleTransform(animation.Scale, easing.Linear, float64(spinner.objData.StartTime), float64(spinner.objData.EndTime), 1.9, 0.1))
	spinner.approach.ResetValuesToTransforms()

	spinner.UpdateCompletion(0.0)

	spinner.clear = sprite.NewSpriteSingle(skin.GetTexture("spinner-clear"), 10.0, vector.NewVec2d(spinner.ScaledWidth/2, 46.5+240), bmath.Origin.Centre)
	spinner.clear.SetAlpha(0.0)

	spinner.frontSprites.Add(spinner.clear)

	spinner.spin = sprite.NewSpriteSingle(skin.GetTexture("spinner-spin"), 10.0, vector.NewVec2d(spinner.ScaledWidth/2, 46.5+536), bmath.Origin.Centre)

	spinner.frontSprites.Add(spinner.spin)

	spinner.spinnerbonus = audio.LoadSample("spinnerbonus")
	spinner.bonusFade = animation.NewGlider(0.0)
	spinner.bonusScale = animation.NewGlider(0.0)

	spinner.rpmBg = sprite.NewSpriteSingle(skin.GetTexture("spinner-rpm"), 0.0, vector.NewVec2d(spinner.ScaledWidth/2-139, spinner.ScaledHeight-56), bmath.Origin.TopLeft)

	//spinner.frontSprites.Add(spinner.rpmBg)
}

func (spinner *Spinner) Update(time int64) bool {
	spinner.fade.Update(float64(time))

	if time >= spinner.objData.StartTime && time <= spinner.objData.EndTime {
		if (!settings.PLAY && !settings.KNOCKOUT) || settings.PLAYERS > 1 {

			rRPMS := rpms * bmath.ClampF32(float32(time-spinner.objData.StartTime)/500, 0.0, 1.0)

			spinner.rad = rRPMS * float32(time-spinner.objData.StartTime) * 2 * math32.Pi

			spinner.rpm = float64(rRPMS) * 1000 * 60

			spinner.SetRotation(float64(spinner.rad))

			spinner.UpdateCompletion(float64(time-spinner.objData.StartTime) / float64(spinner.objData.EndTime-spinner.objData.StartTime))

			if spinner.lastTime < spinner.objData.StartTime {
				spinner.StartSpinSample()
			}
		}

		//frad := float32(easing.InQuad(float64(1.0 - math32.Abs((math32.Mod(spinner.rad, math32.Pi/2)-math32.Pi/4)/(math32.Pi/4)))))
		//a := spinner.rad - math32.Pi/2*math32.Round(spinner.rad*2/math32.Pi)
		//spinner.pos = vector.NewVec2fRad(spinner.rad*1.1, 100/math32.Cos(a) /*+ frad * (50 * (math32.Sqrt(2) - 1))*/).Add(spinner.objData.StartPos)

		//spinner.pos.X = 16 * math32.Pow(math32.Sin(spinner.rad), 3)
		//spinner.pos.Y = 13*math32.Cos(spinner.rad) - 5*math32.Cos(2*spinner.rad) - 2*math32.Cos(3*spinner.rad) - math32.Cos(4*spinner.rad)

		//spinner.pos = spinner.pos.Scl(-6 - 2*math32.Sin(float32(time-spinner.objData.StartTime)/2000*2*math32.Pi)).Add(spinner.objData.StartPos)
		//spinner.GetBasicData().EndPos = spinner.pos
	}

	spinner.pos = spinner.objData.StartPos

	if spinner.lastTime < spinner.objData.EndTime && time >= spinner.objData.EndTime {
		if (!settings.PLAY && !settings.KNOCKOUT) || settings.PLAYERS > 1 {
			spinner.StopSpinSample()
			spinner.Clear()
			spinner.Hit(time, true)
		}
	}

	spinner.sprites.Update(time)

	spinner.frontSprites.Update(time)

	spinner.bonusFade.Update(float64(time))
	spinner.bonusScale.Update(float64(time))

	spinner.lastTime = time

	return true
}

func (spinner *Spinner) Draw(time int64, color color2.Color, batch *batch.QuadBatch) bool {
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

	alpha := spinner.fade.GetValue()

	batch.SetColor(1, 1, 1, alpha)

	scale := batch.GetScale()
	batch.SetScale(1, 1)

	if !spinner.newStyle {
		oldCamera := batch.Projection

		batch.SetCamera(scaledOrtho)

		spinner.background.Draw(time, batch)
		spinner.metre.Draw(time, batch)

		batch.SetCamera(oldCamera)
	}

	batch.SetScale(384.0/480*0.8, 384.0/480*0.8)

	spinner.sprites.Draw(time, batch)

	scoreFont := skin.GetFont("score")

	if spinner.bonusFade.GetValue() > 0.01 {
		batch.SetColor(1.0, 1.0, 1.0, spinner.bonusFade.GetValue())

		scoreFont.DrawCentered(batch, 256, 192+100, spinner.bonusScale.GetValue()*scoreFont.GetSize(), strconv.Itoa(spinner.bonus))
	}

	batch.ResetTransform()
	batch.SetColor(1.0, 1.0, 1.0, alpha)

	oldCamera := batch.Projection

	batch.SetCamera(scaledOrtho)

	spinner.frontSprites.Draw(time, batch)

	batch.SetCamera(mgl32.Ortho(shiftX, float32(spinner.ScaledWidth)+shiftX, float32(spinner.ScaledHeight), 0, -1, 1))

	spinner.rpmBg.Draw(time, batch)

	rpmTxt := fmt.Sprintf("%d", int(spinner.rpm))
	scoreFont.Draw(batch, spinner.ScaledWidth/2+139-scoreFont.GetWidth(scoreFont.GetSize(), rpmTxt), spinner.ScaledHeight-56+scoreFont.GetSize()/2, scoreFont.GetSize(), rpmTxt)

	batch.SetCamera(oldCamera)
	batch.ResetTransform()
	batch.SetScale(scale.X, scale.Y)

	if time >= spinner.objData.EndTime && spinner.fade.GetValue() <= 0.01 {
		return true
	}

	return false
}

func (spinner *Spinner) DrawApproach(time int64, color color2.Color, batch *batch.QuadBatch) {}

func (spinner *Spinner) Hit(_ int64, isHit bool) {
	if !isHit {
		return
	}

	point := spinner.Timings.GetPoint(spinner.objData.EndTime)

	index := spinner.objData.customIndex
	if index == 0 {
		index = point.SampleIndex
	}

	sampleSet := spinner.objData.sampleSet
	if sampleSet == 0 {
		sampleSet = point.SampleSet
	}

	audio.PlaySample(sampleSet, spinner.objData.additionSet, spinner.sample, index, point.SampleVolume, spinner.objData.Number, spinner.GetBasicData().StartPos.X64())
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
		spinner.spin.AddTransform(animation.NewSingleTransform(animation.Fade, easing.Linear, float64(spinner.lastTime), float64(spinner.lastTime+300), 1.0, 0.0))
	}

	spinner.completion = completion

	if skin.GetInfo().SpinnerFrequencyModulate {
		bass.SetRate(spinner.loopSample, math.Min(100000, 20000+(40000*completion)))
	}

	scale := 0.8 + math.Min(1.0, completion)*0.2

	if spinner.newStyle {
		spinner.glow.SetScale(scale)
		spinner.bottom.SetScale(scale)
		spinner.top.SetScale(scale)
		spinner.middle2.SetScale(scale)
		spinner.middle.SetScale(scale)

		spinner.glow.SetAlpha(math32.Min(1.0, float32(completion)))
	} else if spinner.metre != nil {
		bars := int(math.Min(0.99, completion) * 10)

		if skin.GetInfo().SpinnerNoBlink || rand.Float64() < math.Mod(completion*10, 1) {
			bars++
		}

		spinner.metre.SetCutY(1.0 - float64(bars)/10)
	}
}

func (spinner *Spinner) StartSpinSample() {
	if spinner.loopSample == 0 {
		spinner.loopSample = audio.LoadSample("spinnerspin").PlayLoop()
	} else {
		bass.PlaySample(spinner.loopSample)
	}

	bass.SetRate(spinner.loopSample, math.Min(100000, 20000+(40000*spinner.completion)))
}

func (spinner *Spinner) StopSpinSample() {
	bass.PauseSample(spinner.loopSample)
}

func (spinner *Spinner) Clear() {
	spinner.clear.AddTransform(animation.NewSingleTransform(animation.Scale, easing.OutBack, float64(spinner.lastTime), float64(spinner.lastTime+difficulty.HitFadeIn), 2.0, 1.0))
	spinner.clear.AddTransform(animation.NewSingleTransform(animation.Fade, easing.OutQuad, float64(spinner.lastTime), float64(spinner.lastTime+difficulty.HitFadeIn), 0.0, 1.0))
}

func (spinner *Spinner) Bonus() {
	if spinner.glow != nil {
		spinner.glow.AddTransform(animation.NewColorTransform(animation.Color3, easing.OutQuad, float64(spinner.lastTime), float64(spinner.lastTime+difficulty.HitFadeOut), color2.Color{R: 1, G: 1, B: 1, A: 1}, spinnerBlue))
	}

	spinner.spinnerbonus.Play()

	spinner.bonusFade.Reset()
	spinner.bonusFade.AddEventS(float64(spinner.lastTime), float64(spinner.lastTime+800), 1.0, 0.0)

	spinner.bonusScale.Reset()
	spinner.bonusScale.AddEventS(float64(spinner.lastTime), float64(spinner.lastTime+800), 2.0, 1.0)

	spinner.bonus += 1000
}
