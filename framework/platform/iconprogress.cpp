#include "iconprogress.h"

#ifdef _WIN32

#include <shobjidl.h>
#include <stdint.h>

const IID IID_ITaskbarList3 = {0xea1afb91, 0x9e28, 0x4b86, {0x90, 0xe9, 0x9e, 0x9f, 0x8a, 0x5e, 0xef, 0xaf}};

ITaskbarList3 *m_pTaskBarList;

void tryInit() {
    if (m_pTaskBarList == nullptr) {
        CoInitialize(nullptr);
        CoCreateInstance(CLSID_TaskbarList, nullptr, CLSCTX_ALL, IID_ITaskbarList3, (void **) &m_pTaskBarList);
    }
}

void setState(HWND window, TBPFLAG flag) {
    tryInit();
    m_pTaskBarList->SetProgressState(window, flag);
}

void setProgress(HWND window, int32_t progress, int32_t max) {
    tryInit();
    m_pTaskBarList->SetProgressValue(window, progress, max);
}

#endif