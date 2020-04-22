// Package bcmhost contains bindings for the functions declared in bcm_host.h
// (and its includes).
// TODO(robert): Consider refactoring this package to just bind dispmanx.
package bcmhost

//#cgo CFLAGS: -I/opt/vc/include
//#cgo LDFLAGS: -L/opt/vc/lib/ -lbcm_host
//#include "bcm_host.h"
//#include "EGL/eglplatform.h"
import "C"
import (
	"errors"
	"fmt"
	"unsafe"

	"../egl"
)

// Consts from Dispmanx.
const (
	DispmanxIDMainLcd       = 0
	DispmanxProtectionNone  = 0
	DispmanxDefaultResource = 0
)

const dispmanxNoHandle = 0

type Protection int

const (
	ProtectionNone Protection = 0
)

// Init wraps bcm_host_init.
func Init() {
	C.bcm_host_init()
}

// Deinit wraps bcm_host_deinit.
func Deinit() {
	C.bcm_host_deinit()
}

// GraphicsGetDisplaySize wraps graphics_get_display_size and returns width and
// height.
func GraphicsGetDisplaySize(display int) (int, int, error) {
	var w uint32
	var h uint32
	result := C.graphics_get_display_size(
		C.uint16_t(display),
		(*C.uint32_t)(&w),
		(*C.uint32_t)(&h),
	)
	if result < 0 {
		return 0, 0, fmt.Errorf("could not get display size for display \"%d\"", display)
	}
	return int(w), int(h), nil
}

// From https://elinux.org/Raspberry_Pi_VideoCore_APIs#vc_dispmanx_.2A
// vc_dispmanx_* -- Dispmanx is a windowing system in the process of being
// deprecated in favour of OpenWF (or similar), however dispmanx is still used
// in all API demos and it's replacement may not yet be available.

// DispmanxDisplay represents a handle to a Display.
type DispmanxDisplay struct {
	handle C.DISPMANX_DISPLAY_HANDLE_T
	id     int
}

// DispmanxDisplayOpen wraps vc_dispmanx_display_open.
func DispmanxDisplayOpen(display int) (DispmanxDisplay, error) {
	handle := C.vc_dispmanx_display_open(C.uint32_t(display))
	if handle == dispmanxNoHandle {
		return DispmanxDisplay{}, fmt.Errorf("could not open display \"%d\"", display)
	}
	return DispmanxDisplay{handle, display}, nil
}

// Close closes a Display.
func (d DispmanxDisplay) Close() error {
	result := C.vc_dispmanx_display_close(d.handle)
	if result != 0 {
		return fmt.Errorf("could not close display \"%d\"", d.id)
	}
	return nil
}

// DispmanxUpdate represents a handle to a Display update.
type DispmanxUpdate struct {
	display DispmanxDisplay
	handle  C.DISPMANX_UPDATE_HANDLE_T
}

// UpdateStart starts an update.
func (d DispmanxDisplay) UpdateStart(priority int) (DispmanxUpdate, error) {
	handle := C.vc_dispmanx_update_start(C.int32_t(priority))
	if handle == dispmanxNoHandle {
		return DispmanxUpdate{}, errors.New("could not start update")
	}
	return DispmanxUpdate{d, handle}, nil
}

func (d DispmanxUpdate) UpdateSubmit() error {
	result := C.vc_dispmanx_update_submit_sync(d.handle)
	if result != 0 {
		return errors.New("could not submit display update")
	}
	return nil
}

type DispmanxElement struct {
	update DispmanxUpdate
	handle C.DISPMANX_ELEMENT_HANDLE_T
}

// TODO(robert): Bind types for srcHandle.
// TODO(robert): Consider using a parameter struct since so many params.
func (u DispmanxUpdate) ElementAdd(
	layer int, dest Rect, srcHandle int, src Rect, protection Protection) (DispmanxElement, error) {
	// TODO(robert): Avoid extraneous memory allocations.
	cDest := C.VC_RECT_T{
		C.int32_t(dest.X),
		C.int32_t(dest.Y),
		C.int32_t(dest.Width),
		C.int32_t(dest.Height),
	}
	cSrc := C.VC_RECT_T{
		C.int32_t(src.X),
		C.int32_t(src.Y),
		C.int32_t(src.Width),
		C.int32_t(src.Height),
	}
	// TODO(robert): Support alpha, clamp, transform. These are not declared const
	// in the C code, so I'm guessing they will be modified? Need to figure that
	// out.
	alpha := C.VC_DISPMANX_ALPHA_T{}
	clamp := C.DISPMANX_CLAMP_T{}
	handle := C.vc_dispmanx_element_add(
		u.handle,
		u.display.handle,
		C.int32_t(layer),
		&cDest,
		C.uint32_t(srcHandle),
		&cSrc,
		C.uint32_t(protection),
		&alpha,
		&clamp,
		C.DISPMANX_TRANSFORM_T(0))
	if handle == dispmanxNoHandle {
		return DispmanxElement{}, errors.New("could not add element to display update")
	}
	return DispmanxElement{u, handle}, nil
}

type Rect struct {
	X, Y, Width, Height int
}

// DispmanxWindow implements egl.NativeWindow.
type DispmanxWindow struct {
	handle C.EGL_DISPMANX_WINDOW_T
}

func NewDispmanxWindow(element DispmanxElement, w, h int) DispmanxWindow {
	return DispmanxWindow{C.EGL_DISPMANX_WINDOW_T{element.handle, C.int(w), C.int(h)}}
}

func (w DispmanxWindow) Handle() egl.NativeWindowHandle {
	return egl.NativeWindowHandle(unsafe.Pointer(&w.handle))
}
