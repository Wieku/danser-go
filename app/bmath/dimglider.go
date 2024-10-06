package bmath

import (
	"cmp"
	"github.com/wieku/danser-go/framework/math/animation/easing"
	"github.com/wieku/danser-go/framework/math/mutils"
	"slices"
)

type EType int

const (
	Absolute EType = iota
	Intro
	Normal
	Break
)

type event struct {
	startTime float64
	endTime   float64

	hasStartValue bool

	startValue  float64
	targetValue float64

	startSource EType
	endSource   EType

	easeFunc easing.Easing
}

type DimGlider struct {
	eventQueue []event
	current    event
	dirty      bool

	value float64

	easeFunc easing.Easing
}

func NewDimGlider(value float64) *DimGlider {
	return &DimGlider{
		value: value,
		current: event{
			startTime:     -100001,
			endTime:       -100000,
			hasStartValue: true,
			startValue:    value,
			targetValue:   value,
			startSource:   Absolute,
			endSource:     Absolute,
			easeFunc:      easing.Linear,
		},
		easeFunc: easing.Linear,
	}
}

func (glider *DimGlider) SetEasing(easing func(float64) float64) {
	glider.easeFunc = easing
}

func (glider *DimGlider) AddEvent(startTime, endTime float64, endType EType) {
	glider.addEvent(startTime, endTime, 0, 0, false, Absolute, endType, glider.easeFunc)
}

func (glider *DimGlider) AddEventV(startTime, endTime, targetValue float64, endType EType) {
	glider.addEvent(startTime, endTime, 0, targetValue, false, Absolute, endType, glider.easeFunc)
}

func (glider *DimGlider) addEvent(startTime, endTime, startValue, targetValue float64, hasStartValue bool, startType, endType EType, easeFunc easing.Easing) {
	glider.eventQueue = append(glider.eventQueue, event{
		startTime:     startTime,
		endTime:       endTime,
		hasStartValue: hasStartValue,
		startValue:    startValue,
		targetValue:   targetValue,
		startSource:   startType,
		endSource:     endType,
		easeFunc:      easeFunc,
	})

	glider.dirty = true
}

func (glider *DimGlider) Update(time, introD, normalD, breakD float64) {
	if glider.dirty {
		slices.SortFunc(glider.eventQueue, func(a, b event) int {
			return cmp.Compare(a.startTime, b.startTime)
		})

		glider.dirty = false
	}

	glider.updateCurrent(time, introD, normalD, breakD)

	if len(glider.eventQueue) > 0 {
		for i := 0; len(glider.eventQueue) > 0 && glider.eventQueue[i].startTime <= time; i++ {
			e := glider.eventQueue[i]

			if !e.hasStartValue {
				e.startValue = glider.value
				e.startSource = Absolute
			}

			glider.current = e

			glider.updateCurrent(time, introD, normalD, breakD)

			glider.eventQueue = glider.eventQueue[1:]
			i--
		}
	}
}

func (glider *DimGlider) updateCurrent(time, introD, normalD, breakD float64) {
	e := glider.current

	startValue := glider.getValue(e.startSource, e.startValue, introD, normalD, breakD)
	endValue := glider.getValue(e.endSource, e.targetValue, introD, normalD, breakD)

	if time < glider.current.endTime && glider.current.startTime != glider.current.endTime {
		t := mutils.Clamp(0, 1, (time-e.startTime)/(e.endTime-e.startTime))
		glider.value = startValue + e.easeFunc(t)*(endValue-startValue)
	} else {
		glider.value = endValue
	}
}

func (glider *DimGlider) getValue(e EType, absD, introD, normalD, breakD float64) float64 {
	switch e {
	case Absolute:
		return absD
	case Intro:
		return introD
	case Normal:
		return normalD
	case Break:
		return breakD
	}

	return 0
}

func (glider *DimGlider) Reset() {
	glider.eventQueue = glider.eventQueue[:0]
}

func (glider *DimGlider) GetValue() float64 {
	return glider.value
}
