package frame

import (
	"github.com/wieku/danser-go/framework/math/mutils"
	"github.com/wieku/danser-go/framework/qpc"
	"runtime"
	"time"
)

// Full credits: LWJGL Team
// Ported from Java
// Original source: https://github.com/LWJGL/lwjgl/blob/master/src/java/org/lwjgl/opengl/Sync.java

const (
	nanosInSecond   = 1e9
	dampenThreshold = 10e6 // 10ms
	dampenFactor    = 0.9  // don't change: 0.9f is exactly right!
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
		sleepDurations: newRunningAvg(10),
		yieldDurations: newRunningAvg(10),
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
		limiter.yieldDurations.init(int64(-float64(qpc.GetNanoTime()-qpc.GetNanoTime()) * 1.333))

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

	limiter.nextFrame = mutils.Max(limiter.nextFrame+nanosInSecond/int64(fps), qpc.GetNanoTime())
}

type runningAvg struct {
	slots  []int64
	offset int
}

func newRunningAvg(slotCount int) *runningAvg {
	return &runningAvg{
		slots:  make([]int64, slotCount),
		offset: 0,
	}
}

func (ra *runningAvg) init(value int64) {
	for i := 0; i < len(ra.slots); i++ {
		ra.slots[i] = value
	}
}

func (ra *runningAvg) add(value int64) {
	ra.slots[ra.offset] = value
	ra.offset = (ra.offset + 1) % len(ra.slots)
}

func (ra *runningAvg) avg() int64 {
	var sum int64

	for i := 0; i < len(ra.slots); i++ {
		sum += ra.slots[i]
	}

	return sum / int64(len(ra.slots))
}

func (ra *runningAvg) dampenForLowResTicker() {
	if ra.avg() > dampenThreshold {
		for i := 0; i < len(ra.slots); i++ {
			ra.slots[i] = int64(float64(ra.slots[i]) * dampenFactor)
		}
	}
}
