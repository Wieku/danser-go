package launcher

import (
	"encoding/json"
	"github.com/wieku/danser-go/framework/env"
	"github.com/wieku/danser-go/framework/files"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var launcherConfig = &launcherConf{
	Profile:          nil,
	CheckForUpdates:  true,
	ShowFileAfter:    true,
	PreviewSelected:  true,
	PreviewVolume:    0.25,
	SortMapsBy:       Title,
	SortAscending:    true,
	LoadLatestReplay: false,
	SkipMapUpdate:    false,
	AutoRefreshDB:    false,
	ShowJSONPaths:    false,
}

type launcherConf struct {
	Profile          *string
	CheckForUpdates  bool
	ShowFileAfter    bool
	PreviewSelected  bool
	PreviewVolume    float64
	SortMapsBy       SortBy
	SortAscending    bool
	LoadLatestReplay bool
	SkipMapUpdate    bool
	AutoRefreshDB    bool
	ShowJSONPaths    bool
	LastKnockoutPath string
}

func loadLauncherConfig() {
	cPath := filepath.Join(env.ConfigDir(), "launcher.json")

	if file, err := os.Open(cPath); err == nil {
		data, err := io.ReadAll(files.NewUnicodeReader(file))
		if err != nil {
			panic(err)
		}

		if err = json.Unmarshal(data, launcherConfig); err != nil {
			panic(err)
		}
	}

	if launcherConfig.Profile == nil {
		lastModified := time.UnixMilli(0)

		filepath.Walk(env.ConfigDir(), func(path string, info fs.FileInfo, err error) error {
			if !info.IsDir() && strings.HasSuffix(path, ".json") {
				stPath := strings.ReplaceAll(strings.TrimPrefix(strings.TrimSuffix(path, ".json"), env.ConfigDir()+string(os.PathSeparator)), "\\", "/")

				if stPath != "credentials" && stPath != "default" && stPath != "launcher" {
					if info.ModTime().After(lastModified) {
						lastModified = info.ModTime()
						launcherConfig.Profile = &stPath
					}
					info.ModTime()
				}
			}

			return nil
		})
	}

	if launcherConfig.Profile == nil {
		def := "default"
		launcherConfig.Profile = &def
	}

	saveLauncherConfig()
}

func saveLauncherConfig() {
	data, err := json.MarshalIndent(launcherConfig, "", "\t")
	if err != nil {
		panic(err)
	}

	if err = os.MkdirAll(env.ConfigDir(), 0755); err != nil {
		panic(err)
	}

	if err = os.WriteFile(filepath.Join(env.ConfigDir(), "launcher.json"), data, 0644); err != nil {
		panic(err)
	}
}
