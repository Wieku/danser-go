package dance

import (
	"github.com/wieku/danser/beatmap"
	"github.com/wieku/danser/render"
	"github.com/wieku/danser/animation"
	"github.com/Mempler/rplpa"
	"io/ioutil"
	"github.com/wieku/danser/bmath"
	"path/filepath"
	"os"
	"sort"
)

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
	replays     []*rplpa.Replay
	cursors     []*render.Cursor
	controllers []*subControl
}

func NewReplayController(opath string) Controller {
	controller := new(ReplayController)

	filepath.Walk(opath, func(path string, info os.FileInfo, err error) error {
		data, _ := ioutil.ReadFile(path)
		replay, _ := rplpa.ParseReplay(data)
		controller.replays = append(controller.replays, replay)
		return nil
	})

	sort.Slice(controller.replays, func(i, j int) bool {
		return controller.replays[i].Score > controller.replays[j].Score
	})

	for _, replay := range controller.replays {
		control := new(subControl)
		control.xGlider = animation.NewGlider(0)
		control.yGlider = animation.NewGlider(0)
		control.k1Glider = animation.NewGlider(0)
		control.k2Glider = animation.NewGlider(0)
		control.m1Glider = animation.NewGlider(0)
		control.m2Glider = animation.NewGlider(0)
		lastTime := int64(0)
		for _, frame := range replay.ReplayData {
			control.xGlider.AddEvent(float64(lastTime), float64(lastTime+frame.Time), float64(frame.MosueX))
			if replay.Mods&16 > 0 {
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
		controller.controllers = append(controller.controllers, control)
	}

	return controller
}

func (controller *ReplayController) SetBeatMap(beatMap *beatmap.BeatMap) {
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

func (controller *ReplayController) GetReplays() []*rplpa.Replay {
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