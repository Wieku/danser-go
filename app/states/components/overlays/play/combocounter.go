package play

import (
	"fmt"
	"github.com/wieku/danser-go/app/audio"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/app/skin"
	"github.com/wieku/danser-go/framework/bass"
	"github.com/wieku/danser-go/framework/graphics/batch"
	"github.com/wieku/danser-go/framework/graphics/font"
	"github.com/wieku/danser-go/framework/graphics/sprite"
	"github.com/wieku/danser-go/framework/math/animation"
	"github.com/wieku/danser-go/framework/math/animation/easing"
	"github.com/wieku/danser-go/framework/math/vector"
	"math"
)

type ComboCounter struct {
	comboFont *font.Font

	mainCounter *sprite.TextSprite
	popCounter  *sprite.TextSprite

	comboSlide *animation.Glider

	comboBreak *bass.Sample

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
	fnt := skin.GetFont("combo")

	counter := &ComboCounter{
		comboFont:   fnt,
		mainCounter: sprite.NewTextSprite("0x", fnt, 0, vector.NewVec2d(0, 0), vector.BottomLeft),
		popCounter:  sprite.NewTextSprite("0x", fnt, 0, vector.NewVec2d(0, 0), vector.BottomLeft),
		comboSlide:  animation.NewGlider(0),
		comboBreak:  audio.LoadSample("combobreak"),
	}

	counter.mainCounter.SetScale(1.28)

	counter.popCounter.SetScale(1.28)
	counter.popCounter.SetAlpha(0)
	counter.popCounter.SetAdditive(true)

	counter.ScaledHeight = 768
	counter.ScaledWidth = settings.Graphics.GetAspectRatio() * counter.ScaledHeight

	counter.comboSlide.SetEasing(easing.OutQuad)

	return counter
}

func (counter *ComboCounter) Increase() {
	counter.popCounter.ClearTransformations()
	counter.popCounter.AddTransform(animation.NewSingleTransform(animation.Scale, easing.Linear, counter.time, counter.time+300, 1.563, 1))
	counter.popCounter.AddTransform(animation.NewSingleTransform(animation.Fade, easing.Linear, counter.time, counter.time+300, 0.6, 0.0))

	counter.updateMain(counter.newCombo, counter.combo < counter.newCombo)

	counter.newCombo++
	counter.nextEnd = counter.time + 300

	counter.popCounter.SetText(fmt.Sprintf("%dx", counter.newCombo))
}

func (counter *ComboCounter) Reset() {
	if counter.newCombo > 20 && counter.comboBreak != nil && !counter.audioDisabled {
		counter.comboBreak.Play()
	}

	counter.newCombo = 0

	counter.popCounter.SetText(fmt.Sprintf("%dx", counter.newCombo))
}

func (counter *ComboCounter) GetCombo() int {
	return counter.newCombo
}

func (counter *ComboCounter) DisableAudioSubmission(b bool) {
	counter.audioDisabled = b
}

func (counter *ComboCounter) updateMain(combo int, bump bool) {
	counter.combo = combo

	counter.mainCounter.SetText(fmt.Sprintf("%dx", combo))

	if bump {
		counter.mainCounter.ClearTransformationsOfType(animation.Scale)
		counter.mainCounter.AddTransform(animation.NewSingleTransform(animation.Scale, easing.InQuad, counter.time, counter.time+50, 1, 1.094))
		counter.mainCounter.AddTransform(animation.NewSingleTransform(animation.Scale, easing.OutQuad, counter.time+50, counter.time+100, 1.094, 1))
	}
}

func (counter *ComboCounter) Update(time float64) {
	counter.delta += time - counter.time

	if counter.delta >= 16.6667 {
		counter.delta -= 16.6667

		if counter.combo > counter.newCombo && counter.newCombo == 0 {
			counter.updateMain(counter.combo-1, false)
		}
	}

	counter.time = time

	if counter.combo < counter.newCombo && counter.nextEnd < time+140 {
		counter.updateMain(counter.newCombo, true)
		counter.nextEnd = math.MaxInt64
	}

	counter.mainCounter.Update(time)
	counter.popCounter.Update(time)

	counter.comboSlide.Update(time)
}

func (counter *ComboCounter) SlideOut() {
	counter.comboSlide.AddEventEase(counter.time, counter.time+1000, -130, easing.InQuad)
	counter.mainCounter.AddTransform(animation.NewSingleTransform(animation.Fade, easing.Linear, counter.time, counter.time+1000, counter.mainCounter.GetAlpha(), 0.0))
}

func (counter *ComboCounter) SlideIn() {
	counter.comboSlide.AddEventEase(counter.time, counter.time+1000, 0, easing.OutQuad)
	counter.mainCounter.AddTransform(animation.NewSingleTransform(animation.Fade, easing.Linear, counter.time, counter.time+1000, counter.mainCounter.GetAlpha(), 1.0))
}

func (counter *ComboCounter) Draw(batch *batch.QuadBatch, alpha float64) {
	comboAlpha := settings.Gameplay.ComboCounter.Opacity * alpha

	if comboAlpha < 0.001 || !settings.Gameplay.ComboCounter.Show {
		return
	}

	batch.ResetTransform()
	batch.SetColor(1, 1, 1, comboAlpha)

	slideAmount := counter.comboSlide.GetValue()
	if settings.Gameplay.ComboCounter.XOffset > 0.01 {
		slideAmount = 0
	}

	xPos := settings.Gameplay.ComboCounter.XOffset + 3.2 + slideAmount
	yPos := settings.Gameplay.ComboCounter.YOffset + counter.ScaledHeight - 12.8

	batch.SetTranslation(vector.NewVec2d(xPos, yPos))

	scl := settings.Gameplay.ComboCounter.Scale * 1.28

	batch.SetScale(scl, scl)

	origY := counter.comboFont.GetSize()*0.375 - 9

	counter.popCounter.SetPosition(vector.NewVec2d(-3, origY).Scl(scl * counter.popCounter.GetScale().X))
	counter.mainCounter.SetPosition(vector.NewVec2d(0, origY).Scl(scl * counter.mainCounter.GetScale().X))

	counter.popCounter.Draw(0, batch)
	counter.mainCounter.Draw(0, batch)

	batch.SetColor(1, 1, 1, 1)
	batch.ResetTransform()
}
