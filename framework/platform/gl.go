package platform

import "C"
import (
	"fmt"
	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/wieku/danser-go/framework/graphics/hacks"
	"log"
	"os"
	"strings"
	"unsafe"
)

// SetupContext sets glfw hints about OpenGL version
func SetupContext() {
	glfw.WindowHint(glfw.ContextVersionMajor, 3)
	glfw.WindowHint(glfw.ContextVersionMinor, 3)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
}

// GLInit initializes OpenGL, checks for needed extensions, eventually sets up GPU debug logs
func GLInit(debugLogs bool, additionalExtensions ...string) error {
	log.Println("Initializing OpenGL...")

	err := gl.Init()
	if err != nil {
		return err
	}

	err = extensionCheck(additionalExtensions)
	if err != nil {
		return err
	}

	glVendor := C.GoString((*C.char)(unsafe.Pointer(gl.GetString(gl.VENDOR))))
	glRenderer := C.GoString((*C.char)(unsafe.Pointer(gl.GetString(gl.RENDERER))))
	glVersion := C.GoString((*C.char)(unsafe.Pointer(gl.GetString(gl.VERSION))))
	glslVersion := C.GoString((*C.char)(unsafe.Pointer(gl.GetString(gl.SHADING_LANGUAGE_VERSION))))

	lVendor := strings.ToLower(glVendor)

	// HACK HACK HACK: please see github.com/wieku/danser-go/framework/graphics/hacks.IsIntel for more info
	if strings.Contains(lVendor, "intel") {
		hacks.IsIntel = true
	}

	forceAMDHack := false

	if enF, ok := os.LookupEnv("DANSER_FLAGS"); ok && strings.Contains(enF, "AMDHACK") {
		forceAMDHack = true
	}

	// HACK HACK HACK: please see github.com/wieku/danser-go/framework/graphics/hacks.IsOldAMD for more info
	if forceAMDHack || (strings.Contains(lVendor, "amd") || strings.Contains(lVendor, "ati")) &&
		(strings.Contains(glVersion, "15.201.") || strings.Contains(glVersion, "15.200.")) {
		hacks.IsOldAMD = true
	}

	var extensions string

	var numExtensions int32
	gl.GetIntegerv(gl.NUM_EXTENSIONS, &numExtensions)

	for i := int32(0); i < numExtensions; i++ {
		extensions += C.GoString((*C.char)(unsafe.Pointer(gl.GetStringi(gl.EXTENSIONS, uint32(i)))))
		extensions += " "
	}

	log.Println("GL Vendor:    ", glVendor)
	log.Println("GL Renderer:  ", glRenderer)
	log.Println("GL Version:   ", glVersion)
	log.Println("GLSL Version: ", glslVersion)
	log.Println("GL Extensions:", extensions)
	log.Println("OpenGL initialized!")

	if debugLogs {
		gl.Enable(gl.DEBUG_OUTPUT)
		gl.DebugMessageCallback(func(
			source uint32,
			glType uint32,
			id uint32,
			severity uint32,
			length int32,
			message string,
			userParam unsafe.Pointer) {
			log.Println("GL:", message)
		}, gl.Ptr(nil))

		gl.DebugMessageControl(gl.DONT_CARE, gl.DONT_CARE, gl.DONT_CARE, 0, nil, true)
	}

	return nil
}

func extensionCheck(additionalExtensions []string) error {
	extensions := []string{
		"GL_ARB_clear_texture",
		"GL_ARB_direct_state_access",
		"GL_ARB_texture_storage",
		"GL_ARB_vertex_attrib_binding",
		"GL_ARB_buffer_storage",
	}

	if additionalExtensions != nil {
		extensions = append(extensions, additionalExtensions...)
	}

	var notSupported []string

	for _, ext := range extensions {
		if !glfw.ExtensionSupported(ext) {
			notSupported = append(notSupported, ext)
		}
	}

	if len(notSupported) > 0 {
		return fmt.Errorf("your GPU does not support one or more required OpenGL extensions: %s. Please update your graphics drivers or upgrade your GPU", notSupported)
	}

	return nil
}
