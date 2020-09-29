package objects

import (
	"fmt"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/wieku/danser-go/app/audio"
	"github.com/wieku/danser-go/app/bmath"
	"github.com/wieku/danser-go/app/bmath/difficulty"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/app/skin"
	"github.com/wieku/danser-go/framework/bass"
	"github.com/wieku/danser-go/framework/graphics/sprite"
	"github.com/wieku/danser-go/framework/math/animation"
	"github.com/wieku/danser-go/framework/math/animation/easing"
	"github.com/wieku/danser-go/framework/math/math32"
	"github.com/wieku/danser-go/framework/math/vector"
	"math"
	"strconv"
)

const rpms = 0.00795

var spinnerRed = bmath.Color{R: 1, G: 0, B: 0, A: 1}
var spinnerBlue = bmath.Color{R: 0.05, G: 0.5, B: 1.0, A: 1}

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

	newStyle bool
	sprites  *sprite.SpriteManager

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

func (self *Spinner) GetBasicData() *basicData {
	return self.objData
}

func (self *Spinner) GetPosition() vector.Vector2f {
	return self.pos
}

func (self *Spinner) SetTiming(timings *Timings) {
	self.Timings = timings
}

func (self *Spinner) UpdateStacking() {}

func (self *Spinner) SetDifficulty(diff *difficulty.Difficulty) {
	self.ScaledHeight = 768
	self.ScaledWidth = settings.Graphics.GetAspectRatio() * self.ScaledHeight

	self.fade = animation.NewGlider(0)
	self.fade.AddEvent(float64(self.objData.StartTime)-difficulty.HitFadeIn, float64(self.objData.StartTime), 1)
	self.fade.AddEvent(float64(self.objData.EndTime), float64(self.objData.EndTime)+difficulty.HitFadeOut, 0)

	self.sprites = sprite.NewSpriteManager()

	self.newStyle = skin.GetTexture("spinner-background") == nil

	if self.newStyle {
		self.glow = sprite.NewSpriteSingle(skin.GetTexture("spinner-glow"), 0.0, self.objData.StartPos.Copy64(), bmath.Origin.Centre)
		self.glow.SetAdditive(true)
		self.bottom = sprite.NewSpriteSingle(skin.GetTexture("spinner-bottom"), 1.0, self.objData.StartPos.Copy64(), bmath.Origin.Centre)
		self.top = sprite.NewSpriteSingle(skin.GetTexture("spinner-top"), 2.0, self.objData.StartPos.Copy64(), bmath.Origin.Centre)
		self.middle2 = sprite.NewSpriteSingle(skin.GetTexture("spinner-middle2"), 3.0, self.objData.StartPos.Copy64(), bmath.Origin.Centre)
		self.middle = sprite.NewSpriteSingle(skin.GetTexture("spinner-middle"), 4.0, self.objData.StartPos.Copy64(), bmath.Origin.Centre)

		self.sprites.Add(self.glow)
		self.sprites.Add(self.bottom)
		self.sprites.Add(self.top)
		self.sprites.Add(self.middle2)
		self.sprites.Add(self.middle)

		self.glow.SetColor(spinnerBlue)
		self.glow.SetAlpha(0.0)
		self.middle.AddTransform(animation.NewColorTransform(animation.Color3, easing.Linear, float64(self.objData.StartTime), float64(self.objData.EndTime), bmath.Color{R: 1, G: 1, B: 1, A: 1}, spinnerRed))
		self.middle.ResetValuesToTransforms()

	} else {
		self.background = sprite.NewSpriteSingle(skin.GetTexture("spinner-background"), 0.0, vector.NewVec2d(self.ScaledWidth/2-512, 46-24), bmath.Origin.TopLeft)
		self.metre = sprite.NewSpriteSingle(skin.GetTexture("spinner-metre"), 1.0, vector.NewVec2d(self.ScaledWidth/2-512, 46), bmath.Origin.TopLeft)
		self.metre.SetCutOrigin(bmath.Origin.BottomCentre)

		self.middle2 = sprite.NewSpriteSingle(skin.GetTexture("spinner-circle"), 2.0, self.objData.StartPos.Copy64(), bmath.Origin.Centre)

		self.sprites.Add(self.middle2)
	}

	self.approach = sprite.NewSpriteSingle(skin.GetTexture("spinner-approachcircle"), 5.0, self.objData.StartPos.Copy64(), bmath.Origin.Centre)
	self.sprites.Add(self.approach)
	self.approach.AddTransform(animation.NewSingleTransform(animation.Scale, easing.Linear, float64(self.objData.StartTime), float64(self.objData.EndTime), 1.9, 0.1))
	self.approach.ResetValuesToTransforms()

	self.UpdateCompletion(0.0)

	self.clear = sprite.NewSpriteSingle(skin.GetTexture("spinner-clear"), 10.0, self.objData.StartPos.Copy64().SubS(0, 110), bmath.Origin.Centre)
	self.clear.SetAlpha(0.0)

	self.sprites.Add(self.clear)

	self.spin = sprite.NewSpriteSingle(skin.GetTexture("spinner-spin"), 10.0, self.objData.StartPos.Copy64().AddS(0, 110), bmath.Origin.Centre)

	self.sprites.Add(self.spin)

	self.spinnerbonus = audio.LoadSample("assets/sounds/spinnerbonus")
	self.bonusFade = animation.NewGlider(0.0)
	self.bonusScale = animation.NewGlider(0.0)

	self.rpmBg = sprite.NewSpriteSingle(skin.GetTexture("spinner-rpm"), 0.0, vector.NewVec2d(self.ScaledWidth/2-139, self.ScaledHeight-56), bmath.Origin.TopLeft)
}

