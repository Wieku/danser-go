package env

import (
	"github.com/wieku/danser-go/framework/platform"
	"os"
	"path/filepath"
	"strings"
)

var dataDir string
var configDir string
var libDir string

var initialized bool

func Init(pkgName string) {
	execPath := platform.GetExecDir()

	execPathLower := strings.ToLower(execPath)

	if strings.HasPrefix(execPathLower, "/usr/bin") || strings.HasPrefix(execPathLower, "/usr/lib") { //if pkgName is a package
		homePath, err := os.UserHomeDir()
		if err != nil {
			panic(err)
		}

		dataDir = filepath.Join(homePath, ".local", "share", pkgName)
		if env := strings.TrimSpace(os.Getenv("XDG_DATA_HOME")); env != "" {
			dataDir = filepath.Join(env, pkgName)
		}

		configDir = filepath.Join(homePath, ".config", pkgName)
		if env := strings.TrimSpace(os.Getenv("XDG_CONFIG_HOME")); env != "" {
			configDir = filepath.Join(env, pkgName)
		}

		if err = os.MkdirAll(dataDir, 0755); err != nil {
			panic(err)
		}

		if err = os.MkdirAll(configDir, 0755); err != nil {
			panic(err)
		}

		libDir = "/usr/lib/" + pkgName
	} else {
		libDir = execPath
		dataDir = execPath
		configDir = filepath.Join(execPath, "settings")

		if err := os.MkdirAll(configDir, 0755); err != nil {
			panic(err)
		}
	}

	initialized = true
}

func DataDir() string {
	if !initialized {
		panic("Environment not initialized")
	}

	return dataDir
}

func ConfigDir() string {
	if !initialized {
		panic("Environment not initialized")
	}

	return configDir
}

func LibDir() string {
	if !initialized {
		panic("Environment not initialized")
	}

	return libDir
}
