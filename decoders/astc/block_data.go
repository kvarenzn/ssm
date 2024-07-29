package astc

import (
	"encoding/binary"
)

type BlockData struct {
	BlockWidth         int
	BlockHeight        int
	Width              int
	Height             int
	PartCount          int
	DualPlane          bool
	PlaneSelector      int
	WeightRange        int
	WeightCount        int
	Cem                [4]int
	CemRange           int
	EndpointValueCount int
	Endpoints          [4][8]int
	Weights            [144][2]int
	Partition          [144]int
}

func (bd *BlockData) DecodeBlockParams(data []byte) {
	bd.DualPlane = data[1]&4 != 0
	bd.WeightRange = int(data[0])>>4&1 | int(data[1])<<2&8

	if data[0]&3 != 0 {
		bd.WeightRange |= int(data[0]) << 1 & 6
		switch data[0] & 0xc {
		case 0:
			bd.Width = int(binary.LittleEndian.Uint16(data))>>7&3 + 4
			bd.Height = int(data[0])>>5&3 + 2
		case 4:
			bd.Width = int(binary.LittleEndian.Uint16(data))>>7&3 + 8
			bd.Height = int(data[0])>>5&3 + 2
		case 8:
			bd.Width = int(data[0])>>5&3 + 2
			bd.Height = int(binary.LittleEndian.Uint16(data))>>7&3 + 8
		case 12:
			if data[1]&1 != 0 {
				bd.Width = int(data[0])>>7&1 + 2
				bd.Height = int(data[0])>>5&3 + 2
			} else {
				bd.Width = int(data[0])>>5&3 + 2
				bd.Height = int(data[0])>>7&1 + 6
			}
		}
	} else {
		bd.WeightRange |= int(data[0]) >> 1 & 6
		switch binary.LittleEndian.Uint16(data) & 0x180 {
		case 0:
			bd.Width = 12
			bd.Height = int(data[0])>>5&3 + 2
		case 0x80:
			bd.Width = int(data[0])>>5&3 + 2
			bd.Height = 12
		case 0x100:
			bd.Width = int(data[0])>>5&3 + 6
			bd.Height = int(data[1])>>1&3 + 6
			bd.DualPlane = false
			bd.WeightRange &= 7
		case 0x180:
			if data[0]&0x20 != 0 {
				bd.Width = 10
				bd.Height = 6
			} else {
				bd.Width = 6
				bd.Height = 10
			}
		}
	}

	bd.PartCount = int(data[1])>>3&3 + 1
	bd.WeightCount = bd.Width * bd.Height
	if bd.DualPlane {
		bd.WeightCount *= 2
	}

	weightBits := 0
	configBits := 0
	cemBase := 0

	switch weightPrecTableA[bd.WeightRange] {
	case 3:
		weightBits = bd.WeightCount*weightPrecTableB[bd.WeightRange] + (bd.WeightCount*8+4)/5
	case 5:
		weightBits = bd.WeightCount*weightPrecTableB[bd.WeightRange] + (bd.WeightCount*7+2)/3
	default:
		weightBits = bd.WeightCount * weightPrecTableB[bd.WeightRange]
	}

	if bd.PartCount == 1 {
		bd.Cem[0] = int(binary.LittleEndian.Uint16(data[1:])) >> 5 & 0xf
		configBits = 17
	} else {
		cemBase = int(binary.LittleEndian.Uint16(data[2:])) >> 7 & 3
		if cemBase == 0 {
			cem := int(data[3]) >> 1 & 0xf
			for i := 0; i < bd.PartCount; i++ {
				bd.Cem[i] = cem
			}
			configBits = 29
		} else {
			for i := 0; i < bd.PartCount; i++ {
				bd.Cem[i] = ((int(data[3]) >> (i + 1) & 1) + cemBase - 1) << 2
			}

			switch bd.PartCount {
			case 2:
				bd.Cem[0] |= int(data[3]) >> 3 & 0b11
				bd.Cem[1] |= getBits(data, 126-weightBits, 2)
			case 3:
				bd.Cem[0] |= int(data[3]) >> 4 & 1
				bd.Cem[0] |= getBits(data, 122-weightBits, 2) & 2
				bd.Cem[1] |= getBits(data, 124-weightBits, 2)
				bd.Cem[2] |= getBits(data, 126-weightBits, 2)
			case 4:
				for i := 0; i < 4; i++ {
					bd.Cem[i] |= getBits(data, 120+i*2-weightBits, 2)
				}
			}

			configBits = 25 + bd.PartCount*3
		}
	}

	if bd.DualPlane {
		configBits += 2
		var bit int
		if cemBase != 0 {
			bit = 130 - weightBits - bd.PartCount*3
		} else {
			bit = 126 - weightBits
		}
		bd.PlaneSelector = getBits(data, bit, 2)
	}

	remainBits := 128 - configBits - weightBits

	bd.EndpointValueCount = 0
	for i := 0; i < bd.PartCount; i++ {
		bd.EndpointValueCount += bd.Cem[i]>>1&6 + 2
	}

	var endpointBits int
	for i := 0; i < len(cemTableA); i++ {
		switch cemTableA[i] {
		case 3:
			endpointBits = bd.EndpointValueCount*cemTableB[i] + (bd.EndpointValueCount*8+4)/5
		case 5:
			endpointBits = bd.EndpointValueCount*cemTableB[i] + (bd.EndpointValueCount*7+2)/3
		default:
			endpointBits = bd.EndpointValueCount * cemTableB[i]
		}

		if endpointBits <= remainBits {
			bd.CemRange = i
			break
		}
	}
}

