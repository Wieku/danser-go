package utils

import (
	"github.com/wieku/danser-go/app/input"
	"github.com/wieku/danser-go/framework/goroutines"
	"os"
	"os/exec"
)

func QuickRestart() {
	danserPath := os.Args[0]

	arguments := make([]string, 0)

	noDbCheck := false
	quickStart := false

	for _, arg := range os.Args[1:] {
		if arg == "-nodbcheck" {
			noDbCheck = true
		}

		if arg == "-quickstart" {
			quickStart = true
		}

		arguments = append(arguments, arg)
	}

	if !noDbCheck {
		arguments = append(arguments, "-nodbcheck")
	}

	if !quickStart {
		arguments = append(arguments, "-quickstart")
	}

	cmd := exec.Command(danserPath, arguments...)
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Start()

	goroutines.CallNonBlockMain(func() {
		input.Win.SetShouldClose(true)
	})
}
