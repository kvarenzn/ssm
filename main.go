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

	rawEvents := RawVirtualEvents{}
	tick := int64(0)

	for i := 0; i < 1000; i++ {
		rawEvents = append(rawEvents, VirtualEventsItem{
			Timestamp: tick,
			Events: []VirtualTouchEvent{
				{
					X:      250,
					Y:      1430,
					Action: TouchDown,
				},
			},
		})

		tick += 10
		rawEvents = append(rawEvents, VirtualEventsItem{
			Timestamp: tick,
			Events: []VirtualTouchEvent{
				{
					X:      250,
					Y:      1430,
					Action: TouchUp,
				},
			},
		})

		tick += 100

		rawEvents = append(rawEvents, VirtualEventsItem{
			Timestamp: tick,
			Events: []VirtualTouchEvent{
				{
					X:      300,
					Y:      DEVICE_HEIGHT / 2,
					Action: TouchDown,
				},
			},
		})

		tick += 10
		rawEvents = append(rawEvents, VirtualEventsItem{
			Timestamp: tick,
			Events: []VirtualTouchEvent{
				{
					X:      300,
					Y:      DEVICE_HEIGHT / 2,
					Action: TouchUp,
				},
			},
		})

		tick += 100

		rawEvents = append(rawEvents, VirtualEventsItem{
			Timestamp: tick,
			Events: []VirtualTouchEvent{
				{
					X:      160,
					Y:      1750,
					Action: TouchDown,
				},
			},
		})

		tick += 10
		rawEvents = append(rawEvents, VirtualEventsItem{
			Timestamp: tick,
			Events: []VirtualTouchEvent{
				{
					X:      160,
					Y:      1750,
					Action: TouchUp,
				},
			},
		})

		tick += 100
	}

	vEvents := preprocess(func(x, y float64) (int, int) {
		return int(x), int(y)
	}, rawEvents)

	start := time.Now()

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

var (
	mainPageStack  int32
	sashPos1       float32 = 200
	sashPos2       float32 = 500
	sashPos3       float32 = 80
	songSearchText string
)

var songID = flag.Int("n", -1, "Song ID")
var difficulty = flag.String("d", "", "Difficulty of song")
var enableTapper = flag.Bool("t", false, "Tapper mode")
var extract = flag.String("e", "", "Extract assets from assets folder <path>")

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

	if *songID == -1 || *difficulty == "" {
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

	baseFolder := "./assets/star/forassetbundle/startapp/musicscore/"
	pathResults, err := filepath.Glob(filepath.Join(baseFolder, fmt.Sprintf("musicscore*/%03d/*_%s.txt", *songID, *difficulty)))
	if err != nil {
		log.Fatal(err)
	}

	if len(pathResults) < 1 {
		log.Fatal("not found")
	}

	fmt.Println("path:", pathResults[0])

	text, err := os.ReadFile(pathResults[0])
	if err != nil {
		log.Fatal(err)
	}

	chart := Parse(string(text))
	config := NewVTEGenerateConfig()
	rawEvents := GenerateTouchEvent(config, chart)
	vEvents := preprocess(func(x, y float64) (int, int) {
		return 1080 - int(math.Round((540-844)*y+844)), int(math.Round((1800-540)*x + 540))
	}, rawEvents)

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
