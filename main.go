package main

import (
	"crypto"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kvarenzn/ssm/adb"
	"github.com/kvarenzn/ssm/common"
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
	backend      = flag.String("b", "hid", "Specify ssm backend, possible values: `hid`, `adb`")
	songID       = flag.Int("n", -1, "Song ID")
	difficulty   = flag.String("d", "", "Difficulty of song")
	extract      = flag.String("e", "", "Extract assets from assets folder <path>")
	direction    = flag.String("r", "left", "Direction of device, possible values: `left` (↺, anti-clockwise), `right` (↻, clockwise). Note: this takes no effect when the backend is `adb`")
	chartPath    = flag.String("p", "", "Custom chart path (if this is provided, song ID and difficulty will be ignored)")
	deviceSerial = flag.String("s", "", "Specify the device serial (if not provided, ssm will use the first device serial)")
	showDebugLog = flag.Bool("g", false, "Display useful information for debugging")
	showVersion  = flag.Bool("v", false, "Show ssm's version number and exit")
)

const (
	SERVER_FILE_VERSION      = "3.3.1"
	SERVER_FILE              = "scrcpy-server-v" + SERVER_FILE_VERSION
	SERVER_FILE_DOWNLOAD_URL = "https://github.com/Genymobile/scrcpy/releases/download/v" + SERVER_FILE_VERSION + "/" + SERVER_FILE
	SERVER_FILE_SHA256       = "a0f70b20aa4998fbf658c94118cd6c8dab6abbb0647a3bdab344d70bc1ebcbb8"
)

func downloadServer() {
	log.Infoln("To use ADB as the backend, the third-party component `scrcpy-server` (version 3.3.1) is required.")
	log.Infoln("This component is developed by Genymobile and licensed under Apache License 2.0.")
	log.Infoln()
	log.Infoln("Please download it from the official release page and place it in the same directory as `ssm.exe`.")
	log.Infoln("Download link:", SERVER_FILE_DOWNLOAD_URL)
	log.Infoln()
	log.Infoln("Alternatively, SSM can automatically handle this process for you.")
	log.Info("Would you like to proceed with automatic download? [Y/N]: ")
	var input string
	_, err := fmt.Scanln(&input)
	if err != nil {
		log.Die("Failed to get input from user:", err)
	}

	if input != "Y" && input != "y" {
		log.Die("`scrcpy-server` is required. To use `adb` as the backend, you should download it manually.")
	}

	log.Infoln("Downloading... Please wait.")

	res, err := http.Get(SERVER_FILE_DOWNLOAD_URL)
	if err != nil {
		log.Dieln("Failed to download `scrcpy-server`.",
			fmt.Sprintf("Error: %s", err),
			"You may try again later, download it manually or choose `hid` as backend.")
	}

	data, err := io.ReadAll(res.Body)
	if err != nil {
		log.Dieln("Failed to download `scrcpy-server`.",
			fmt.Sprintf("Error: %s", err),
			"You may try again later, download it manually or choose `hid` as backend.")
	}

	h := crypto.SHA256.New()
	if _, err := h.Write(data); err != nil {
		log.Die("Failed to calculate sha256 of `scrcpy-server`:", err)
	}

	if fmt.Sprintf("%x", h.Sum(nil)) != SERVER_FILE_SHA256 {
		log.Die("Hashsum mismatch. You may try again later.")
	}

	if err := os.WriteFile(SERVER_FILE, data, 0o644); err != nil {
		log.Die("Failed to save `scrcpy-server` to disk:", err)
	}
}

func checkOrDownload() {
	if _, err := os.Stat(SERVER_FILE); err != nil {
		if !os.IsNotExist(err) {
			log.Die("Failed to locate server file:", err)
		}

		downloadServer()
	} else {
		data, err := os.ReadFile(SERVER_FILE)
		if err != nil {
			log.Die("Failed to read the content of `scrcpy-server`:", err)
		}

		h := crypto.SHA256.New()
		if _, err := h.Write(data); err != nil {
			log.Die("Failed to calculate sha256 of `scrcpy-server`:", err)
		}

		if fmt.Sprintf("%x", h.Sum(nil)) != SERVER_FILE_SHA256 {
			log.Warn("Hashsum mismatch. File might be corrupted.")
			downloadServer()
		}
	}
}

const (
	errNoDevice = "Plug your gaming android device to this device."
)

