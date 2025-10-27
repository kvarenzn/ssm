// Copyright (C) 2024, 2025 kvarenzn
// SPDX-License-Identifier: GPL-3.0-or-later

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

	"github.com/kvarenzn/ssm/locale"
)

var SSM_VERSION = "(unknown)"

// flags
var (
	backend      string
	songID       int
	difficulty   string
	extract      string
	direction    string
	chartPath    string
	deviceSerial string
	showDebugLog bool
	showVersion  bool
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
	log.Infof("To use adb as the backend, the third-party component `scrcpy-server` (version %s) is required.", SERVER_FILE_VERSION)
	log.Infoln("This component is developed by Genymobile and licensed under Apache License 2.0.")
	log.Infoln()
	log.Infoln("Please download it from the official release page and place it in the same directory as `ssm.exe`.")
	log.Infoln("Download link:", SERVER_FILE_DOWNLOAD_URL)
	log.Infoln()
	log.Infoln("Alternatively, ssm can automatically handle this process for you.")
	log.Info("Proceed with automatic download? [Y/n]: ")
	var input string
	_, err := fmt.Scanln(&input)
	if err != nil {
		log.Die("Failed to get input:", err)
	}

	if input == "N" || input == "n" {
		log.Die("`scrcpy-server` is required. To use `adb` as the backend, you should download it manually.")
	}

	log.Infoln("Downloading... Please wait.")

	res, err := http.Get(SERVER_FILE_DOWNLOAD_URL)
	if err != nil {
		log.Dieln("Failed to download `scrcpy-server`.",
			locale.P.Sprintf("Error: %s", err),
			"You may try again later, download it manually, or use `hid` backend instead.")
	}

	data, err := io.ReadAll(res.Body)
	if err != nil {
		log.Dieln("Failed to download `scrcpy-server`.",
			fmt.Sprintf("Error: %s", err),
			"You may try again later, download it manually, or use `hid` backend instead.")
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
	size           *term.TermSize
	playing        bool
	start          time.Time
	offset         int
	controller     controllers.Controller
	events         []common.ViscousEventItem
	firstTick      int64
	loadFailed     bool
	orignal        image.Image
	scaled         image.Image
	graphicsMethod term.GraphicsMethod
	renderMutex    *sync.Mutex
	sigwinch       chan os.Signal
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

	term.SetWindowTitle(locale.P.Sprintf("ssm: READY"))

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

	if chartPath != "" {
		return fmt.Errorf("No song ID provided")
	}

	path := songsData.Jacket(songID)
	if path == "" {
		return fmt.Errorf("Jacket not found")
	}

	t.graphicsMethod = term.GetGraphicsMethod()
	switch t.graphicsMethod {
	case term.HALF_BLOCK:
		fallthrough
	case term.OVERSTRIKED_DOTS:
		path = filepath.Join(path, "thumb.png")
	case term.SIXEL_PROTOCOL:
		fallthrough
	case term.ITERM2_GRAPHICS_PROTOCOL:
		fallthrough
	case term.KITTY_GRAPHICS_PROTOCOL:
		path = filepath.Join(path, "jacket.png")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	t.orignal, err = term.DecodeImage(data)
	if err != nil {
		return err
	}

	var length int
	switch t.graphicsMethod {
	case term.HALF_BLOCK:
		length = jacketHeight * 2
	case term.OVERSTRIKED_DOTS:
		length = jacketHeight * 4
	case term.SIXEL_PROTOCOL:
		fallthrough
	case term.KITTY_GRAPHICS_PROTOCOL:
		length = t.size.CellHeight * jacketHeight
	}

	if length > 0 {
		scaled := image.NewNRGBA(image.Rect(0, 0, length, length))
		draw.BiLinear.Scale(scaled, scaled.Rect, t.orignal, t.orignal.Bounds(), draw.Src, nil)
		t.scaled = scaled
	}
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
		var length int
		switch t.graphicsMethod {
		case term.HALF_BLOCK:
			if t.scaled != nil {
				length = jacketHeight * 2
			}
		case term.OVERSTRIKED_DOTS:
			if t.scaled != nil {
				length = jacketHeight * 4
			}
		case term.SIXEL_PROTOCOL:
			fallthrough
		case term.KITTY_GRAPHICS_PROTOCOL:
			if t.scaled == nil || t.size == nil || newSize.CellHeight != t.size.CellHeight {
				length = newSize.CellHeight * jacketHeight
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
	fmt.Print(strings.Repeat(" ", max((cols-width)/2, 0)))
	fmt.Print(s)
	term.ClearToRight()
	fmt.Println()
}

func displayDifficulty() string {
	switch difficulty {
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
	fmt.Println()
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

	if full && (t.scaled != nil || t.graphicsMethod == term.ITERM2_GRAPHICS_PROTOCOL && t.orignal != nil) {
		switch t.graphicsMethod {
		case term.HALF_BLOCK:
			term.DisplayImageUsingHalfBlock(t.scaled, false, (t.size.Col-jacketHeight*2)/2)
		case term.OVERSTRIKED_DOTS:
			term.DisplayImageUsingOverstrikedDots(t.scaled, 0, 0, (t.size.Col-jacketHeight*2)/2)
		case term.SIXEL_PROTOCOL:
			term.DisplayImageUsingSixelProtocol(t.scaled, t.size, jacketHeight)
		case term.ITERM2_GRAPHICS_PROTOCOL:
			term.DisplayImageUsingITerm2Protocol(t.orignal, t.size, jacketHeight)
		case term.KITTY_GRAPHICS_PROTOCOL:
			term.DisplayImageUsingKittyProtocol(t.scaled, t.size, jacketHeight)
		}
	} else {
		term.MoveDownAndReset(jacketHeight)
	}

	t.emptyLine()

	if chartPath == "" {
		t.pcenterln(fmt.Sprintf("%s%s", displayDifficulty(), songsData.Title(songID, "\x1b[1m%title\x1b[0m")))
		t.pcenterln(songsData.Title(songID, "%artist"))
	} else {
		t.pcenterln(chartPath)
	}

	t.emptyLine()

	if !t.playing {
		t.pcenterln(locale.P.Sprintf("ui line 0"))
		t.emptyLine()
		t.emptyLine()
	} else {
		t.pcenterln(locale.P.Sprintf("Offset: %d ms", t.offset))
		t.pcenterln(locale.P.Sprintf("ui line 1"))
		t.pcenterln(locale.P.Sprintf("ui line 2"))
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
	if len(chartPath) == 0 {
		term.SetWindowTitle(locale.P.Sprintf("ssm: Autoplaying %s (%s)", songsData.Title(songID, "%title :: %artist"), strings.ToUpper(difficulty)))
	} else {
		term.SetWindowTitle(locale.P.Sprintf("ssm: Autoplaying %s", chartPath))
	}
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
	if deviceSerial == "" {
		device = adb.FirstAuthorizedDevice(devices)
		if device == nil {
			log.Die("No authorized devices.")
		}
	} else {
		for _, d := range devices {
			if d.Serial() == deviceSerial {
				device = d
				break
			}
		}

		if device == nil {
			log.Dief("No device has serial `%s`", deviceSerial)
		}

		if !device.Authorized() {
			log.Dief("Found device with serial number `%s`, but that device is not authorized.", deviceSerial)
		}
	}

	log.Debugln("Selected device:", device)
	controller := controllers.NewScrcpyController(device)
	if err := controller.Open("./scrcpy-server-v3.3.1", "3.3.1"); err != nil {
		log.Die("Failed to connect to device:", err)
	}
	defer controller.Close()

	dc := conf.Get(device.Serial())
	events := controller.Preprocess(rawEvents, direction == "right", dc)

	t.init(controller, events)

	t.begin()

	go t.waitForKey()

	t.autoplay()
}

func (t *tui) hidBackend(conf *config.Config, rawEvents common.RawVirtualEvents) {
	if deviceSerial == "" {
		serials := controllers.FindHIDDevices()
		log.Debugln("Recognized devices:", serials)

		if len(serials) == 0 {
			log.Die(errNoDevice)
		}

		deviceSerial = serials[0]
	}

	dc := conf.Get(deviceSerial)
	controller := controllers.NewHIDController(dc)
	controller.Open()
	defer controller.Close()

	events := controller.Preprocess(rawEvents, direction == "right")
	t.init(controller, events)

	t.begin()

	go t.waitForKey()

	t.autoplay()
}

func main() {
	log.Debugf("LANG: %s", locale.LanguageString)
	p := locale.P

	var err error

	flag.Usage = func() {
		fmt.Fprintln(flag.CommandLine.Output(), p.Sprintf("Usage of %s:", os.Args[0]))
		flag.PrintDefaults()
	}

	flag.StringVar(&backend, "b", "hid", p.Sprintf("usage.b"))
	flag.IntVar(&songID, "n", -1, p.Sprintf("usage.n"))
	flag.StringVar(&difficulty, "d", "", p.Sprintf("usage.d"))
	flag.StringVar(&extract, "e", "", p.Sprintf("usage.e"))
	flag.StringVar(&direction, "r", "left", p.Sprintf("usage.r"))
	flag.StringVar(&chartPath, "p", "", p.Sprintf("usage.p"))
	flag.StringVar(&deviceSerial, "s", "", p.Sprintf("usage.s"))
	flag.BoolVar(&showDebugLog, "g", false, p.Sprintf("usage.g"))
	flag.BoolVar(&showVersion, "v", false, p.Sprintf("usage.v"))

	flag.Parse()

	term.Hello()
	defer term.Bye()

	log.ShowDebug(showDebugLog)

	if extract != "" {
		db, err := Extract(extract, func(path string) bool {
			return (strings.Contains(path, "musicscore") || strings.Contains(path, "musicjacket")) && !strings.HasSuffix(path, ".acb.bytes")
		})
		if err != nil {
			log.Die(err)
		}

		data, err := json.MarshalIndent(db, "", "\t")
		if err != nil {
			log.Die(err)
		}

		os.WriteFile("./extract.json", data, 0o644)
		return
	}

	songsData, err = LoadSongsData()
	if err != nil {
		songsData = nil
	}

	switch locale.LanguageString[:2] {
	case "zh":
		if locale.LanguageString == "zh_TW" || locale.LanguageString == "zh-TW" {
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

	if showVersion {
		fmt.Println(p.Sprintf("ssm version: %s", p.Sprintf(SSM_VERSION)))
		fmt.Println(p.Sprintf("copyright info"))
		return
	}

	const CONFIG_PATH = "./config.json"

	conf, err := config.Load(CONFIG_PATH)
	if err != nil {
		log.Die(err)
	}

	if chartPath == "" && (songID == -1 || difficulty == "") {
		log.Die("Song id and difficulty are both required")
	}

	var chartText []byte
	if chartPath == "" {
		const BaseFolder = "./assets/star/forassetbundle/startapp/musicscore/"
		pathResults, err := filepath.Glob(filepath.Join(BaseFolder, fmt.Sprintf("musicscore*/%03d/*_%s.txt", songID, difficulty)))
		if err != nil {
			log.Die("Failed to find musicscore file:", err)
		}

		if len(pathResults) < 1 {
			log.Die("Musicscore not found")
		}

		log.Debugln("Musicscore loaded:", pathResults[0])
		chartText, err = os.ReadFile(pathResults[0])
	} else {
		log.Debugln("Musicscore loaded:", chartPath)
		chartText, err = os.ReadFile(chartPath)
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
		switch backend {
		case "adb":
			t.adbBackend(conf, rawEvents)
		case "hid":
			t.hidBackend(conf, rawEvents)
		default:
			log.Dief("Unknown backend: %q", backend)
		}
		stop()
	}()

	<-ctx.Done()

	if err := t.deinit(); err != nil {
		log.Die(err)
	}
}
