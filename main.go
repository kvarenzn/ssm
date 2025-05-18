package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/kvarenzn/ssm/adb"
	"github.com/kvarenzn/ssm/config"
	"github.com/kvarenzn/ssm/controllers"
	"github.com/kvarenzn/ssm/log"
)

type Point struct {
	X int
	Y int
}

var SSM_VERSION = "(unknown)"

var (
	songID       = flag.Int("n", -1, "Song ID")
	difficulty   = flag.String("d", "", "Difficulty of song")
	extract      = flag.String("e", "", "Extract assets from assets folder <path>")
	direction    = flag.String("r", "left", "Direction of device, possible values: left (turn left), right (turn right)")
	chartPath    = flag.String("p", "", "Custom chart path (if this is provided, songID and difficulty will be ignored)")
	deviceSerial = flag.String("s", "", "Specify the device serial (if not provided, ssm will use the first device serial)")
	showVersion  = flag.Bool("v", false, "Show ssm's version number and exit")
)

func TryListen(host string, port int) (net.Listener, int) {
	for {
		addr := fmt.Sprintf("%s:%d", host, port)
		listen, err := net.Listen("tcp", addr)
		if err == nil {
			return listen, port
		}

		port++
	}
}

func TestAdb() {
	if err := adb.StartADBServer("localhost", 5037); err != nil && err != adb.ErrADBServerRunning {
		log.Fatal(err)
	}

	client := adb.NewDefaultClient()
	devices, err := client.Devices()
	if err != nil {
		log.Fatal(err)
	}

	if len(devices) == 0 {
		log.Fatal("no android devices recognized")
	}

	log.Info("adb devices:", devices)

	device := adb.FirstAuthorizedDevice(devices)
	if device == nil {
		log.Fatal("no authorized device")
	}

	log.Info("selected device:", device)

	localName := "localabstract:scrcpy_11451419"

	// open reverse socket
	listener, port := TryListen("localhost", 27188)
	err = device.Forward(
		localName,
		fmt.Sprintf("tcp:%d", port),
		true,
		false)
	if err != nil {
		log.Fatal(err)
	}

	// try upload a file
	f, err := os.Open("./scrcpy-server-v3.1")
	if err != nil {
		log.Fatal(err)
	}

	if err := device.Push(f, "/data/local/tmp/scrcpy-server.jar"); err != nil {
		log.Fatal(err)
	}

	log.Info("scrcpy server uploaded")

	go func() {
		result, err := device.Sh(
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
		if err != nil {
			log.Fatal(err)
		}
		log.Info(result)
	}()

	videoSocket, err := listener.Accept()
	if err != nil {
		log.Fatal(err)
	}

	controlSocket, err := listener.Accept()
	if err != nil {
		log.Fatal(err)
	}

	err = client.KillForward(localName, true)
	if err != nil {
		log.Fatal(err)
	}

	deviceName := make([]byte, 64)
	videoSocket.Read(deviceName)
	log.Info(string(deviceName))

	buf := make([]byte, 4)
	videoSocket.Read(buf)
	codecID := string(buf)
	log.Info("codecID:", codecID)
	videoSocket.Read(buf)
	width := binary.BigEndian.Uint32(buf)
	videoSocket.Read(buf)
	height := binary.BigEndian.Uint32(buf)
	log.Info("width:", width)
	log.Info("height:", height)

	go func() {
		msgTypeBuf := make([]byte, 1)
		sizeBuf := make([]byte, 4)
		for {
			controlSocket.Read(msgTypeBuf)
			controlSocket.Read(sizeBuf)
			size := binary.BigEndian.Uint32(sizeBuf)
			bodyBuf := make([]byte, size)
			controlSocket.Read(bodyBuf)
		}
	}()

	go func() {
		ptsBuf := make([]byte, 8)
		sizeBuf := make([]byte, 4)
		for {
			videoSocket.Read(ptsBuf)
			videoSocket.Read(sizeBuf)
		}
	}()

	videoSocket.Close()
	controlSocket.Close()
	listener.Close()
	os.Exit(0)
}

func main() {
	TestAdb()

	flag.Parse()

	if *showVersion {
		fmt.Printf("ssm version: %s\n", SSM_VERSION)
		return
	}

	const CONFIG_PATH = "./config.json"

	config, err := config.Load(CONFIG_PATH)
	if err != nil {
		log.Fatal(err)
	}

	if *extract != "" {
		Extract(*extract)
		return
	}

	if len(*chartPath) == 0 && (*songID == -1 || *difficulty == "") {
		log.Fatal("Song id and difficulty are both required")
	}

	if *deviceSerial == "" {
		serials := controllers.FindDevices()
		log.Info("Recognized devices:", serials)

		if len(serials) == 0 {
			log.Fatal("plug your gaming device to pc")
		}

		*deviceSerial = serials[0]
	}

	dc := config.Get(*deviceSerial)
	controller := controllers.NewHIDController(dc)
	controller.Open()
	defer controller.Close()

	var text []byte
	if len(*chartPath) == 0 {
		baseFolder := "./assets/star/forassetbundle/startapp/musicscore/"
		pathResults, err := filepath.Glob(filepath.Join(baseFolder, fmt.Sprintf("musicscore*/%03d/*_%s.txt", *songID, *difficulty)))
		if err != nil {
			log.Fatal(err)
		}

		if len(pathResults) < 1 {
			log.Fatal("not found")
		}

		log.Info("Music score loaded:", pathResults[0])
		text, err = os.ReadFile(pathResults[0])
	} else {
		log.Info("Music score loaded:", *chartPath)
		text, err = os.ReadFile(*chartPath)
	}

	if err != nil {
		log.Fatal(err)
	}

	chart := Parse(string(text))
	rawEvents := GenerateTouchEvent(VTEGenerateConfig{
		TapDuration:         10,
		FlickDuration:       60,
		FlickReportInterval: 5,
		SlideReportInterval: 10,
	}, chart)

	vEvents := controller.Preprocess(rawEvents, *direction == "right")

	log.Info("Ready. Press ENTER to start autoplay.")
	fmt.Scanln()

	firstEvent := vEvents[0]
	tick := firstEvent.Timestamp
	start := time.Now().Add(-time.Duration(tick) * time.Millisecond)

	current := 0
	for current < len(vEvents) {
		delta := time.Now().Sub(start).Milliseconds()
		events := vEvents[current]
		if delta >= events.Timestamp {
			controller.Send(events.Data)
			current++
		}
	}
}
