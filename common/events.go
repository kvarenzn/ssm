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
	PointerID int
	Action    TouchAction
	X         float64
	Y         float64
}

type VirtualEventsItem struct {
	Timestamp int64
	Events    []VirtualTouchEvent
}

type RawVirtualEvents []VirtualEventsItem

type ViscousEventItem struct {
	Timestamp int64
	Data      []byte
}
