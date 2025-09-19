// Copyright (C) 2024, 2025 kvarenzn
// SPDX-License-Identifier: GPL-3.0-or-later

package ocr

import "github.com/kvarenzn/ssm/ort"

type TextDetector struct {
	session *ort.Session
	memInfo *ort.MemoryInfo
}

func NewTextDetector() (*TextDetector, error) {
	session, err := ortEnv.NewSession("./models/det.onnx", nil)
	if err != nil {
		return nil, err
	}

	memInfo, err := ort.NewMemoryInfo(ort.DeviceCPU, ort.ArenaAllocator, 0, ort.MemTypeDefault)
	if err != nil {
		return nil, err
	}

	return &TextDetector{
		session: session,
		memInfo: memInfo,
	}, nil
}
