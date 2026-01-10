// Copyright (C) 2024, 2025 kvarenzn
// SPDX-License-Identifier: GPL-3.0-or-later

package common

type TouchAction uint8

const (
	TouchDown TouchAction = iota
	TouchUp
	TouchMove
	TouchCancel
	TouchOutside
	TouchPointerDown
	TouchPointerUp
	TouchHoverMove
)

type VirtualTouchEvent struct {
	PointerID int         `json:"pointerId"`
	Action    TouchAction `json:"action"`
	X         float64     `json:"x"`
	Y         float64     `json:"y"`
}

type VirtualEventsItem struct {
	Timestamp int64                `json:"timestamp"`
	Events    []*VirtualTouchEvent `json:"events"`
}

type RawVirtualEvents []*VirtualEventsItem

type ViscousEventItem struct {
	Timestamp int64
	Data      []byte
}
