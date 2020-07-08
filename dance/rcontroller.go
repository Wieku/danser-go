package dance

import (
	"github.com/Mempler/rplpa"
	"github.com/thehowl/go-osuapi"
	"github.com/wieku/danser-go/animation"
	"github.com/wieku/danser-go/beatmap"
	"github.com/wieku/danser-go/bmath"
	"github.com/wieku/danser-go/bmath/difficulty"
	"github.com/wieku/danser-go/render"
	"github.com/wieku/danser-go/rulesets/osu"
	"github.com/wieku/danser-go/settings"
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
	xGlider    *animation.Glider
	yGlider    *animation.Glider
	k1Glider   *animation.Glider
	k2Glider   *animation.Glider
	m1Glider   *animation.Glider
	m2Glider   *animation.Glider
	frame      *animation.Glider
	lolControl Controller
	replayIndex int
	replayTime int64
	frames []*rplpa.ReplayData
}

func NewSubControl() *subControl {
	control := new(subControl)
	control.xGlider = animation.NewGlider(0)
	control.xGlider.SetSorting(false)
	control.yGlider = animation.NewGlider(0)
	control.yGlider.SetSorting(false)
	control.k1Glider = animation.NewGlider(0)
	control.k1Glider.SetSorting(false)
	control.k2Glider = animation.NewGlider(0)
	control.k2Glider.SetSorting(false)
	control.m1Glider = animation.NewGlider(0)
	control.m1Glider.SetSorting(false)
	control.m2Glider = animation.NewGlider(0)
	control.m2Glider.SetSorting(false)
	control.frame = animation.NewGlider(0)
	control.frame.SetSorting(false)
	return control
}

type ReplayController struct {
	bMap              *beatmap.BeatMap
	replays           []RpData
	cursors           []*render.Cursor
	controllers       []*subControl
	ruleset           *osu.OsuRuleSet
	lastTime int64
	//counter int64
}

func NewReplayController() Controller {
	return new(ReplayController)
}

