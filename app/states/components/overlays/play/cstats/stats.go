package cstats

import (
	"fmt"
	"github.com/spf13/cast"
	"github.com/wieku/danser-go/app/beatmap"
	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/framework/graphics/batch"
	"github.com/wieku/danser-go/framework/math/animation"
	"math"
	"strconv"
	"sync"
	"text/template"
)

type StatDisplay struct {
	engine     *template.Template
	stats      []*Stat
	statHolder *StatHolder

	counterMap map[string]*rollingCounter
	lastTime   int64

	mtx *sync.RWMutex // To not risk a crash during settings reload
}

type rollingCounter struct {
	src, target string
	toStr       bool
	decimals    int
	prevValue   float64
	glider      *animation.TargetGlider
}

func NewStatDisplay(bMap *beatmap.BeatMap, diff *difficulty.Difficulty) *StatDisplay {
	display := &StatDisplay{
		engine:     template.New("engine"),
		statHolder: NewStatHolder(bMap, diff),
		counterMap: make(map[string]*rollingCounter),
		lastTime:   math.MinInt64,
		mtx:        &sync.RWMutex{},
	}

	display.engine.Funcs(templateFuncs)

	display.loadDisplays()

	settings.AddReloadListener(display.loadDisplays)

	return display
}

func (statDisplay *StatDisplay) loadDisplays() {
	statDisplay.mtx.Lock()

	statDisplay.stats = make([]*Stat, 0)

	for i, c := range settings.Gameplay.Statistics {
		stat := NewStat(statDisplay, c, statDisplay.engine.New(strconv.Itoa(i)))

		if stat != nil {
			statDisplay.stats = append(statDisplay.stats, stat)
		}
	}

	statDisplay.mtx.Unlock()
}

func (statDisplay *StatDisplay) GetStatHolder() *StatHolder {
	return statDisplay.statHolder
}

func (statDisplay *StatDisplay) Update(audioTime, normalTime float64) {
	statDisplay.statHolder.UpdateBPM()
	statDisplay.statHolder.UpdateTime(audioTime)

	statDisplay.updateRolling(normalTime)
	statDisplay.updateBTimes(audioTime)

	statDisplay.mtx.RLock()
	for _, stat := range statDisplay.stats {
		stat.Update(statDisplay.statHolder.stats)
	}
	statDisplay.mtx.RUnlock()
}

func (statDisplay *StatDisplay) updateRolling(time float64) {
	for _, counter := range statDisplay.counterMap {
		counter.glider.SetValue(cast.ToFloat64(statDisplay.statHolder.stats[counter.src]), false)
		counter.glider.Update(time)

		statDisplay.statHolder.stats[counter.target] = counter.glider.GetValue()

		if counter.toStr && counter.prevValue != counter.glider.GetValue() {
			statDisplay.statHolder.stats[counter.target+"S"] = fmt.Sprintf("%.*f", counter.decimals, counter.glider.GetValue())
		}

		counter.prevValue = counter.glider.GetValue()
	}
}

func (statDisplay *StatDisplay) updateBTimes(time float64) {
	hitObjects := statDisplay.statHolder.bMap.HitObjects
	startTime := hitObjects[0].GetStartTime()
	endTime := hitObjects[len(hitObjects)-1].GetEndTime()

	currentTime := int64(math.Floor((time - startTime) / 1000))

	statDisplay.statHolder.stats["timeFromStart"] = currentTime

	if statDisplay.lastTime != currentTime {
		statDisplay.statHolder.stats["timeToEnd"] = max(0, int64(math.Floor((endTime-time)/1000)))
		statDisplay.lastTime = currentTime
	}
}

func (statDisplay *StatDisplay) registerRollingValue(field string, decimals int, toStr bool) {
	rollingName := field + "Roll" + strconv.Itoa(decimals)

	counter, ok := statDisplay.counterMap[rollingName]

	if !ok {
		counter = &rollingCounter{
			src:       field,
			target:    rollingName,
			prevValue: cast.ToFloat64(statDisplay.statHolder.stats[field]),
			decimals:  decimals,
			glider:    animation.NewTargetGlider(cast.ToFloat64(statDisplay.statHolder.stats[field]), decimals),
		}

		statDisplay.counterMap[rollingName] = counter
		statDisplay.statHolder.stats[rollingName] = counter.glider.GetValue()
	}

	if !counter.toStr && toStr {
		counter.toStr = toStr

		statDisplay.statHolder.stats[rollingName+"S"] = fmt.Sprintf("%.*f", decimals, counter.glider.GetValue())
	}
}

func (statDisplay *StatDisplay) Draw(batch *batch.QuadBatch, alpha float64, sclWidth, sclHeight float64) {
	statDisplay.mtx.RLock()
	for _, stat := range statDisplay.stats {
		stat.Draw(batch, alpha, sclWidth, sclHeight)
	}
	statDisplay.mtx.RUnlock()
}
