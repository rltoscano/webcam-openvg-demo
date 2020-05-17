package ffmpeg

import (
	"errors"
	"fmt"
	"strings"
	"unsafe"
)

/*
  #cgo amd64,windows LDFLAGS: -lavformat
  #cgo arm,linux pkg-config: libavformat libavcodec
  #include <libavformat/avformat.h>
  #include <libavutil/error.h>
*/
import "C"

// InputFormat wraps a AVInputFormat.
type InputFormat struct {
	cptr *C.AVInputFormat
}

// FormatContext wraps a AVFormatContext.
type FormatContext struct {
	cptr           *C.AVFormatContext
	SourceFilename string
}

// Packet wraps a AVPacket.
type Packet struct {
	cptr *C.AVPacket
}

// NewInputFormat wraps av_find_input_format.
func NewInputFormat(shortName string) (InputFormat, error) {
	cstr := C.CString(shortName)
	defer C.free(unsafe.Pointer(cstr))
	return InputFormat{C.av_find_input_format(cstr)}, nil
}

// NewFormatContext wraps avformat_open_input().
func NewFormatContext(filename string, inputFormat InputFormat) (FormatContext, error) {
	var ctxp *C.AVFormatContext
	filenamecs := C.CString(filename)
	defer C.free(unsafe.Pointer(filenamecs))
	if result := C.avformat_open_input(&ctxp, filenamecs, inputFormat.cptr, nil /*options*/); result < 0 {
		return FormatContext{}, fmt.Errorf("ffmpeg: failed to create context: %s", getErrStr(result))
	}
	return FormatContext{ctxp, filename}, nil
}

// Close wraps avformat_close_input.
func (ctx FormatContext) Close() {
	C.avformat_close_input(&ctx.cptr)
}

// ReadFrame wraps av_read_frame.
func (ctx FormatContext) ReadFrame() (Packet, error) {
	var p Packet
	if p.cptr = C.av_packet_alloc(); p.cptr == nil {
		return p, errors.New("ffmpeg: failed to alloc packet")
	}
	if result := C.av_read_frame(ctx.cptr, p.cptr); result < 0 {
		defer p.Free()
		return p, fmt.Errorf("ffmpeg: failed to read frame: [%d] %v", result, getErrStr(result))
	}
	return p, nil
}

// Dump wraps av_dump_format.
func (ctx FormatContext) Dump() {
	cstr := C.CString(ctx.SourceFilename)
	defer C.free(unsafe.Pointer(cstr))
	C.av_dump_format(ctx.cptr, 0, cstr, 0 /*is_output*/)
}

// Free unref counts the packet and frees it.
func (p Packet) Free() {
	C.av_packet_unref(p.cptr)
	C.av_packet_free(&p.cptr)
}

// Stream wraps an AVStream.
type Stream struct {
	cptr *C.AVStream
}

func (ctx FormatContext) GetStream(i int) (Stream, error) {
	q := uintptr(unsafe.Pointer(*ctx.cptr.streams)) + (uintptr(i) * unsafe.Sizeof(*ctx.cptr.streams))
	return Stream{(*C.AVStream)(unsafe.Pointer(q))}, nil
}

func (s Stream) Codecpar() CodecParameters {
	return CodecParameters{s.cptr.codecpar}
}

// getErrStr gets the corresponding error message for the given result code.
func getErrStr(result C.int) string {
	errStr := C.CString(strings.Repeat(" ", C.AV_ERROR_MAX_STRING_SIZE))
	defer C.free(unsafe.Pointer(errStr))
	C.av_strerror(result, errStr, C.AV_ERROR_MAX_STRING_SIZE)
	return C.GoString(errStr)
}
