package env

import (
	"os"
	"path/filepath"
	"strings"
)

var dataDir string
var configDir string
var libDir string

var initialized bool

func Init(pkgName string) {
	execPath := GetExecDir()

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

func GetExecPath() string {
	exec, err := os.Executable()
	if err != nil {
		panic(err)
	}

	if exec, err = filepath.EvalSymlinks(exec); err != nil {
		panic(err)
	}

	return exec
}

func GetExecDir() string {
	return filepath.Dir(GetExecPath())
}
