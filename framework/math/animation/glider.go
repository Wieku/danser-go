package animation

import (
	"github.com/wieku/danser-go/framework/math/animation/easing"
	"math"
	"sort"
)

type event struct {
	startTime, endTime, targetValue float64
	hasStartValue                   bool
	startValue                      float64
	easing                          func(float64) float64
}

type Glider struct {
	eventqueue              []event
	time, value, startValue float64
	current                 event
	easing                  func(float64) float64
	sorting                 bool
	dirty                   bool
}

func NewGlider(value float64) *Glider {
	return &Glider{value: value, startValue: value, current: event{-1, 0, value, false, 0, easing.Linear}, easing: easing.Linear, sorting: true}
}

func (glider *Glider) SetSorting(sorting bool) {
	glider.sorting = sorting
}

func (glider *Glider) SetEasing(easing func(float64) float64) {
	glider.easing = easing
}

func (glider *Glider) AddEvent(startTime, endTime, targetValue float64) {
	glider.eventqueue = append(glider.eventqueue, event{startTime, endTime, targetValue, false, 0, glider.easing})
	glider.dirty = true
}

func (glider *Glider) AddEventS(startTime, endTime, startValue, targetValue float64) {
	glider.eventqueue = append(glider.eventqueue, event{startTime, endTime, targetValue, true, startValue, glider.easing})
	glider.dirty = true
}

func (glider *Glider) AddEventEase(startTime, endTime, targetValue float64, easing func(float64) float64) {
	glider.eventqueue = append(glider.eventqueue, event{startTime, endTime, targetValue, false, 0, easing})
	glider.dirty = true
}

func (glider *Glider) AddEventSEase(startTime, endTime, startValue, targetValue float64, easing func(float64) float64) {
	glider.eventqueue = append(glider.eventqueue, event{startTime, endTime, targetValue, true, startValue, easing})
	glider.dirty = true
}

func (glider *Glider) Update(time float64) {
	if glider.dirty && glider.sorting {
		sort.Slice(glider.eventqueue, func(i, j int) bool { return glider.eventqueue[i].startTime < glider.eventqueue[j].startTime })
		glider.dirty = false
	}
	glider.time = time

	glider.updateCurrent(time)

	if len(glider.eventqueue) > 0 {
		/*for i := 0; i < len(glider.eventqueue); i++ {
			if e := glider.eventqueue[i]; (e.startTime <= time && (e.endTime >= time || len(glider.eventqueue) == 1 || e.startTime == e.endTime)) || (i < len(glider.eventqueue)-1 && time > e.endTime && glider.eventqueue[i+1].startTime > time) {
				if e.hasStartValue {
					glider.startValue = e.startValue
				} else if i > 0 {
					glider.startValue = glider.eventqueue[i-1].targetValue
				} else if glider.current.endTime <= e.startTime {
					glider.startValue = glider.current.targetValue
				} else {
					glider.startValue = glider.value
				}
				glider.current = e
				glider.updateCurrent(time)
				glider.eventqueue = glider.eventqueue[i+1:]
				i--
			} else if e.startTime > time {
				break
			}
		}*/
		for i := 0; len(glider.eventqueue) > 0 && glider.eventqueue[i].startTime <= time; i++ {
			e := glider.eventqueue[i]
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
			glider.eventqueue = glider.eventqueue[1:]
			i--
		}
	}
}

func (glider *Glider) updateCurrent(time float64) {
	if time < glider.current.endTime {
		e := glider.current
		t := math.Min(1.0, math.Max(0.0, time-e.startTime)/(e.endTime-e.startTime))
		glider.value = glider.startValue + e.easing(t)*(e.targetValue-glider.startValue)
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
	glider.eventqueue = make([]event, 0)
	glider.current.targetValue = glider.value
	glider.current.endTime = glider.time
}

func (glider *Glider) RemoveLast() {
	if len(glider.eventqueue) > 1 {
		glider.eventqueue = glider.eventqueue[:len(glider.eventqueue)-1]
	}
}

func (glider *Glider) GetValue() float64 {
	return glider.value
}
