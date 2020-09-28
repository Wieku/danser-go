package objects

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/wieku/danser-go/app/audio"
	"github.com/wieku/danser-go/app/bmath/difficulty"
	"github.com/wieku/danser-go/app/graphics"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/framework/graphics/sprite"
	"github.com/wieku/danser-go/framework/math/animation"
	"github.com/wieku/danser-go/framework/math/math32"
	"github.com/wieku/danser-go/framework/math/vector"
	"math"
	"strconv"
)

const rpms = 0.00795

type Spinner struct {
	objData  *basicData
	Timings  *Timings
	sample   int
	rad      float32
	pos      vector.Vector2f
	fade     *animation.Glider
	lastTime int64
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
	self.fade = animation.NewGlider(0)
	self.fade.AddEvent(float64(self.objData.StartTime)-diff.Preempt, float64(self.objData.StartTime)-(diff.Preempt-difficulty.HitFadeIn), 1)
	self.fade.AddEvent(float64(self.objData.EndTime), float64(self.objData.EndTime)+difficulty.HitFadeOut, 0)
}

func (self *Spinner) Update(time int64) bool {
	self.fade.Update(float64(time))

	if time >= self.objData.StartTime && time <= self.objData.EndTime {
		self.rad = rpms * float32(time-self.objData.StartTime) * 2 * math32.Pi

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

	self.lastTime = time

	return true
}

func (self *Spinner) Draw(time int64, color mgl32.Vec4, batch *sprite.SpriteBatch) bool {

	batch.SetTranslation(self.objData.StartPos.Copy64())

	alpha := self.fade.GetValue()

	if settings.DIVIDES >= settings.Objects.MandalaTexturesTrigger {
		alpha *= settings.Objects.MandalaTexturesAlpha
	}

	batch.SetColor(1, 1, 1, alpha)
	scale := batch.GetScale()
	batch.SetScale(1, 1)

	batch.SetRotation(float64(self.rad))
	batch.SetSubScale(20*10, 20*10)

	batch.DrawUnit(*graphics.SpinnerMiddle)
	batch.DrawUnit(*graphics.SpinnerMiddle2)

	scl := 16 + math.Min(220, math.Max(0, (1.0-float64(time-self.objData.StartTime)/float64(self.objData.EndTime-self.objData.StartTime))*220))

	batch.SetSubScale(scl, scl)

	batch.DrawUnit(*graphics.SpinnerAC)

	batch.SetSubScale(1, 1)
	batch.SetRotation(0)
	batch.SetTranslation(vector.NewVec2d(0, 0))
	batch.SetScale(scale.X, scale.Y)

	if time >= self.objData.EndTime && self.fade.GetValue() <= 0.01 {
		return true
	}

	return false
}

func (self *Spinner) DrawApproach(time int64, color mgl32.Vec4, batch *sprite.SpriteBatch) {}
