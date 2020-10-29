package dance

import (
	"github.com/Mempler/rplpa"
	"github.com/thehowl/go-osuapi"
	"github.com/wieku/danser-go/app/beatmap"
	"github.com/wieku/danser-go/app/bmath"
	"github.com/wieku/danser-go/app/bmath/difficulty"
	"github.com/wieku/danser-go/app/graphics"
	"github.com/wieku/danser-go/app/rulesets/osu"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/framework/math/math32"
	"github.com/wieku/danser-go/framework/math/vector"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

type RpData struct {
	Name     string
	Mods     string
	ModsV    difficulty.Modifier
	Accuracy float64
	Combo    int64
	MaxCombo int64
	Grade    osu.Grade
}

type subControl struct {
	danceController Controller
	replayIndex     int
	replayTime      int64
	frames          []*rplpa.ReplayData
}

func NewSubControl() *subControl {
	control := new(subControl)
	return control
}

type ReplayController struct {
	bMap        *beatmap.BeatMap
	replays     []RpData
	cursors     []*graphics.Cursor
	controllers []*subControl
	ruleset     *osu.OsuRuleSet
	lastTime    int64
	//counter int64
}

func NewReplayController() Controller {
	return new(ReplayController)
}

func getReplay(scoreID int64) ([]byte, error) {
	values := url.Values{}
	values.Add("c", strconv.FormatInt(scoreID, 10))
	values.Add("m", "0")
	values.Add("u", settings.Knockout.Username)
	values.Add("h", settings.Knockout.MD5Pass)
	request, err := http.NewRequest(http.MethodGet, "https://osu.ppy.sh/web/osu-getreplay.php?"+values.Encode(), nil)

	if err != nil {
		return nil, err
	}

	client := new(http.Client)

	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}

	return ioutil.ReadAll(response.Body)
}

