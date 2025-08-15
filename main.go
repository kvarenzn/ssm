package main

import (
	"context"
	"crypto"
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"io"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/kvarenzn/ssm/adb"
	"github.com/kvarenzn/ssm/common"
	"github.com/kvarenzn/ssm/config"
	"github.com/kvarenzn/ssm/controllers"
	"github.com/kvarenzn/ssm/log"
	"github.com/kvarenzn/ssm/term"
	"golang.org/x/image/draw"
)

var SSM_VERSION = "(unknown)"

var (
	backend      = flag.String("b", "hid", "Specify ssm backend, possible values: `hid`, `adb`")
	songID       = flag.Int("n", -1, "Song ID")
	difficulty   = flag.String("d", "", "Difficulty of song")
	extract      = flag.String("e", "", "Extract assets from assets folder <path>")
	direction    = flag.String("r", "left", "Device orientation, options: `left` (↺, counter-clockwise), `right` (↻, clockwise). Note: ignored when using `adb` backend")
	chartPath    = flag.String("p", "", "Custom chart path (if this is provided, song ID and difficulty will be ignored)")
	deviceSerial = flag.String("s", "", "Specify the device serial (if not provided, ssm will use the first device serial)")
	showDebugLog = flag.Bool("g", false, "Display useful information for debugging")
	showVersion  = flag.Bool("v", false, "Show ssm's version number and exit")
)

var (
	songsData    *SongsData
	preferLocale int
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
	log.Info("Proceed with automatic download? [Y/n]: ")
	var input string
	_, err := fmt.Scanln(&input)
	if err != nil {
		log.Die("Failed to get input from user:", err)
	}

	if input == "N" || input == "n" {
		log.Die("`scrcpy-server` is required. To use `adb` as the backend, you should download it manually.")
	}

	log.Infoln("Downloading... Please wait.")

	res, err := http.Get(SERVER_FILE_DOWNLOAD_URL)
	if err != nil {
		log.Dieln("Failed to download `scrcpy-server`.",
			fmt.Sprintf("Error: %s", err),
			"Try again later, download manually, or use `hid` backend instead.")
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
		log.Die("Checksum mismatch. Please try again later.")
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
			log.Warn("Checksum mismatch. File may be corrupted.")
			downloadServer()
		}
	}
}

const (
	errNoDevice = "Please connect your Android device to this computer."
)

const jacketHeight = 15

type tui struct {
	size        *term.TermSize
	playing     bool
	start       time.Time
	offset      int
	controller  controllers.Controller
	events      []common.ViscousEventItem
	firstTick   int64
	loadFailed  bool
	orignal     image.Image
	scaled      image.Image
	graphics    bool
	renderMutex *sync.Mutex
	sigwinch    chan os.Signal
}

func newTui() *tui {
	return &tui{
		renderMutex: &sync.Mutex{},
		sigwinch:    make(chan os.Signal, 1),
	}
}

func (t *tui) init(controller controllers.Controller, events []common.ViscousEventItem) error {
	if err := term.PrepareTerminal(); err != nil {
		return err
	}

	log.SetBeforeDie(func() {
		t.deinit()
	})

	if err := t.onResize(); err != nil {
		return err
	}

	t.controller = controller
	t.events = events

	t.startListenResize()

	return nil
}

