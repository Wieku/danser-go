package storyboard

import (
	"github.com/wieku/danser/bmath"
	"math"
	"github.com/go-gl/mathgl/mgl32"
)

type Command struct {
	start, end            int64
	command               string
	easing                func(float64) float64
	startVal, endVal, val []float64
	custom                string
	constant              bool
}

func (command *Command) Update(time int64) {

	if command.command == "P" {
		return
	}

	if command.constant {
		copy(command.val, command.startVal)
	} else {
		t := command.easing(math.Min(math.Max(float64(time-command.start)/float64(command.end-command.start), 0), 1))

		for i := range command.val {
			command.val[i] = command.startVal[i] + t*(command.endVal[i]-command.startVal[i])
		}
	}
}

func (command *Command) Apply(obj Object) {
	switch command.command {
	case "F":
		obj.SetAlpha(command.val[0])
		break
	case "R":
		obj.SetRotation(command.val[0])
		break
	case "S":
		obj.SetScale(bmath.NewVec2d(command.val[0], command.val[0]))
		break
	case "V":
		obj.SetScale(bmath.NewVec2d(command.val[0], command.val[1]))
		break
	case "M":
		obj.SetPosition(bmath.NewVec2d(command.val[0], command.val[1]))
		break
	case "MX":
		obj.SetPosition(bmath.NewVec2d(command.val[0], obj.GetPosition().Y))
		break
	case "MY":
		obj.SetPosition(bmath.NewVec2d(obj.GetPosition().X, command.val[0]))
		break
	case "C":
		obj.SetColor(mgl32.Vec3{float32(command.val[0]), float32(command.val[1]), float32(command.val[2])})
		break
	case "P":
		switch command.custom {
		case "H":
			obj.SetHFlip()
			break
		case "V":
			obj.SetVFlip()
			break
		case "A":
			obj.SetAdditive()
			break
		}
		break
	}
}
