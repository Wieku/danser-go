package settings

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/itchio/lzma"
	"github.com/wieku/danser-go/framework/files"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type defaultsFactory struct{}

var DefaultsFactory = &defaultsFactory{}

type Config struct {
	srcPath string
	srcData []byte

	General     *general     `icon:"\uF0AD"`                   // wrench
	Graphics    *graphics    `icon:"\uE163"  liveedit:"false"` // display
	Audio       *audio       `icon:"\uF028"`                   // volume-high
	Input       *input       `icon:"\uF11C"`                   // keyboard
	Gameplay    *gameplay    `icon:"\uF192"`                   // circle-dot
	Skin        *skin        `icon:"\uF1FC"`                   // paintbrush
	Cursor      *cursor      `icon:"\uF245"`                   // arrow-pointer
	Objects     *objects     `icon:"\uF1E0"`                   // share-nodes
	Playfield   *playfield   `icon:"\uF43C"`                   // chess-board
	CursorDance *cursorDance `icon:"\uE599"`                   // worm
	Knockout    *knockout    `icon:"\uF0CB"`                   // list-ol
	Recording   *recording   `icon:"\uF03D"`                   // video
	Debug       *debug       `icon:"\uF188"`                   // bug
	Dance       *danceOld    `json:",omitempty" icon:"\uF5B7"`
}

type CombinedConfig struct {
	Credentials *credentials `icon:"\uF084" label:"Credentials (Global)" liveedit:"false"` // key
	General     *general     `icon:"\uF0AD" liveedit:"false"`                              // wrench
	Graphics    *graphics    `icon:"\uE163"`                                               // display
	Audio       *audio       `icon:"\uF028"`                                               // volume-high
	Input       *input       `icon:"\uF11C"`                                               // keyboard
	Gameplay    *gameplay    `icon:"\uF192"`                                               // circle-dot
	Skin        *skin        `icon:"\uF1FC"`                                               // paintbrush
	Cursor      *cursor      `icon:"\uF245"`                                               // arrow-pointer
	Objects     *objects     `icon:"\uF1E0"`                                               // share-nodes
	Playfield   *playfield   `icon:"\uF43C"`                                               // chess-board
	CursorDance *cursorDance `icon:"\uE599"`                                               // worm
	Knockout    *knockout    `icon:"\uF0CB"`                                               // list-ol
	Recording   *recording   `icon:"\uF03D"`                                               // video
	Debug       *debug       `icon:"\uF188"`                                               // bug
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

	if err = json.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("SettingsManager: Failed to parse %s! Please re-check the file for mistakes. Error: %s", file.Name(), err)
	}

	config.migrateCursorDance()
	config.migrateHitCounterColors()
	config.migrateBlendWeights()

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
		Debug:       initDebug(),
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

	if config.Dance.Bezier != nil {
		config.CursorDance.MoverSettings.Bezier = []*bezier{
			config.Dance.Bezier,
		}
	}

	if config.Dance.Flower != nil {
		config.CursorDance.MoverSettings.Flower = []*flower{
			config.Dance.Flower,
		}
	}

	if config.Dance.HalfCircle != nil {
		config.CursorDance.MoverSettings.HalfCircle = []*circular{
			config.Dance.HalfCircle,
		}
	}

	if config.Dance.Spline != nil {
		config.CursorDance.MoverSettings.Spline = []*spline{
			config.Dance.Spline,
		}
	}

	if config.Dance.Momentum != nil {
		config.CursorDance.MoverSettings.Momentum = []*momentum{
			config.Dance.Momentum,
		}
	}

	if config.Dance.ExGon != nil {
		config.CursorDance.MoverSettings.ExGon = []*exgon{
			config.Dance.ExGon,
		}
	}

	config.Dance = nil
}

func (config *Config) migrateHitCounterColors() {
	if config.Gameplay.HitCounter.Color == nil {
		return
	}

	idx := 0

	ln := len(config.Gameplay.HitCounter.Color)

	if config.Gameplay.HitCounter.Show300 {
		config.Gameplay.HitCounter.Color300 = config.Gameplay.HitCounter.Color[idx%ln]
		idx++
	}

	config.Gameplay.HitCounter.Color100 = config.Gameplay.HitCounter.Color[idx%ln]
	idx++

	config.Gameplay.HitCounter.Color50 = config.Gameplay.HitCounter.Color[idx%ln]
	idx++

	config.Gameplay.HitCounter.ColorMiss = config.Gameplay.HitCounter.Color[idx%ln]
	idx++

	if config.Gameplay.HitCounter.ShowSliderBreaks {
		config.Gameplay.HitCounter.ColorSB = config.Gameplay.HitCounter.Color[idx%ln]
		idx++
	}

	config.Gameplay.HitCounter.Color = nil
}

func (config *Config) migrateBlendWeights() {
	if config.Recording.MotionBlur.BlendWeights == nil {
		return
	}

	config.Recording.MotionBlur.BlendFunctionID = config.Recording.MotionBlur.BlendWeights.AutoWeightsID
	config.Recording.MotionBlur.GaussWeightsMult = config.Recording.MotionBlur.BlendWeights.GaussWeightsMult

	config.Recording.MotionBlur.BlendWeights = nil
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
	Debug = config.Debug
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
		CursorDance: config.CursorDance,
		Knockout:    config.Knockout,
		Recording:   config.Recording,
		Debug:       config.Debug,
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

func (config *Config) GetCompressedString() string {
	data, err := json.MarshalIndent(config, "", "\t")
	if err != nil {
		panic(err)
	}

	buf := new(bytes.Buffer)

	writer := lzma.NewWriter(buf)

	_, _ = writer.Write(data)
	_ = writer.Close()

	return base64.StdEncoding.EncodeToString(buf.Bytes())
}
