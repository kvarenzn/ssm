// Copyright (C) 2024, 2025 kvarenzn
// SPDX-License-Identifier: GPL-3.0-or-later

package av

// #cgo pkg-config: libavformat libavcodec libavutil libswscale
// #include <libavformat/avformat.h>
// #include <libavcodec/avcodec.h>
// #include <libavutil/error.h>
// #include <libavutil/avutil.h>
// #include <libswscale/swscale.h>
import "C"

import (
	"errors"
	"image"
	"sync"
	"sync/atomic"
	"unsafe"
)

// ref: app/src/demuxer.c @ Genymobile/scrcpy
const (
	SC_PACKET_FLAG_CONFIG    = 1 << 63
	SC_PACKET_FLAG_KEY_FRAME = 1 << 62
	SC_PACKET_PTS_MASK       = SC_PACKET_FLAG_KEY_FRAME - 1
)

func PtrAdd[T any](ptr *T, offset int) unsafe.Pointer {
	var zero T
	p := unsafe.Pointer(ptr)
	off := uintptr(offset)
	return unsafe.Pointer(uintptr(p) + off*unsafe.Sizeof(zero))
}

type slot struct {
	Pic *image.NRGBA

	users     atomic.Int32
	busy      atomic.Bool
	available atomic.Bool
}

func (f *slot) acquire() {
	f.users.Add(1)
}

func (f *slot) release() {
	f.users.Add(-1)
}

func (f *slot) idle() bool {
	return f.users.Load() == 0
}

type AVDecoder struct {
	config    []byte
	codec     *C.AVCodec
	ctx       *C.AVCodecContext
	swsCtx    *C.SwsContext
	rgbFrame  *C.AVFrame
	needMerge bool

	slots        []*slot
	latestFrame  atomic.Int32
	nextIdx      int32 // 指向下一个待写入的槽位
	cond         *sync.Cond
	frameVersion atomic.Uint64
}

var (
	ErrCodecOpenFailed     = errors.New("failed to open avcodec")
	ErrOutOfMemory         = errors.New("out of memory")
	ErrSendPacketFailed    = errors.New("failed to send packet")
	ErrDecodeFailed        = errors.New("decode error")
	ErrCannotAllocateFrame = errors.New("failed to allocate frame buffer")
)

func NewAVDecoder(id string) (*AVDecoder, error) {
	const ringBufferSize = 10

	var codecId uint32 = C.AV_CODEC_ID_NONE
	needMerge := false
	switch id {
	case "h264":
		codecId = C.AV_CODEC_ID_H264
		needMerge = true
	case "h265":
		codecId = C.AV_CODEC_ID_H265
		needMerge = true
	case "av1\x00":
		codecId = C.AV_CODEC_ID_AV1
	case "opus":
		codecId = C.AV_CODEC_ID_OPUS
	case "aac\x00":
		codecId = C.AV_CODEC_ID_AAC
	case "flac":
		codecId = C.AV_CODEC_ID_FLAC
	case "raw\x00":
		codecId = C.AV_CODEC_ID_PCM_S16LE
	}

	C.av_log_set_level(C.AV_LOG_QUIET) // no log please

	codec := C.avcodec_find_decoder(codecId)
	ctx := C.avcodec_alloc_context3(codec)

	if C.avcodec_open2(ctx, codec, nil) != 0 {
		return nil, ErrCodecOpenFailed
	}

	frames := make([]*slot, ringBufferSize)
	for i := range ringBufferSize {
		frames[i] = &slot{}
	}

	return &AVDecoder{
		needMerge: needMerge,
		codec:     codec,
		ctx:       ctx,

		slots: frames,
		cond:  sync.NewCond(&sync.Mutex{}),
	}, nil
}

func (d *AVDecoder) Drop() {
	if d.ctx != nil {
		C.avcodec_free_context(&d.ctx)
	}

	if d.swsCtx != nil {
		C.sws_freeContext(d.swsCtx)
	}

	if d.rgbFrame != nil {
		C.av_frame_unref(d.rgbFrame)
		C.av_frame_free(&d.rgbFrame)
	}
}

func (d *AVDecoder) Get() *slot {
	f := d.slots[d.latestFrame.Load()]
	f.acquire()
	return f
}

func (d *AVDecoder) Put(f *slot) {
	f.release()
}

func (d *AVDecoder) WaitForNewFrame(lastFrameVersion *uint64) {
	for {
		current := d.frameVersion.Load()
		if current != *lastFrameVersion {
			*lastFrameVersion = current
			return
		}

		d.cond.L.Lock()
		d.cond.Wait()
		d.cond.L.Unlock()
	}
}

func (d *AVDecoder) notifyAll() {
	d.frameVersion.Add(1)

	d.cond.L.Lock()
	d.cond.Broadcast()
	d.cond.L.Unlock()
}

