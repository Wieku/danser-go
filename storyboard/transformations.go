package storyboard

import (
	"sort"
	"math"
)

type Transformations struct {
	object                       Object
	commands                     []*Command
	queue                        []*Command
	processed                    []*Command
	startTime, endTime, lastTime int64
}

func NewTransformations(obj Object) *Transformations {
	return &Transformations{object: obj, startTime: math.MaxInt64, endTime: math.MinInt64}
}

func (trans *Transformations) Add(command *Command) {

	if command.command != "P" {
		exists := false

		for _, e := range trans.queue {
			if e.command == command.command && e.start < command.start {
				exists = true
				break
			}
		}

		if !exists {
			if trans.object != nil {
				command.Init(trans.object)
			}
		}
	}

	trans.commands = append(trans.commands, command)
	trans.queue = append(trans.queue, command)

	sort.Slice(trans.queue, func(i, j int) bool {
		return trans.queue[i].start < trans.queue[j].start
	})

	sort.Slice(trans.commands, func(i, j int) bool {
		return trans.commands[i].start < trans.commands[j].start
	})

	if command.start < trans.startTime {
		trans.startTime = command.start
	}

	if command.end > trans.endTime {
		trans.endTime = command.end
	}
}

func (trans *Transformations) Update(time int64) {

	if time < trans.lastTime {
		trans.queue = make([]*Command, len(trans.commands))
		copy(trans.queue, trans.commands)
		trans.processed = make([]*Command, 0)
	}

	for i := 0; i < len(trans.queue); i++ {
		c := trans.queue[i]
		if c.start <= time {
			trans.processed = append(trans.processed, c)
			trans.queue = append(trans.queue[:i], trans.queue[i+1:]...)
			i--
		}
	}

	for i := 0; i < len(trans.processed); i++ {
		c := trans.processed[i]
		c.Update(time)
		if trans.object != nil {
			c.Apply(trans.object)
		}

		if time > c.end {
			trans.processed = append(trans.processed[:i], trans.processed[i+1:]...)
			i--
		}
	}

	trans.lastTime = time
}
