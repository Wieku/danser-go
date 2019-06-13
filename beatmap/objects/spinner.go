package objects

import (
	"strconv"
	"math"
	"github.com/wieku/danser-go/bmath"
	"github.com/wieku/danser-go/audio"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/wieku/danser-go/settings"
	"github.com/wieku/danser-go/render/batches"
	"github.com/wieku/danser-go/render"
	"github.com/wieku/danser-go/animation"
)

const rpms = 0.00795

type Spinner struct {
	objData *basicData
	Timings *Timings
	sample  int
	rad     float64
	pos bmath.Vector2d
	fade *animation.Glider
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

func (self *Spinner) GetPosition() bmath.Vector2d {
	return self.pos
}

func (self *Spinner) SetTiming(timings *Timings) {
	self.Timings = timings
}


func (self *Spinner) SetDifficulty(preempt, fadeIn float64) {
	self.fade = animation.NewGlider(0)
	self.fade.AddEvent(float64(self.objData.StartTime)-preempt, float64(self.objData.StartTime)-(preempt-fadeIn), 1)
	self.fade.AddEvent(float64(self.objData.EndTime), float64(self.objData.EndTime)+fadeIn, 0)
}

func (self *Spinner) Update(time int64) bool {
	if time < self.objData.EndTime {
		self.rad = rpms * float64(time-self.objData.StartTime) * 2 * math.Pi
		/*self.pos = bmath.NewVec2dRad(self.rad, 10).Add(self.objData.StartPos)*/

		self.pos.X = 16*math.Pow(math.Sin(self.rad), 3)
		self.pos.Y = 13*math.Cos(self.rad) - 5*math.Cos(2*self.rad) - 2*math.Cos(3 *self.rad) - math.Cos(4 *self.rad)

		self.pos = self.pos.Scl(-8).Add(self.objData.StartPos)

		return false
	}

	index := self.objData.customIndex

	if index == 0 {
		index = self.Timings.Current.SampleIndex
	}

	if self.objData.sampleSet == 0 {
		audio.PlaySample(self.Timings.Current.SampleSet, self.objData.additionSet, self.sample, index, self.Timings.Current.SampleVolume)
	} else {
		audio.PlaySample(self.objData.sampleSet, self.objData.additionSet, self.sample, index, self.Timings.Current.SampleVolume)
	}

	return true
}

func (self *Spinner) Draw(time int64, color mgl32.Vec4, batch *batches.SpriteBatch) bool {
	self.fade.Update(float64(time))

	batch.SetTranslation(self.objData.StartPos)

	alpha := self.fade.GetValue()

	if settings.DIVIDES >= settings.Objects.MandalaTexturesTrigger {
		alpha *= settings.Objects.MandalaTexturesAlpha
	}

	batch.SetColor(1, 1, 1, alpha)
	scale := batch.GetScale()
	batch.SetScale(1, 1)

	batch.SetRotation(self.rad)
	batch.SetSubScale(20,20)

	batch.DrawUnit(*render.SpinnerMiddle)
	batch.DrawUnit(*render.SpinnerMiddle2)

	scl := 16 + math.Min(220, math.Max(0, (1.0 - float64(time - self.objData.StartTime)/float64(self.objData.EndTime - self.objData.StartTime)) * 220))

	batch.SetSubScale(scl, scl)

	batch.DrawUnit(*render.SpinnerAC)

	batch.SetSubScale(1, 1)
	batch.SetRotation(0)
	batch.SetScale(scale.X, scale.Y)

	if time >= self.objData.EndTime && self.fade.GetValue() <= 0.01 {
		return true
	}

	return false
}

func (self *Spinner) DrawApproach(time int64, color mgl32.Vec4, batch *batches.SpriteBatch) {}