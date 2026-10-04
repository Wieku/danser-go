//go:build !windows

package gcontext

import "github.com/Zyko0/go-sdl3/sdl"

func SetSwapInterval(interval int) {
	if err := sdl.GL_SetSwapInterval(int32(interval)); err != nil {
		panic(err)
	}
}

func SwapBuffers() {
	if err := sdl.GL_SwapWindow(sdlWindow); err != nil {
		panic(err)
	}
}
