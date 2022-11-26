#ifndef ICONPROGRESS_H
#define ICONPROGRESS_H

#ifdef _WIN32

#ifdef __cplusplus
extern "C" {
#endif

#include <stdint.h>
#include <shobjidl.h>

void setState(HWND window, TBPFLAG flag);

void setProgress(HWND window, int32_t progress, int32_t max);

#ifdef __cplusplus
}
#endif
#endif
#endif