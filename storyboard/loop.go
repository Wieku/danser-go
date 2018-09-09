package storyboard

import (
	"strconv"
)

type Loop struct {
	start, end, repeats int64
	transformations     *Transformations
}

func NewLoop(data []string, object Object) *Loop {
	loop := &Loop{transformations: NewTransformations(object)}
	loop.start, _ = strconv.ParseInt(data[1], 10, 64)
	loop.repeats, _ = strconv.ParseInt(data[2], 10, 64)
	return loop
}

func (loop *Loop) Add(command *Command) {
	loop.transformations.Add(command)
	loop.end = loop.start + loop.repeats*loop.transformations.endTime
}

func (loop *Loop) Update(time int64) {
	sTime := int64(0)
	if time-loop.start > loop.transformations.endTime {
		sTime = loop.transformations.startTime
	}

	local := (time - loop.start - sTime) % (loop.transformations.endTime - sTime)
	loop.transformations.Update(sTime + local)
}
