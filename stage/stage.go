package stage

type screen struct {
	Width, Height float64
}

func NewScreen(width, height int) *screen {
	return &screen{float64(width), float64(height)}
}

func (s *screen) isWideScreen() bool {
	return 9*s.Width > 16*s.Height
}

func (s *screen) limitedRatio() float64 {
	return min(s.Width/s.Height, 2)
}

func (s *screen) laneYOffset() float64 {
	if s.isWideScreen() {
		return s.Height * (115 - 9*s.limitedRatio()) / 110
	} else {
		return 9 * s.Width / 16
	}
}

func (s *screen) laneWidth() float64 {
	if s.isWideScreen() {
		return s.Height * (s.limitedRatio()/4 + 2.0/3)
	} else {
		return s.Width * 9 / 13
	}
}

func (s *screen) Y() float64 {
	return s.Height/2 + 26*s.laneYOffset()/81
}

func (s *screen) X() (float64, float64) {
	half := s.laneWidth() / 2
	middle := s.Width / 2
	return middle - half, middle + half
}
