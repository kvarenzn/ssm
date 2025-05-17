package main

import (
	"context"
	"flag"
	"fmt"
	"math"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/kvarenzn/ssm/term"
	"golang.org/x/sys/unix"
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
	enableServer = flag.Bool("server", false, "Enable server (EXPERIMENTAL. DO NOT USE)")
)

type ssm struct {
	events     []ViscousEventItem
	eventCount int
	ctl        *HIDController
	dc         DeviceConfig
	sz         *unix.Winsize
	ts         *unix.Termios
	current    int
	delay      int64
	done       chan struct{}
	start      chan struct{}
	running    bool
	wg         *sync.WaitGroup
}

func newSsm(c *Config, serial string) *ssm {
	dc := c.Query(serial)
	return &ssm{
		ctl: NewHIDController(dc.Width, dc.Height, serial),
		dc:  dc,
	}
}

func (s *ssm) load(path string, right bool) error {
	text, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	chart := Parse(string(text))
	rawEvents := GenerateTouchEvent(VTEGenerateConfig{
		TapDuration:         10,
		FlickDuration:       60,
		FlickReportInterval: 5,
		SlideReportInterval: 10,
	}, chart)

	var processFn func(float64, float64) (int, int)
	if right {
		processFn = func(x, y float64) (int, int) {
			ix, iy := processFn(x, y)
			return s.dc.Width - ix, s.dc.Height - iy
		}
	} else {
		processFn = func(x, y float64) (int, int) {
			return int(math.Round(float64(s.dc.Width-s.dc.Line.Y) + float64(s.dc.Line.Y-s.dc.Width/2)*y)), int(math.Round(float64(s.dc.Line.X1) + float64(s.dc.Line.X2-s.dc.Line.X1)*x))
		}
	}

	s.events = preprocess(processFn, rawEvents)
	s.eventCount = len(s.events)

	return nil
}

func (s *ssm) prepare() (err error) {
	if err = s.ctl.Open(); err != nil {
		return
	}

	s.sz, err = term.GetTerminalSize()
	if err != nil {
		return
	}

	term.HideCursor()
	s.ts, err = term.GetTermios()
	if err != nil {
		return
	}

	return term.PrepareTerminal()
}

func (s *ssm) loop() {
	Info("Ready. Press ENTER to start autoplay.")
	buf := make([]byte, 1)
	for {
		_, err := os.Stdin.Read(buf)
		if err != nil {
			continue
		}

		if buf[0] == '\r' || buf[0] == '\n' {
			break
		}
	}
	s.start <- struct{}{}

	firstEvent := s.events[0]
	tick := firstEvent.Timestamp
	start := time.Now().Add(-time.Duration(tick) * time.Millisecond)

	s.current = 0
	for s.current < s.eventCount {
		delta := time.Now().Sub(start).Milliseconds() - s.delay
		events := s.events[s.current]
		if delta >= events.Timestamp {
			s.ctl.Send(events.Data)
			s.current++
		}
		time.Sleep(1 * time.Millisecond)
	}
	s.running = false
	s.wg.Done()
}

func (s *ssm) progress() {
	<-s.start
	columns := int(s.sz.Col) - 4
	idx := 0
	for s.running {
		term.MoveToColumn(0)
		count := s.current * columns / s.eventCount
		print(" ")
		for i := range columns {
			if i == 0 {
				if i < count {
					print("\uee03")
				} else {
					print("\uee00")
				}
			} else if i == columns-1 {
				if i <= count {
					print("\uee05")
				} else {
					print("\uee02")
				}
			} else {
				if i < count {
					print("\uee04")
				} else {
					print("\uee01")
				}
			}
		}
		print(" ")
		print(string(rune('\uee06' + idx)))
		idx = (idx + 1) % 6
		time.Sleep(100 * time.Millisecond)
	}
	println()

	s.wg.Done()
}

func (s *ssm) run() chan struct{} {
	if s.done != nil {
		close(s.done)
	}
	s.done = make(chan struct{})
	s.start = make(chan struct{})
	s.wg = new(sync.WaitGroup)
	s.wg.Add(2)
	s.running = true

	go s.loop()
	go s.progress()
	go func() {
		s.wg.Wait()
		s.done <- struct{}{}
	}()
	return s.done
}

func (s *ssm) cleanup() {
	if s.ts != nil {
		if err := term.SetTermios(s.ts); err != nil {
			Fatal(err)
		}
	}

	term.ShowCursor()
}

func main() {
	flag.Parse()

	if *showVersion {
		fmt.Printf("ssm version: %s\n", SSM_VERSION)
		return
	}

	const CONFIG_PATH = "./config.json"

	if *extract != "" {
		Extract(*extract)
		return
	}

	if len(*chartPath) == 0 && (*songID == -1 || *difficulty == "") {
		Fatal("Song id and difficulty are both required")
	}

	if *deviceSerial == "" {
		serials, err := FindDevices()
		if err != nil {
			Fatal(err)
		}

		Info("Recognized devices:", serials)

		if len(serials) == 0 {
			Fatal("plug your gaming device to pc")
		}

		*deviceSerial = serials[0]
	}

	config, err := LoadConfig(CONFIG_PATH)
	if err != nil {
		Fatal(err)
	}

	s := newSsm(config, *deviceSerial)
	right := *direction == "right"

	var path string
	if len(*chartPath) == 0 {
		baseFolder := "./assets/star/forassetbundle/startapp/musicscore/"
		pathResults, err := filepath.Glob(filepath.Join(baseFolder, fmt.Sprintf("musicscore*/%03d/*_%s.txt", *songID, *difficulty)))
		if err != nil {
			Fatal(err)
		}

		if len(pathResults) < 1 {
			Fatal("not found")
		}

		path = pathResults[0]
	} else {
		path = *chartPath
	}

	if err = s.load(path, right); err != nil {
		Fatal(err)
	}
	Info("Music score loaded:", path)

	if err := s.prepare(); err != nil {
		Fatal(err)
	}

	done := s.run()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT)
	defer stop()

	select {
	case <-ctx.Done():
		break
	case <-done:
		break
	}

	s.cleanup()
}
