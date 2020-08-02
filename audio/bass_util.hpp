#ifndef BASS_UTIL_H
#define BASS_UTIL_H
#include "bass.h"
#ifdef __cplusplus
extern "C" {
#endif

HSTREAM CreateBassStream(char* file, DWORD flags);
HSAMPLE LoadBassSample(char* file, DWORD max, DWORD flags);

#ifdef __cplusplus
}
#endif
#endif