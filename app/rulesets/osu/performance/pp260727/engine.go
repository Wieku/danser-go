package pp260727

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/rulesets/osu/performance/api"
)

type calculatorEngine struct {
	mu sync.Mutex

	beatmapPath string
	process     *exec.Cmd
	input       io.WriteCloser
	output      *bufio.Reader

	variants map[string]*loadedVariant
}

type loadedVariant struct {
	id         int
	attributes []api.Attributes
	strains    api.StrainPeaks
}

type engineResponse struct {
	OK     bool            `json:"ok"`
	Result json.RawMessage `json:"result"`
	Error  string          `json:"error"`
}

type externalAttributes struct {
	Stars                      float64 `json:"stars"`
	MaxCombo                   int     `json:"maxCombo"`
	Aim                        float64 `json:"aim"`
	Speed                      float64 `json:"speed"`
	Flashlight                 float64 `json:"flashlight"`
	SliderFactor               float64 `json:"sliderFactor"`
	SpeedNoteCount             float64 `json:"speedNoteCount"`
	AimDifficultSliderCount    float64 `json:"aimDifficultSliderCount"`
	AimTopWeightedSliderFactor float64 `json:"aimTopWeightedSliderFactor"`
	AimDifficultStrainCount    float64 `json:"aimDifficultStrainCount"`
	SpeedDifficultStrainCount  float64 `json:"speedDifficultStrainCount"`
	Reading                    float64 `json:"reading"`
	NCircles                   int     `json:"nCircles"`
	NSliders                   int     `json:"nSliders"`
	NSpinners                  int     `json:"nSpinners"`
	MaximumLegacyComboScore    int64   `json:"maximumLegacyComboScore"`
	NestedScorePerObject       float64 `json:"nestedScorePerObject"`
}

type externalStrains struct {
	Aim        []float64 `json:"aim"`
	Speed      []float64 `json:"speed"`
	Flashlight []float64 `json:"flashlight"`
	Total      []float64 `json:"total"`
}

type loadResult struct {
	ID         int                  `json:"id"`
	Attributes []externalAttributes `json:"attributes"`
	Strains    externalStrains      `json:"strains"`
}

type externalPerformance struct {
	PP         float64 `json:"pp"`
	Aim        float64 `json:"aim"`
	Speed      float64 `json:"speed"`
	Accuracy   float64 `json:"accuracy"`
	Flashlight float64 `json:"flashlight"`
	Reading    float64 `json:"reading"`
}

type performanceResult struct {
	Accuracy    float64             `json:"accuracy"`
	Performance externalPerformance `json:"performance"`
}

var currentEngine = &calculatorEngine{variants: make(map[string]*loadedVariant)}

func (engine *calculatorEngine) ensureVariantForPath(path string, diff *difficulty.Difficulty) (*loadedVariant, error) {
	absolutePath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	engine.mu.Lock()
	defer engine.mu.Unlock()

	engine.setBeatmapPathLocked(absolutePath)
	return engine.ensureVariantLocked(diff)
}

func (engine *calculatorEngine) setBeatmapPathLocked(path string) {
	path = filepath.Clean(path)
	if engine.beatmapPath != path {
		if engine.process != nil {
			var result map[string]any
			if err := engine.call(map[string]any{"command": "clear"}, &result); err != nil {
				log.Printf("Failed to clear current PP runtime cache: %v", err)
			}
		}
		engine.beatmapPath = path
		clear(engine.variants)
	}
}

func (engine *calculatorEngine) variantKey(diff *difficulty.Difficulty) (string, error) {
	mods, err := json.Marshal(diff.ExportMods2())
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s\x00%t\x00%s", engine.beatmapPath, diff.CheckModActive(difficulty.Lazer), mods), nil
}

func (engine *calculatorEngine) ensureVariant(diff *difficulty.Difficulty) (*loadedVariant, error) {
	engine.mu.Lock()
	defer engine.mu.Unlock()

	return engine.ensureVariantLocked(diff)
}

