//go:build !windows

package platform

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

func ShowFileInManager(path string) {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("busctl", "--user", "org.freedesktop.FileManager1", "/org/freedesktop/FileManager1", "org.freedesktop.FileManager1", "ShowItems", "ass", "1", "file://"+path, "").Start()

		if err != nil {
			log.Println("Failed to run busctl: ", err)
			log.Println("Trying an alternative method...")

			if stat, err2 := os.Stat(path); err2 == nil && !stat.IsDir() {
				path = filepath.Dir(path)
			}

			err = exec.Command("xdg-open", path).Start()
		}
	case "darwin":
		err = exec.Command("open", "-R", path).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}

	if err != nil {
		panic(err)
	}
}
