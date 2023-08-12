package animation

import (
	"github.com/wieku/danser-go/framework/math/animation/easing"
	"github.com/wieku/danser-go/framework/math/mutils"
	"sort"
)

type event struct {
	startTime     float64
	endTime       float64
	hasStartValue bool
	startValue    float64
	targetValue   float64
	easeFunc      easing.Easing
}

type Glider struct {
	eventQueue []event
	current    event

	time       float64
	value      float64
	startValue float64

	easeFunc easing.Easing

	sorting bool
	dirty   bool
}

func NewGlider(value float64) *Glider {
	return &Glider{
		value:      value,
		startValue: value,
		current: event{
			startTime:     -1,
			endTime:       0,
			targetValue:   value,
			hasStartValue: false,
			startValue:    0,
			easeFunc:      easing.Linear,
		},
		easeFunc: easing.Linear,
		sorting:  true,
	}
}

func (glider *Glider) SetSorting(sorting bool) {
	glider.sorting = sorting
}

func (glider *Glider) SetEasing(easing func(float64) float64) {
	glider.easeFunc = easing
}

func (glider *Glider) AddEvent(startTime, endTime, targetValue float64) {
	glider.addEvent(startTime, endTime, 0, targetValue, false, glider.easeFunc)
}

func (glider *Glider) AddEventS(startTime, endTime, startValue, targetValue float64) {
	glider.addEvent(startTime, endTime, startValue, targetValue, true, glider.easeFunc)
}

func (glider *Glider) AddEventEase(startTime, endTime, targetValue float64, easeFunc easing.Easing) {
	glider.addEvent(startTime, endTime, 0, targetValue, false, easeFunc)
}

func (glider *Glider) AddEventSEase(startTime, endTime, startValue, targetValue float64, easeFunc easing.Easing) {
	glider.addEvent(startTime, endTime, startValue, targetValue, true, easeFunc)
}

func (glider *Glider) addEvent(startTime, endTime, startValue, targetValue float64, hasStartValue bool, easeFunc easing.Easing) {
	glider.eventQueue = append(glider.eventQueue, event{
		startTime:     startTime,
		endTime:       endTime,
		hasStartValue: hasStartValue,
		startValue:    startValue,
		targetValue:   targetValue,
		easeFunc:      easeFunc,
	})

	glider.dirty = true
}

func (glider *Glider) Update(time float64) {
	if glider.dirty && glider.sorting {
		sort.Slice(glider.eventQueue, func(i, j int) bool {
			return glider.eventQueue[i].startTime < glider.eventQueue[j].startTime
		})

		glider.dirty = false
	}

	glider.time = time

	glider.updateCurrent(time)

	if len(glider.eventQueue) > 0 {
		for i := 0; len(glider.eventQueue) > 0 && glider.eventQueue[i].startTime <= time; i++ {
			e := glider.eventQueue[i]

			if e.hasStartValue {
				glider.startValue = e.startValue
			} else if glider.current.endTime <= e.startTime {
				glider.startValue = glider.current.targetValue
			} else {
				glider.startValue = glider.value
			}

			if glider.current.endTime > e.startTime && e.hasStartValue {
				glider.value = e.startValue
			}

			glider.current = e
			if glider.startValue == glider.current.targetValue {
				glider.value = glider.current.targetValue
			}

			glider.updateCurrent(time)

			glider.eventQueue = glider.eventQueue[1:]
			i--
		}
	}
}

func (glider *Glider) updateCurrent(time float64) {
	if time < glider.current.endTime && glider.current.startTime != glider.current.endTime {
		e := glider.current
		t := mutils.Clamp(0, 1, (time-e.startTime)/(e.endTime-e.startTime))
		glider.value = glider.startValue + e.easeFunc(t)*(e.targetValue-glider.startValue)
	} else {
		glider.value = glider.current.targetValue
		glider.startValue = glider.value
	}
}

func (glider *Glider) UpdateD(delta float64) {
	glider.Update(glider.time + delta)
}

func (glider *Glider) GetTime() float64 {
	return glider.time
}

func (glider *Glider) SetValue(value float64) {
	glider.value = value
	glider.current.targetValue = value
	glider.startValue = value
}

func (glider *Glider) Reset() {
	glider.eventQueue = glider.eventQueue[:0]
	glider.SetValue(glider.value)
}

func (glider *Glider) RemoveLast() {
	if len(glider.eventQueue) > 1 {
		glider.eventQueue = glider.eventQueue[:len(glider.eventQueue)-1]
	}
}

func (glider *Glider) GetValue() float64 {
	return glider.value
}