func (d *AVDecoder) writeBuffer() {
	N := int32(len(d.slots))

	// find empty buffer to write
	selected := int32(-1)
	for range N {
		if d.slots[d.nextIdx].idle() {
			selected = d.nextIdx
			break
		}

		d.nextIdx = (d.nextIdx + 1) % N
	}

	if selected == -1 {
		selected = N
		d.slots = append(d.slots, &slot{})
		N++
	}

	width, height := int(d.rgbFrame.width), int(d.rgbFrame.height)

	f := d.slots[selected]
	f.busy.Store(true)
	if f.Pic == nil || !f.Pic.Rect.Eq(image.Rect(0, 0, width, height)) {
		f.Pic = image.NewNRGBA(image.Rect(0, 0, width, height))
	}

	pix := f.Pic.Pix
	data := unsafe.Slice(d.rgbFrame.data[0], d.rgbFrame.linesize[0]*d.rgbFrame.height)
	for y := range height {
		for x := range width {
			pix[((y*width)+x)*4+0] = uint8(data[y*int(d.rgbFrame.linesize[0])+x*3+0])
			pix[((y*width)+x)*4+1] = uint8(data[y*int(d.rgbFrame.linesize[0])+x*3+1])
			pix[((y*width)+x)*4+2] = uint8(data[y*int(d.rgbFrame.linesize[0])+x*3+2])
			pix[((y*width)+x)*4+3] = 0xff
		}
	}

	f.busy.Store(false)
	d.latestFrame.Store(selected)
	d.nextIdx = (selected + 1) % N

	d.notifyAll()
}

func (d *AVDecoder) Decode(pts uint64, data []byte) error {
	packet := C.av_packet_alloc()
	frame := C.av_frame_alloc()

	C.av_new_packet(packet, C.int(len(data)))
	C.memcpy(unsafe.Pointer(packet.data), unsafe.Pointer(&data[0]), C.size_t(len(data)))

	if pts&SC_PACKET_FLAG_CONFIG != 0 {
		packet.pts = C.AV_NOPTS_VALUE
		d.config = data
	} else {
		packet.pts = C.int64_t(pts & SC_PACKET_PTS_MASK)

		// merge config packet if needed
		if d.config != nil {
			if C.av_grow_packet(packet, C.int(len(d.config))) != 0 {
				return ErrOutOfMemory
			}

			C.memmove(PtrAdd(packet.data, len(d.config)), unsafe.Pointer(packet.data), C.size_t(len(data)))
			C.memcpy(unsafe.Pointer(packet.data), unsafe.Pointer(&d.config[0]), C.size_t(len(d.config)))
			d.config = nil
		}
	}

	if pts&SC_PACKET_FLAG_KEY_FRAME != 0 {
		packet.flags |= C.AV_PKT_FLAG_KEY
	}

	packet.dts = packet.pts

	ret := C.avcodec_send_packet(d.ctx, packet)
	if ret < 0 {
		return ErrSendPacketFailed
	}

	C.av_packet_unref(packet)

	for {
		ret = C.avcodec_receive_frame(d.ctx, frame)
		if ret == C.AVERROR_EOF || ret == -C.EAGAIN {
			break
		} else if ret < 0 {
			return ErrDecodeFailed
		}

		if d.swsCtx == nil || d.swsCtx.src_w != frame.width || d.swsCtx.src_h != frame.height {
			if d.swsCtx != nil {
				C.sws_freeContext(d.swsCtx)
			}

			d.swsCtx = C.sws_getContext(frame.width, frame.height, d.ctx.pix_fmt, frame.width, frame.height, C.AV_PIX_FMT_RGB24, C.SWS_BILINEAR, nil, nil, nil)
		}

		if d.rgbFrame == nil || d.rgbFrame.width != frame.width || d.rgbFrame.height != frame.height {
			if d.rgbFrame != nil {
				C.av_frame_unref(d.rgbFrame)
				C.av_frame_free(&d.rgbFrame)
			}

			d.rgbFrame = C.av_frame_alloc()
			d.rgbFrame.format = C.AV_PIX_FMT_RGB24
			d.rgbFrame.width = frame.width
			d.rgbFrame.height = frame.height

			if C.av_frame_get_buffer(d.rgbFrame, 0) < 0 {
				return ErrCannotAllocateFrame
			}
		}

		C.sws_scale(d.swsCtx, &frame.data[0], &frame.linesize[0], 0, frame.height, &d.rgbFrame.data[0], &d.rgbFrame.linesize[0])

		d.writeBuffer()

		C.av_frame_unref(frame)
	}

	C.av_packet_free(&packet)
	C.av_frame_free(&frame)

	return nil
}