func adbBackend(conf *config.Config, rawEvents common.RawVirtualEvents) {
	checkOrDownload()
	if err := adb.StartADBServer("localhost", 5037); err != nil && err != adb.ErrADBServerRunning {
		log.Fatal(err)
	}

	client := adb.NewDefaultClient()
	devices, err := client.Devices()
	if err != nil {
		log.Fatal(err)
	}

	if len(devices) == 0 {
		log.Die(errNoDevice)
	}

	log.Debugln("ADB devices:", devices)

	var device *adb.Device
	if *deviceSerial == "" {
		device = adb.FirstAuthorizedDevice(devices)
		if device == nil {
			log.Die("No authorized devices.")
		}
	} else {
		for _, d := range devices {
			if d.Serial() == *deviceSerial {
				device = d
				break
			}
		}

		if device == nil {
			log.Dief("No device has serial `%s`", *deviceSerial)
		}

		if !device.Authorized() {
			log.Dief("Found device with serial number `%s`, but that device is not authorized", *deviceSerial)
		}
	}

	log.Debugln("Selected device:", device)
	controller := controllers.NewScrcpyController(device)
	if err := controller.Open("./scrcpy-server-v3.3.1", "3.3.1"); err != nil {
		log.Die("Failed to connect to device:", err)
	}
	defer controller.Close()

	dc := conf.Get(device.Serial())
	events := controller.Preprocess(rawEvents, *direction == "right", dc)

	firstTick := events[0].Timestamp

	log.Infoln("Ready. Press ENTER to start autoplay.")
	fmt.Scanln()
	log.Infoln("Autoplaying... Press Ctrl-C to interrupt.")

	start := time.Now().Add(-time.Duration(firstTick) * time.Millisecond)

	current := 0
	for current < len(events) {
		now := time.Since(start).Milliseconds()
		event := events[current]
		remaining := event.Timestamp - now

		if remaining <= 0 {
			controller.Send(event.Data)
			current++
			continue
		}

		if remaining > 10 {
			time.Sleep(time.Duration(remaining-5) * time.Millisecond)
		} else if remaining > 4 {
			time.Sleep(1 * time.Millisecond)
		}
	}
}

func hidBackend(conf *config.Config, rawEvents common.RawVirtualEvents) {
	if *deviceSerial == "" {
		serials := controllers.FindHIDDevices()
		log.Debugln("Recognized devices:", serials)

		if len(serials) == 0 {
			log.Die(errNoDevice)
		}

		*deviceSerial = serials[0]
	}

	dc := conf.Get(*deviceSerial)
	controller := controllers.NewHIDController(dc)
	controller.Open()
	defer controller.Close()

	events := controller.Preprocess(rawEvents, *direction == "right")
	firstTick := events[0].Timestamp

	log.Infoln("Ready. Press ENTER to start autoplay.")
	fmt.Scanln()
	log.Infoln("Autoplaying... Press Ctrl-C to interrupt.")

	start := time.Now().Add(-time.Duration(firstTick) * time.Millisecond)

	current := 0
	for current < len(events) {
		now := time.Since(start).Milliseconds()
		event := events[current]
		remaining := event.Timestamp - now

		if remaining <= 0 {
			controller.Send(event.Data)
			current++
			continue
		}

		if remaining > 10 {
			time.Sleep(time.Duration(remaining-5) * time.Millisecond)
		} else if remaining > 4 {
			time.Sleep(1 * time.Millisecond)
		}
	}
}

func main() {
	flag.Parse()

	if *showVersion {
		fmt.Printf("ssm version: %s\n", SSM_VERSION)
		return
	}

	log.ShowDebug(*showDebugLog)

	const CONFIG_PATH = "./config.json"

	conf, err := config.Load(CONFIG_PATH)
	if err != nil {
		log.Die(err)
	}

	if *extract != "" {
		db, err := Extract(*extract, func(path string) bool {
			return (strings.Contains(path, "musicscore") || strings.Contains(path, "musicjacket")) && !strings.HasSuffix(path, ".acb.bytes")
		})
		if err != nil {
			log.Die(err)
		}

		data, err := json.Marshal(db)
		if err != nil {
			log.Die(err)
		}

		os.WriteFile("./extract.json", data, 0o644)
		return
	}

	if len(*chartPath) == 0 && (*songID == -1 || *difficulty == "") {
		log.Die("Song id and difficulty are both required")
	}

	var chartText []byte
	if len(*chartPath) == 0 {
		const BaseFolder = "./assets/star/forassetbundle/startapp/musicscore/"
		pathResults, err := filepath.Glob(filepath.Join(BaseFolder, fmt.Sprintf("musicscore*/%03d/*_%s.txt", *songID, *difficulty)))
		if err != nil {
			log.Die("Failed to find musicscore file:", err)
		}

		if len(pathResults) < 1 {
			log.Die("Musicscore file not found")
		}

		log.Debugln("Musicscore loaded:", pathResults[0])
		chartText, err = os.ReadFile(pathResults[0])
	} else {
		log.Debugln("Musicscore loaded:", *chartPath)
		chartText, err = os.ReadFile(*chartPath)
	}

	if err != nil {
		log.Die("Failed to load musicscore:", err)
	}

	chart := Parse(string(chartText))
	rawEvents := GenerateTouchEvent(VTEGenerateConfig{
		TapDuration:         10,
		FlickDuration:       60,
		FlickReportInterval: 5,
		SlideReportInterval: 10,
	}, chart)

	switch *backend {
	case "hid":
		hidBackend(conf, rawEvents)
	case "adb":
		adbBackend(conf, rawEvents)
	default:
		log.Dief("Unknown backend: %q", *backend)
	}
}
