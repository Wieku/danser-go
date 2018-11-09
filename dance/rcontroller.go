package dance

import (
	"github.com/wieku/danser/beatmap"
	"github.com/wieku/danser/render"
	"github.com/wieku/danser/animation"
	"github.com/Mempler/rplpa"
	"github.com/wieku/danser/bmath"
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
	"github.com/wieku/danser/dance/schedulers"
)

type RpData struct {
	Name string
	Mods string
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
	scheduler 	schedulers.Scheduler
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
	//log.Println(request)
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
	replayDir := filepath.Join("replays",beatMap.MD5)
	os.Mkdir(replayDir, os.ModeDir)

	client := osuapi.NewClient(strings.Split(settings.KNOCKOUT,":")[2])
	beatMapO, _ := client.GetBeatmaps(osuapi.GetBeatmapsOpts{BeatmapHash: beatMap.MD5})

	scores, _ := client.GetScores(osuapi.GetScoresOpts{BeatmapID: beatMapO[0].BeatmapID, Limit:50})

	for _, score := range scores {

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
		control.yGlider = animation.NewGlider(0)
		control.k1Glider = animation.NewGlider(0)
		control.k2Glider = animation.NewGlider(0)
		control.m1Glider = animation.NewGlider(0)
		control.m2Glider = animation.NewGlider(0)
		lastTime := int64(0)

		data, err := ioutil.ReadFile(fileName)
		if err != nil {
			panic(err)
		}

		replay, _ := rplpa.ParseCompressed(data)
		for _, frame := range replay {
			control.xGlider.AddEvent(float64(lastTime), float64(lastTime+frame.Time), float64(frame.MosueX))
			if strings.Contains(score.Mods.String(), "HR") {
				control.yGlider.AddEvent(float64(lastTime), float64(lastTime+frame.Time), float64(384-frame.MouseY))
			} else {
				control.yGlider.AddEvent(float64(lastTime), float64(lastTime+frame.Time), float64(frame.MouseY))
			}

			press := frame.KeyPressed

			translate := func(k bool) float64 {
				if k {
					return 1.0
				} else {
					return 0.0
				}
			}

			control.k1Glider.AddEventS(float64(lastTime+frame.Time), float64(lastTime+frame.Time), translate(press.Key1), translate(press.Key1))
			control.k2Glider.AddEventS(float64(lastTime+frame.Time), float64(lastTime+frame.Time), translate(press.Key2), translate(press.Key2))
			control.m1Glider.AddEventS(float64(lastTime+frame.Time), float64(lastTime+frame.Time), translate(press.LeftClick && !press.Key1), translate(press.LeftClick && !press.Key1))
			control.m2Glider.AddEventS(float64(lastTime+frame.Time), float64(lastTime+frame.Time), translate(press.RightClick && !press.Key2), translate(press.RightClick && !press.Key2))

			lastTime += frame.Time
		}

		controller.replays = append(controller.replays, RpData{score.Username, strings.Replace(score.Mods.String(), "NV", "", -1)})
		controller.controllers = append(controller.controllers, control)
	}

	controller.bMap = beatMap
}

func (controller *ReplayController) InitCursors() {
	for range controller.controllers {
		controller.cursors = append(controller.cursors, render.NewCursor())
	}
}

func (controller *ReplayController) Update(time int64, delta float64) {
	for i, c := range controller.controllers {
		c.xGlider.Update(float64(time))
		c.yGlider.Update(float64(time))
		c.k1Glider.Update(float64(time))
		c.k2Glider.Update(float64(time))
		c.m1Glider.Update(float64(time))
		c.m2Glider.Update(float64(time))
		controller.cursors[i].SetPos(bmath.NewVec2d(c.xGlider.GetValue(), c.yGlider.GetValue()))
		controller.cursors[i].Update(delta)
	}
}

func (controller *ReplayController) GetCursors() []*render.Cursor {
	return controller.cursors
}

func (controller *ReplayController) GetReplays() []RpData {
	return controller.replays
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