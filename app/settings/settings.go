package settings

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/wieku/danser-go/framework/files"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type defaultsFactory struct{}

var DefaultsFactory = &defaultsFactory{}

type Config struct {
	srcPath string
	srcData []byte

	General     *general     `icon:"\uF0AD"`
	Graphics    *graphics    `icon:"\uF108"`
	Audio       *audio       `icon:"\uF028"`
	Input       *input       `icon:"\uF11C"`
	Gameplay    *gameplay    `icon:"\uF140"`
	Skin        *skin        `icon:"\uF53F"`
	Cursor      *cursor      `icon:"\uF245"`
	Objects     *objects     `icon:"\uF1CD"`
	Playfield   *playfield   `icon:"\uF853"`
	Dance       *danceOld    `json:",omitempty" icon:"\uF5B7"`
	CursorDance *cursorDance `icon:"\uE599"`
	Knockout    *knockout    `icon:"\uF0CB"`
	Recording   *recording   `icon:"\uF03D"`
}

type CombinedConfig struct {
	Credentials *credentials `icon:"\uF084" label:"Credentials (Global)"`
	General     *general     `icon:"\uF0AD"`
	Graphics    *graphics    `icon:"\uF108"`
	Audio       *audio       `icon:"\uF028"`
	Input       *input       `icon:"\uF11C"`
	Gameplay    *gameplay    `icon:"\uF140"`
	Skin        *skin        `icon:"\uF53F"`
	Cursor      *cursor      `icon:"\uF245"`
	Objects     *objects     `icon:"\uF1CD"`
	Playfield   *playfield   `icon:"\uF853"`
	Dance       *danceOld    `json:",omitempty" icon:"\uF5B7"`
	CursorDance *cursorDance `icon:"\uE599"`
	Knockout    *knockout    `icon:"\uF0CB"`
	Recording   *recording   `icon:"\uF03D"`
}

func LoadConfig(file *os.File) (*Config, error) {
	log.Println(fmt.Sprintf(`SettingsManager: Loading "%s"`, file.Name()))

	data, err := io.ReadAll(files.NewUnicodeReader(file))
	if err != nil {
		return nil, fmt.Errorf("SettingsManager: Failed to read %s! Error: %s", file.Name(), err)
	}

	config := NewConfigFile()

	config.General.OsuReplaysDir = "" // Clear Replay path, so we can migrate it from Songs if JSON misses it

	config.srcPath = file.Name()
	config.srcData = data

	str := string(data)

	// I hope it won't backfire, replacing \ or \\\\\\\ with \\ so JSON can parse it as \

	str = regexp.MustCompile(`\\+`).ReplaceAllString(str, `\`)
	str = strings.ReplaceAll(str, `\`, `\\`)

	if err = json.Unmarshal([]byte(str), config); err != nil {
		return nil, fmt.Errorf("SettingsManager: Failed to parse %s! Please re-check the file for mistakes. Error: %s", file.Name(), err)
	}

	config.migrateCursorDance()

	if config.General.OsuReplaysDir == "" { // Set the replay directory if it hasn't been loaded
		config.General.OsuReplaysDir = filepath.Join(filepath.Dir(config.General.OsuSongsDir), "Replays")
	}

	log.Println(fmt.Sprintf(`SettingsManager: "%s" loaded!`, file.Name()))

	return config, nil
}

func NewConfigFile() *Config {
	return &Config{
		General:     initGeneral(),
		Graphics:    initGraphics(),
		Audio:       initAudio(),
		Input:       initInput(),
		Gameplay:    initGameplay(),
		Skin:        initSkin(),
		Cursor:      initCursor(),
		Objects:     initObjects(),
		Playfield:   initPlayfield(),
		CursorDance: initCursorDance(),
		Knockout:    initKnockout(),
		Recording:   initRecording(),
	}
}

func (config *Config) migrateCursorDance() {
	if config.Dance == nil {
		return
	}

	movers := make([]*mover, 0, len(config.Dance.Movers))
	spinners := make([]*spinner, 0, len(config.Dance.Spinners))

	for _, m := range config.Dance.Movers {
		movers = append(movers, &mover{
			Mover:             m,
			SliderDance:       config.Dance.SliderDance,
			RandomSliderDance: config.Dance.RandomSliderDance,
		})
	}

	for _, m := range config.Dance.Spinners {
		spinners = append(spinners, &spinner{
			Mover:  m,
			Radius: config.Dance.SpinnerRadius,
		})
	}

	config.CursorDance.Movers = movers
	config.CursorDance.Spinners = spinners

	config.CursorDance.Battle = config.Dance.Battle
	config.CursorDance.DoSpinnersTogether = config.Dance.DoSpinnersTogether
	config.CursorDance.TAGSliderDance = config.Dance.TAGSliderDance

	config.CursorDance.MoverSettings.Bezier = []*bezier{
		config.Dance.Bezier,
	}

	config.CursorDance.MoverSettings.Flower = []*flower{
		config.Dance.Flower,
	}

	config.CursorDance.MoverSettings.HalfCircle = []*circular{
		config.Dance.HalfCircle,
	}

	config.CursorDance.MoverSettings.Spline = []*spline{
		config.Dance.Spline,
	}

	config.CursorDance.MoverSettings.Momentum = []*momentum{
		config.Dance.Momentum,
	}

	config.CursorDance.MoverSettings.ExGon = []*exgon{
		config.Dance.ExGon,
	}

	config.Dance = nil
}

func (config *Config) attachToGlobals() {
	General = config.General
	Graphics = config.Graphics
	Audio = config.Audio
	Input = config.Input
	Gameplay = config.Gameplay
	Skin = config.Skin
	Cursor = config.Cursor
	Objects = config.Objects
	Playfield = config.Playfield
	CursorDance = config.CursorDance
	Knockout = config.Knockout
	Recording = config.Recording
}

func (config *Config) GetCombined() *CombinedConfig {
	return &CombinedConfig{
		Credentials: Credentails,
		General:     config.General,
		Graphics:    config.Graphics,
		Audio:       config.Audio,
		Input:       config.Input,
		Gameplay:    config.Gameplay,
		Skin:        config.Skin,
		Cursor:      config.Cursor,
		Objects:     config.Objects,
		Playfield:   config.Playfield,
		Dance:       config.Dance,
		CursorDance: config.CursorDance,
		Knockout:    config.Knockout,
		Recording:   config.Recording,
	}
}

func (config *Config) Save(path string, forceSave bool) {
	if strings.TrimSpace(path) == "" {
		path = config.srcPath
	}

	data, err := json.MarshalIndent(config, "", "\t")
	if err != nil {
		panic(err)
	}

	if forceSave || !bytes.Equal(data, config.srcData) { // Don't rewrite the file unless necessary
		log.Println(fmt.Sprintf(`SettingsManager: Saving settings to "%s"`, path))

		config.srcData = data

		if err = os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			panic(err)
		}

		if err = os.WriteFile(path, data, 0644); err != nil {
			panic(err)
		}

		config.srcPath = path
	}
}