func getReplay(scoreID int64) ([]byte, error) {
	vals := url.Values{}
	vals.Add("c", strconv.FormatInt(scoreID, 10))
	vals.Add("m", "0")
	vals.Add("u", strings.Split(settings.KNOCKOUT, ":")[0])
	vals.Add("h", strings.Split(settings.KNOCKOUT, ":")[1])
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
	os.Mkdir("replays", os.ModeDir)
	replayDir := filepath.Join("replays", beatMap.MD5)
	os.Mkdir(replayDir, os.ModeDir)

	client := osuapi.NewClient(strings.Split(settings.KNOCKOUT, ":")[2])
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
	counter := 0

	sort.SliceStable(scores, func(i, j int) bool {
		return scores[i].Score.Score > scores[j].Score.Score
	})

	//bot
	if settings.KNOCKOUTDANCE {
		control := NewSubControl()

		control.lolControl = NewGenericController()
		control.lolControl.SetBeatMap(beatMap)

		controller.replays = append(controller.replays, RpData{"serand", "HD", difficulty.Hidden, 100, 0, 0, osu.NONE})
		controller.controllers = append(controller.controllers, control)
		counter++
	}

	filepath.Walk(replayDir, func(path string, f os.FileInfo, err error) error {
		if strings.HasSuffix(f.Name(), ".osr") {
			data, err := ioutil.ReadFile(path)
			if err != nil {
				panic(err)
			}

			replayD, _ := rplpa.ParseReplay(data)

			control := NewSubControl()

			lastTime := int64(0)


			control.frames = replayD.ReplayData
			for _, frame := range replayD.ReplayData {
				if frame.Time == -12345 {
					continue
				}

				time1 := float64(lastTime)
				time2 := math.Max(time1, float64(lastTime+frame.Time))
				mY := float64(frame.MouseY)
				/*if strings.Contains(replayD., "HR") {
					mY = 384 - mY
				}*/

				control.xGlider.AddEventS(time2, time2, float64(frame.MosueX), float64(frame.MosueX))
				control.yGlider.AddEventS(time2, time2, mY, mY)

				//log.Println(time2, frame.Time, frame.MosueX, frame.MouseY, frame.KeyPressed)

				press := frame.KeyPressed

				translate := func(k bool) float64 {
					if k {
						return 1.0
					} else {
						return 0.0
					}
				}

				control.k1Glider.AddEventS(time2, time2, translate(press.Key1), translate(press.Key1))
				control.k2Glider.AddEventS(time2, time2, translate(press.Key2), translate(press.Key2))
				control.m1Glider.AddEventS(time2, time2, translate(press.LeftClick && !press.Key1), translate(press.LeftClick && !press.Key1))
				control.m2Glider.AddEventS(time2, time2, translate(press.RightClick && !press.Key2), translate(press.RightClick && !press.Key2))
				if frame.Time == 0 {
					control.frame.RemoveLast()
				}
				control.frame.AddEventS(time2, time2, 1, 1)
				control.frame.AddEventS(time2+1, time2+1, 0, 0)

				//lastX = float64(frame.MosueX)
				//lastY = mY
				lastTime += frame.Time
			}

			mxCombo := replayD.MaxCombo

			controller.replays = append(controller.replays, RpData{replayD.Username + "Ç", /*strings.Replace(strings.Replace(score.Mods.String(), "NF", "NF", -1), "NV", "TD", -1)*/"", difficulty.Modifier(0), 100, 0, int64(mxCombo), osu.NONE})
			controller.controllers = append(controller.controllers, control)

			counter++
		}
		return nil
	})

	/*if settings.KNOCKOUTDANCE {
		control := NewSubControl()

		control.lolControl = NewGenericController()
		control.lolControl.SetBeatMap(beatMap)

		controller.replays = append(controller.replays, RpData{"resnad", "HD", difficulty.Hidden, 100, 0, 0, osu.NONE})
		controller.controllers = append(controller.controllers, control)
		counter++
	}*/

	for _, score := range scores {
		if score.Mods&osuapi.ModHalfTime > 0 || score.Mods&osuapi.ModEasy > 0 || counter >= 50 {
			continue
		}
		if score.Username != "FlyingTuna" && score.Username != "WhiteCat" {
			continue
		}
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

		control := NewSubControl()

		lastTime := int64(0)
		lastX := 0.0
		lastY := 0.0

		log.Println(score.Username, score.PP)

		data, err := ioutil.ReadFile(fileName)
		if err != nil {
			panic(err)
		}

		ad1 := int64(0)
		ad2 := 0

		replay, _ := rplpa.ParseCompressed(data)
		control.frames = replay
		for _, frame := range replay {
			if frame.Time == -12345 {
				continue
			}

			time1 := float64(lastTime)
			time2 := math.Max(time1, float64(lastTime+frame.Time))
			mY := float64(frame.MouseY)
			if strings.Contains(score.Mods.String(), "HR") {
				mY = 384 - mY
			}

			control.xGlider.AddEventS(time1, time2, lastX, float64(frame.MosueX))
			control.yGlider.AddEventS(time1, time2, lastY, mY)

			//log.Println(time2, frame.Time, frame.MosueX, frame.MouseY, frame.KeyPressed)

			press := frame.KeyPressed

			translate := func(k bool) float64 {
				if k {
					return 1.0
				} else {
					return 0.0
				}
			}

			control.k1Glider.AddEventS(time2, time2, translate(press.Key1), translate(press.Key1))
			control.k2Glider.AddEventS(time2, time2, translate(press.Key2), translate(press.Key2))
			control.m1Glider.AddEventS(time2, time2, translate(press.LeftClick && !press.Key1), translate(press.LeftClick && !press.Key1))
			control.m2Glider.AddEventS(time2, time2, translate(press.RightClick && !press.Key2), translate(press.RightClick && !press.Key2))
			if frame.Time == 0 {
				control.frame.RemoveLast()
			}
			control.frame.AddEventS(time2, time2, 1, 1)
			control.frame.AddEventS(time2+1, time2+1, 0, 0)

			if frame.Time > 0 {
				ad1 += frame.Time
				ad2++
			}
			lastX = float64(frame.MosueX)
			lastY = mY
			lastTime += frame.Time
		}

		aa := 1000.0 / (float64(ad1)/float64(ad2))

		log.Println("Framerate", aa)

		mxCombo := score.MaxCombo

		controller.replays = append(controller.replays, RpData{score.Username, strings.Replace(strings.Replace(score.Mods.String(), "NF", "NF", -1), "NV", "TD", -1), difficulty.Modifier(score.Mods), 100, 0, int64(mxCombo), osu.NONE})
		controller.controllers = append(controller.controllers, control)

		counter++
	}

	settings.PLAYERS = len(controller.replays)

	controller.bMap = beatMap
	controller.lastTime = -200
}

