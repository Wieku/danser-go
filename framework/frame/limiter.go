package frame

import (
	"github.com/wieku/danser-go/app/bmath"
	"github.com/wieku/danser-go/framework/qpc"
	"runtime"
	"time"
)

type Limiter struct {
	FPS               int
	variableYieldTime int64
	lastTime          int64
}

func NewLimiter(fps int) *Limiter {
	return &Limiter{fps, 0, 0}
}

/**
 * An accurate sync method that adapts automatically
 * to the system it runs on to provide reliable results.
 *
 * @author kappa (On the LWJGL Forums)
 */
func (limiter *Limiter) Sync() {
	if limiter.FPS <= 0 {
		return
	}

	sleepTime := int64(1000000000) / int64(limiter.FPS) // nanoseconds to sleep this frame
	// yieldTime + remainder micro & nano seconds if smaller than sleepTime
	yieldTime := bmath.MinI64(sleepTime, limiter.variableYieldTime+sleepTime%int64(1000*1000))
	overSleep := int64(0) // time the sync goes over by

	for {
		t := qpc.GetNanoTime() - limiter.lastTime

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

	limiter.lastTime = qpc.GetNanoTime() - bmath.MinI64(overSleep, sleepTime)

	// auto tune the time sync should yield
	if overSleep > limiter.variableYieldTime {
		// increase by 200 microseconds (1/5 a ms)
		limiter.variableYieldTime = bmath.MinI64(limiter.variableYieldTime+200*1000, sleepTime)
	} else if overSleep < limiter.variableYieldTime-200*1000 {
		// decrease by 2 microseconds
		limiter.variableYieldTime = bmath.MaxI64(limiter.variableYieldTime-2*1000, 0)
	}

}