func (controller *ReplayController) SetBeatMap(beatMap *beatmap.BeatMap) {
	replayDir := filepath.Join("replays", beatMap.MD5)

	err := os.MkdirAll(replayDir, os.ModeDir)
	if err != nil {
		panic(err)
	}

	counter := settings.Knockout.MaxPlayers

	if settings.Knockout.LocalReplays {
		filepath.Walk(replayDir, func(path string, f os.FileInfo, err error) error {
			if strings.HasSuffix(f.Name(), ".osr") {
				if counter == 0 {
					return nil
				}

				data, err := ioutil.ReadFile(path)
				if err != nil {
					panic(err)
				}

				replayD, _ := rplpa.ParseReplay(data)

				log.Println("Loading replay for:", replayD.Username)

				control := NewSubControl()

				loadFrames(control, replayD.ReplayData)

				mxCombo := replayD.MaxCombo

				controller.replays = append(controller.replays, RpData{replayD.Username + string(unicode.MaxRune), difficulty.Modifier(replayD.Mods).String(), difficulty.Modifier(replayD.Mods), 100, 0, int64(mxCombo), osu.NONE})
				controller.controllers = append(controller.controllers, control)

				log.Println("Expected score:", replayD.Score)
				log.Println("Expected pp:", math.NaN())
				log.Println("Replay loaded!")
				counter--
			}
			return nil
		})
	}

	if settings.Knockout.OnlineReplays && counter > 0 {
		client := osuapi.NewClient(settings.Knockout.ApiKey)
		beatMapO, err := client.GetBeatmaps(osuapi.GetBeatmapsOpts{BeatmapHash: beatMap.MD5})
		if len(beatMapO) == 0 || err != nil {
			log.Println("Online beatmap not found")
			if err != nil {
				log.Println(err)
			}

			goto skip
		}

		scores, err := client.GetScores(osuapi.GetScoresOpts{BeatmapID: beatMapO[0].BeatmapID, Limit: 200})
		if len(scores) == 0 || err != nil {
			log.Println("Can't find online scores")
			if err != nil {
				log.Println(err)
			}

			goto skip
		}

		sort.SliceStable(scores, func(i, j int) bool {
			return scores[i].Score.Score > scores[j].Score.Score
		})

		excludedMods := osuapi.ParseMods(settings.Knockout.ExcludeMods)

		for _, score := range scores {
			if counter == 0 {
				break
			}

			if score.Mods&excludedMods > 0 {
				continue
			}

			fileName := filepath.Join(replayDir, strconv.FormatInt(score.ScoreID, 10)+".dsr")

			file, err := os.Open(fileName)
			if file != nil {
				file.Close()
			}

			if os.IsNotExist(err) {
				data, err := getReplay(score.ScoreID)
				if err != nil {
					panic(err)
				} else if len(data) == 0 {
					log.Println("Replay for:", score.Username, "doesn't exist. Skipping...")
					continue
				} else {
					log.Println("Downloaded replay for:", score.Username)
				}

				err = ioutil.WriteFile(fileName, data, 0644)
				if err != nil {
					panic(err)
				}
			}

			log.Println("Loading replay for:", score.Username)

			control := NewSubControl()

			data, err := ioutil.ReadFile(fileName)
			if err != nil {
				panic(err)
			}

			replay, _ := rplpa.ParseCompressed(data)

			loadFrames(control, replay)

			mxCombo := score.MaxCombo

			controller.replays = append(controller.replays, RpData{score.Username, strings.Replace(strings.Replace(score.Mods.String(), "NF", "NF", -1), "NV", "TD", -1), difficulty.Modifier(score.Mods), 100, 0, int64(mxCombo), osu.NONE})
			controller.controllers = append(controller.controllers, control)

			log.Println("Expected score:", score.Score.Score)
			log.Println("Expected pp:", score.PP)
			log.Println("Replay loaded!")

			counter--
		}
	}

skip:

	if settings.Knockout.AddDanser || counter == settings.Knockout.MaxPlayers {
		control := NewSubControl()

		control.danceController = NewGenericController()
		control.danceController.SetBeatMap(beatMap)

		controller.replays = append([]RpData{{settings.Knockout.DanserName, "", difficulty.None, 100, 0, 0, osu.NONE}}, controller.replays...)
		controller.controllers = append([]*subControl{control}, controller.controllers...)
	}

	settings.PLAYERS = len(controller.replays)

	controller.bMap = beatMap
	controller.lastTime = -200
}

func loadFrames(subController *subControl, frames []*rplpa.ReplayData) {
	maniaFrameIndex := 0
	for i, frame := range frames {
		if frame.Time == -12345 {
			maniaFrameIndex = i
			break
		}
	}

	frames = append(frames[:maniaFrameIndex], frames[maniaFrameIndex+1:]...)

	subController.frames = frames
}

func (controller *ReplayController) InitCursors() {
	var modifiers []difficulty.Modifier
	for i := range controller.controllers {
		if controller.controllers[i].danceController != nil {
			controller.controllers[i].danceController.InitCursors()
			controller.controllers[i].danceController.GetCursors()[0].IsPlayer = true

			cursors := controller.controllers[i].danceController.GetCursors()

			for _, cursor := range cursors {
				cursor.Name = controller.replays[i].Name
			}

			controller.cursors = append(controller.cursors, cursors...)
		} else {
			cursor := graphics.NewCursor()
			cursor.Name = controller.replays[i].Name
			controller.cursors = append(controller.cursors, cursor)
		}

		modifiers = append(modifiers, controller.replays[i].ModsV)
	}
	controller.ruleset = osu.NewOsuRuleset(controller.bMap, controller.cursors, modifiers)

	//controller.Update(480000, 1)
}

