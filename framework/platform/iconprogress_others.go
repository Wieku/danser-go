//go:build !windows

package platform

import "github.com/go-gl/glfw/v3.3/glfw"

func StopProgress(_ *glfw.Window) {}

func StartProgress(_ *glfw.Window) {}

func SetProgress(_ *glfw.Window, _, _ int) {}