func setEndpoint(endpoint []int, r1, g1, b1, a1, r2, g2, b2, a2 int) {
	endpoint[0] = r1
	endpoint[1] = g1
	endpoint[2] = b1
	endpoint[3] = a1
	endpoint[4] = r2
	endpoint[5] = g2
	endpoint[6] = b2
	endpoint[7] = a2
}

func setEndpointClamp(endpoint []int, r1, g1, b1, a1, r2, g2, b2, a2 int) {
	endpoint[0] = int(clampU8(r1))
	endpoint[1] = int(clampU8(g1))
	endpoint[2] = int(clampU8(b1))
	endpoint[3] = int(clampU8(a1))
	endpoint[4] = int(clampU8(r2))
	endpoint[5] = int(clampU8(g2))
	endpoint[6] = int(clampU8(b2))
	endpoint[7] = int(clampU8(a2))
}

func setEndpointBlue(endpoint []int, r1, g1, b1, a1, r2, g2, b2, a2 int) {
	endpoint[0] = (r1 + b1) >> 1
	endpoint[1] = (g1 + b1) >> 1
	endpoint[2] = b1
	endpoint[3] = a1
	endpoint[4] = (r2 + b2) >> 1
	endpoint[5] = (g2 + b2) >> 1
	endpoint[6] = b2
	endpoint[7] = a2
}

func setEndpointBlueClamp(endpoint []int, r1, g1, b1, a1, r2, g2, b2, a2 int) {
	endpoint[0] = int(clampU8((r1 + b1) >> 1))
	endpoint[1] = int(clampU8((g1 + b1) >> 1))
	endpoint[2] = int(clampU8(b1))
	endpoint[3] = int(clampU8(a1))
	endpoint[4] = int(clampU8((r2 + b2) >> 1))
	endpoint[5] = int(clampU8((g2 + b2) >> 1))
	endpoint[6] = int(clampU8(b2))
	endpoint[7] = int(clampU8(a2))
}

func setEndpointHdrClamp(endpoint []int, r1, g1, b1, a1, r2, b2, g2, a2 int) {
	endpoint[0] = int(clampHdr(r1))
	endpoint[1] = int(clampHdr(g1))
	endpoint[2] = int(clampHdr(b1))
	endpoint[3] = int(clampHdr(a1))
	endpoint[4] = int(clampHdr(r2))
	endpoint[5] = int(clampHdr(g2))
	endpoint[6] = int(clampHdr(b2))
	endpoint[7] = int(clampHdr(a2))
}

type IntSeqData struct {
	Bits    int
	NonBits int
}

func DecodeIntSeq(data []byte, offset, a, b, count int, reverse bool, out []IntSeqData) {
	if count <= 0 {
		return
	}

	n := 0

	if a == 3 {
		mask := 1<<b - 1
		blockCount := (count + 4) / 5
		lastBlockCount := (count+4)%5 + 1
		blockSize := 8 + 5*b
		lastBlockSize := (blockSize*lastBlockCount + 4) / 5

		if reverse {
			for i, p := 0, offset; i < blockCount; i, p = i+1, p-blockSize {
				var nowSize int
				if i < blockCount-1 {
					nowSize = blockSize
				} else {
					nowSize = lastBlockSize
				}

				d := bitReverseU64(getBits64(data, p-nowSize, nowSize), nowSize)
				x := int(d>>b&3 | d>>(b*2)&0xc | d>>(b*3)&0x10 | d>>(b*4)&0x60 | d>>(b*5)&0x80)
				for j := 0; j < 5 && n < count; j, n = j+1, n+1 {
					out[n] = IntSeqData{
						Bits:    int(d >> uint64(intSeqMt[j]+b*j) & uint64(mask)),
						NonBits: intSeqTrintsTable[j][x],
					}
				}
			}
		} else {
			for i, p := 0, offset; i < blockCount; i, p = i+1, p+blockSize {
				var l int
				if i < blockCount-1 {
					l = blockSize
				} else {
					l = lastBlockSize
				}

				d := getBits64(data, p, l)
				x := int(d>>b&3 | d>>(b*2)&0xc | d>>(b*3)&0x10 | d>>(b*4)&0x60 | d>>(b*5)&0x80)
				for j := 0; j < 5 && n < count; j, n = j+1, n+1 {
					out[n] = IntSeqData{
						Bits:    int(d >> uint64(intSeqMt[j]+b*j) & uint64(mask)),
						NonBits: intSeqTrintsTable[j][x],
					}
				}
			}
		}
	} else if a == 5 {
		mask := 1<<b - 1
		blockCount := (count + 2) / 3
		lastBlockCount := (count+2)%3 + 1
		blockSize := 7 + 3*b
		lastBlockSize := (blockSize*lastBlockCount + 2) / 3

		if reverse {
			for i, p := 0, offset; i < blockCount; i, p = i+1, p-blockSize {
				var nowSize int
				if i < blockCount-1 {
					nowSize = blockSize
				} else {
					nowSize = lastBlockSize
				}

				d := bitReverseU64(getBits64(data, p-nowSize, nowSize), nowSize)
				x := int(d>>b&7 | d>>(b*2)&0x18 | d>>(b*3)&0x60)
				for j := 0; j < 3 && n < count; j, n = j+1, n+1 {
					out[n] = IntSeqData{
						Bits:    int(d >> uint64(intSeqMq[j]+b*j) & uint64(mask)),
						NonBits: intSeqQuintsTable[j][x],
					}
				}
			}
		} else {
			for i, p := 0, offset; i < blockCount; i, p = i+1, p+blockSize {
				var l int
				if i < blockCount-1 {
					l = blockSize
				} else {
					l = lastBlockSize
				}

				d := getBits64(data, p, l)
				x := int(d>>b&7 | d>>(b*2)&0x18 | d>>(b*3)&0x60)
				for j := 0; j < 3 && n < count; j, n = j+1, n+1 {
					out[n] = IntSeqData{
						Bits:    int(d >> uint64(intSeqMq[j]+b*j) & uint64(mask)),
						NonBits: intSeqQuintsTable[j][x],
					}
				}
			}
		}
	} else {
		if reverse {
			for p := offset - b; n < count; n, p = n+1, p-b {
				out[n] = IntSeqData{
					Bits:    int(bitReverseU8(byte(getBits(data, p, b)), b)),
					NonBits: 0,
				}
			}
		} else {
			for p := offset; n < count; n, p = n+1, p+b {
				out[n] = IntSeqData{
					Bits:    getBits(data, p, b),
					NonBits: 0,
				}
			}
		}
	}
}

