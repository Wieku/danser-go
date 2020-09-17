package storyboard

import (
	"github.com/wieku/danser-go/framework/math/animation"
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
				transforms = append(transforms, NewCommand(command).GenerateTransformations()...)
			}
		}

		if command[0] == "L" {

			currentLoop = NewLoopProcessor(command)

			loopDepth = removed + 1

		} else if removed == loopDepth {
			currentLoop.Add(NewCommand(command))
		}
	}

	if currentLoop != nil {
		transforms = append(transforms, currentLoop.Unwind()...)

		currentLoop = nil
		loopDepth = -1
	}

	return transforms
}
