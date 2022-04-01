package play

import (
	"fmt"
	"github.com/wieku/danser-go/app/audio"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/app/skin"
	"github.com/wieku/danser-go/framework/bass"
	"github.com/wieku/danser-go/framework/graphics/batch"
	"github.com/wieku/danser-go/framework/graphics/font"
	"github.com/wieku/danser-go/framework/math/animation"
	"github.com/wieku/danser-go/framework/math/animation/easing"
	"github.com/wieku/danser-go/framework/math/vector"
	"math"
)

type ComboCounter struct {
	comboFont *font.Font

	comboSlide     *animation.Glider
	newComboScale  *animation.Glider
	newComboScaleB *animation.Glider
	newComboFadeB  *animation.Glider

	combobreak *bass.Sample

	time    float64
	delta   float64
	nextEnd float64

	newCombo int
	combo    int

	audioDisabled bool

	ScaledWidth  float64
	ScaledHeight float64
}

func NewComboCounter() *ComboCounter {
	counter := &ComboCounter{
		comboFont:      skin.GetFont("combo"),
		comboSlide:     animation.NewGlider(0),
		newComboScale:  animation.NewGlider(1.28),
		newComboScaleB: animation.NewGlider(1.28),
		newComboFadeB:  animation.NewGlider(0),
		combobreak:     audio.LoadSample("combobreak"),
	}

	counter.ScaledHeight = 768
	counter.ScaledWidth = settings.Graphics.GetAspectRatio() * counter.ScaledHeight

	counter.comboSlide.SetEasing(easing.OutQuad)

	return counter
}

func (counter *ComboCounter) Increase() {
	counter.newComboScaleB.Reset()
	counter.newComboScaleB.AddEventS(counter.time, counter.time+300, 2, 1.28)

	counter.newComboFadeB.Reset()
	counter.newComboFadeB.AddEventS(counter.time, counter.time+300, 0.6, 0.0)

	if counter.combo < counter.newCombo {
		counter.animate(counter.time)
	}

	counter.combo = counter.newCombo
	counter.newCombo++
	counter.nextEnd = counter.time + 300
}

func (counter *ComboCounter) Reset() {
	if counter.newCombo > 20 && counter.combobreak != nil && !counter.audioDisabled {
		counter.combobreak.Play()
	}

	counter.newCombo = 0
}

func (counter *ComboCounter) GetCombo() int {
	return counter.newCombo
}

func (counter *ComboCounter) DisableAudioSubmission(b bool) {
	counter.audioDisabled = b
}

func (counter *ComboCounter) animate(time float64) {
	counter.newComboScale.Reset()
	counter.newComboScale.AddEventSEase(time, time+50, 1.28, 1.4, easing.InQuad)
	counter.newComboScale.AddEventSEase(time+50, time+100, 1.4, 1.28, easing.OutQuad)
}

func (counter *ComboCounter) Update(time float64) {

	counter.delta += time - counter.time
	if counter.delta >= 16.6667 {
		counter.delta -= 16.6667
		if counter.combo > counter.newCombo && counter.newCombo == 0 {
			counter.combo--
		}
	}

	if counter.combo < counter.newCombo && counter.nextEnd < time+140 {
		counter.animate(time)
		counter.combo = counter.newCombo
		counter.nextEnd = math.MaxInt64
	}

	counter.time = time

	counter.newComboScale.Update(time)
	counter.newComboScaleB.Update(time)
	counter.newComboFadeB.Update(time)
	counter.comboSlide.Update(time)
}

func (counter *ComboCounter) SlideOut() {
	counter.comboSlide.AddEvent(counter.time, counter.time+500, -1)
}

func (counter *ComboCounter) SlideIn() {
	counter.comboSlide.AddEvent(counter.time, counter.time+500, 0)
}

func (counter *ComboCounter) Draw(batch *batch.QuadBatch, alpha float64) {
	comboAlpha := settings.Gameplay.ComboCounter.Opacity * alpha

	if comboAlpha < 0.001 || !settings.Gameplay.ComboCounter.Show {
		return
	}

	cmbSize := counter.comboFont.GetSize() * settings.Gameplay.ComboCounter.Scale

	slideAmount := counter.comboSlide.GetValue()

	if settings.Gameplay.ComboCounter.XOffset > 0.01 {
		slideAmount = 0
		comboAlpha *= 1 + counter.comboSlide.GetValue()
	}

	posX := slideAmount*counter.comboFont.GetWidth(cmbSize*counter.newComboScale.GetValue(), fmt.Sprintf("%dx", counter.combo)) + 2.5
	posY := counter.ScaledHeight - 12.8
	origY := counter.comboFont.GetSize()*0.375 - 9

	batch.ResetTransform()
	batch.SetTranslation(vector.NewVec2d(settings.Gameplay.ComboCounter.XOffset, settings.Gameplay.ComboCounter.YOffset))

	batch.SetAdditive(true)

	batch.SetColor(1, 1, 1, counter.newComboFadeB.GetValue()*comboAlpha)
	counter.comboFont.DrawOrigin(batch, posX-2.4*counter.newComboScaleB.GetValue()*settings.Gameplay.ComboCounter.Scale, posY+origY*counter.newComboScaleB.GetValue()*settings.Gameplay.ComboCounter.Scale, vector.BottomLeft, cmbSize*counter.newComboScaleB.GetValue(), false, fmt.Sprintf("%dx", counter.newCombo))

	batch.SetAdditive(false)

	batch.SetColor(1, 1, 1, comboAlpha)
	counter.comboFont.DrawOrigin(batch, posX, posY+origY*counter.newComboScale.GetValue()*settings.Gameplay.ComboCounter.Scale, vector.BottomLeft, cmbSize*counter.newComboScale.GetValue(), false, fmt.Sprintf("%dx", counter.combo))

	batch.ResetTransform()
}
