#ifndef WINSTUFF_H
#define WINSTUFF_H

#ifdef _WIN32

#ifdef __cplusplus
extern "C" {
#endif

#include <stdint.h>
#include <shobjidl.h>

HRESULT openInExplorer(const wchar_t* filePath);

#ifdef __cplusplus
}
#endif
#endif
#endif