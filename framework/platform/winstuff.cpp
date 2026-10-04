#include "winstuff.h"

#ifdef _WIN32

#include <shlobj.h>
#include <shobjidl.h>
#include <stdint.h>

bool isInit{false};

void tryInit() {
    if (!isInit) {
        CoInitializeEx(nullptr, COINIT_MULTITHREADED);
        isInit = true;
    }
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