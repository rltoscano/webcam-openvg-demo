package openvg

// #cgo CFLAGS: -I/opt/vc/include
// #cgo LDFLAGS: -L/opt/vc/lib -lbrcmOpenVG -lbrcmGLESv2
// #include "VG/openvg.h"
import "C"
import "errors"

func SetClearColor() {
	clearColor := []C.float{1, 1, 1, 1}
	C.vgSetfv(C.VG_CLEAR_COLOR, 4, &clearColor[0])
}

func Clear(x, y, w, h int) {
	C.vgClear(C.VGint(x), C.VGint(y), C.VGint(w), C.VGint(h))
}

type ImageFormat C.VGImageFormat

const (
	ImageFormatSrgbx8888 ImageFormat = C.VG_sRGBX_888
)

type ImageQuality C.VGImageQuality

const (
	ImageQualityNonantialiased = C.VG_IMAGE_QUALITY_NONANTIALIASED
	ImageQualityFaster         = C.VG_IMAGE_QUALITY_FASTER
	ImageQualityBetter         = C.VG_IMAGE_QUALITY_BETTER
)

type Image struct {
	handle C.VGImage
}

func CreateImage(format ImageFormat, width, height int, quality []ImageQuality) (Image, error) {
	qualityBitfield = C.VGbitfield(0)
	for _, q := range quality {
		qualityBitField |= q
	}
	handle := C.vgCreateImage(format, C.VGint(width), C.VGint(height), qualityBitField)
	if handle == C.VG_INVALID_HANDLE {
		return Image{}, errors.New("failed to create image: %s", errNames[C.vgGetError()])
	}
	return Image{handle}, nil
}

func (img Image) Destroy() error {
	C.vgDestroyImage(img.handle)
	err := C.vgGetError()
	if err != C.VG_NO_ERROR {
		return errors.New("failed to destroy image: %s", errNames[err])
	}
	return nil
}

func (img Image) WriteSubData() {
	// TODO
}

var errNames = map[C.VGErrorCode]string{
	C.VG_NO_ERROR:                       "VG_NO_ERROR",
	C.VG_BAD_HANDLE_ERROR:               "VG_BAD_HANDLE_ERROR",
	C.VG_UNSUPPORTED_IMAGE_FORMAT_ERROR: "VG_UNSUPPORTED_IMAGE_FORMAT_ERROR",
	C.VG_ILLEGAL_ARGUMENT_ERROR:         "VG_ILLEGAL_ARGUMENT_ERROR",
}