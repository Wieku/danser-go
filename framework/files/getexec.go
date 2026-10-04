package files

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"syscall"

	"github.com/wieku/danser-go/framework/env"
)

func GetCommandExec(pkg, cmd string) (string, error) {
	if pkg != "" {
		if ex, err := getLocalExec(filepath.Join(env.LibDir(), pkg, "bin", cmd)); err == nil {
			return ex, nil
		}

		if ex, err := getLocalExec(filepath.Join(env.LibDir(), pkg, cmd)); err == nil {
			return ex, nil
		}
	}

	if ex, err := getLocalExec(filepath.Join(env.LibDir(), cmd)); err == nil {
		return ex, nil
	}

	return exec.LookPath(cmd)
}

func getLocalExec(path string) (string, error) {
	if runtime.GOOS == "windows" {
		return exec.LookPath(path)
	}

	// For linux we're not using LookPath because it checks for execute permissions

	d, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	m := d.Mode()
	if m.IsDir() {
		return "", syscall.EISDIR
	}

	return path, nil
}
