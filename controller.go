package main

import (
	"bytes"
	"encoding/binary"

	"github.com/google/gousb"
)

var _REPORT_DESC_HEAD = []byte{
	0x05, 0x0d, // Usage Page (Digitalizers)
	0x09, 0x04, // Usage (Touch Screen)
	0xa1, 0x01, // Collection (Application)
	0x15, 0x00, //		Logical Mininum (0)
}

var _REPORT_DESC_BODY_PART1 = []byte{
	0x09, 0x22, //		Usage (Fingers)
	0xa1, 0x02, //		Collection (Logical)
	0x09, 0x51, //			Usage (Contact Identifier)
	0x75, 0x04, //			Report Size (4)
	0x95, 0x01, //			Report Count (1)
	0x25, 0x09, //			Logical Maximum (9)
	0x81, 0x02, //			Input (Data, Variable, Absolute)
	0x09, 0x42, //			Usage (Tip Switch)
	0x25, 0x01, //			Logical Maximum (1)
	0x75, 0x01, //			Report Size (1)
	0x81, 0x02, //			Input (Data, Variable, Absolute)
	0x09, 0x32, //			Usage (In Range)
	0x25, 0x01, //			Logical Maximum (1)
	0x81, 0x02, //			Input (Data, Variable, Absolute)
	0x75, 0x02, //			Report Size (2)
	0x81, 0x01, //			Input (Constant)
	0x05, 0x01, //			Usage Page (Generic Desktop Page)
	0x09, 0x30, //			Usage (X)
	0x26, //				Logical Maximum (Currently Unknown)
}

var _REPORT_DESC_BODY_PART2 = []byte{
	0x75, 0x10, //			Report Size (16)
	0x81, 0x02, //			Input (Data, Variable, Absolute)
	0x09, 0x31, //			Usage (Y)
	0x26, //				Logical Maximum (Currently Unknown)
}

var _REPORT_DESC_BODY_PART3 = []byte{
	0x81, 0x02, //			Input (Data, Variable, Absolute)
	0x05, 0x0d, //			Usage Page (Digitalizers)
	0xc0, //			End Collection
}

var _REPORT_DESC_TAIL = []byte{
	0xc0, //		End Collection
}

func fingerEvent(id int, onScreen bool, x, y int) []byte {
	result := make([]byte, 5)
	result[0] = byte(id & 0b1111)
	if onScreen {
		result[0] |= 0b110000
	}

	binary.LittleEndian.PutUint16(result[1:3], uint16(x))
	binary.LittleEndian.PutUint16(result[3:5], uint16(y))

	return result
}

type PointerStatus struct {
	X        int
	Y        int
	OnScreen bool
}

func genEventData(pointers [10]PointerStatus) []byte {
	result := bytes.NewBuffer([]byte{})
	for i, s := range pointers {
		result.Write(fingerEvent(i, s.OnScreen, s.X, s.Y))
	}
	return result.Bytes()
}

type HIDController struct {
	AccessoryID       int
	Serial            string
	DeviceWidth       int
	DeviceHeight      int
	device            *gousb.Device
	reportDescription []byte
	usbContext        *gousb.Context
}

func NewHIDController(width, height int, serial string) *HIDController {
	usbContext := gousb.NewContext()
	// defer usbContext.Close()

	devs, _ := usbContext.OpenDevices(func(desc *gousb.DeviceDesc) bool {
		if desc.Class != gousb.ClassPerInterface || desc.SubClass != gousb.ClassPerInterface {
			return false
		}
		return true
	})

	var device *gousb.Device = nil

	for _, dev := range devs {
		s, err := dev.SerialNumber()
		if err != nil || s != serial {
			dev.Close()
			continue
		}

		if device == nil {
			device = dev
		} else {
			dev.Close()
		}
	}

	uint16Buffer := make([]byte, 2)

	reportDescBody := bytes.NewBuffer(nil)
	reportDescBody.Write(_REPORT_DESC_BODY_PART1)
	binary.LittleEndian.PutUint16(uint16Buffer, uint16(width))
	reportDescBody.Write(uint16Buffer)
	reportDescBody.Write(_REPORT_DESC_BODY_PART2)
	binary.LittleEndian.PutUint16(uint16Buffer, uint16(height))
	reportDescBody.Write(uint16Buffer)
	reportDescBody.Write(_REPORT_DESC_BODY_PART3)

	reportDescription := bytes.NewBuffer(nil)
	reportDescription.Write(_REPORT_DESC_HEAD)
	for range 10 {
		reportDescription.Write(reportDescBody.Bytes())
	}
	reportDescription.Write(_REPORT_DESC_TAIL)

	return &HIDController{
		AccessoryID:       114514,
		Serial:            serial,
		DeviceWidth:       width,
		DeviceHeight:      height,
		device:            device,
		reportDescription: reportDescription.Bytes(),
		usbContext:        usbContext,
	}
}

