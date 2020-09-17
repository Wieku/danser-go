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
	"github.com/wieku/danser-go/framework/math/animation"
	"github.com/wieku/danser-go/framework/math/math32"
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
	k1Glider *animation.Glider
	k2Glider *animation.Glider
	m1Glider *animation.Glider
	m2Glider *animation.Glider

	danceController Controller
	replayIndex     int
	replayTime      int64
	frames          []*rplpa.ReplayData
}

func NewSubControl() *subControl {
	control := new(subControl)
	control.k1Glider = animation.NewGlider(0)
	control.k1Glider.SetSorting(false)
	control.k2Glider = animation.NewGlider(0)
	control.k2Glider.SetSorting(false)
	control.m1Glider = animation.NewGlider(0)
	control.m1Glider.SetSorting(false)
	control.m2Glider = animation.NewGlider(0)
	control.m2Glider.SetSorting(false)
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
	vals := url.Values{}
	vals.Add("c", strconv.FormatInt(scoreID, 10))
	vals.Add("m", "0")
	vals.Add("u", settings.Knockout.Username)
	vals.Add("h", settings.Knockout.MD5Pass)
	request, err := http.NewRequest(http.MethodGet, "https://osu.ppy.sh/web/osu-getreplay.php?"+vals.Encode(), nil)

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
	os.MkdirAll(replayDir, os.ModeDir)

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

				controller.replays = append(controller.replays, RpData{replayD.Username + "Ç" /*strings.Replace(strings.Replace(score.Mods.String(), "NF", "NF", -1), "NV", "TD", -1)*/, "", difficulty.Hidden, 100, 0, int64(mxCombo), osu.NONE})
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
		beatMapO, _ := client.GetBeatmaps(osuapi.GetBeatmapsOpts{BeatmapHash: beatMap.MD5})

		var scores []osuapi.GSScore
		if len(beatMapO) > 0 {
			scores, _ = client.GetScores(osuapi.GetScoresOpts{BeatmapID: beatMapO[0].BeatmapID, Limit: 200})
			/*emptyMods := osuapi.Mods(0)
			scores1, _ := client.GetScores(osuapi.GetScoresOpts{BeatmapID: beatMapO[0].BeatmapID, Limit: 200, Mods: &emptyMods})
			for _, s := range scores1 {
				washere := false
				for _, s1 := range scores  {
					if s.ScoreID == s1.ScoreID || (s.Username == s1.Username) {
						washere = true
						break
					}
				}
				if !washere {
					scores = append(scores, s)
				}
			}*/
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

			//if score.Username != "WhiteCat"/*"eyeball"*//*"Rhythm blue"*/ {
			//	continue
			//}

			//if score.Username != "Freddie Benson"/*"itsamemarioo"*//*"Teppi"*//*"ThePooN"*//*"Kosmonautas"*/ /*"idke"*/ /*"Vaxei"*/ /*"Dustice"*//*"WalkingTuna"*/ {
			//	continue
			//}
			fileName := filepath.Join(replayDir, strconv.FormatInt(score.ScoreID, 10)+".dsr")
			file, err := os.Open(fileName)
			file.Close()

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

				ioutil.WriteFile(fileName, data, 644)
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

	if settings.Knockout.AddDanser || counter == settings.Knockout.MaxPlayers {
		control := NewSubControl()

		control.danceController = NewGenericController()
		control.danceController.SetBeatMap(beatMap)

		controller.replays = append([]RpData{{settings.Knockout.DanserName /*"HD"*/, "HD", difficulty.Hidden, 100, 0, 0, osu.NONE}}, controller.replays...)
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

	translate := func(k bool) float64 {
		if k {
			return 1.0
		}
		return 0.0
	}

	lastTime := int64(0)
	for _, frame := range frames {
		eventTime := math.Max(float64(lastTime), float64(lastTime+frame.Time))

		//log.Println(eventTime, frame.Time, frame.MosueX, frame.MouseY, frame.KeyPressed)

		press := frame.KeyPressed

		subController.k1Glider.AddEventS(eventTime, eventTime, translate(press.Key1), translate(press.Key1))
		subController.k2Glider.AddEventS(eventTime, eventTime, translate(press.Key2), translate(press.Key2))
		subController.m1Glider.AddEventS(eventTime, eventTime, translate(press.LeftClick && !press.Key1), translate(press.LeftClick && !press.Key1))
		subController.m2Glider.AddEventS(eventTime, eventTime, translate(press.RightClick && !press.Key2), translate(press.RightClick && !press.Key2))

		lastTime += frame.Time
	}
}

func (controller *ReplayController) InitCursors() {
	var modifiers []difficulty.Modifier
	for i := range controller.controllers {
		if controller.controllers[i].danceController != nil {
			/*if i == 1 {
				Mover = movers.NewAngleOffsetMover
				settings.Dance.SliderDance = true
			}*/

			controller.controllers[i].danceController.InitCursors()
			//settings.Dance.SliderDance = false
			controller.controllers[i].danceController.GetCursors()[0].IsPlayer = true

			cursors := controller.controllers[i].danceController.GetCursors()

			for _, crsr := range cursors {
				crsr.Name = controller.replays[i].Name
			}

			controller.cursors = append(controller.cursors, cursors...)
		} else {
			crsr := graphics.NewCursor()
			crsr.Name = controller.replays[i].Name
			controller.cursors = append(controller.cursors, crsr)
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
				/*if i == 1 {
					settings.Dance.SliderDance = true
				}*/
				c.danceController.Update(nTime, 1)
				//settings.Dance.SliderDance = false
				if controller.cursors[i].LeftButton {

					c.k1Glider.Reset()
					c.k1Glider.AddEventS(float64(nTime), float64(nTime), 1.0, 1.0)
					c.k1Glider.AddEventS(float64(nTime+50), float64(nTime+50), 0.0, 0.0)
				}
				if controller.cursors[i].RightButton {
					c.k2Glider.Reset()
					c.k2Glider.AddEventS(float64(nTime), float64(nTime), 1.0, 1.0)
					c.k2Glider.AddEventS(float64(nTime+50), float64(nTime+50), 0.0, 0.0)
				}

				c.k1Glider.Update(float64(nTime))
				c.k2Glider.Update(float64(nTime))

				if nTime%12 == 0 {
					controller.cursors[i].LastFrameTime = nTime - 12
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

					controller.cursors[i].SetPos(bmath.NewVec2f(frame.MosueX, mY))
					controller.cursors[i].LeftButton = frame.KeyPressed.LeftClick || frame.KeyPressed.Key1
					controller.cursors[i].RightButton = frame.KeyPressed.RightClick || frame.KeyPressed.Key2
					controller.cursors[i].LastFrameTime = controller.cursors[i].CurrentFrameTime
					controller.cursors[i].CurrentFrameTime = c.replayTime
					controller.cursors[i].IsReplayFrame = true

					controller.ruleset.UpdateClickFor(controller.cursors[i], c.replayTime)
					controller.ruleset.UpdateNormalFor(controller.cursors[i], c.replayTime)
					controller.ruleset.UpdatePostFor(controller.cursors[i], c.replayTime)
					wasUpdated = true

					c.replayIndex++
				}

				c.k1Glider.Update(float64(nTime))
				c.k2Glider.Update(float64(nTime))
				c.m1Glider.Update(float64(nTime))
				c.m2Glider.Update(float64(nTime))

				if !wasUpdated {
					localIndex := bmath.ClampI(c.replayIndex, 0, len(c.frames)-1)

					progress := math32.Min(float32(nTime-c.replayTime), float32(c.frames[localIndex].Time)) / float32(c.frames[localIndex].Time)

					prevIndex := bmath.MaxI(0, localIndex-1)

					mX := (c.frames[localIndex].MosueX-c.frames[prevIndex].MosueX)*progress + c.frames[prevIndex].MosueX
					mY := (c.frames[localIndex].MouseY-c.frames[prevIndex].MouseY)*progress + c.frames[prevIndex].MouseY

					if controller.replays[i].ModsV&difficulty.HardRock > 0 {
						mY = 384 - mY
					}

					controller.cursors[i].SetPos(bmath.NewVec2f(mX, mY))
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
		return controller.controllers[player].k1Glider.GetValue() > 0.5
	case 1:
		return controller.controllers[player].k2Glider.GetValue() > 0.5
	case 2:
		return controller.controllers[player].m1Glider.GetValue() > 0.5
	case 3:
		return controller.controllers[player].m2Glider.GetValue() > 0.5
	}
	return false
}
