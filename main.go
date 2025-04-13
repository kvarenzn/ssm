package main

import (
	"flag"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"time"

	"github.com/kvarenzn/ssm/adb"
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

func TestAdb() {
	if err := adb.StartADBServer("localhost", 5037); err != nil && err != adb.ErrADBServerRunning {
		Fatal(err)
	}

	client := adb.NewDefaultClient()
	devices, err := client.Devices()
	if err != nil {
		Fatal(err)
	}

	if len(devices) == 0 {
		Fatal("no android devices recognized")
	}

	device := devices[0]
	if state, err := device.State(); err != nil {
		Fatal(err)
	} else if state != "device" {
		Fatal("not authorized")
	}

	// try upload a file
	f, err := os.Open("./scrcpy-server-v3.1")
	if err != nil {
		Fatal(err)
	}

	if err := device.Push(f, "/data/local/tmp/scrcpy-server.jar"); err != nil {
		Fatal(err)
	}

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
		Fatal(err)
	}

	Info(result)

	os.Exit(0)
}

func main() {
	// TestAdb()

	flag.Parse()

	if *showVersion {
		fmt.Printf("ssm version: %s\n", SSM_VERSION)
		return
	}

	const CONFIG_PATH = "./config.json"

	if err := LoadConfig(CONFIG_PATH); err != nil {
		Fatal(err)
	}

	if *extract != "" {
		Extract(*extract)
		return
	}

	if len(*chartPath) == 0 && (*songID == -1 || *difficulty == "") {
		Fatal("Song id and difficulty are both required")
	}

	if *deviceSerial == "" {
		serials := FindDevices()
		Info("Recognized devices:", serials)

		if len(serials) == 0 {
			Fatal("plug your gaming device to pc")
		}

		*deviceSerial = serials[0]
	}

	dc, ok := GlobalConfig.Devices[*deviceSerial]
	if !ok {
		dc = GlobalConfig.AskForSerial(*deviceSerial)
		SaveConfig(CONFIG_PATH)
	}

	controller := NewHIDController(dc.Width, dc.Height, *deviceSerial)
	controller.Open()
	defer controller.Close()

	var text []byte
	var err error
	if len(*chartPath) == 0 {
		baseFolder := "./assets/star/forassetbundle/startapp/musicscore/"
		pathResults, err := filepath.Glob(filepath.Join(baseFolder, fmt.Sprintf("musicscore*/%03d/*_%s.txt", *songID, *difficulty)))
		if err != nil {
			Fatal(err)
		}

		if len(pathResults) < 1 {
			Fatal("not found")
		}

		Info("Music score loaded:", pathResults[0])
		text, err = os.ReadFile(pathResults[0])
	} else {
		Info("Music score loaded:", *chartPath)
		text, err = os.ReadFile(*chartPath)
	}

	if err != nil {
		Fatal(err)
	}

	chart := Parse(string(text))
	config := VTEGenerateConfig{
		TapDuration:         10,
		FlickDuration:       60,
		FlickReportInterval: 5,
		SlideReportInterval: 10,
	}
	rawEvents := GenerateTouchEvent(config, chart)

	processFn := func(x, y float64) (int, int) {
		return int(math.Round(float64(dc.Width-dc.Line.Y) + float64(dc.Line.Y-dc.Width/2)*y)), int(math.Round(float64(dc.Line.X1) + float64(dc.Line.X2-dc.Line.X1)*x))
	}

	if *direction == "right" {
		processFn = func(x, y float64) (int, int) {
			ix, iy := processFn(x, y)
			return dc.Width - ix, dc.Height - iy
		}
	}

	vEvents := preprocess(processFn, rawEvents)

	Info("Ready. Press ENTER to start autoplay.")
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
