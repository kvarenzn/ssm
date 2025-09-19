// Copyright (C) 2024, 2025 kvarenzn
// SPDX-License-Identifier: GPL-3.0-or-later

package ocr

import (
	"log"

	"github.com/kvarenzn/ssm/ort"
)

var ortEnv *ort.Env

func init() {
	var err error
	ortEnv, err = ort.NewEnv(ort.LoggingLevelFatal, "ssm")
	if err != nil {
		log.Fatalf("failed to create env: %s", err)
	}
}
