package settings

import (
	"os"
	"path/filepath"
	"runtime"
)

var General = initGeneral()

func initGeneral() *general {
	osuBaseDir := ""
	if runtime.GOOS == "windows" {
		osuBaseDir = filepath.Join(os.Getenv("localappdata"), "osu!")
	} else {
		dir, _ := os.UserHomeDir()
		osuBaseDir = filepath.Join(dir, ".osu")
	}

	return &general{
		OsuSongsDir:       filepath.Join(osuBaseDir, "Songs"),
		OsuSkinsDir:       filepath.Join(osuBaseDir, "Skins"),
		DiscordPresenceOn: true,
		UnpackOszFiles:    true,
	}
}

type general struct {

	// Directory that contains osu! songs,
	OsuSongsDir string

	// Directory that contains osu! skins,
	OsuSkinsDir string

	// Whether discord should show that danser is on
	DiscordPresenceOn bool

	// Whether danser should unpack .osz files in Songs folder, osu! may complain about it
	UnpackOszFiles bool
}
