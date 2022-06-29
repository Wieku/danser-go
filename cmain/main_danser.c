#include <stdlib.h>
#include <string.h>
#include <danser-core.h> //auto-generated during dist

#ifdef _WIN32
#include <windows.h>
// force switch to the high performance gpu in multi-gpu systems (mostly laptops)
__declspec(dllexport) DWORD NvOptimusEnablement = 0x00000001; // http://developer.download.nvidia.com/devzone/devcenter/gamegraphics/files/OptimusRenderingPolicies.pdf
__declspec(dllexport) DWORD AmdPowerXpressRequestHighPerformance = 0x00000001; // https://community.amd.com/thread/169965

int wmain(int argc, wchar_t *argv[], wchar_t *envp[]) {
    #ifdef LAUNCHER
        HWND consoleWnd = GetConsoleWindow();
        DWORD dwProcessId;
        GetWindowThreadProcessId(consoleWnd, &dwProcessId);
        if (GetCurrentProcessId() == dwProcessId) { //hide the console window only if's been created by executable
            ShowWindow(consoleWnd, SW_HIDE);
        }
    #endif

#else

int main(int argc, char *argv[]) {

#endif

    GoString *gS = malloc(argc * sizeof(GoString));

    for (int i = 0; i < argc; i++) {
        GoString s;

#ifdef _WIN32
        size_t bSize = WideCharToMultiByte(CP_UTF8, 0, argv[i], -1, NULL, 0, NULL, NULL);

        char* utf8Arg = calloc(bSize, sizeof(char));

        WideCharToMultiByte(CP_UTF8, 0, argv[i], -1, utf8Arg, bSize, NULL, NULL);

        s.p = utf8Arg;
        s.n = bSize-1;
#else
        s.p = argv[i];
        s.n = strlen(argv[i]);
#endif

        gS[i] = s;
    }

    GoSlice slc;
    slc.data = gS;
    slc.len = argc;
    slc.cap = argc;

#ifdef LAUNCHER
    danserMain(1, slc);
#else
    danserMain(0, slc);
#endif

    return 0;
}