func decodeEndpointsHdr7(endpoints []int, v []int) {
	modeValue := v[2]>>4&0x8 | v[1]>>5&0x4 | v[0]>>6
	var majorComponent, mode int
	if modeValue&0xc != 0xc {
		majorComponent = modeValue >> 2
		mode = modeValue & 3
	} else if modeValue != 0xf {
		majorComponent = modeValue & 3
		mode = 4
	} else {
		majorComponent = 0
		mode = 5
	}

	c := []int{v[0] & 0x3f, v[1] & 0x1f, v[2] & 0x1f, v[3] & 0x1f}

	switch mode {
	case 0:
		c[3] |= v[3] & 0x60
		c[0] |= v[3] >> 1 & 0x40
		c[0] |= v[2] << 1 & 0x80
		c[0] |= v[1] << 3 & 0x300
		c[0] |= v[2] << 5 & 0x400
		c[0] <<= 1
		c[1] <<= 1
		c[2] <<= 1
		c[3] <<= 1
	case 1:
		c[1] |= v[1] & 0x20
		c[2] |= v[2] & 0x20
		c[0] |= v[3] >> 1 & 0x40
		c[0] |= v[2] << 1 & 0x80
		c[0] |= v[1] << 2 & 0x100
		c[0] |= v[3] << 4 & 0x600
		c[0] <<= 1
		c[1] <<= 1
		c[2] <<= 1
		c[3] <<= 1
	case 2:
		c[3] |= v[3] & 0xe0
		c[0] |= v[2] << 1 & 0xc0
		c[0] |= v[1] << 3 & 0x300
		c[0] <<= 2
		c[1] <<= 2
		c[2] <<= 2
		c[3] <<= 2
	case 3:
		c[1] |= v[1] & 0x20
		c[2] |= v[2] & 0x20
		c[3] |= v[3] & 0x60
		c[0] |= v[3] >> 1 & 0x40
		c[0] |= v[2] << 1 & 0x80
		c[0] |= v[1] << 2 & 0x100
		c[0] <<= 3
		c[1] <<= 3
		c[2] <<= 3
		c[3] <<= 3
	case 4:
		c[1] |= v[1] & 0x60
		c[2] |= v[2] & 0x60
		c[3] |= v[3] & 0x20
		c[0] |= v[3] >> 1 & 0x40
		c[0] |= v[3] << 1 & 0x80
		c[0] <<= 4
		c[1] <<= 4
		c[2] <<= 4
		c[3] <<= 4
	case 5:
		c[1] |= v[1] & 0x60
		c[2] |= v[2] & 0x60
		c[3] |= v[3] & 0x60
		c[0] |= v[3] >> 1 & 0x40
		c[0] <<= 5
		c[1] <<= 5
		c[2] <<= 5
		c[3] <<= 5
	}

	if mode != 5 {
		c[1] = c[0] - c[1]
		c[2] = c[0] - c[2]
	}

	if majorComponent == 1 {
		setEndpointHdrClamp(endpoints, c[1]-c[3], c[0]-c[3], c[2]-c[3], 0x780, c[1], c[0], c[2], 0x780)
	} else if majorComponent == 2 {
		setEndpointHdrClamp(endpoints, c[2]-c[3], c[1]-c[3], c[0]-c[3], 0x780, c[2], c[1], c[0], 0x780)
	} else {
		setEndpointHdrClamp(endpoints, c[0]-c[3], c[1]-c[3], c[2]-c[3], 0x780, c[0], c[1], c[2], 0x780)
	}
}

