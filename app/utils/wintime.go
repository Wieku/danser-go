package utils

import (
	"runtime"
	"time"
)

/*
#ifdef _WIN32
#include <windows.h>

int started = 0;
double PCFreq = 0.0;
__int64 CounterStart = 0;

void startCounter() {
    LARGE_INTEGER li;
    QueryPerformanceFrequency(&li);

    PCFreq = (double)(li.QuadPart) / 1000000000.0;

    QueryPerformanceCounter(&li);
    CounterStart = li.QuadPart;
}

double getTime() {
	if (!started) {
		startCounter();
		started = 1;
	}
    LARGE_INTEGER li;
    QueryPerformanceCounter(&li);
    return (double)(li.QuadPart-CounterStart)/PCFreq;
}

long long getNanoTime() {
	return (long long)(getTime());
}
#else
long long getNanoTime() {
	return 0;
}
#endif
*/
import "C"

func GetNanoTime() int64 {
	if runtime.GOOS == "windows" {
		return int64(C.getNanoTime())
	}
	return time.Now().UnixNano()
}
