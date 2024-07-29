package decoders

func CopyBlockBuffer(bx, by, w, h, bw, bh int, buffer, out []byte) {
	x := bw * bx
	var xl int
	if bw*(bx+1) > w {
		xl = w - bw*bx
	} else {
		xl = bw
	}
	xl *= 4

	end := bw * bh * 4
	ptr := 0

	for y := by * bh; ptr < end && y < h; ptr, y = ptr+bw*4, y+1 {
		start := ((h-y-1)*w + x) * 4
		copy(out[start:start+xl], buffer[ptr:ptr+xl])
	}
}