func decodeEndpointsHdr11(endpoints []int, v []int, alpha1, alpha2 int) {
	majorComponent := (v[4] >> 7) | (v[5] >> 6 & 2)
	if majorComponent == 3 {
		setEndpoint(endpoints, v[0]<<4, v[2]<<4, v[4]<<5&0xfe0, alpha1, v[1]<<4, v[3]<<4, v[5]<<5&0xfe0, alpha2)
		return
	}

	mode := v[1]>>7 | v[2]>>6&2 | v[3]>>5&4
	va := v[0] | v[1]<<2&0x100
	vb0, vb1 := v[2]&0x3f, v[3]&0x3f
	vc := v[1] & 0x3f

	var vd0, vd1 int

	switch mode {
	case 0, 2:
		vd0 = v[4] & 0x7f
		if vd0&0x40 != 0 {
			vd0 |= 0xff80
		}
		vd1 = v[5] & 0x7f
		if vd0&0x40 != 0 {
			vd1 |= 0xff80
		}
	case 1, 3, 5, 7:
		vd0 = v[4] & 0x3f
		if vd0&0x20 != 0 {
			vd0 |= 0xffc0
		}
		vd1 = v[5] & 0x3f
		if vd1&0x20 != 0 {
			vd1 |= 0xffc0
		}
	default:
		vd0 = v[4] & 0x1f
		if vd0&0x10 != 0 {
			vd0 |= 0xffe0
		}
		vd1 = v[5] & 0x1f
		if vd1&0x10 != 0 {
			vd1 |= 0xffe0
		}
	}

	switch mode {
	case 0:
		vb0 |= v[2] & 0x40
		vb1 |= v[3] & 0x40
	case 1:
		vb0 |= v[2]&0x40 | v[4]<<1&0x80
		vb1 |= v[3]&0x40 | v[5]<<1&0x80
	case 2:
		va |= v[2] << 3 & 0x200
		vc |= v[3] & 0x40
	case 3:
		va |= v[4] << 3 & 0x200
		vc |= v[5] & 0x40
		vb0 |= v[2] & 0x40
		vb1 |= v[3] & 0x40
	case 4:
		va |= v[4]<<4&0x200 | v[5]<<5&0x400
		vb0 |= v[2]&0x40 | v[4]<<1&0x80
		vb1 |= v[3]&0x40 | v[5]<<1&0x80
	case 5:
		va |= v[2]<<3&0x200 | v[3]<<4&0x400
		vc |= v[5]&0x40 | v[4]<<1&0x80
	case 6:
		va |= v[4]<<4&0x200 | v[5]<<5&0x400 | v[4]<<5&0x800
		vc |= v[5] & 0x40
		vb0 |= v[2] & 0x40
		vb1 |= v[3] & 0x40
	case 7:
		va |= v[2]<<3&0x200 | v[3]<<4&0x400 | v[4]<<5&0x800
		vc |= v[5] & 0x40
	}

	shamt := mode>>1 ^ 3
	va <<= shamt
	vb0 <<= shamt
	vb1 <<= shamt
	vc <<= shamt
	mult := 1 << shamt
	vd0 *= mult
	vd1 *= mult

	if majorComponent == 1 {
		setEndpointHdrClamp(endpoints, va-vb0-vc-vd0, va-vc, va-vb1-vc-vd1, alpha1, va-vb0, va, va-vb1, alpha2)
	} else if majorComponent == 2 {
		setEndpointHdrClamp(endpoints, va-vb1-vc-vd1, va-vb0-vc-vd0, va-vc, alpha1, va-vb1, va-vb0, va, alpha2)
	} else {
		setEndpointHdrClamp(endpoints, va-vc, va-vb0-vc-vd0, va-vb1-vc-vd1, alpha1, va, va-vb0, va-vb1, alpha2)
	}
}

