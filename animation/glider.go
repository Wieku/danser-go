package animation

import (
	"danser/animation/easing"
)

type event struct {
	startTime, endTime, targetValue float64
	hasStartValue                   bool
	startValue                      float64
}

type Glider struct {
	eventqueue              []event
	time, value, startValue float64
	current                 event
	easing                  func(float64) float64
}

func NewGlider(value float64) *Glider {
	return &Glider{value: value, startValue: value, current: event{-1, 0, value, false, 0}, easing: easing.Linear}
}

func (glider *Glider) SetEasing(easing func(float64) float64) {
	glider.easing = easing
}

func (glider *Glider) AddEvent(startTime, endTime, targetValue float64) {
	glider.eventqueue = append(glider.eventqueue, event{startTime, endTime, targetValue, false, 0})
}

func (glider *Glider) AddEventS(startTime, endTime, startValue, targetValue float64) {
	glider.eventqueue = append(glider.eventqueue, event{startTime, endTime, targetValue, true, startValue})
}

func (glider *Glider) Update(time float64) {
	glider.time = time
	if len(glider.eventqueue) > 0 {
		if e := glider.eventqueue[0]; e.startTime <= time {
			glider.current = e
			glider.eventqueue = glider.eventqueue[1:]
		}
	}

	if time <= glider.current.endTime {
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
