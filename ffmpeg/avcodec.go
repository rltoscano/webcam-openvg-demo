package ffmpeg

/*
  #cgo pkg-config: libavcodec
  #include <libavcodec/avcodec.h>
*/
import "C"
import (
	"errors"
	"fmt"
	"unsafe"
)

func init() {
	C.avcodec_register_all()
}

// CodecID represents an AVCodecID.
type CodecID C.enum_AVCodecID

// Supported codec IDs.
const (
	CodecIDRawVideo = CodecID(C.AV_CODEC_ID_RAWVIDEO)
)

// CodecParameters wraps a AVCodecParameters.
type CodecParameters struct {
	cptr *C.AVCodecParameters
}

// CodecID wraps AVCodecParameters.codec_id.
func (p CodecParameters) CodecID() CodecID {
	return CodecID(C.enum_AVCodecID(p.cptr.codec_id))
}

// Codec wraps an AVCodec.
type Codec struct {
	cptr *C.AVCodec
}

// FindDecoder wraps avcodec_find_decoder.
func FindDecoder(id CodecID) (Codec, error) {
	cptr := C.avcodec_find_decoder(C.enum_AVCodecID(id))
	if cptr == nil {
		return Codec{}, fmt.Errorf("ffmpeg: could not find decoder for %v", id)
	}
	return Codec{cptr}, nil
}

// CodecID wraps AVCodec.id.
func (c Codec) CodecID() CodecID {
	return CodecID(c.cptr.id)
}

// CodecContext wraps an AVCodecContext.
type CodecContext struct {
	cptr *C.AVCodecContext
}

// NewContext wraps avcodec_alloc_context3.
func (c Codec) NewContext(params CodecParameters) (CodecContext, error) {
	cptr := C.avcodec_alloc_context3(c.cptr)
	if cptr == nil {
		return CodecContext{}, fmt.Errorf("ffmpeg: failed creating context for codec ID %v", c.CodecID())
	}
	if result := C.avcodec_parameters_to_context(cptr, params.cptr); result < 0 {
		return CodecContext{}, fmt.Errorf("ffmpeg: failed to copy parameters to new context: %v", getErrStr(result))
	}
	if result := C.avcodec_open2(cptr, c.cptr, nil /*options*/); result < 0 {
		return CodecContext{}, fmt.Errorf("ffmpeg: failed to open codec context: %v", getErrStr(result))
	}
	return CodecContext{cptr}, nil
}

// Free wraps avcodec_free_context.
func (ctx CodecContext) Free() {
	C.avcodec_free_context(&ctx.cptr)
}

// Width wraps AVCodecContext.width.
func (ctx CodecContext) Width() int {
	return int(ctx.cptr.width)
}

// Height wraps AVCodecContext.height.
func (ctx CodecContext) Height() int {
	return int(ctx.cptr.height)
}

// SendPacket wraps avcodec_send_packet.
func (ctx CodecContext) SendPacket(p Packet) error {
	if result := C.avcodec_send_packet(ctx.cptr, p.cptr); result < 0 {
		return fmt.Errorf("ffmpeg: failed to send package: [%d] %s", result, getErrStr(result))
	}
	return nil
}

// ReceiveFrame wraps avcodec_receive_frame.
func (ctx CodecContext) ReceiveFrame(frame *Frame) error {
	if result := C.avcodec_receive_frame(ctx.cptr, frame.cptr); result < 0 {
		return fmt.Errorf("ffmpeg: failed receiving frame: [%d] %s", result, getErrStr(result))
	}
	return nil
}

// Frame wraps an AVFrame.
type Frame struct {
	cptr *C.AVFrame
}

// NewFrame wraps av_frame_alloc.
func NewFrame() (Frame, error) {
	f := Frame{C.av_frame_alloc()}
	if f.cptr == nil {
		return f, errors.New("ffmpeg: failed to allocate a new frame")
	}
	return f, nil
}

// Free wraps av_frame_free.
func (f Frame) Free() {
	C.av_frame_free(&f.cptr)
}

// Linesize wraps AVFrame.linesize[0].
func (f Frame) Linesize() int {
	return int(f.cptr.linesize[0])
}

// Data wraps AVFrame.data[0].
func (f Frame) Data() unsafe.Pointer {
	return unsafe.Pointer(f.cptr.data[0])
}