func (bd *BlockData) DecodeEndpoints(data []byte) {
	seq := [32]IntSeqData{}
	ev := [32]int{}
	var offset int
	if bd.PartCount == 1 {
		offset = 17
	} else {
		offset = 29
	}

	DecodeIntSeq(data, offset, cemTableA[bd.CemRange], cemTableB[bd.CemRange], bd.EndpointValueCount, false, seq[:])

	switch cemTableA[bd.CemRange] {
	case 3:
		var b int
		for i, c := 0, tritsTable[cemTableB[bd.CemRange]]; i < bd.EndpointValueCount; i++ {
			a := seq[i].Bits & 1 * 0x1ff
			x := seq[i].Bits >> 1

			switch cemTableB[bd.CemRange] {
			case 1:
				b = 0
			case 2:
				b = 278 * x
			case 3:
				b = x<<7 | x<<2 | x
			case 4:
				b = x<<6 | x
			case 5:
				b = x<<5 | x>>2
			case 6:
				b = x<<4 | x>>4
			}
			ev[i] = a&0x80 | ((seq[i].NonBits*c+b)^a)>>2
		}
	case 5:
		var b int
		for i, c := 0, quintsTable[cemTableB[bd.CemRange]]; i < bd.EndpointValueCount; i++ {
			a := seq[i].Bits & 1 * 0x1ff
			x := seq[i].Bits >> 1

			switch cemTableB[bd.CemRange] {
			case 1:
				b = 0
			case 2:
				b = 268 * x
			case 3:
				b = x<<7 | x<<1 | x>>1
			case 4:
				b = x<<6 | x>>1
			case 5:
				b = x<<5 | x>>3
			}

			ev[i] = a&0x80 | ((seq[i].NonBits*c+b)^a)>>2
		}
	default:
		switch cemTableB[bd.CemRange] {
		case 1:
			for i := 0; i < bd.EndpointValueCount; i++ {
				ev[i] = seq[i].Bits * 0xff
			}
		case 2:
			for i := 0; i < bd.EndpointValueCount; i++ {
				ev[i] = seq[i].Bits * 0x55
			}
		case 3:
			for i := 0; i < bd.EndpointValueCount; i++ {
				ev[i] = seq[i].Bits<<5 | seq[i].Bits<<2 | seq[i].Bits>>1
			}
		case 4:
			for i := 0; i < bd.EndpointValueCount; i++ {
				ev[i] = seq[i].Bits<<4 | seq[i].Bits
			}
		case 5:
			for i := 0; i < bd.EndpointValueCount; i++ {
				ev[i] = seq[i].Bits<<3 | seq[i].Bits>>2
			}
		case 6:
			for i := 0; i < bd.EndpointValueCount; i++ {
				ev[i] = seq[i].Bits<<2 | seq[i].Bits>>4
			}
		case 7:
			for i := 0; i < bd.EndpointValueCount; i++ {
				ev[i] = seq[i].Bits<<1 | seq[i].Bits>>6
			}
		case 8:
			for i := 0; i < bd.EndpointValueCount; i++ {
				ev[i] = seq[i].Bits
			}
		}
	}

	v := 0
	for cem := 0; cem < bd.PartCount; v, cem = v+(bd.Cem[cem]/4+1)*2, cem+1 {
		switch bd.Cem[cem] {
		case 0:
			setEndpoint(bd.Endpoints[cem][:], ev[v+0], ev[v+0], ev[v+0], 0xff, ev[v+1], ev[v+1], ev[v+1], 0xff)
		case 1:
			l0 := ev[v+0]>>2 | ev[v+1]&0xc0
			l1 := int(clampU8(l0 + ev[v+1]&0x3f))
			setEndpoint(bd.Endpoints[cem][:], l0, l0, l0, 0xff, l1, l1, l1, 0xff)
		case 2:
			var y0, y1 int
			if ev[v+0] <= ev[v+1] {
				y0 = ev[v+0] << 4
				y1 = ev[v+1] << 4
			} else {
				y0 = ev[v+1]<<4 + 8
				y1 = ev[v+0]<<4 - 8
			}
			setEndpoint(bd.Endpoints[cem][:], y0, y0, y0, 0x780, y1, y1, y1, 0x780) // setEndpointHdr
		case 3:
			var y0, d int
			if ev[v+0]&0x80 != 0 {
				y0 = ev[v+1]&0xe0<<4 | ev[v+0]&0x7f<<2
				d = ev[v+1] & 0x1f << 2
			} else {
				y0 = ev[v+1]&0xf0<<4 | ev[v+0]&0x7f<<1
				d = ev[v+1] & 0x0f << 1
			}
			y1 := int(clampHdr(y0 + d))
			setEndpoint(bd.Endpoints[cem][:], y0, y0, y0, 0x780, y1, y1, y1, 0x780) // setEndpointHdr
		case 4:
			setEndpoint(bd.Endpoints[cem][:], ev[v+0], ev[v+0], ev[v+0], ev[v+2], ev[v+1], ev[v+1], ev[v+1], ev[v+3])
		case 5:
			bitTransferSigned(&ev[v+1], &ev[v+0])
			bitTransferSigned(&ev[v+3], &ev[v+2])
			ev[v+1] += ev[v+0]
			setEndpointClamp(bd.Endpoints[cem][:], ev[v+0], ev[v+0], ev[v+0], ev[v+2], ev[v+1], ev[v+1], ev[v+1], ev[v+2]+ev[v+3])
		case 6:
			setEndpoint(bd.Endpoints[cem][:], ev[v+0]*ev[v+3]>>8, ev[v+1]*ev[v+3]>>8, ev[v+2]*ev[v+3]>>8, 0xff, ev[v+0], ev[v+1], ev[v+2], 0xff)
		case 7:
			decodeEndpointsHdr7(bd.Endpoints[cem][:], ev[v:])
		case 8:
			if ev[v+0]+ev[v+2]+ev[v+4] <= ev[v+1]+ev[v+3]+ev[v+5] {
				setEndpoint(bd.Endpoints[cem][:], ev[v+0], ev[v+2], ev[v+4], 0xff, ev[v+1], ev[v+3], ev[v+5], 0xff)
			} else {
				setEndpointBlue(bd.Endpoints[cem][:], ev[v+1], ev[v+3], ev[v+5], 0xff, ev[v+0], ev[v+2], ev[v+4], 0xff)
			}
		case 9:
			bitTransferSigned(&ev[v+1], &ev[v+0])
			bitTransferSigned(&ev[v+3], &ev[v+2])
			bitTransferSigned(&ev[v+5], &ev[v+4])
			if ev[v+1]+ev[v+3]+ev[v+5] >= 0 {
				setEndpointClamp(bd.Endpoints[cem][:], ev[v+0], ev[v+2], ev[v+4], 0xff, ev[v+0]+ev[v+1], ev[v+2]+ev[v+3], ev[v+4]+ev[v+5], 0xff)
			} else {
				setEndpointBlueClamp(bd.Endpoints[cem][:], ev[v+0]+ev[v+1], ev[v+2]+ev[v+3], ev[v+4]+ev[v+5], 0xff, ev[v+0], ev[v+2], ev[v+4], 0xff)
			}
		case 10:
			setEndpoint(bd.Endpoints[cem][:], ev[v+0]*ev[v+3]>>8, ev[v+1]*ev[v+3]>>8, ev[v+2]*ev[v+3]>>8, ev[v+4], ev[v+0], ev[v+1], ev[v+2], ev[v+5])
		case 11:
			decodeEndpointsHdr11(bd.Endpoints[cem][:], ev[v:], 0x780, 0x780)
		case 12:
			if ev[v+0]+ev[v+2]+ev[v+4] <= ev[v+1]+ev[v+3]+ev[v+5] {
				setEndpoint(bd.Endpoints[cem][:], ev[v+0], ev[v+2], ev[v+4], ev[v+6], ev[v+1], ev[v+3], ev[v+5], ev[v+7])
			} else {
				setEndpointBlue(bd.Endpoints[cem][:], ev[v+1], ev[v+3], ev[v+5], ev[v+7], ev[v+0], ev[v+2], ev[v+4], ev[v+6])
			}
		case 13:
			bitTransferSigned(&ev[v+1], &ev[v+0])
			bitTransferSigned(&ev[v+3], &ev[v+2])
			bitTransferSigned(&ev[v+5], &ev[v+4])
			bitTransferSigned(&ev[v+7], &ev[v+6])
			if ev[v+1]+ev[v+3]+ev[v+5] >= 0 {
				setEndpointClamp(bd.Endpoints[cem][:], ev[v+0], ev[v+2], ev[v+4], ev[v+6], ev[v+0]+ev[v+1], ev[v+2]+ev[v+3], ev[v+4]+ev[v+5], ev[v+6]+ev[v+7])
			} else {
				setEndpointBlueClamp(bd.Endpoints[cem][:], ev[v+0]+ev[v+1], ev[v+2]+ev[v+3], ev[v+4]+ev[v+5], ev[v+6]+ev[v+7], ev[v+0], ev[v+2], ev[v+4], ev[v+6])
			}
		case 14:
			decodeEndpointsHdr11(bd.Endpoints[cem][:], ev[v:], ev[v+6], ev[v+7])
		case 15:
			mode := ev[v+6]>>7&1 | ev[v+7]>>6&2
			ev[v+6] &= 0x7f
			ev[v+7] &= 0x7f
			if mode == 3 {
				decodeEndpointsHdr11(bd.Endpoints[cem][:], ev[v:], ev[v+6]<<5, ev[v+7]<<5)
			} else {
				ev[v+6] |= ev[v+7] << (mode + 1) & 0x780
				ev[v+7] = ev[v+7]&(0x3f>>mode) ^ 0x20>>mode - 0x20>>mode
				ev[v+6] <<= 4 - mode
				ev[v+7] <<= 4 - mode
				decodeEndpointsHdr11(bd.Endpoints[cem][:], ev[v:], ev[v+6], int(clampHdr(ev[v+6]+ev[v+7])))
			}
		}
	}
}