func (t *tui) loadJacket() error {
	var err error
	if t.size == nil {
		t.size, err = term.GetTerminalSize()
		if err != nil {
			return err
		}
	}

	if *chartPath != "" {
		return fmt.Errorf("No song ID provided")
	}

	path := songsData.Jacket(*songID)
	if path == "" {
		return fmt.Errorf("Jacket not found")
	}

	t.graphics = term.SupportsGraphics()
	if t.graphics {
		path = filepath.Join(path, "jacket.png")
	} else {
		path = filepath.Join(path, "thumb.png")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	t.orignal, err = term.DecodeImage(data)
	if err != nil {
		return err
	}

	var scaled *image.NRGBA
	if t.graphics {
		length := t.size.CellHeight * jacketHeight
		scaled = image.NewNRGBA(image.Rect(0, 0, length, length))
	} else {
		scaled = image.NewNRGBA(image.Rect(0, 0, 30, 30))
	}
	draw.BiLinear.Scale(scaled, scaled.Rect, t.orignal, t.orignal.Bounds(), draw.Src, nil)
	t.scaled = scaled
	return nil
}

func (t *tui) startListenResize() {
	term.StartWatchResize(t.sigwinch)
	go func() {
		for range t.sigwinch {
			t.onResize()
		}
	}()
}

func (t *tui) onResize() error {
	newSize, err := term.GetTerminalSize()
	if err != nil {
		return err
	}

	if t.orignal == nil && !t.loadFailed {
		if err := t.loadJacket(); err != nil {
			log.Debugf("Failed to load music jacket: %s", err)
			t.loadFailed = true
		}
	}

	if t.orignal != nil {
		length := 0
		if t.graphics {
			if t.scaled == nil || t.size == nil || newSize.CellHeight != t.size.CellHeight {
				length = newSize.CellHeight * jacketHeight
			}
		} else {
			if t.scaled != nil {
				length = 30
			}
		}

		if length > 0 {
			s := image.NewNRGBA(image.Rect(0, 0, length, length))
			draw.BiLinear.Scale(s, s.Rect, t.orignal, t.orignal.Bounds(), draw.Src, nil)
			t.scaled = s
		}
	}

	t.size = newSize

	term.ClearScreen()

	t.render(true)
	return nil
}

func (t *tui) pcenterln(s string) {
	if t.size == nil {
		return
	}

	term.MoveHome()
	cols := t.size.Col
	width := term.WidthOf(s)
	print(strings.Repeat(" ", max((cols-width)/2, 0)))
	print(s)
	term.ClearToRight()
	println()
}

func displayDifficulty() string {
	switch *difficulty {
	case "easy":
		return "\x1b[0;44m EASY \x1b[0m "
	case "normal":
		return "\x1b[0;42m NORMAL \x1b[0m "
	case "hard":
		return "\x1b[0;43m HARD \x1b[0m "
	case "expert":
		return "\x1b[0;41m EXPERT \x1b[0m "
	default:
		return ""
	}
}

func (t *tui) emptyLine() {
	term.ClearCurrentLine()
	println()
}

func (t *tui) render(full bool) {
	if t.size == nil {
		return
	}

	if ok := t.renderMutex.TryLock(); !ok {
		return
	}

	term.ResetCursor()
	t.emptyLine()

	if full && t.scaled != nil {
		if term.SupportsGraphics() {
			padLeftPixels := (t.size.Xpixel - t.scaled.Bounds().Dx()) / 2
			print(strings.Repeat(" ", padLeftPixels/t.size.CellWidth))
			term.ClearToRight()
			term.DisplayImageUsingKittyProtocol(t.scaled, true, padLeftPixels%t.size.CellWidth, 0)
		} else {
			term.DisplayImageUsingHalfBlock(t.scaled, false, (t.size.Col-jacketHeight*2)/2)
		}
	} else {
		term.MoveDownAndReset(jacketHeight)
	}

	t.emptyLine()

	if *chartPath == "" {
		t.pcenterln(fmt.Sprintf("%s%s", displayDifficulty(), songsData.Title(*songID, "\x1b[1m%title\x1b[0m")))
		t.pcenterln(songsData.Title(*songID, "%artist"))
	} else {
		t.pcenterln(*chartPath)
	}

	t.emptyLine()

	if !t.playing {
		t.pcenterln("\x1b[7m\x1b[1m ENTER/SPACE \x1b[0m GO!!!!!")
		t.emptyLine()
		t.emptyLine()
	} else {
		t.pcenterln(fmt.Sprintf("Offset: %d ms", t.offset))
		t.pcenterln("\x1b[7m\x1b[1m ← \x1b[0m -10ms   \x1b[7m\x1b[1m Shift-← \x1b[0m -50ms   \x1b[7m\x1b[1m Ctrl-← \x1b[0m -100ms   \x1b[7m\x1b[1m Ctrl-C \x1b[0m Stop")
		t.pcenterln("\x1b[7m\x1b[1m → \x1b[0m +10ms   \x1b[7m\x1b[1m Shift-→ \x1b[0m +50ms   \x1b[7m\x1b[1m Ctrl-→ \x1b[0m +100ms                ")
	}

	t.renderMutex.Unlock()
}

func (t *tui) begin() {
	t.firstTick = t.events[0].Timestamp

	for {
		key, err := term.ReadKey(os.Stdin, 10*time.Millisecond)
		if err != nil {
			log.Dief("Failed to get key from stdin: %s", err)
		}

		if key == term.KEY_ENTER || key == term.KEY_SPACE {
			break
		}
	}

	t.playing = true
	t.start = time.Now().Add(-time.Duration(t.firstTick) * time.Millisecond)
	t.offset = 0
	t.render(false)
}

func (t *tui) addOffset(delta int) {
	t.offset += delta
	t.start = t.start.Add(time.Duration(-delta) * time.Millisecond)
	t.render(false)
}

func (t *tui) waitForKey() {
	for {
		key, err := term.ReadKey(os.Stdin, 10*time.Millisecond)
		if err != nil {
			log.Dief("Failed to get key from stdin: %s", err)
		}

		switch key {
		case term.KEY_LEFT:
			t.addOffset(-10)
		case term.KEY_SHIFT_LEFT:
			t.addOffset(-50)
		case term.KEY_CTRL_LEFT:
			t.addOffset(-100)
		case term.KEY_RIGHT:
			t.addOffset(10)
		case term.KEY_SHIFT_RIGHT:
			t.addOffset(50)
		case term.KEY_CTRL_RIGHT:
			t.addOffset(100)
		}
	}
}

func (t *tui) deinit() error {
	if err := term.RestoreTerminal(); err != nil {
		return err
	}

	term.Bye()
	return nil
}

func (t *tui) autoplay() {
	current := 0
	n := len(t.events)
	for current < n {
		now := time.Since(t.start).Milliseconds()
		event := t.events[current]
		remaining := event.Timestamp - now

		if remaining <= 0 {
			t.controller.Send(event.Data)
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

func (t *tui) adbBackend(conf *config.Config, rawEvents common.RawVirtualEvents) {
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

	t.init(controller, events)

	t.begin()

	go t.waitForKey()

	t.autoplay()
}

func (t *tui) hidBackend(conf *config.Config, rawEvents common.RawVirtualEvents) {
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
	t.init(controller, events)

	t.begin()

	go t.waitForKey()

	t.autoplay()
}

func main() {
	flag.Parse()

	term.Hello()

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
		term.Bye()
		return
	}

	log.ShowDebug(*showDebugLog)

	var err error
	songsData, err = LoadSongsData()
	if err != nil {
		songsData = nil
	}

	lang, err := GetSystemLocale()
	if err != nil {
		lang = ""
	}
	log.Debugf("LANG: %s", lang)

	switch lang[:2] {
	case "zh":
		if lang == "zh_TW" || lang == "zh-TW" {
			preferLocale = 2 // LANG: zh_TW
		} else {
			preferLocale = 3 // LANG: zh_CN
		}
	case "ko":
		preferLocale = 4 // LANG: ko_KR
	case "en":
		preferLocale = 1 // LANG: en_US
	default:
		preferLocale = 0 // LANG: ja_JP
	}

	if *showVersion {
		fmt.Printf("ssm version: %s\n", SSM_VERSION)
		return
	}

	const CONFIG_PATH = "./config.json"

	conf, err := config.Load(CONFIG_PATH)
	if err != nil {
		log.Die(err)
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
			log.Die("Musicscore not found")
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

	t := newTui()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT)
	defer stop()

	go func() {
		switch *backend {
		case "adb":
			t.adbBackend(conf, rawEvents)
		case "hid":
			t.hidBackend(conf, rawEvents)
		default:
			log.Dief("Unknown backend: %q", *backend)
		}
		stop()
	}()

	<-ctx.Done()

	if err := t.deinit(); err != nil {
		log.Die(err)
	}

	term.Bye()
}
