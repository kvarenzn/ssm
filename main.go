package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"time"
)

type Point struct {
	X int
	Y int
}

func tapper() {
	seriales := FindDevices()
	fmt.Println(seriales)

	if len(seriales) == 0 {
		fmt.Println("plug your gaming device to pc")
		return
	}

	const DEVICE_WIDTH = 1080
	const DEVICE_HEIGHT = 2340

	controller := NewHidController(DEVICE_WIDTH, DEVICE_HEIGHT, seriales[0])
	controller.Open()
	defer controller.Close()

	time.Sleep(1 * time.Second)

	rawEvents := RawVirtualEvents{
		{
			Timestamp: 0,
			Events: []VirtualTouchEvent{
				{
					X:      250,
					Y:      DEVICE_HEIGHT / 2,
					Action: TouchDown,
				},
			},
		},
		{
			Timestamp: 1,
			Events: []VirtualTouchEvent{
				{
					X:      250,
					Y:      DEVICE_HEIGHT / 2,
					Action: TouchUp,
				},
			},
		},
		{
			Timestamp: 2,
			Events: []VirtualTouchEvent{
				{
					X:      300,
					Y:      DEVICE_HEIGHT / 2,
					Action: TouchDown,
				},
			},
		},
		{
			Timestamp: 3,
			Events: []VirtualTouchEvent{
				{
					X:      300,
					Y:      DEVICE_HEIGHT / 2,
					Action: TouchUp,
				},
			},
		},
		{
			Timestamp: 4,
			Events: []VirtualTouchEvent{
				{
					X:      160,
					Y:      1750,
					Action: TouchDown,
				},
			},
		},
		{
			Timestamp: 5,
			Events: []VirtualTouchEvent{
				{
					X:      160,
					Y:      1750,
					Action: TouchUp,
				},
			},
		},
	}

	vEvents := preprocess(func(x, y float64) (int, int) {
		return int(x), int(y)
	}, rawEvents)

	current := 0
	for current < len(vEvents) {
		events := vEvents[current]
		controller.Send(events.Data)
		if current%2 == 0 {
			time.Sleep(10 * time.Millisecond)
		} else {
			time.Sleep(100 * time.Millisecond)
		}

		current = (current + 1) % 6
	}
}

var (
	mainPageStack  int32
	sashPos1       float32 = 200
	sashPos2       float32 = 500
	sashPos3       float32 = 80
	songSearchText string
)

var (
	songID       = flag.Int("n", -1, "Song ID")
	difficulty   = flag.String("d", "", "Difficulty of song")
	enableTapper = flag.Bool("t", false, "Tapper mode")
	extract      = flag.String("e", "", "Extract assets from assets folder <path>")
	direction    = flag.String("r", "left", "Direction of device, possible values: left (turn left), right (turn right)")
	chartPath    = flag.String("p", "", "Custom chart path (if this is provided, songID and difficulty will be ignored)")
)

func main() {
	flag.Parse()

	if *extract != "" {
		Extract(*extract)
		return
	}

	if *enableTapper {
		tapper()
		return
	}

	if len(*chartPath) == 0 && (*songID == -1 || *difficulty == "") {
		fmt.Println("Both song id and difficulty must be provided")
		os.Exit(1)
	}

	seriales := FindDevices()
	fmt.Println(seriales)

	if len(seriales) == 0 {
		fmt.Println("plug your gaming device to pc")
		return
	}

	const DEVICE_WIDTH = 1080
	const DEVICE_HEIGHT = 2340

	controller := NewHidController(DEVICE_WIDTH, DEVICE_HEIGHT, seriales[0])
	controller.Open()
	defer controller.Close()

	var text []byte
	var err error
	if len(*chartPath) == 0 {
		baseFolder := "./assets/star/forassetbundle/startapp/musicscore/"
		pathResults, err := filepath.Glob(filepath.Join(baseFolder, fmt.Sprintf("musicscore*/%03d/*_%s.txt", *songID, *difficulty)))
		if err != nil {
			log.Fatal(err)
		}

		if len(pathResults) < 1 {
			log.Fatal("not found")
		}

		fmt.Println("path:", pathResults[0])
		text, err = os.ReadFile(pathResults[0])
	} else {
		fmt.Println("path:", *chartPath)
		text, err = os.ReadFile(*chartPath)
	}

	if err != nil {
		log.Fatal(err)
	}

	chart := Parse(string(text))
	config := VTEGenerateConfig{
		TapDuration:         10,
		FlickDuration:       60,
		FlickReportInterval: 5,
		SlideReportInterval: 10,
	}
	rawEvents := GenerateTouchEvent(config, chart)

	err = drawMain(chart, rawEvents, "out.svg")
	if err != nil {
		log.Fatal(err)
	}

	processFn := func(x, y float64) (int, int) {
		return DEVICE_WIDTH - int(math.Round((540-844)*y+844)), int(math.Round((1800-540)*x + 540))
	}

	if *direction == "right" {
		processFn = func(x, y float64) (int, int) {
			return int(math.Round(540-844)*y + 844), DEVICE_HEIGHT - int(math.Round((1800-540)*x+540))
		}
	}

	vEvents := preprocess(processFn, rawEvents)

	fmt.Println("Ready.")
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
