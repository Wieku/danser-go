#include "time.h"
#include <stdint.h>
#include <stdbool.h>

#ifdef _WIN32

#include <windows.h>

bool started = false;
double pcFreq = 0.0;
int64_t counterStart = 0;

void startCounter() {
    LARGE_INTEGER li;
    QueryPerformanceFrequency(&li);

    pcFreq = li.QuadPart / 1000000000.0;

    QueryPerformanceCounter(&li);

    counterStart = li.QuadPart;
}

double getTime() {
    if (!started) {
        startCounter();
        started = true;
    }

    LARGE_INTEGER li;
    QueryPerformanceCounter(&li);

    return (li.QuadPart - counterStart) / pcFreq;
}

int64_t getNanoTime() {
    return (int64_t) getTime();
}

#else

int64_t getNanoTime() {
    return 0;
}

#endif