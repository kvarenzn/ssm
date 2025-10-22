// Copyright (C) 2024, 2025 kvarenzn
// SPDX-License-Identifier: GPL-3.0-or-later

package controllers

import (
	"encoding/binary"
	"fmt"
	"math"
	"math/rand"
	"net"
	"os"

	"github.com/kvarenzn/ssm/adb"
	"github.com/kvarenzn/ssm/common"
	"github.com/kvarenzn/ssm/config"
	"github.com/kvarenzn/ssm/decoders/av"
	"github.com/kvarenzn/ssm/log"
	"github.com/kvarenzn/ssm/stage"
)

type ScrcpyController struct {
	device    *adb.Device
	sessionID string

	listener      net.Listener
	videoSocket   net.Conn
	controlSocket net.Conn

	width    int
	height   int
	codecID  string
	decoder  *av.AVDecoder
	cRunning bool
	vRunning bool
}

func NewScrcpyController(device *adb.Device) *ScrcpyController {
	return &ScrcpyController{
		device:    device,
		sessionID: fmt.Sprintf("%08x", rand.Int31()),
	}
}

func tryListen(host string, port int) (net.Listener, int) {
	for {
		addr := fmt.Sprintf("%s:%d", host, port)
		listen, err := net.Listen("tcp", addr)
		if err == nil {
			return listen, port
		}

		port++
	}
}

const testFromPort = 27188

func (c *ScrcpyController) Open(filepath string, version string) error {
	listener, port := tryListen("localhost", testFromPort)
	c.listener = listener
	log.Debugf("Listening at localhost:%d", port)

	localName := fmt.Sprintf("localabstract:scrcpy_%s", c.sessionID)
	err := c.device.Forward(localName, fmt.Sprintf("tcp:%d", port), true, false)
	if err != nil {
		return err
	}
	log.Debugf("ADB reverse socket `%s` created.", localName)

	f, err := os.Open(filepath)
	if err != nil {
		return err
	}

	log.Debugln("`scrcpy-server` loaded.")

	if err := c.device.Push(f, "/data/local/tmp/scrcpy-server.jar"); err != nil {
		return err
	}

	log.Debugln("`scrcpy-server` pushed to gaming device.")

	go func() {
		result, err := c.device.Sh(
			"CLASSPATH=/data/local/tmp/scrcpy-server.jar",
			"app_process",
			"/",
			"com.genymobile.scrcpy.Server",
			version,
			fmt.Sprintf("scid=%s", c.sessionID), // session id
			"log_level=info",                    // log level
			"audio=false",                       // disable audio sync
			"clipboard_autosync=false",          // disable clipboard
		)
		if err != nil {
			log.Fatalln("Failed to start `scrcpy-server`:", err)
		}

		log.Debugln(result)
	}()

	videoSocket, err := listener.Accept()
	if err != nil {
		return err
	}
	c.videoSocket = videoSocket

	log.Debugln("Video socket accepted.")

	controlSocket, err := listener.Accept()
	if err != nil {
		return err
	}
	c.controlSocket = controlSocket

	log.Debugln("Control socket accepted.")

	err = c.device.Client().KillForward(localName, true)
	if err != nil {
		return err
	}

	log.Debugf("ADB reverse socket `%s` removed.", localName)

	deviceName := make([]byte, 64)
	videoSocket.Read(deviceName)

	buf := make([]byte, 4)
	videoSocket.Read(buf)
	c.codecID = string(buf)

	c.decoder, err = av.NewAVDecoder(c.codecID)
	if err != nil {
		return err
	}

	videoSocket.Read(buf)
	c.width = int(binary.BigEndian.Uint32(buf))

	videoSocket.Read(buf)
	c.height = int(binary.BigEndian.Uint32(buf))

	c.cRunning = true
	c.vRunning = true

	go func() {
		msgTypeBuf := make([]byte, 1)
		sizeBuf := make([]byte, 4)
		for c.cRunning {
			if n, err := controlSocket.Read(msgTypeBuf); err != nil || n != 1 {
				break
			}

			if n, err := controlSocket.Read(sizeBuf); err != nil || n != 4 {
				break
			}

			size := binary.BigEndian.Uint32(sizeBuf)
			bodyBuf := make([]byte, size)
			if n, err := controlSocket.Read(bodyBuf); err != nil || n != int(size) {
				break
			}
		}

		c.cRunning = false
	}()

	go func() {
		ptsBuf := make([]byte, 8)
		sizeBuf := make([]byte, 4)
		for c.vRunning {
			if n, err := videoSocket.Read(ptsBuf); err != nil || n != 8 {
				break
			}

			pts := binary.BigEndian.Uint64(ptsBuf)

			if n, err := videoSocket.Read(sizeBuf); err != nil || n != 4 {
				break
			}

			size := binary.BigEndian.Uint32(sizeBuf)

			data := make([]byte, size)

			if n, err := videoSocket.Read(data); err != nil || n != int(size) {
				break
			}

			c.decoder.Decode(pts, data)
		}

		c.vRunning = false
	}()

	return nil
}

