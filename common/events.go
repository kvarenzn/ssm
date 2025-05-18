package common

type TouchAction byte

const (
	TouchDown TouchAction = iota
	TouchMove
	TouchUp
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
