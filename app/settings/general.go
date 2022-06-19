package settings

import (
	"github.com/wieku/danser-go/framework/env"
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
		OsuReplaysDir:     filepath.Join(osuBaseDir, "Replays"),
		DiscordPresenceOn: true,
		UnpackOszFiles:    true,
		VerboseImportLogs: false,
	}
}

type general struct {
	// Directory that contains osu! songs
	OsuSongsDir string `long:"true" label:"osu! Songs directory" path:"Select osu! Songs directory"`

	// Directory that contains osu! skins
	OsuSkinsDir string `long:"true" label:"osu! Skins directory" path:"Select osu! Skins directory"`

	// Directory that contains osu! replays
	OsuReplaysDir string `long:"true" label:"osu! Replays directory" path:"Select osu! Replays directory" tooltip:"Don't use replays directory inside danser's directory!"`

	// Whether discord should show that danser is on
	DiscordPresenceOn bool `label:"Discord Rich Presence"`

	// Whether danser should unpack .osz files in Songs folder, osu! may complain about it
	UnpackOszFiles bool

	// Whether import details should be shown. If false, only failures will be logged.
	VerboseImportLogs bool

	songsDir   *string
	skinsDir   *string
	replaysDir *string
}

func (g *general) GetSongsDir() string {
	if g.songsDir == nil {
		dir := filepath.Join(env.DataDir(), g.OsuSongsDir)

		if filepath.IsAbs(g.OsuSongsDir) {
			dir = g.OsuSongsDir
		}

		g.songsDir = &dir
	}

	return *g.songsDir
}

func (g *general) GetSkinsDir() string {
	if g.skinsDir == nil {
		dir := filepath.Join(env.DataDir(), g.OsuSkinsDir)

		if filepath.IsAbs(g.OsuSkinsDir) {
			dir = g.OsuSkinsDir
		}

		g.skinsDir = &dir
	}

	return *g.skinsDir
}

func (g *general) GetReplaysDir() string {
	if g.replaysDir == nil {
		dir := filepath.Join(env.DataDir(), g.OsuReplaysDir)

		if filepath.IsAbs(g.OsuReplaysDir) {
			dir = g.OsuReplaysDir
		}

		g.replaysDir = &dir
	}

	return *g.replaysDir
}
