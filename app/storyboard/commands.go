package storyboard

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/wieku/danser-go/app/bmath"
	"github.com/wieku/danser-go/framework/math/animation"
	"github.com/wieku/danser-go/framework/math/animation/easing"
	"github.com/wieku/danser-go/framework/math/vector"
	"log"
	"math"
	"strconv"
)

type Command struct {
	start, end  int64
	command     string
	easing      func(float64) float64
	val         []float64
	numSections int64
	sectionTime int64
	sections    [][]float64
	custom      string
	constant    bool
}

func NewCommand(data []string) *Command {
	command := &Command{}
	command.command = data[0]

	easingID, err := strconv.ParseInt(data[1], 10, 32)

	if err != nil {
		log.Println(err)
	}

	command.easing = easing.GetEasing(easingID)

	command.start, err = strconv.ParseInt(data[2], 10, 64)

	if err != nil {
		panic(err)
	}

	if data[3] == "" {
		command.end = command.start
	} else {
		command.end, err = strconv.ParseInt(data[3], 10, 64)

		if err != nil {
			panic(err)
		}
	}

	if command.end < command.start {
		command.end = command.start
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
		command.val = make([]float64, 1)
		return command
	}

	numSections := len(parameters) / arguments
	command.numSections = int64(numSections) - 1
	command.sectionTime = command.end - command.start

	if command.numSections > 1 {
		command.end = command.start + command.sectionTime*command.numSections
	}

	command.sections = make([][]float64, numSections)
	command.val = make([]float64, arguments)

	for i := 0; i < numSections; i++ {
		command.sections[i] = make([]float64, arguments)
		for j := 0; j < arguments; j++ {
			var err error
			command.sections[i][j], err = strconv.ParseFloat(parameters[arguments*i+j], 64)

			if command.command == "C" {
				command.sections[i][j] /= 255
			}

			if err != nil {
				log.Println(err)
			}
		}
	}

	copy(command.val, command.sections[0])

	if numSections == 1 || command.sectionTime == 0 {
		command.constant = true
	}

	return command
}

func (command *Command) Update(time int64) {

	if command.command == "P" {
		if command.start != command.end {
			if time >= command.start && time <= command.end {
				command.val[0] = 1
			} else {
				command.val[0] = 0
			}
		} else {
			command.val[0] = 1
		}

		return
	}

	if command.constant {
		copy(command.val, command.sections[len(command.sections)-1])
	} else {

		section := int64(0)

		if command.sectionTime > 0 {
			section = (time - command.start) / command.sectionTime
		}

		if section >= command.numSections {
			section = command.numSections - 1
		}

		t := command.easing(math.Min(math.Max(float64((time-command.start)-section*command.sectionTime)/float64(command.sectionTime), 0), 1))

		for i := range command.val {
			command.val[i] = command.sections[section][i] + t*(command.sections[section+1][i]-command.sections[section][i])
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
		obj.SetScale(vector.NewVec2d(command.val[0], command.val[0]))
		break
	case "V":
		obj.SetScale(vector.NewVec2d(command.val[0], command.val[1]))
		break
	case "M":
		obj.SetPosition(vector.NewVec2d(command.val[0], command.val[1]))
		break
	case "MX":
		obj.SetPosition(vector.NewVec2d(command.val[0], obj.GetPosition().Y))
		break
	case "MY":
		obj.SetPosition(vector.NewVec2d(obj.GetPosition().X, command.val[0]))
		break
	case "C":
		obj.SetColor(mgl32.Vec3{float32(command.val[0]), float32(command.val[1]), float32(command.val[2])})
		break
	case "P":
		switch command.custom {
		case "H":
			obj.SetHFlip(command.val[0] == 1)
			break
		case "V":
			obj.SetVFlip(command.val[0] == 1)
			break
		case "A":
			obj.SetAdditive(command.val[0] == 1)
			break
		}
		break
	}
}

func (command *Command) Init(obj Object) {

	if command.command == "P" {
		return
	}

	copy(command.val, command.sections[0])
	command.Apply(obj)
}

func (command *Command) GenerateTransformations() []*animation.Transformation {
	if command.command == "P" {
		switch command.custom {
		case "H":
			return []*animation.Transformation{animation.NewBooleanTransform(animation.HorizontalFlip, float64(command.start), float64(command.end))}
		case "V":
			return []*animation.Transformation{animation.NewBooleanTransform(animation.VerticalFlip, float64(command.start), float64(command.end))}
		case "A":
			return []*animation.Transformation{animation.NewBooleanTransform(animation.Additive, float64(command.start), float64(command.end))}
		}
	}

	var transformations []*animation.Transformation

	nSections := bmath.MaxI(1, int(command.numSections))

	for section := 0; section < nSections; section++ {
		start := float64(command.start + int64(section)*command.sectionTime)
		end := float64(command.start + int64(section+1)*command.sectionTime)
		nextSection := (section + 1) % int(command.numSections+1)
		switch command.command {
		case "F":
			transformations = append(transformations, animation.NewSingleTransform(animation.Fade, command.easing, start, end, command.sections[section][0], command.sections[nextSection][0]))
		case "R":
			transformations = append(transformations, animation.NewSingleTransform(animation.Rotate, command.easing, start, end, command.sections[section][0], command.sections[nextSection][0]))
		case "S":
			transformations = append(transformations, animation.NewSingleTransform(animation.Scale, command.easing, start, end, command.sections[section][0], command.sections[nextSection][0]))
		case "V":
			transformations = append(transformations, animation.NewVectorTransform(animation.ScaleVector, command.easing, start, end, command.sections[section][0], command.sections[section][1], command.sections[nextSection][0], command.sections[nextSection][1]))
		case "M":
			transformations = append(transformations, animation.NewVectorTransform(animation.Move, command.easing, start, end, command.sections[section][0], command.sections[section][1], command.sections[nextSection][0], command.sections[nextSection][1]))
		case "MX":
			transformations = append(transformations, animation.NewSingleTransform(animation.MoveX, command.easing, start, end, command.sections[section][0], command.sections[nextSection][0]))
			break
		case "MY":
			transformations = append(transformations, animation.NewSingleTransform(animation.MoveY, command.easing, start, end, command.sections[section][0], command.sections[nextSection][0]))
			break
		case "C":
			color1 := bmath.Color{
				R: command.sections[section][0],
				G: command.sections[section][1],
				B: command.sections[section][2],
			}
			color2 := bmath.Color{
				R: command.sections[nextSection][0],
				G: command.sections[nextSection][1],
				B: command.sections[nextSection][2],
			}
			transformations = append(transformations, animation.NewColorTransform(animation.Color3, command.easing, start, end, color1, color2))
		}
	}

	return transformations
}

//TODO: TRIGGER command