func (engine *calculatorEngine) ensureVariantLocked(diff *difficulty.Difficulty) (*loadedVariant, error) {
	if engine.beatmapPath == "" {
		return nil, errors.New("current PP calculator has no beatmap path")
	}

	key, err := engine.variantKey(diff)
	if err != nil {
		return nil, err
	}

	if variant := engine.variants[key]; variant != nil {
		return variant, nil
	}

	var result loadResult
	err = engine.call(map[string]any{
		"command": "load",
		"path":    engine.beatmapPath,
		"mods":    diff.ExportMods2(),
		"isLazer": diff.CheckModActive(difficulty.Lazer),
	}, &result)
	if err != nil {
		return nil, err
	}

	attributes := make([]api.Attributes, len(result.Attributes))
	for i, attr := range result.Attributes {
		attributes[i] = api.Attributes{
			Total:                      attr.Stars,
			Aim:                        attr.Aim,
			AimNoSliders:               attr.Aim * attr.SliderFactor,
			Speed:                      attr.Speed,
			Reading:                    attr.Reading,
			SpeedNoteCount:             attr.SpeedNoteCount,
			AimDifficultStrainCount:    attr.AimDifficultStrainCount,
			AimDifficultSliderCount:    attr.AimDifficultSliderCount,
			AimTopWeightedSliderFactor: attr.AimTopWeightedSliderFactor,
			SpeedDifficultStrainCount:  attr.SpeedDifficultStrainCount,
			Flashlight:                 attr.Flashlight,
			SliderFactor:               attr.SliderFactor,
			ObjectCount:                attr.NCircles + attr.NSliders + attr.NSpinners,
			Circles:                    attr.NCircles,
			Sliders:                    attr.NSliders,
			Spinners:                   attr.NSpinners,
			MaxCombo:                   attr.MaxCombo,
			MaximumLegacyComboScore:    attr.MaximumLegacyComboScore,
			NestedScorePerObject:       attr.NestedScorePerObject,
		}
	}

	variant := &loadedVariant{
		id:         result.ID,
		attributes: attributes,
		strains: api.StrainPeaks{
			Aim:        result.Strains.Aim,
			Speed:      result.Strains.Speed,
			Flashlight: result.Strains.Flashlight,
			Total:      result.Strains.Total,
		},
	}
	engine.variants[key] = variant

	return variant, nil
}

func (engine *calculatorEngine) calculate(diff *difficulty.Difficulty, index int, score map[string]any) (performanceResult, error) {
	variant, err := engine.ensureVariant(diff)
	if err != nil {
		return performanceResult{}, err
	}

	engine.mu.Lock()
	defer engine.mu.Unlock()

	var result performanceResult
	err = engine.call(map[string]any{
		"command": "performance",
		"id":      variant.id,
		"index":   index,
		"score":   score,
	}, &result)

	return result, err
}

func (engine *calculatorEngine) call(request any, result any) error {
	if engine.process == nil {
		if err := engine.start(); err != nil {
			return err
		}
	}

	data, err := json.Marshal(request)
	if err != nil {
		return err
	}

	data = append(data, '\n')
	if _, err = engine.input.Write(data); err != nil {
		engine.stop()
		return err
	}

	line, err := engine.output.ReadBytes('\n')
	if err != nil {
		engine.stop()
		return err
	}

	var response engineResponse
	if err = json.Unmarshal(line, &response); err != nil {
		return err
	}

	if !response.OK {
		return errors.New(response.Error)
	}

	return json.Unmarshal(response.Result, result)
}

func (engine *calculatorEngine) start() error {
	runtimeDir, nodePath, serverPath, err := findRuntime()
	if err != nil {
		return err
	}

	command := exec.Command(nodePath, serverPath)
	command.Dir = runtimeDir
	command.Stderr = os.Stderr

	input, err := command.StdinPipe()
	if err != nil {
		return err
	}

	output, err := command.StdoutPipe()
	if err != nil {
		return err
	}

	if err = command.Start(); err != nil {
		return err
	}

	engine.process = command
	engine.input = input
	engine.output = bufio.NewReader(output)
	log.Printf("Started osu! difficulty/PP runtime from %s", runtimeDir)

	return nil
}

func (engine *calculatorEngine) stop() {
	if engine.input != nil {
		_ = engine.input.Close()
	}
	if engine.process != nil && engine.process.Process != nil {
		_ = engine.process.Process.Kill()
		_, _ = engine.process.Process.Wait()
	}

	engine.process = nil
	engine.input = nil
	engine.output = nil
}

func findRuntime() (runtimeDir, nodePath, serverPath string, err error) {
	serverName := "pp-server.cjs"
	nodeName := "node"
	if runtime.GOOS == "windows" {
		nodeName = "node.exe"
	}

	candidates := make([]string, 0, 4)
	if configured := os.Getenv("DANSER_PP_RUNTIME"); configured != "" {
		candidates = append(candidates, configured)
	}
	if executable, executableErr := os.Executable(); executableErr == nil {
		candidates = append(candidates, filepath.Join(filepath.Dir(executable), "pp-runtime"))
	}
	if workingDirectory, workingErr := os.Getwd(); workingErr == nil {
		candidates = append(candidates,
			filepath.Join(workingDirectory, "pp-runtime"),
			filepath.Join(workingDirectory, "..", "pp-runtime"),
		)
	}

	for _, candidate := range candidates {
		candidate, _ = filepath.Abs(candidate)
		server := filepath.Join(candidate, serverName)
		node := filepath.Join(candidate, nodeName)
		if fileExists(server) && fileExists(node) {
			return candidate, node, server, nil
		}
	}

	if node, lookupErr := exec.LookPath("node"); lookupErr == nil {
		for _, candidate := range candidates {
			candidate, _ = filepath.Abs(candidate)
			server := filepath.Join(candidate, serverName)
			if fileExists(server) {
				return candidate, node, server, nil
			}
		}
	}

	return "", "", "", errors.New("current PP runtime not found; expected pp-runtime/node and pp-runtime/pp-server.cjs")
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