func (controller *ReplayController) InitCursors() {
	var modifiers []difficulty.Modifier
	for i := range controller.controllers {
		if controller.controllers[i].lolControl != nil {
			/*if i == 1 {
				Mover = movers.NewAngleOffsetMover
				settings.Dance.SliderDance = true
			}*/

			controller.controllers[i].lolControl.InitCursors()
			//settings.Dance.SliderDance = false
			controller.controllers[i].lolControl.GetCursors()[0].IsPlayer = true
			controller.cursors = append(controller.cursors, controller.controllers[i].lolControl.GetCursors()...)
		} else {
			crsr := render.NewCursor()
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
			if c.lolControl != nil {
				/*if i == 1 {
					settings.Dance.SliderDance = true
				}*/
				c.lolControl.Update(nTime, 1)
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

				/*controller.counter += 1 //nTime - controller.lastTime

				if controller.counter >= 12 {
					controller.cursors[i].LastFrameTime = nTime - 12
					controller.cursors[i].CurrentFrameTime = nTime
					controller.cursors[i].IsReplayFrame = true
					controller.counter -= 12
				} else {
					controller.cursors[i].IsReplayFrame = false
				}*/

				if nTime % 12 == 0 {
					controller.cursors[i].LastFrameTime = nTime - 12
					controller.cursors[i].CurrentFrameTime = nTime
					controller.cursors[i].IsReplayFrame = true
				} else {
					controller.cursors[i].IsReplayFrame = false
				}

				controller.ruleset.UpdateFor(controller.cursors[i], nTime)

			} else {

				for ; c.replayIndex < len(c.frames) && c.replayTime + c.frames[c.replayIndex].Time <= nTime ; {

					frame := c.frames[c.replayIndex]
					c.replayTime += frame.Time

					mY := float64(frame.MouseY)

					if controller.replays[i].ModsV & difficulty.HardRock > 0 {
						mY = 384 - mY
					}

					controller.cursors[i].SetPos(bmath.NewVec2d(float64(frame.MosueX), mY))
					controller.cursors[i].LeftButton = frame.KeyPressed.LeftClick || frame.KeyPressed.Key1
					controller.cursors[i].RightButton = frame.KeyPressed.RightClick || frame.KeyPressed.Key2
					controller.cursors[i].LastFrameTime = controller.cursors[i].CurrentFrameTime
					controller.cursors[i].CurrentFrameTime = c.replayTime
					controller.cursors[i].IsReplayFrame = true

					controller.ruleset.UpdateFor(controller.cursors[i], c.replayTime)
					c.replayIndex++
				}

				c.xGlider.Update(float64(nTime))
				c.yGlider.Update(float64(nTime))
				c.k1Glider.Update(float64(nTime))
				c.k2Glider.Update(float64(nTime))
				c.m1Glider.Update(float64(nTime))
				c.m2Glider.Update(float64(nTime))
				c.frame.Update(float64(nTime))
				controller.cursors[i].SetPos(bmath.NewVec2d(c.xGlider.GetValue(), c.yGlider.GetValue()))
				//controller.cursors[i].LeftButton = controller.GetClick(i, 0) || controller.GetClick(i, 2)
				//controller.cursors[i].RightButton = controller.GetClick(i, 1) || controller.GetClick(i, 3)
				if c.frame.GetValue() > 0.5 {
					//controller.cursors[i].LastFrameTime = controller.cursors[i].CurrentFrameTime
					//controller.cursors[i].CurrentFrameTime = nTime
					controller.cursors[i].IsReplayFrame = true
				} else {
					controller.cursors[i].IsReplayFrame = false
				}
			}
		}

		controller.ruleset.Update(nTime)
		controller.lastTime = nTime
	}

	for i := range controller.controllers {
		if controller.controllers[i].lolControl == nil {
			controller.cursors[i].Update(delta)
		}
		accuracy, combo, _, grade := controller.ruleset.GetResults(controller.cursors[i])
		controller.replays[i].Accuracy = accuracy
		controller.replays[i].Combo = combo
		controller.replays[i].Grade = grade
	}

}

func (controller *ReplayController) GetCursors() []*render.Cursor {
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
