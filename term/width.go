package term

import (
	"regexp"

	"golang.org/x/text/width"
)

func WcWidth(r rune) int {
	switch {
	case r == '\t':
		return 4 // tab width is 4
	case r == '\n' || r == '\r':
		return 0 // carriage return is 0
	case r < 127:
		return 1 // ASCII characters are 1
	default:
		prop := width.LookupRune(r)
		if prop.Kind() == width.EastAsianFullwidth || prop.Kind() == width.EastAsianWide {
			return 2
		}
		return 1
	}
}

var commonANSIPattern = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

func Width(text string) int {
	text = commonANSIPattern.ReplaceAllString(text, "")

	s := 0
	for _, r := range text {
		s += WcWidth(r)
	}
	return s
}
