#ifndef BASS_UTIL_H
#define BASS_UTIL_H
#include "bass.h"

HSTREAM CreateBassStream(char* file, DWORD flags);
HSAMPLE LoadBassSample(char* file, DWORD max, DWORD flags);

#endif