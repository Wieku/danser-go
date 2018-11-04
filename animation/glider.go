package animation

import (
	"github.com/wieku/danser/animation/easing"
	"sort"
)

type event struct {
	startTime, endTime, targetValue float64
	hasStartValue bool
	startValue float64
}

type Glider struct {
	eventqueue              []event
	time, value, startValue float64
	current                 event
	easing                  func(float64) float64
	dirty bool
}

func NewGlider(value float64) *Glider {
	return &Glider{value: value, startValue: value, current: event{-1, 0, value, false, 0}, easing: easing.Linear}
}

func (glider *Glider) SetEasing(easing func(float64) float64) {
	glider.easing = easing
}

func (glider *Glider) AddEvent(startTime, endTime, targetValue float64) {
	glider.eventqueue = append(glider.eventqueue, event{startTime, endTime, targetValue, false, 0})
	glider.dirty = true
}

func (glider *Glider) AddEventS(startTime, endTime, startValue, targetValue float64) {
	glider.eventqueue = append(glider.eventqueue, event{startTime, endTime, targetValue, true, startValue})
	glider.dirty = true
}

func (glider *Glider) Update(time float64) {
	if glider.dirty {
		sort.Slice(glider.eventqueue, func(i, j int) bool { return glider.eventqueue[i].startTime < glider.eventqueue[j].startTime })
		glider.dirty = false
	}
	glider.time = time
	if len(glider.eventqueue) > 0 {
		if e := glider.eventqueue[0]; e.startTime <= time {
			glider.current = e
			glider.eventqueue = glider.eventqueue[1:]
			if e.hasStartValue {
				glider.startValue = e.startValue
			} else {
				glider.startValue = glider.value
			}
		}
	}

	if time < glider.current.endTime {
		e := glider.current
		t := (time - e.startTime) / (e.endTime - e.startTime)
		glider.value = glider.startValue + glider.easing(t)*(e.targetValue-glider.startValue)
	} else {
		glider.value = glider.current.targetValue
		glider.startValue = glider.value
	}
}

func (glider *Glider) UpdateD(delta float64) {
	glider.Update(glider.time + delta)
}
func (glider *Glider) SetValue(value float64) {
	glider.value = value
	glider.current.targetValue = value
	glider.startValue = value
}

func (glider *Glider) GetValue() float64 {
	return glider.value
}
