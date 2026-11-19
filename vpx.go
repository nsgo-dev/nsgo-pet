package main

import (
	"errors"
	"fmt"
	"runtime"
	"sync"
	"unsafe"

	"github.com/ebitengine/purego"
)

// libvpx's decoder functions, bound by loadVPX through purego, so the app needs no cgo.
var vpx struct {
	vp9      func() uintptr
	initVer  func(ctx unsafe.Pointer, iface uintptr, cfg unsafe.Pointer, flags int, ver int32) int32
	decode   func(ctx unsafe.Pointer, data unsafe.Pointer, size uint32, userPriv uintptr, deadline int) int32
	getFrame func(ctx unsafe.Pointer, iter *uintptr) unsafe.Pointer
	destroy  func(ctx unsafe.Pointer) int32
	errText  func(err int32) string
}

var (
	vpxOnce sync.Once
	vpxErr  error
)

// loadVPX opens libvpx (openVPXLibrary) and binds its functions, once.
func loadVPX() error {
	vpxOnce.Do(func() {
		lib, err := openVPXLibrary()
		if err != nil {
			vpxErr = fmt.Errorf("loading libvpx: %w", err)
			return
		}
		purego.RegisterLibFunc(&vpx.vp9, lib, "vpx_codec_vp9_dx")
		purego.RegisterLibFunc(&vpx.initVer, lib, "vpx_codec_dec_init_ver")
		purego.RegisterLibFunc(&vpx.decode, lib, "vpx_codec_decode")
		purego.RegisterLibFunc(&vpx.getFrame, lib, "vpx_codec_get_frame")
		purego.RegisterLibFunc(&vpx.destroy, lib, "vpx_codec_destroy")
		purego.RegisterLibFunc(&vpx.errText, lib, "vpx_codec_err_to_string")
	})
	return vpxErr
}

// vpxDecoderABI is VPX_DECODER_ABI_VERSION of libvpx's headers, 3 + VPX_CODEC_ABI_VERSION (4 + VPX_IMAGE_ABI_VERSION, 5), which libvpx checks so a mismatched vpxImage below fails to open rather than misreads.
const vpxDecoderABI = 3 + 4 + 5

// vpxImgI420 is VPX_IMG_FMT_I420, 8-bit 4:2:0, the only format the clips use.
const vpxImgI420 = 0x102

// vpxCS709 and vpxFullRange are VPX_CS_BT_709 and VPX_CR_FULL_RANGE.
const (
	vpxCS709     = 2
	vpxFullRange = 1
)

// vpxImage is the start of vpx_image_t, up to the planes and their strides.
type vpxImage struct {
	fmt, cs, colorRange        uint32
	w, h, bitDepth             uint32
	dw, dh                     uint32
	rw, rh                     uint32
	xChromaShift, yChromaShift uint32
	planes                     [4]*byte
	stride                     [4]int32
}

// vp9Decoder is one libvpx VP9 decoder; a clip with alpha takes two, one for each stream. It is used on one goroutine.
type vp9Decoder struct {
	// ctx is a vpx_codec_ctx_t, 56 bytes on 64-bit Windows and Unix, with room to spare; Go's collector does not move it.
	ctx *[16]uint64
}

func newVP9Decoder() (*vp9Decoder, error) {
	if err := loadVPX(); err != nil {
		return nil, err
	}
	d := &vp9Decoder{ctx: new([16]uint64)}
	if e := vpx.initVer(unsafe.Pointer(d.ctx), vpx.vp9(), nil, 0, vpxDecoderABI); e != 0 {
		return nil, fmt.Errorf("opening a VP9 decoder: %s", vpx.errText(e))
	}
	return d, nil
}

// decode decodes one frame and returns its picture, valid until the next call; ok is false when the frame shows nothing.
func (d *vp9Decoder) decode(frame []byte) (p yuvPlanes, ok bool, err error) {
	if len(frame) == 0 {
		return p, false, errors.New("an empty frame")
	}
	if e := vpx.decode(unsafe.Pointer(d.ctx), unsafe.Pointer(&frame[0]), uint32(len(frame)), 0, 0); e != 0 {
		return p, false, fmt.Errorf("decoding: %s", vpx.errText(e))
	}
	runtime.KeepAlive(frame)
	var iter uintptr
	ptr := vpx.getFrame(unsafe.Pointer(d.ctx), &iter)
	if ptr == nil {
		return p, false, nil
	}
	img := (*vpxImage)(ptr)
	if img.fmt != vpxImgI420 {
		return p, false, fmt.Errorf("the picture's format is %#x, not 8-bit 4:2:0", img.fmt)
	}
	h, ch := int(img.dh), int(img.dh+1)/2
	p = yuvPlanes{
		width: int(img.dw), height: h,
		y: unsafe.Slice(img.planes[0], int(img.stride[0])*h), yStride: int(img.stride[0]),
		u: unsafe.Slice(img.planes[1], int(img.stride[1])*ch), uStride: int(img.stride[1]),
		v: unsafe.Slice(img.planes[2], int(img.stride[2])*ch), vStride: int(img.stride[2]),
		bt709: img.cs == vpxCS709, fullRange: img.colorRange == vpxFullRange,
	}
	return p, true, nil
}

// close frees the decoder.
func (d *vp9Decoder) close() {
	if d != nil && d.ctx != nil {
		vpx.destroy(unsafe.Pointer(d.ctx))
		d.ctx = nil
	}
}
