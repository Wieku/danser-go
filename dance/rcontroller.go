package dance

import (
	"github.com/wieku/danser/beatmap"
	"github.com/wieku/danser/render"
	"github.com/wieku/danser/animation"
	"github.com/Mempler/rplpa"
	"github.com/thehowl/go-osuapi"
	"github.com/wieku/danser/settings"
	"log"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"net/url"
	"strconv"
	"net/http"
	"github.com/wieku/danser/rulesets/osu"
	"github.com/wieku/danser/bmath/difficulty"
	"math"
	"github.com/wieku/danser/bmath"
)

type RpData struct {
	Name string
	Mods string
	ModsV difficulty.Modifier
	Accuracy float64
	Combo int64
	Grade osu.Grade
}

type subControl struct {
	xGlider  *animation.Glider
	yGlider  *animation.Glider
	k1Glider *animation.Glider
	k2Glider *animation.Glider
	m1Glider *animation.Glider
	m2Glider *animation.Glider
}

type ReplayController struct {
	bMap        *beatmap.BeatMap
	replays     []RpData
	cursors     []*render.Cursor
	controllers []*subControl
	ruleset     *osu.OsuRuleSet
	lastTime int64
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

	scores, _ := client.GetScores(osuapi.GetScoresOpts{BeatmapID: beatMapO[0].BeatmapID, Limit: 100})
	counter := 0
	for _, score := range scores {
		if score.Mods&osuapi.ModHalfTime > 0 || counter >= 50 {
			continue
		}
		//if score.Username != /*"itsamemarioo"*//*"nathan on osu"*//*"ThePooN"*//*"Kosmonautas"*/ {
		//	continue
		//}
		fileName := filepath.Join(replayDir, strconv.FormatInt(score.ScoreID, 10)+".dsr")
		file, err := os.Open(fileName)
		file.Close()

		if os.IsNotExist(err) {
			data, err := getReplay(score.ScoreID)
			if err != nil {
				panic(err)
			} else {
				log.Println("Downloaded replay for:", score.Username)
			}

			ioutil.WriteFile(fileName, data, 644)
		}

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

		lastTime := int64(0)
		lastX := 0.0
		lastY := 0.0

		data, err := ioutil.ReadFile(fileName)
		if err != nil {
			panic(err)
		}

		replay, _ := rplpa.ParseCompressed(data)
		for _, frame := range replay {
			if frame.Time == -12345 {
				continue
			}

			time1 := float64(lastTime)
			time2 := math.Max(time1, float64(lastTime+frame.Time))
			mY := float64(frame.MouseY)
			if strings.Contains(score.Mods.String(), "HR") {
				mY = 384-mY
			}

			control.xGlider.AddEventS(time1, time2, lastX, float64(frame.MosueX))
			control.yGlider.AddEventS(time1, time2, lastY, mY)

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

			lastX = float64(frame.MosueX)
			lastY = mY
			lastTime += frame.Time
		}

		controller.replays = append(controller.replays, RpData{score.Username, strings.Replace(score.Mods.String(), "NV", "", -1), difficulty.Modifier(score.Mods), 100, 0, 0})
		controller.controllers = append(controller.controllers, control)

		counter++
	}

	controller.bMap = beatMap
}

func (controller *ReplayController) InitCursors() {
	var modifiers []difficulty.Modifier
	for i := range controller.controllers {
		controller.cursors = append(controller.cursors, render.NewCursor())
		modifiers = append(modifiers, controller.replays[i].ModsV)
	}
	controller.ruleset = osu.NewOsuRuleset(controller.bMap, controller.cursors, modifiers)
}

func (controller *ReplayController) Update(time int64, delta float64) {
	
	for nTime:=controller.lastTime; nTime < time; nTime++ {
		for i, c := range controller.controllers {
			c.xGlider.Update(float64(nTime))
			c.yGlider.Update(float64(nTime))
			c.k1Glider.Update(float64(nTime))
			c.k2Glider.Update(float64(nTime))
			c.m1Glider.Update(float64(nTime))
			c.m2Glider.Update(float64(nTime))
			controller.cursors[i].SetPos(bmath.NewVec2d(c.xGlider.GetValue(), c.yGlider.GetValue()))
			controller.cursors[i].LeftButton = controller.GetClick(i, 0) || controller.GetClick(i, 2)
			controller.cursors[i].RightButton = controller.GetClick(i, 1) || controller.GetClick(i, 3)
		}

		controller.ruleset.Update(nTime)
	}

	for i := range controller.controllers {
		controller.cursors[i].Update(delta)
		accuracy, combo, grade := controller.ruleset.GetResults(controller.cursors[i])
		controller.replays[i].Accuracy = accuracy
		controller.replays[i].Combo = combo
		controller.replays[i].Grade = grade
	}
	controller.lastTime = time
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
