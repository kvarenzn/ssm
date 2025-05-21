package controllers

import (
	"encoding/binary"
	"fmt"
	"math/rand"
	"net"
	"os"

	"github.com/kvarenzn/ssm/adb"
	"github.com/kvarenzn/ssm/decoders/av"
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

const testFromPort = 27188

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

func (c *ScrcpyController) Open(filepath string) error {
	listener, port := tryListen("localhost", 27188)
	c.listener = listener
	localName := fmt.Sprintf("localabstract:scrcpy_%s", c.sessionID)
	err := c.device.Forward(localName, fmt.Sprintf("tcp:%d", port), true, false)
	if err != nil {
		return err
	}

	f, err := os.Open(filepath)
	if err != nil {
		return err
	}

	if err := c.device.Push(f, "/data/local/tmp/scrcpy-server.jar"); err != nil {
		return err
	}

	go func() {
		c.device.Sh(
			"CLASSPATH=/data/local/tmp/scrcpy-server.jar",
			"app_process",
			"/",
			"com.genymobile.scrcpy.Server",
			"3.1",                      // version
			"scid=11451419",            // session id
			"log_level=info",           // log level
			"audio=false",              // disable audio sync
			"clipboard_autosync=false", // disable clipboard
		)
	}()

	videoSocket, err := listener.Accept()
	if err != nil {
		return err
	}
	c.videoSocket = videoSocket

	controlSocket, err := listener.Accept()
	if err != nil {
		return err
	}
	c.controlSocket = controlSocket

	err = c.device.Client().KillForward(localName, true)
	if err != nil {
		return err
	}

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