func (controller *ReplayController) Update(time int64, delta float64) {

	for nTime := controller.lastTime + 1; nTime <= time; nTime++ {
		controller.bMap.Update(nTime)
		for i, c := range controller.controllers {
			if c.danceController != nil {
				c.danceController.Update(nTime, 1)

				if nTime%17 == 0 {
					controller.cursors[i].LastFrameTime = nTime - 17
					controller.cursors[i].CurrentFrameTime = nTime
					controller.cursors[i].IsReplayFrame = true
				} else {
					controller.cursors[i].IsReplayFrame = false
				}

				controller.ruleset.UpdateClickFor(controller.cursors[i], nTime)
				controller.ruleset.UpdateNormalFor(controller.cursors[i], nTime)
				controller.ruleset.UpdatePostFor(controller.cursors[i], nTime)

			} else {
				wasUpdated := false
				for c.replayIndex < len(c.frames) && c.replayTime+c.frames[c.replayIndex].Time <= nTime {

					frame := c.frames[c.replayIndex]
					c.replayTime += frame.Time

					mY := frame.MouseY

					if controller.replays[i].ModsV&difficulty.HardRock > 0 {
						mY = 384 - mY
					}

					controller.cursors[i].SetPos(vector.NewVec2f(frame.MosueX, mY))
					controller.cursors[i].LeftKey = frame.KeyPressed.LeftClick && frame.KeyPressed.Key1
					controller.cursors[i].RightKey = frame.KeyPressed.RightClick && frame.KeyPressed.Key2

					controller.cursors[i].LeftMouse = frame.KeyPressed.LeftClick && !frame.KeyPressed.Key1
					controller.cursors[i].RightMouse = frame.KeyPressed.RightClick && !frame.KeyPressed.Key2

					controller.cursors[i].LeftButton = frame.KeyPressed.LeftClick
					controller.cursors[i].RightButton = frame.KeyPressed.RightClick

					controller.cursors[i].LastFrameTime = controller.cursors[i].CurrentFrameTime
					controller.cursors[i].CurrentFrameTime = c.replayTime
					controller.cursors[i].IsReplayFrame = true

					controller.ruleset.UpdateClickFor(controller.cursors[i], c.replayTime)
					controller.ruleset.UpdateNormalFor(controller.cursors[i], c.replayTime)
					controller.ruleset.UpdatePostFor(controller.cursors[i], c.replayTime)
					wasUpdated = true

					c.replayIndex++
				}

				if !wasUpdated {
					localIndex := bmath.ClampI(c.replayIndex, 0, len(c.frames)-1)

					progress := math32.Min(float32(nTime-c.replayTime), float32(c.frames[localIndex].Time)) / float32(c.frames[localIndex].Time)

					prevIndex := bmath.MaxI(0, localIndex-1)

					mX := (c.frames[localIndex].MosueX-c.frames[prevIndex].MosueX)*progress + c.frames[prevIndex].MosueX
					mY := (c.frames[localIndex].MouseY-c.frames[prevIndex].MouseY)*progress + c.frames[prevIndex].MouseY

					if controller.replays[i].ModsV&difficulty.HardRock > 0 {
						mY = 384 - mY
					}

					controller.cursors[i].SetPos(vector.NewVec2f(mX, mY))
					controller.cursors[i].IsReplayFrame = false
					//controller.ruleset.UpdateNormalFor(controller.cursors[i], nTime)
				}
			}
		}

		controller.ruleset.Update(nTime)

		controller.lastTime = nTime
	}

	for i := range controller.controllers {
		if controller.controllers[i].danceController == nil {
			controller.cursors[i].Update(delta)
		}

		accuracy, combo, _, grade := controller.ruleset.GetResults(controller.cursors[i])
		controller.replays[i].Accuracy = accuracy
		controller.replays[i].Combo = combo
		controller.replays[i].Grade = grade
	}

}

func (controller *ReplayController) GetCursors() []*graphics.Cursor {
	return controller.cursors
}

func (controller *ReplayController) GetReplays() []RpData {
	return controller.replays
}

func (controller *ReplayController) GetRuleset() *osu.OsuRuleSet {
	return controller.ruleset
}

func (controller *ReplayController) GetBeatMap() *beatmap.BeatMap {
	return controller.bMap
}

func (controller *ReplayController) GetClick(player, key int) bool {
	switch key {
	case 0:
		return controller.cursors[player].LeftKey
	case 1:
		return controller.cursors[player].RightKey
	case 2:
		return controller.cursors[player].LeftButton
	case 3:
		return controller.cursors[player].RightButton
	}

	return false
}