func (c *HIDController) registerHID() {
	_, err := c.device.Control(
		64, // ENDPOINT_OUT | REQUEST_TYPE_VENDOR
		54, // ACCESSORY_REGISTER_HID
		uint16(c.AccessoryID),
		uint16(len(c.reportDescription)),
		nil,
	)
	if err != nil {
		Fatal(err)
	}
}

func (c *HIDController) unregisterHID() {
	_, err := c.device.Control(
		64, // ENDPOINT_OUT | REQUEST_TYPE_VENDOR
		55, // ACCESSORY_UNREGISTER_ID
		uint16(c.AccessoryID),
		0,
		nil,
	)
	if err != nil {
		Fatal(err)
	}
}

func (c *HIDController) setHIDReportDescription() {
	_, err := c.device.Control(
		64, // ENDPOINT_OUT | REQUEST_TYPE_VENDOR
		56, // ACCESSORY_SET_HID_REPORT_DESC
		uint16(c.AccessoryID),
		0,
		c.reportDescription,
	)
	if err != nil {
		Fatal(err)
	}
}

func (c *HIDController) sendHIDEvent(event []byte) {
	_, err := c.device.Control(
		64, // ENDPOINT_OUT | REQUEST_TYPE_VENDOR
		57, // ACCESSORY_SEND_HID_EVENT
		uint16(c.AccessoryID),
		0,
		event,
	)
	if err != nil {
		Fatal(err)
	}
}

func (c *HIDController) Open() {
	c.registerHID()
	c.setHIDReportDescription()
}

func (c *HIDController) Send(event []byte) {
	c.sendHIDEvent(event)
}

func (c *HIDController) Close() {
	c.unregisterHID()
	c.device.Close()
	c.usbContext.Close()
}

func FindDevices() []string {
	result := []string{}

	ctx := gousb.NewContext()
	defer ctx.Close()

	devs, _ := ctx.OpenDevices(func(desc *gousb.DeviceDesc) bool {
		if desc.Class != gousb.ClassPerInterface || desc.SubClass != gousb.ClassPerInterface {
			return false
		}

		return true
	})

	for _, dev := range devs {
		serial, err := dev.SerialNumber()
		if err == nil {
			result = append(result, serial)
		}

		err = dev.Close()
		if err != nil {
			Fatal(err)
		}
	}

	return result
}

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

type ViscousEventItem struct {
	Timestamp int64
	Data      []byte
}

type VirtualEventsItem struct {
	Timestamp int64
	Events    []VirtualTouchEvent
}

type (
	CoordMapper      func(x, y float64) (int, int)
	RawVirtualEvents []VirtualEventsItem
)

func preprocess(mapper CoordMapper, rawEvents RawVirtualEvents) []ViscousEventItem {
	result := []ViscousEventItem{}
	currentFingers := [10]PointerStatus{}
	for _, events := range rawEvents {
		for _, event := range events.Events {
			x, y := mapper(event.X, event.Y)
			status := currentFingers[event.PointerID]
			switch event.Action {
			case TouchDown:
				if status.OnScreen {
					Fatalf("pointer id: %d is already on screen", event.PointerID)
				}
				status.X = x
				status.Y = y
				status.OnScreen = true
				currentFingers[event.PointerID] = status
				break
			case TouchMove:
				if !status.OnScreen {
					Fatalf("pointer id: %d is not on screen", event.PointerID)
				}
				status.X = x
				status.Y = y
				status.OnScreen = true
				currentFingers[event.PointerID] = status
				break
			case TouchUp:
				if !status.OnScreen {
					Fatalf("pointer id: %d is not on screen", event.PointerID)
				}
				status.X = x
				status.Y = y
				status.OnScreen = false
				currentFingers[event.PointerID] = status
				break
			default:
				Fatalf("unknown touch action: %d\n", event.Action)
			}
		}
		result = append(result, ViscousEventItem{
			Timestamp: events.Timestamp,
			Data:      genEventData(currentFingers),
		})
	}
	return result
}
