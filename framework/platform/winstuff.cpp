#include "winstuff.h"

#ifdef _WIN32

#include <shlobj.h>
#include <shobjidl.h>
#include <stdint.h>

const IID IID_ITaskbarList3 = {0xea1afb91, 0x9e28, 0x4b86, {0x90, 0xe9, 0x9e, 0x9f, 0x8a, 0x5e, 0xef, 0xaf}};

bool isInit{false};
ITaskbarList3 *m_pTaskBarList;

void tryInit() {
    if (!isInit) {
        CoInitializeEx(nullptr, COINIT_MULTITHREADED);
        isInit = true;
    }
}

void tryInstance() {
    tryInit();

    if (!m_pTaskBarList) {
        CoCreateInstance(CLSID_TaskbarList, nullptr, CLSCTX_ALL, IID_ITaskbarList3, (void **) &m_pTaskBarList);
    }
}

void setState(HWND window, TBPFLAG flag) {
    tryInstance();
    m_pTaskBarList->SetProgressState(window, flag);
}

void setProgress(HWND window, int32_t progress, int32_t max) {
    tryInstance();
    m_pTaskBarList->SetProgressValue(window, progress, max);
}

HRESULT openInExplorer(const wchar_t* filePath) {
    tryInit();

    PIDLIST_ABSOLUTE pidl = ILCreateFromPathW(filePath);

    if (pidl) {
        HRESULT res = SHOpenFolderAndSelectItems(pidl, 0, nullptr, 0);

        ILFree(pidl);

        return res;
    }

    return -1;
}

#endif