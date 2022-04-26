package files

import (
	"github.com/wieku/danser-go/framework/env"
	"os/exec"
	"path/filepath"
)

func GetCommandExec(pkg, cmd string) (string, error) {
	if pkg != "" {
		if ex, err := exec.LookPath(filepath.Join(env.LibDir(), pkg, "bin", cmd)); err == nil {
			return ex, nil
		}

		if ex, err := exec.LookPath(filepath.Join(env.LibDir(), pkg, cmd)); err == nil {
			return ex, nil
		}
	}

	if ex, err := exec.LookPath(filepath.Join(env.LibDir(), cmd)); err == nil {
		return ex, nil
	}

	return exec.LookPath(cmd)
}
