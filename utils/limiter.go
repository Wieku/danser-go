package utils

import (
	"runtime"
	"time"
)

type FpsLimiter struct {
	FPS               int
	variableYieldTime int64
	lastTime          int64
}

func NewFpsLimiter(fps int) *FpsLimiter {
	return &FpsLimiter{fps, 0, 0}
}

/**
 * An accurate sync method that adapts automatically
 * to the system it runs on to provide reliable results.
 *
 * @author kappa (On the LWJGL Forums)
 */
func (limiter *FpsLimiter) Sync() {
	if limiter.FPS <= 0 {
		return
	}

	sleepTime := int64(1000000000) / int64(limiter.FPS) // nanoseconds to sleep this frame
	// yieldTime + remainder micro & nano seconds if smaller than sleepTime
	yieldTime := Minint64(sleepTime, limiter.variableYieldTime+sleepTime%int64(1000*1000))
	overSleep := int64(0) // time the sync goes over by

	for ; ; {
		t := GetNanoTime() - limiter.lastTime

		if t < sleepTime-yieldTime {
			time.Sleep(time.Millisecond)
		} else if t < sleepTime {
			// burn the last few CPU cycles to ensure accuracy
			runtime.Gosched()
		} else {
			overSleep = t - sleepTime
			break // exit while loop
		}
	}

	limiter.lastTime = GetNanoTime() - Minint64(overSleep, sleepTime)

	// auto tune the time sync should yield
	if overSleep > limiter.variableYieldTime {
		// increase by 200 microseconds (1/5 a ms)
		limiter.variableYieldTime = Minint64(limiter.variableYieldTime+200*1000, sleepTime)
	} else if overSleep < limiter.variableYieldTime-200*1000 {
		// decrease by 2 microseconds
		limiter.variableYieldTime = Maxint64(limiter.variableYieldTime-2*1000, 0)
	}

}

func Minint64(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

func Maxint64(a, b int64) int64 {
	if a < b {
		return b
	}
	return a
}
