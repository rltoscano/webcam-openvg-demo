package egl

/*
  #cgo CFLAGS: -I/opt/vc/include
  #cgo LDFLAGS: -L/opt/vc/lib -lbrcmGLESv2 -lbrcmEGL
  #include "EGL/egl.h"
*/
import "C"
import (
	"errors"
	"fmt"
	"log"
	"unsafe"
)

type NativeDisplayType C.EGLNativeDisplayType

var (
	DefaultDisplay = NativeDisplayType(C.EGLNativeDisplayType(C.EGL_DEFAULT_DISPLAY))
)

type Api C.uint

const (
	APIOpenVG Api = C.EGL_OPENVG_API
)

type NativeWindowHandle C.EGLNativeWindowType

type NativeWindow interface {
	Handle() NativeWindowHandle
}

type Config struct {
	handle C.EGLConfig
}

var errNames = map[C.EGLint]string{
	C.EGL_NOT_INITIALIZED: "EGL_NOT_INITIALIZED",
	C.EGL_BAD_DISPLAY:     "EGL_BAD_DISPLAY",
}

type Display struct {
	handle C.EGLDisplay
}

func GetDisplay(display NativeDisplayType) (Display, error) {
	handle := C.eglGetDisplay(C.EGLNativeDisplayType(display))
	if handle == C.EGLDisplay(C.EGL_NO_DISPLAY) {
		return Display{}, fmt.Errorf("could not get display for \"%d\"", display)
	}
	major := C.EGLint(0)
	minor := C.EGLint(0)
	if C.eglInitialize(handle, &major, &minor) == C.EGL_FALSE {
		log.Printf("xxxfailed to initialize display: %s", errNames[C.eglGetError()])
	}
	return Display{handle}, nil
}

func (d Display) Initialize() (string, error) {
	major := C.EGLint(0)
	minor := C.EGLint(0)
	if C.eglInitialize(d.handle, &major, &minor) == C.EGL_FALSE {
		return "", fmt.Errorf("failed to initialize display: %s", errNames[C.eglGetError()])
	}
	return fmt.Sprintf("%d.%d", major, minor), nil
}

func (d Display) Terminate() error {
	if C.eglTerminate(d.handle) == C.EGL_FALSE {
		return errors.New("failed to terminate display")
	}
	return nil
}

func (d Display) ChooseConfig() (Config, error) {
	attribs := []C.EGLint{
		C.EGL_RED_SIZE, 8,
		C.EGL_GREEN_SIZE, 8,
		C.EGL_BLUE_SIZE, 8,
		C.EGL_ALPHA_SIZE, 8,
		C.EGL_LUMINANCE_SIZE, C.EGL_DONT_CARE,
		C.EGL_SAMPLES, 1,
		C.EGL_NONE,
	}
	cfg := C.EGLConfig(unsafe.Pointer(uintptr(0)))
	var numConfigs C.EGLint
	if C.eglChooseConfig(d.handle, &attribs[0], &cfg, 1 /*size*/, &numConfigs) == C.EGL_FALSE {
		return Config{}, fmt.Errorf("failed to choose configs: %s", errNames[C.eglGetError()])
	}
	if numConfigs != 1 {
		return Config{}, fmt.Errorf("wanted 1 matching config, got %d", numConfigs)
	}
	return Config{cfg}, nil
}

func (d Display) CreateWindowSurface(config Config, window NativeWindow) (Surface, error) {
	handle := C.eglCreateWindowSurface(d.handle, config.handle, C.EGLNativeWindowType(window.Handle()), (*C.EGLint)(unsafe.Pointer(uintptr(0))))
	if handle == C.EGLSurface(C.EGL_NO_SURFACE) {
		return Surface{}, fmt.Errorf("failed creating window surface: %v", errNames[C.eglGetError()])
	}
	return Surface{handle}, nil
}

func (d Display) CreateContext(config Config) (Context, error) {
	ctx := C.eglCreateContext(d.handle, config.handle, C.EGLContext(unsafe.Pointer(uintptr(0))), (*C.EGLint)(unsafe.Pointer(uintptr(0))))
	if ctx == C.EGLContext(C.EGL_NO_CONTEXT) {
		return Context{}, errors.New("failed to create context")
	}
	return Context{ctx}, nil
}

type Context struct {
	handle C.EGLContext
}

func (d Display) MakeCurrent(surface Surface, ctx Context) error {
	if C.eglMakeCurrent(d.handle, surface.handle, surface.handle, ctx.handle) == C.EGL_FALSE {
		return fmt.Errorf("failed to make context current: %s", errNames[C.eglGetError()])
	}
	return nil
}

func (d Display) SwapBuffers(surface Surface) error {
	if C.eglSwapBuffers(d.handle, surface.handle) == C.EGL_FALSE {
		return fmt.Errorf("failed to swap buffers: %s", errNames[C.eglGetError()])
	}
	return nil
}

type Surface struct {
	handle C.EGLSurface
}

func BindAPI(api Api) error {
	if C.eglBindAPI(C.uint(api)) == C.EGL_FALSE {
		return fmt.Errorf("could not bind API: %s", errNames[C.eglGetError()])
	}
	return nil
}
