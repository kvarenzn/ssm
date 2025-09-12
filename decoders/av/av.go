// Copyright (C) 2024, 2025 kvarenzn
// SPDX-License-Identifier: GPL-3.0-or-later

package av

// #cgo pkg-config: libavformat libavcodec libavutil
// #include <libavformat/avformat.h>
// #include <libavcodec/avcodec.h>
// #include <libavutil/error.h>
// #include <libavutil/avutil.h>
import "C"

import (
	"errors"
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

type AVDecoder struct {
	config    []byte
	codec     *C.AVCodec
	ctx       *C.AVCodecContext
	needMerge bool
}

var (
	ErrCodecOpenFailed  = errors.New("failed to open avcodec")
	ErrOutOfMemory      = errors.New("out of memory")
	ErrSendPacketFailed = errors.New("failed to send packet")
	ErrDecodeFailed     = errors.New("decode error")
)

func NewAVDecoder(id string) (*AVDecoder, error) {
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

	codec := C.avcodec_find_decoder(codecId)
	ctx := C.avcodec_alloc_context3(codec)

	if C.avcodec_open2(ctx, codec, nil) != 0 {
		return nil, ErrCodecOpenFailed
	}

	return &AVDecoder{
		needMerge: needMerge,
		codec:     codec,
		ctx:       ctx,
	}, nil
}

func (d *AVDecoder) Drop() {
	if d.ctx != nil {
		C.avcodec_free_context(&d.ctx)
	}
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

		// [TODO] process decoded frame

		C.av_frame_unref(frame)
	}

	C.av_packet_free(&packet)
	C.av_frame_free(&frame)

	return nil
}
