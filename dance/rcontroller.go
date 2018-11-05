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
)

type subControl struct {
	xGlider *animation.Glider
	yGlider *animation.Glider
}

type ReplayController struct {
	bMap    *beatmap.BeatMap
	replays []*rplpa.Replay
	cursors  []*render.Cursor
	controllers  []*subControl

}

func NewReplayController(opath string) Controller {
	controller := new(ReplayController)

	filepath.Walk(opath, func(path string, info os.FileInfo, err error) error {
		control := new(subControl)
		control.xGlider = animation.NewGlider(0)
		control.yGlider = animation.NewGlider(0)
		data, _ := ioutil.ReadFile(path)
		replay, _ := rplpa.ParseReplay(data)
		lastTime := int64(0)
		for _, frame := range replay.ReplayData {
			control.xGlider.AddEvent(float64(lastTime), float64(lastTime+frame.Time), float64(frame.MosueX))
			if replay.Mods & 16 > 0 {
				control.yGlider.AddEvent(float64(lastTime), float64(lastTime+frame.Time), float64(384-frame.MouseY))
			} else {
				control.yGlider.AddEvent(float64(lastTime), float64(lastTime+frame.Time), float64(frame.MouseY))
			}

			lastTime += frame.Time
		}
		controller.controllers = append(controller.controllers, control)
		controller.replays = append(controller.replays, replay)
		return nil
	})

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
		controller.cursors[i].SetPos(bmath.NewVec2d(c.xGlider.GetValue(), c.yGlider.GetValue()))
		controller.cursors[i].Update(delta)
	}
}

func (controller *ReplayController) GetCursors() []*render.Cursor {
	return controller.cursors
}