func (bd *BlockData) DecodeWeights(data []byte) {
	seq := [128]IntSeqData{}
	wv := [128]int{}

	DecodeIntSeq(data, 128, weightPrecTableA[bd.WeightRange], weightPrecTableB[bd.WeightRange], bd.WeightCount, true, seq[:])

	if weightPrecTableA[bd.WeightRange] == 0 {
		switch weightPrecTableB[bd.WeightRange] {
		case 1:
			for i := 0; i < bd.WeightCount; i++ {
				if seq[i].Bits != 0 {
					wv[i] = 63
				} else {
					wv[i] = 0
				}
			}
		case 2:
			for i := 0; i < bd.WeightCount; i++ {
				wv[i] = seq[i].Bits<<4 | seq[i].Bits<<2 | seq[i].Bits
			}
		case 3:
			for i := 0; i < bd.WeightCount; i++ {
				wv[i] = seq[i].Bits<<3 | seq[i].Bits
			}
		case 4:
			for i := 0; i < bd.WeightCount; i++ {
				wv[i] = seq[i].Bits<<2 | seq[i].Bits>>2
			}
		case 5:
			for i := 0; i < bd.WeightCount; i++ {
				wv[i] = seq[i].Bits<<1 | seq[i].Bits>>4
			}
		}

		for i := 0; i < bd.WeightCount; i++ {
			if wv[i] > 32 {
				wv[i]++
			}
		}
	} else if weightPrecTableB[bd.WeightRange] == 0 {
		var s int
		if weightPrecTableA[bd.WeightRange] == 3 {
			s = 32
		} else {
			s = 16
		}

		for i := 0; i < bd.WeightCount; i++ {
			wv[i] = seq[i].NonBits * s
		}
	} else {
		if weightPrecTableA[bd.WeightRange] == 3 {
			switch weightPrecTableB[bd.WeightRange] {
			case 1:
				for i := 0; i < bd.WeightCount; i++ {
					wv[i] = seq[i].NonBits * 50
				}
			case 2:
				for i := 0; i < bd.WeightCount; i++ {
					wv[i] = seq[i].NonBits * 23
					if seq[i].Bits&2 != 0 {
						wv[i] += 69
					}
				}
			case 3:
				for i := 0; i < bd.WeightCount; i++ {
					wv[i] = seq[i].NonBits*11 + (seq[i].Bits<<4|seq[i].Bits>>1)&0b1100011
				}
			}
		} else if weightPrecTableA[bd.WeightRange] == 5 {
			switch weightPrecTableB[bd.WeightRange] {
			case 1:
				for i := 0; i < bd.WeightCount; i++ {
					wv[i] = seq[i].NonBits * 28
				}
			case 2:
				for i := 0; i < bd.WeightCount; i++ {
					wv[i] = seq[i].NonBits * 13
					if seq[i].Bits&2 != 0 {
						wv[i] += 66
					}
				}
			}
		}

		for i := 0; i < bd.WeightCount; i++ {
			a := seq[i].Bits & 1 * 0x7f
			wv[i] = a&0x20 | (wv[i]^a)>>2
			if wv[i] > 32 {
				wv[i]++
			}
		}
	}

	ds := (1024 + bd.BlockWidth/2) / (bd.BlockWidth - 1)
	dt := (1024 + bd.BlockHeight/2) / (bd.BlockHeight - 1)
	var pn int
	if bd.DualPlane {
		pn = 2
	} else {
		pn = 1
	}

	for t, i := 0, 0; t < bd.BlockHeight; t++ {
		for s := 0; s < bd.BlockWidth; s, i = s+1, i+1 {
			gs := (ds*s*(bd.Width-1) + 32) >> 6
			gt := (dt*t*(bd.Height-1) + 32) >> 6
			fs := gs & 0xf
			ft := gt & 0xf
			v := gs>>4 + gt>>4*bd.Width
			w11 := (fs*ft + 8) >> 4
			w10 := ft - w11
			w01 := fs - w11
			w00 := 16 - fs - ft + w11

			for p := 0; p < pn; p++ {
				p00 := wv[v*pn+p]
				p01 := wv[(v+1)*pn+p]
				p10 := wv[(v+bd.Width)*pn+p]
				p11 := wv[(v+bd.Width+1)*pn+p]
				bd.Weights[i][p] = (p00*w00 + p01*w01 + p10*w10 + p11*w11 + 8) >> 4
			}
		}
	}
}

