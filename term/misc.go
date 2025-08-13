package term

type TermSize struct {
	Col        int // terminal width (in cell)
	Row        int // terminal height (in cell)
	Xpixel     int // terminal width (in pixel)
	Ypixel     int // terminal height (in pixel)
	CellWidth  int // font width (in pixel)
	CellHeight int // font height (in pixel)
}