func (self *Spinner) Update(time int64) bool {
	self.fade.Update(float64(time))

	if time >= self.objData.StartTime && time <= self.objData.EndTime {
		if (!settings.PLAY && !settings.KNOCKOUT) || settings.PLAYERS > 1 {
			self.rad = rpms * float32(time-self.objData.StartTime) * 2 * math32.Pi
		}

		//frad := float32(easing.InQuad(float64(1.0 - math32.Abs((math32.Mod(self.rad, math32.Pi/2)-math32.Pi/4)/(math32.Pi/4)))))
		a := self.rad - math32.Pi/2*math32.Round(self.rad*2/math32.Pi)
		self.pos = vector.NewVec2fRad(self.rad*1.1, 100/math32.Cos(a) /*+ frad * (50 * (math32.Sqrt(2) - 1))*/).Add(self.objData.StartPos)

		//self.pos.X = 16 * math32.Pow(math32.Sin(self.rad), 3)
		//self.pos.Y = 13*math32.Cos(self.rad) - 5*math32.Cos(2*self.rad) - 2*math32.Cos(3*self.rad) - math32.Cos(4*self.rad)

		//self.pos = self.pos.Scl(-6 - 2*math32.Sin(float32(time-self.objData.StartTime)/2000*2*math32.Pi)).Add(self.objData.StartPos)
		self.GetBasicData().EndPos = self.pos
	}

	if self.lastTime < self.objData.EndTime && time >= self.objData.EndTime {
		index := self.objData.customIndex

		point := self.Timings.GetPoint(self.objData.EndTime)

		if index == 0 {
			index = point.SampleIndex
		}

		if self.objData.sampleSet == 0 {
			audio.PlaySample(point.SampleSet, self.objData.additionSet, self.sample, index, point.SampleVolume, self.objData.Number, self.GetBasicData().StartPos.X64())
		} else {
			audio.PlaySample(self.objData.sampleSet, self.objData.additionSet, self.sample, index, point.SampleVolume, self.objData.Number, self.GetBasicData().StartPos.X64())
		}
	}

	self.sprites.Update(time)

	self.rpmBg.Update(time)

	self.bonusFade.Update(float64(time))
	self.bonusScale.Update(float64(time))

	self.lastTime = time

	return true
}

