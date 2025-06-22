package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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
	controller := controllers.NewScrcpyController(device)
	if controller.Open("./scrcpy-server-v3.3"); err != nil {
		log.Fatal(err)
	}

	controller.Close()

	os.Exit(0)
}

func main() {
	TestAdb()
	return

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
		db, err := Extract(*extract, func(path string) bool {
			return (strings.Contains(path, "musicscore") || strings.Contains(path, "musicjacket")) && !strings.HasSuffix(path, ".acb.bytes")
		})
		if err != nil {
			log.Fatal(err)
		}

		data, err := json.Marshal(db)
		if err != nil {
			log.Fatal(err)
		}

		os.WriteFile("./extract.json", data, 0o644)
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
