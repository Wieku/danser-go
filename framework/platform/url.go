package platform

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

func OpenURL(url string) {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}

	if err != nil {
		panic(err)
	}
}

func ShowFileInManager(path string) {
	var err error

	switch runtime.GOOS {
	case "linux":
		if stat, err2 := os.Stat(path); err2 == nil && !stat.IsDir() {
			path = filepath.Dir(path)
		}

		err = exec.Command("xdg-open", path).Start()
	case "windows":
		err = exec.Command("explorer", "/select,", path).Start()
	case "darwin":
		err = exec.Command("open", "-R", path).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}

	if err != nil {
		panic(err)
	}
}
