package storyboard

import (
	"github.com/wieku/danser/bmath"
	"math"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/Wieku/danser/animation/easing"
	"strconv"
	"log"
)

type Command struct {
	start, end            int64
	command               string
	easing                func(float64) float64
	startVal, endVal, val []float64
	custom                string
	constant              bool
}

func NewCommand(data []string) *Command {
	command := &Command{}
	command.command = data[0]

	easingID, err := strconv.ParseInt(data[1], 10, 32)

	if err != nil {
		log.Println(err)
	}

	command.easing = easing.Easings[easingID]

	command.start, err = strconv.ParseInt(data[2], 10, 64)

	if err != nil {
		panic(err)
	}

	if data[3] == "" { //if this field is empty, this command will run until the next command of this same type is called or till the end of the last command in object
		command.end = command.start - 1000
	} else {
		command.end, err = strconv.ParseInt(data[3], 10, 64)

		if err != nil {
			panic(err)
		}
	}

	arguments := 0

	switch command.command {
	case "F", "R", "S", "MX", "MY":
		arguments = 1
		break
	case "M", "V":
		arguments = 2
		break
	case "C":
		arguments = 3
		break
	}

	parameters := data[4:]

	if arguments == 0 {
		command.custom = parameters[0]
		return command
	}

	if arguments < len(parameters) {
		command.endVal = make([]float64, arguments)

		for i := range command.endVal {
			var err error
			command.endVal[i], err = strconv.ParseFloat(parameters[arguments+i], 64)

			if err != nil {
				log.Println(err)
			}
		}
	} else {
		command.constant = true
	}

	command.startVal = make([]float64, arguments)
	command.val = make([]float64, arguments)

	for i := range command.startVal {
		var err error
		command.startVal[i], err = strconv.ParseFloat(parameters[i], 64)

		if err != nil {
			log.Println(err)
		}
	}

	return command
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

//TODO: LOOP and TRIGGER commands