func (self *Spinner) Draw(time int64, color mgl32.Vec4, batch *sprite.SpriteBatch) bool {
	batch.SetTranslation(vector.NewVec2d(0, 0))
	//percent := bmath.ClampF64(float64(time-self.objData.StartTime)/float64(self.objData.EndTime-self.objData.StartTime), 0.0, 1.0)

	alpha := self.fade.GetValue()

	if settings.DIVIDES >= settings.Objects.MandalaTexturesTrigger {
		alpha *= settings.Objects.MandalaTexturesAlpha
	}

	batch.SetColor(1, 1, 1, alpha)

	scale := batch.GetScale()
	batch.SetScale(1, 1)

	if !self.newStyle {
		oldCamera := batch.Projection
		batch.SetCamera(mgl32.Ortho(0, float32(self.ScaledWidth), float32(self.ScaledHeight), 0, -1, 1))

		self.background.Draw(time, batch)
		self.metre.Draw(time, batch)

		batch.SetCamera(oldCamera)
	}

	batch.SetScale(384.0/480*0.8, 384.0/480*0.8)

	self.sprites.Draw(time, batch)

	scoreFont := skin.GetFont("score")

	if self.bonusFade.GetValue() > 0.01 {
		batch.SetColor(1.0, 1.0, 1.0, self.bonusFade.GetValue())

		scoreFont.DrawCentered(batch, 256, 192+100, self.bonusScale.GetValue()*scoreFont.GetSize(), strconv.Itoa(self.bonus))
	}

	batch.ResetTransform()
	batch.SetColor(1.0, 1.0, 1.0, alpha)

	oldCamera := batch.Projection

	batch.SetCamera(mgl32.Ortho(0, float32(self.ScaledWidth), float32(self.ScaledHeight), 0, -1, 1))

	self.rpmBg.Draw(time, batch)

	rpmTxt := fmt.Sprintf("%d", int(self.rpm))
	scoreFont.Draw(batch, self.ScaledWidth/2+139-scoreFont.GetWidth(scoreFont.GetSize(), rpmTxt), self.ScaledHeight-56+scoreFont.GetSize()/2, scoreFont.GetSize(), rpmTxt)

	batch.SetCamera(oldCamera)
	batch.ResetTransform()
	batch.SetScale(scale.X, scale.Y)

	if time >= self.objData.EndTime && self.fade.GetValue() <= 0.01 {
		return true
	}

	return false
}

func (self *Spinner) DrawApproach(time int64, color mgl32.Vec4, batch *sprite.SpriteBatch) {}

func (self *Spinner) SetRotation(f float64) {
	self.rad = float32(f)

	if self.newStyle {
		self.bottom.SetRotation(f / 3)
		self.top.SetRotation(f * 0.5)
		self.middle2.SetRotation(f)
	} else if self.middle2 != nil {
		self.middle2.SetRotation(f)
	}
}

func (self *Spinner) SetRPM(rpm float64) {
	self.rpm = rpm
}

func (self *Spinner) UpdateCompletion(completion float64) {
	if completion > 0 && self.completion == 0 {
		self.spin.AddTransform(animation.NewSingleTransform(animation.Fade, easing.Linear, float64(self.lastTime), float64(self.lastTime+300), 1.0, 0.0))
	}

	self.completion = completion

	bass.SetRate(self.loopSample, math.Min(100000, 20000+(40000*completion)))

	scale := 0.8 + math.Min(1.0, completion)*0.2

	if self.newStyle {
		self.glow.SetScale(scale)
		self.bottom.SetScale(scale)
		self.top.SetScale(scale)
		self.middle2.SetScale(scale)
		self.middle.SetScale(scale)

		self.glow.SetAlpha(math.Min(1.0, completion))
	} else if self.metre != nil {
		self.metre.SetCutY(1.0 - math.Floor(math.Min(1.0, completion)*10)/10)
	}
}

func (self *Spinner) StartSpinSample() {
	if self.loopSample == 0 {
		self.loopSample = audio.LoadSampleLoop("assets/sounds/spinnerspin").Play()
	} else {
		bass.PlaySample(self.loopSample)
	}

	bass.SetRate(self.loopSample, math.Min(100000, 20000+(40000*self.completion)))
}

func (self *Spinner) StopSpinSample() {
	bass.PauseSample(self.loopSample)
}

func (self *Spinner) Clear() {
	self.clear.AddTransform(animation.NewSingleTransform(animation.Scale, easing.OutBack, float64(self.lastTime), float64(self.lastTime+difficulty.HitFadeIn), 2.0, 1.0))
	self.clear.AddTransform(animation.NewSingleTransform(animation.Fade, easing.OutQuad, float64(self.lastTime), float64(self.lastTime+difficulty.HitFadeIn), 0.0, 1.0))
}

func (self *Spinner) Bonus() {
	if self.glow != nil {
		self.glow.AddTransform(animation.NewColorTransform(animation.Color3, easing.OutQuad, float64(self.lastTime), float64(self.lastTime+difficulty.HitFadeOut), bmath.Color{R: 1, G: 1, B: 1, A: 1}, spinnerBlue))
	}

	self.spinnerbonus.Play()

	self.bonusFade.Reset()
	self.bonusFade.AddEventS(float64(self.lastTime), float64(self.lastTime+800), 1.0, 0.0)

	self.bonusScale.Reset()
	self.bonusScale.AddEventS(float64(self.lastTime), float64(self.lastTime+800), 2.0, 1.0)

	self.bonus += 1000
}
