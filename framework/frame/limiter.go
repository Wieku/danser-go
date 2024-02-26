package frame

import (
	"github.com/wieku/danser-go/framework/qpc"
	"runtime"
	"time"
)

// Full credits: LWJGL Team
// Ported from Java
// Original source: https://github.com/LWJGL/lwjgl/blob/master/src/java/org/lwjgl/opengl/Sync.java

const (
	nanosInSecond        = 1e9
	dampenThresholdSleep = 1.5e6 // 1.5ms
	dampenThresholdYield = 5e4   // 50us

	rFactorSleep = 30
	rFactorYield = 10

	dampenFactor = 0.9 // don't change: 0.9f is exactly right!
)

type Limiter struct {
	FPS         int
	nextFrame   int64
	initialised bool

	sleepDurations *runningAvg
	yieldDurations *runningAvg
}

func NewLimiter(fps int) *Limiter {
	return &Limiter{
		FPS:            fps,
		sleepDurations: newRunningAvg(10, dampenThresholdSleep, rFactorSleep),
		yieldDurations: newRunningAvg(10, dampenThresholdYield, rFactorYield),
	}
}

func (limiter *Limiter) Sync() {
	fps := limiter.FPS

	if fps <= 0 {
		return
	}

	if !limiter.initialised {
		limiter.initialised = true

		limiter.sleepDurations.init(1000 * 1000)
		limiter.yieldDurations.init(12_000) //int64(-float64(qpc.GetNanoTime()-qpc.GetNanoTime()) * 1.333))

		limiter.nextFrame = qpc.GetNanoTime()
	}

	for t0, t1 := qpc.GetNanoTime(), int64(0); (limiter.nextFrame - t0) > limiter.sleepDurations.avg(); t0 = t1 {
		time.Sleep(time.Millisecond)

		t1 = qpc.GetNanoTime()

		limiter.sleepDurations.add(t1 - t0)
	}

	limiter.sleepDurations.dampenForLowResTicker()

	for t0, t1 := qpc.GetNanoTime(), int64(0); (limiter.nextFrame - t0) > limiter.yieldDurations.avg(); t0 = t1 {
		runtime.Gosched()

		t1 = qpc.GetNanoTime()

		limiter.yieldDurations.add(t1 - t0)
	}

	limiter.yieldDurations.dampenForLowResTicker()

	limiter.nextFrame = max(limiter.nextFrame+nanosInSecond/int64(fps), qpc.GetNanoTime())
}

type runningAvg struct {
	slots  []int64
	offset int

	dampenThreshold int64
	anomalyRatio    int64
	average         int64
}

func newRunningAvg(slotCount int, dampenThreshold, anomalyRatio int64) *runningAvg {
	return &runningAvg{
		slots:           make([]int64, slotCount),
		offset:          0,
		dampenThreshold: dampenThreshold,
		anomalyRatio:    anomalyRatio,
	}
}

func (ra *runningAvg) init(value int64) {
	for i := 0; i < len(ra.slots); i++ {
		ra.slots[i] = value
	}

	ra.cAvg()
}

func (ra *runningAvg) add(value int64) {
	if value <= ra.anomalyRatio*ra.average {
		ra.slots[ra.offset] = value
		ra.offset = (ra.offset + 1) % len(ra.slots)

		ra.cAvg()
	}
}

func (ra *runningAvg) cAvg() {
	var sum int64

	for i := 0; i < len(ra.slots); i++ {
		sum += ra.slots[i]
	}

	ra.average = sum / int64(len(ra.slots))
}

func (ra *runningAvg) avg() int64 {
	return ra.average
}

func (ra *runningAvg) dampenForLowResTicker() {
	if ra.average > ra.dampenThreshold {
		for i := 0; i < len(ra.slots); i++ {
			ra.slots[i] = int64(float64(ra.slots[i]) * dampenFactor)
		}

		ra.cAvg()
	}
}
