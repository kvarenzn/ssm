// Copyright (C) 2024, 2025, 2026 kvarenzn
// SPDX-License-Identifier: GPL-3.0-or-later

package utils

import (
	"fmt"
)

func FormatSeconds(seconds float64) string {
	m := int(seconds / 60)
	s := int(seconds) % 60
	ms := int(seconds*1000) % 1000
	return fmt.Sprintf("%d:%02d.%03d", m, s, ms)
}
