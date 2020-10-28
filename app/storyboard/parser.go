package storyboard

import (
	"github.com/wieku/danser-go/app/bmath"
	"github.com/wieku/danser-go/framework/math/animation"
	"github.com/wieku/danser-go/framework/math/animation/easing"
	color2 "github.com/wieku/danser-go/framework/math/color"
	"log"
	"strconv"
	"strings"
	"unicode"
)

func cutWhites(text string) (string, int) {
	for i, c := range text {
		if unicode.IsLetter(c) || unicode.IsNumber(c) {
			return text[i:], i
		}
	}

	return text, 0
}

func parseCommands(commands []string) []*animation.Transformation {
	transforms := make([]*animation.Transformation, 0)

	var currentLoop *LoopProcessor = nil

	loopDepth := -1

	for _, subCommand := range commands {
		command := strings.Split(subCommand, ",")

		var removed int
		command[0], removed = cutWhites(command[0])

		if command[0] == "T" {
			continue
		}

		if removed == 1 {
			if currentLoop != nil {
				transforms = append(transforms, currentLoop.Unwind()...)

				currentLoop = nil
				loopDepth = -1
			}

			if command[0] != "L" {
				transforms = append(transforms, parseCommand(command)...)
			}
		}

		if command[0] == "L" {
			currentLoop = NewLoopProcessor(command)
			loopDepth = removed + 1
		} else if removed == loopDepth && currentLoop != nil {
			currentLoop.Add(command)
		}
	}

	if currentLoop != nil {
		transforms = append(transforms, currentLoop.Unwind()...)

		currentLoop = nil
		loopDepth = -1
	}

	return transforms
}

func parseCommand(data []string) []*animation.Transformation {
	transforms := make([]*animation.Transformation, 0)

	checkError := func(err error) {
		if err != nil {
			log.Println("Failed to parse: ", data)
			panic(err)
		}
	}

	command := data[0]

	easingID, err := strconv.ParseInt(data[1], 10, 32)
	checkError(err)

	easeFunc := easing.GetEasing(easingID)

	startTime, err := strconv.ParseInt(data[2], 10, 64)
	checkError(err)

	var endTime int64
	if data[3] != "" {
		endTime, err = strconv.ParseInt(data[3], 10, 64)
		checkError(err)
	}

	endTime = bmath.MaxI64(endTime, startTime)

	arguments := 0

	switch command {
	case "F", "R", "S", "MX", "MY":
		arguments = 1
	case "M", "V":
		arguments = 2
	case "C":
		arguments = 3
	}

	parameters := data[4:]

	if arguments == 0 {
		typ := animation.TransformationType(0)

		switch parameters[0] {
		case "H":
			typ = animation.HorizontalFlip
		case "V":
			typ = animation.VerticalFlip
		case "A":
			typ = animation.Additive
		}

		return []*animation.Transformation{animation.NewBooleanTransform(typ, float64(startTime), float64(endTime))}
	}

	numSections := len(parameters) / arguments

	sectionTime := endTime - startTime

	if numSections > 2 {
		endTime = startTime + sectionTime*int64(numSections-1)
	}

	sections := make([][]float64, numSections)

	for i := 0; i < numSections; i++ {
		sections[i] = make([]float64, arguments)

		for j := 0; j < arguments; j++ {
			sections[i][j], err = strconv.ParseFloat(parameters[arguments*i+j], 64)
			checkError(err)

			if command == "C" {
				sections[i][j] /= 255
			}
		}
	}

	if numSections == 1 {
		sections = append(sections, sections[0])
	}

	for i := 0; i < bmath.MaxI(1, numSections-1); i++ {
		start := float64(startTime + int64(i)*sectionTime)
		end := float64(startTime + int64(i+1)*sectionTime)

		section := sections[i]
		nextSection := sections[i+1]

		switch command {
		case "F":
			transforms = append(transforms, animation.NewSingleTransform(animation.Fade, easeFunc, start, end, section[0], nextSection[0]))
		case "R":
			transforms = append(transforms, animation.NewSingleTransform(animation.Rotate, easeFunc, start, end, section[0], nextSection[0]))
		case "S":
			transforms = append(transforms, animation.NewSingleTransform(animation.Scale, easeFunc, start, end, section[0], nextSection[0]))
		case "V":
			transforms = append(transforms, animation.NewVectorTransform(animation.ScaleVector, easeFunc, start, end, section[0], section[1], nextSection[0], nextSection[1]))
		case "M":
			transforms = append(transforms, animation.NewVectorTransform(animation.Move, easeFunc, start, end, section[0], section[1], nextSection[0], nextSection[1]))
		case "MX":
			transforms = append(transforms, animation.NewSingleTransform(animation.MoveX, easeFunc, start, end, section[0], nextSection[0]))
		case "MY":
			transforms = append(transforms, animation.NewSingleTransform(animation.MoveY, easeFunc, start, end, section[0], nextSection[0]))
		case "C":
			color1 := color2.Color{
				R: float32(section[0]),
				G: float32(section[1]),
				B: float32(section[2]),
			}
			color2 := color2.Color{
				R: float32(nextSection[0]),
				G: float32(nextSection[1]),
				B: float32(nextSection[2]),
			}
			transforms = append(transforms, animation.NewColorTransform(animation.Color3, easeFunc, start, end, color1, color2))
		}
	}

	return transforms
}
