package storyboard

import (
	"github.com/wieku/danser-go/framework/math/animation"
	"github.com/wieku/danser-go/framework/math/animation/easing"
	color2 "github.com/wieku/danser-go/framework/math/color"
	"github.com/wieku/danser-go/framework/math/mutils"
	"log"
	"math"
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
				if parsed := parseCommand(command); parsed != nil {
					transforms = append(transforms, parsed...)
				}
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
	}

	return transforms
}

func parseCommand(data []string) []*animation.Transformation {
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

	startTime, err := strconv.ParseFloat(data[2], 64)
	checkError(err)

	endTime := -math.MaxFloat64

	if data[3] != "" {
		endTime, err = strconv.ParseFloat(data[3], 64)
		checkError(err)
	}

	endTime = math.Max(endTime, startTime)

	var arguments int

	switch command {
	case "P":
		arguments = 0
	case "F", "R", "S", "MX", "MY":
		arguments = 1
	case "M", "V":
		arguments = 2
	case "C":
		arguments = 3
	default:
		return nil
	}

	parameters := data[4:]

	if arguments == 0 {
		switch parameters[0] {
		case "H":
			return []*animation.Transformation{animation.NewBooleanTransform(animation.HorizontalFlip, startTime, endTime)}
		case "V":
			return []*animation.Transformation{animation.NewBooleanTransform(animation.VerticalFlip, startTime, endTime)}
		case "A":
			return []*animation.Transformation{animation.NewBooleanTransform(animation.Additive, startTime, endTime)}
		}

		return nil
	}

	numSections := len(parameters) / arguments

	sectionTime := endTime - startTime

	sections := make([][]float64, numSections)

	for i := 0; i < numSections; i++ {
		sections[i] = make([]float64, arguments)

		for j := 0; j < arguments; j++ {
			sections[i][j], err = strconv.ParseFloat(parameters[arguments*i+j], 64)
			checkError(err)
		}
	}

	if numSections == 1 {
		sections = append(sections, sections[0])
	}

	var transforms []*animation.Transformation

	for i := 0; i < mutils.MaxI(1, numSections-1); i++ {
		start := startTime + float64(i)*sectionTime
		end := startTime + float64(i+1)*sectionTime

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
			col1 := color2.NewRGB(float32(section[0]/255), float32(section[1]/255), float32(section[2]/255))
			col2 := color2.NewRGB(float32(nextSection[0]/255), float32(nextSection[1]/255), float32(nextSection[2]/255))
			transforms = append(transforms, animation.NewColorTransform(animation.Color3, easeFunc, start, end, col1, col2))
		}
	}

	return transforms
}