func (c *ScrcpyController) Encode(action common.TouchAction, x, y int32, pointerID uint64) []byte {
	data := make([]byte, 32)
	data[0] = 2 // type: SC_CONTROL_MSG_TYPE_INJECT_TOUCH_EVENT
	data[1] = byte(action)
	binary.BigEndian.PutUint64(data[2:], pointerID)
	binary.BigEndian.PutUint32(data[10:], uint32(x))
	binary.BigEndian.PutUint32(data[14:], uint32(y))
	binary.BigEndian.PutUint16(data[18:], uint16(c.width))
	binary.BigEndian.PutUint16(data[20:], uint16(c.height))
	binary.BigEndian.PutUint16(data[22:], 0xffff)
	binary.BigEndian.PutUint32(data[24:], 1) // AMOTION_EVENT_BUTTON_PRIMARY
	binary.BigEndian.PutUint32(data[28:], 1) // AMOTION_EVENT_BUTTON_PRIMARY
	return data
}

func (c *ScrcpyController) touch(action common.TouchAction, x, y int32, pointerID uint64) {
	c.Send(c.Encode(action, x, y, pointerID))
}

func (c *ScrcpyController) Down(pointerID uint64, x, y int) {
	c.touch(common.TouchDown, int32(x), int32(y), pointerID)
}

func (c *ScrcpyController) Move(pointerID uint64, x, y int) {
	c.touch(common.TouchMove, int32(x), int32(y), pointerID)
}

func (c *ScrcpyController) Up(pointerID uint64, x, y int) {
	c.touch(common.TouchUp, int32(x), int32(y), pointerID)
}

func (c *ScrcpyController) Close() error {
	c.cRunning = false
	c.vRunning = false

	if err := c.videoSocket.Close(); err != nil {
		return err
	}

	if err := c.controlSocket.Close(); err != nil {
		return err
	}

	return c.listener.Close()
}

func (c *ScrcpyController) Preprocess(rawEvents common.RawVirtualEvents, turnRight bool, dc *config.DeviceConfig) []common.ViscousEventItem {
	s := stage.NewScreen(dc.Height, dc.Width)
	x1, x2 := s.X()
	yy := s.Y()
	mapper := func(x, y float64) (int, int) {
		return int(math.Round(x1 + (x2-x1)*x)), int(math.Round(yy - (yy-s.Height/2)*y))
	}

	result := []common.ViscousEventItem{}
	currentFingers := make([]bool, 10)
	for _, events := range rawEvents {
		var data []byte
		for _, event := range events.Events {
			x, y := mapper(event.X, event.Y)
			switch event.Action {
			case common.TouchDown:
				if currentFingers[event.PointerID] {
					log.Fatalf("pointer `%d` is already on screen", event.PointerID)
				}
				currentFingers[event.PointerID] = true
			case common.TouchMove:
				if !currentFingers[event.PointerID] {
					log.Fatalf("pointer `%d` is not on screen", event.PointerID)
				}
			case common.TouchUp:
				if !currentFingers[event.PointerID] {
					log.Fatalf("pointer `%d` is not on screen", event.PointerID)
				}
				currentFingers[event.PointerID] = false
			default:
				log.Fatalf("unknown touch action: %d\n", event.Action)
			}

			data = append(data, c.Encode(event.Action, int32(x), int32(y), uint64(event.PointerID))...)
		}

		result = append(result, common.ViscousEventItem{
			Timestamp: events.Timestamp,
			Data:      data,
		})
	}

	return result
}

func (c *ScrcpyController) Send(data []byte) {
	n, err := c.controlSocket.Write(data)
	if err != nil {
		log.Fatalln("Failed to send control data through control socket:", err)
	}

	if n != len(data) {
		log.Fatalf("Failed to send control data through control socket: expect to send %d bytes, but %d bytes were sent", len(data), n)
	}
}
