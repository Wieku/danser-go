package qpc

/*
#include "time.h"
*/
import "C"

import (
	"runtime"
	"time"
)

func GetNanoTime() int64 {
	if runtime.GOOS == "windows" {
		return int64(C.getNanoTime())
	}
	return time.Now().UnixNano()
}

func GetNanoTimeF() float64 {
	return float64(GetNanoTime())
}

func GetMicroTime() int64 {
	return GetNanoTime() / 1e3
}

func GetMicroTimeF() float64 {
	return GetNanoTimeF() / 1e3
}

func GetMilliTime() int64 {
	return GetNanoTime() / 1e6
}

func GetMilliTimeF() float64 {
	return GetNanoTimeF() / 1e6
}