func (bd *BlockData) SelectPartition(data []byte) {
	smallBlock := bd.BlockWidth*bd.BlockHeight < 31
	seed := int(binary.LittleEndian.Uint32(data))>>13&0x3ff | (bd.PartCount-1)<<10

	rnum := uint32(seed)
	rnum ^= rnum >> 15
	rnum -= rnum << 17
	rnum += rnum << 7
	rnum += rnum << 4
	rnum ^= rnum >> 5
	rnum += rnum << 16
	rnum ^= rnum >> 7
	rnum ^= rnum >> 3
	rnum ^= rnum << 6
	rnum ^= rnum >> 17

	seeds := [8]int{}
	for i := 0; i < 8; i++ {
		seeds[i] = int(rnum >> (i * 4) & 0xf)
		seeds[i] *= seeds[i]
	}

	sh := [2]int{}
	if seed&2 != 0 {
		sh[0] = 4
	} else {
		sh[0] = 5
	}

	if bd.PartCount == 3 {
		sh[1] = 6
	} else {
		sh[1] = 5
	}

	if seed&1 != 0 {
		for i := 0; i < 8; i++ {
			seeds[i] >>= sh[i%2]
		}
	} else {
		for i := 0; i < 8; i++ {
			seeds[i] >>= sh[1-i%2]
		}
	}

	if smallBlock {
		for t, i := 0, 0; t < bd.BlockHeight; t++ {
			for s := 0; s < bd.BlockWidth; s, i = s+1, i+1 {
				x := s << 1
				y := t << 1
				a := (seeds[0]*x + seeds[1]*y + int(rnum>>14)) & 0x3f
				b := (seeds[2]*x + seeds[3]*y + int(rnum>>10)) & 0x3f
				c, d := 0, 0
				if bd.PartCount >= 3 {
					c = (seeds[4]*x + seeds[5]*y + int(rnum>>6)) & 0x3f
				}
				if bd.PartCount >= 4 {
					d = (seeds[6]*x + seeds[7]*y + int(rnum>>2)) & 0x3f
				}

				if a >= b && a >= c && a >= d {
					bd.Partition[i] = 0
				} else if b >= c && b >= d {
					bd.Partition[i] = 1
				} else if c >= d {
					bd.Partition[i] = 2
				} else {
					bd.Partition[i] = 3
				}
			}
		}
	} else {
		for y, i := 0, 0; y < bd.BlockHeight; y++ {
			for x := 0; x < bd.BlockWidth; x, i = x+1, i+1 {
				a := (seeds[0]*x + seeds[1]*y + int(rnum>>14)) & 0x3f
				b := (seeds[2]*x + seeds[3]*y + int(rnum>>10)) & 0x3f
				c, d := 0, 0
				if bd.PartCount >= 3 {
					c = (seeds[4]*x + seeds[5]*y + int(rnum>>6)) & 0x3f
				}
				if bd.PartCount >= 4 {
					d = (seeds[6]*x + seeds[7]*y + int(rnum>>2)) & 0x3f
				}

				if a >= b && a >= c && a >= d {
					bd.Partition[i] = 0
				} else if b >= c && b >= d {
					bd.Partition[i] = 1
				} else if c >= d {
					bd.Partition[i] = 2
				} else {
					bd.Partition[i] = 3
				}
			}
		}
	}
}

