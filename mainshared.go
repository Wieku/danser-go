package main

import "C"

import (
	"github.com/wieku/danser-go/app"
	"github.com/wieku/danser-go/framework/env"
	"github.com/wieku/danser-go/launcher"
	"os"
)

//export danserMain
func danserMain(isLauncher bool, args []string) {
	os.Args = args

	env.Init("danser")
	if isLauncher {
		launcher.StartLauncher()
	} else {
		app.Run()
	}
}
