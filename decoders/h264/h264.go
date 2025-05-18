package h264

// #cgo pkg-config: libavformat libavcodec libavutil

// #include <libavformat/avformat.h>
// #include <libavcodec/avcodec.h>
// #include <libavutil/error.h>
// #include <libavutil/avutil.h>
import "C"

import (
	"unsafe"

	"github.com/kvarenzn/ssm/log"
)

// ref: app/demuxer.c in Genymobile/scrcpy
const (
	SC_PACKET_FLAG_CONFIG    = 1 << 63
	SC_PACKET_FLAG_KEY_FRAME = 1 << 62
	SC_PACKET_PTS_MASK       = SC_PACKET_FLAG_KEY_FRAME - 1
)

func DecodeH264(pts uint64, data []byte) {
	codec := C.avcodec_find_decoder(C.AV_CODEC_ID_H264)

	codecCtx := C.avcodec_alloc_context3(codec)

	if C.avcodec_open2(codecCtx, codec, nil) != 0 {
		log.Fatal("failed to open avcodec")
	}

	// [TODO] we might need to apply config data from other packet

	packet := C.av_packet_alloc()
	frame := C.av_frame_alloc()

	C.av_new_packet(packet, C.int(len(data)))
	C.memcpy(unsafe.Pointer(packet.data), unsafe.Pointer(&data[0]), C.size_t(len(data)))

	if pts&SC_PACKET_FLAG_CONFIG != 0 {
		packet.pts = C.AV_NOPTS_VALUE
	} else {
		packet.pts = C.int64_t(pts & SC_PACKET_PTS_MASK)
	}

	if pts&SC_PACKET_FLAG_KEY_FRAME != 0 {
		packet.flags |= C.AV_PKT_FLAG_KEY
	}

	packet.dts = packet.pts

	ret := C.avcodec_send_packet(codecCtx, packet)
	if ret < 0 {
		log.Fatal("failed to send packet")
	}

	C.av_packet_unref(packet)

	for {
		ret = C.avcodec_receive_frame(codecCtx, frame)
		if ret == C.AVERROR_EOF || ret == -C.EAGAIN {
			break
		} else if ret < 0 {
			log.Fatal("decode error")
		}

		// [TODO] process decoded frame

		C.av_frame_unref(frame)
	}

	C.av_packet_free(&packet)
	C.av_frame_free(&frame)
	C.avcodec_free_context(&codecCtx)
}
