package settings

import (
	"io/ioutil"
	"path/filepath"
	"sort"
	"strings"
)

var Skin = initSkin()

func initSkin() *skin {
	return &skin{
		CurrentSkin:       "default",
		FallbackSkin:      "default",
		UseColorsFromSkin: false,
		UseBeatmapColors:  false,
		Cursor: &skinCursor{
			UseSkinCursor:    false,
			Scale:            1.0,
			TrailScale:       1.0,
			ForceLongTrail:   false,
			LongTrailLength:  2048,
			LongTrailDensity: 1.0,
		},
	}
}

type skin struct {
	CurrentSkin       string `combo:"true" comboSrc:"SkinOptions" search:"true"`
	FallbackSkin      string `combo:"true" comboSrc:"SkinOptions" search:"true"`
	UseColorsFromSkin bool
	UseBeatmapColors  bool

	Cursor *skinCursor
}

var skinPath string
var skinCache []string

func (d *defaultsFactory) SkinOptions() []string {
	if General.GetSkinsDir() != skinPath {
		skinPath = General.GetSkinsDir()

		skinCache = []string{}

		fs, err := ioutil.ReadDir(skinPath)
		if err == nil {
			for _, f := range fs {
				if f.IsDir() {
					skinCache = append(skinCache, filepath.Base(f.Name()))
				}
			}

			sort.Slice(skinCache, func(i, j int) bool {
				return strings.ToLower(skinCache[i]) < strings.ToLower(skinCache[j])
			})
		}

		skinCache = append([]string{"default"}, skinCache...)
	}

	return skinCache
}

type skinCursor struct {
	UseSkinCursor    bool
	Scale            float64 `max:"2"`
	TrailScale       float64 `max:"2"`
	ForceLongTrail   bool
	LongTrailLength  int64   `max:"10000"`
	LongTrailDensity float64 `min:"0.1" max:"3"`
}