func (bd *BlockData) ApplicateColor(out []byte) {
	if bd.DualPlane {
		ps := []int{0, 0, 0, 0}
		ps[bd.PlaneSelector] = 1
		if bd.PartCount > 1 {
			for i := 0; i < bd.BlockWidth*bd.BlockHeight; i++ {
				p := bd.Partition[i]
				r := funcTableC[bd.Cem[p]](bd.Endpoints[p][0], bd.Endpoints[p][4], bd.Weights[i][ps[0]])
				g := funcTableC[bd.Cem[p]](bd.Endpoints[p][1], bd.Endpoints[p][5], bd.Weights[i][ps[1]])
				b := funcTableC[bd.Cem[p]](bd.Endpoints[p][2], bd.Endpoints[p][6], bd.Weights[i][ps[2]])
				a := funcTableA[bd.Cem[p]](bd.Endpoints[p][3], bd.Endpoints[p][7], bd.Weights[i][ps[3]])
				out[i*4+0] = r
				out[i*4+1] = g
				out[i*4+2] = b
				out[i*4+3] = a
			}
		} else {
			for i := 0; i < bd.BlockWidth*bd.BlockHeight; i++ {
				r := funcTableC[bd.Cem[0]](bd.Endpoints[0][0], bd.Endpoints[0][4], bd.Weights[i][ps[0]])
				g := funcTableC[bd.Cem[0]](bd.Endpoints[0][1], bd.Endpoints[0][5], bd.Weights[i][ps[1]])
				b := funcTableC[bd.Cem[0]](bd.Endpoints[0][2], bd.Endpoints[0][6], bd.Weights[i][ps[2]])
				a := funcTableA[bd.Cem[0]](bd.Endpoints[0][3], bd.Endpoints[0][7], bd.Weights[i][ps[3]])
				out[i*4+0] = r
				out[i*4+1] = g
				out[i*4+2] = b
				out[i*4+3] = a
			}
		}
	} else if bd.PartCount > 1 {
		for i := 0; i < bd.BlockWidth*bd.BlockHeight; i++ {
			p := bd.Partition[i]
			r := funcTableC[bd.Cem[p]](bd.Endpoints[p][0], bd.Endpoints[p][4], bd.Weights[i][0])
			g := funcTableC[bd.Cem[p]](bd.Endpoints[p][1], bd.Endpoints[p][5], bd.Weights[i][0])
			b := funcTableC[bd.Cem[p]](bd.Endpoints[p][2], bd.Endpoints[p][6], bd.Weights[i][0])
			a := funcTableA[bd.Cem[p]](bd.Endpoints[p][3], bd.Endpoints[p][7], bd.Weights[i][0])
			out[i*4+0] = r
			out[i*4+1] = g
			out[i*4+2] = b
			out[i*4+3] = a
		}
	} else {
		for i := 0; i < bd.BlockWidth*bd.BlockHeight; i++ {
			r := funcTableC[bd.Cem[0]](bd.Endpoints[0][0], bd.Endpoints[0][4], bd.Weights[i][0])
			g := funcTableC[bd.Cem[0]](bd.Endpoints[0][1], bd.Endpoints[0][5], bd.Weights[i][0])
			b := funcTableC[bd.Cem[0]](bd.Endpoints[0][2], bd.Endpoints[0][6], bd.Weights[i][0])
			a := funcTableA[bd.Cem[0]](bd.Endpoints[0][3], bd.Endpoints[0][7], bd.Weights[i][0])
			out[i*4+0] = r
			out[i*4+1] = g
			out[i*4+2] = b
			out[i*4+3] = a
		}
	}
}
