package settings

import (
	"os"
	"path/filepath"
	"runtime"
)

var General = initGeneral()

func initGeneral() *general {
	osuDir := ""
	if runtime.GOOS == "windows" {
		osuDir = filepath.Join(os.Getenv("localappdata"), "osu!", "Songs")
	} else {
		dir, _ := os.UserHomeDir()
		osuDir = filepath.Join(dir, ".osu", "Songs")
	}

	return &general{
		OsuSongsDir:       osuDir,
		DiscordPresenceOn: true,
	}
}

type general struct {

	// Directory that contains osu! songs,
	OsuSongsDir string

	// Whether discord should show that danser is on
	DiscordPresenceOn bool
}
