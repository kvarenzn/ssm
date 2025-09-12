// Copyright (C) 2024, 2025 kvarenzn
// SPDX-License-Identifier: GPL-3.0-or-later

package term

import (
	"os"
	"time"
)

type KeyCode int16

const KEY_UNKNOWN KeyCode = -1
const (
	KEY_NUL   KeyCode = iota // Null Character, CTRL-@
	KEY_SOH                  // Start of Heading, CTRL-A
	KEY_STX                  // Start of Text, CTRL-B
	KEY_ETX                  // End of Text, CTRL-C
	KEY_EOT                  // End of Transmission, CTRL-D
	KEY_ENQ                  // Enquiry, CTRL-E
	KEY_ACK                  // Acknowledge, CTRL-F
	KEY_BEL                  // Bell, CTRL-G
	KEY_BS                   // Backspace, CTRL-H
	KEY_TAB                  // Horizontal Tab, CTRL-I
	KEY_LF                   // Line Feed, CTRL-J
	KEY_VT                   // Vertical Tab, CTRL-K
	KEY_FF                   // Form Feed, CTRL-L
	KEY_ENTER                // Carriage Return, CTRL-M
	KEY_SO                   // Shift Out, CTRL-N
	KEY_SI                   // Shift In, CTRL-O
	KEY_DLE                  // Data Link Escape, CTRL-P
	KEY_DC1                  // Device Control 1, CTRL-Q
	KEY_DC2                  // Device Control 2, CTRL-R
	KEY_DC3                  // Device Control 3, CTRL-S
	KEY_DC4                  // Device Control 4, CTRL-T
	KEY_NAK                  // Negative Acknowledge, CTRL-U
	KEY_SYN                  // Synchronous Idle, CTRL-V
	KEY_ETB                  // End of Transmission Block, CTRL-W
	KEY_CAN                  // Cancel, CTRL-X
	KEY_EM                   // End of medium, CTRL-Y
	KEY_SUB                  // Substitute, CTRL-Z
	KEY_ESC                  // CTRL-[
	KEY_FS                   // File Separator, CTRL-/
	KEY_GS                   // Group Separator, CTRL-]
	KEY_RS                   // Record Separator
	KEY_US                   // Unit Separator
	KEY_SPACE
	KEY_BANG
	KEY_DQUOTE
	KEY_SHARP
	KEY_DOLLAR
	KEY_PERCENT
	KEY_AND
	KEY_SQUOTE
	KEY_LPAREN
	KEY_RPAREN
	KEY_STAR
	KEY_PLUS
	KEY_COMMA
	KEY_DASH
	KEY_DOT
	KEY_SLASH
	KEY_0
	KEY_1
	KEY_2
	KEY_3
	KEY_4
	KEY_5
	KEY_6
	KEY_7
	KEY_8
	KEY_9
	KEY_COLON
	KEY_SEMI
	KEY_LT
	KEY_EQUAL
	KEY_GT
	KEY_QUEST
	KEY_AT
	KEY_A
	KEY_B
	KEY_C
	KEY_D
	KEY_E
	KEY_F
	KEY_G
	KEY_H
	KEY_I
	KEY_J
	KEY_K
	KEY_L
	KEY_M
	KEY_N
	KEY_O
	KEY_P
	KEY_Q
	KEY_R
	KEY_S
	KEY_T
	KEY_U
	KEY_V
	KEY_W
	KEY_X
	KEY_Y
	KEY_Z
	KEY_LSQ
	KEY_BSLASH
	KEY_RSQ
	KEY_HAT
	KEY_UNDERSCORE
	KEY_BQUOTE
	KEY_a
	KEY_b
	KEY_c
	KEY_d
	KEY_e
	KEY_f
	KEY_g
	KEY_h
	KEY_i
	KEY_j
	KEY_k
	KEY_l
	KEY_m
	KEY_n
	KEY_o
	KEY_p
	KEY_q
	KEY_r
	KEY_s
	KEY_t
	KEY_u
	KEY_v
	KEY_w
	KEY_x
	KEY_y
	KEY_z
	KEY_LBRACE
	KEY_BAR
	KEY_RBRACE
	KEY_TILDE
	KEY_BACKSPACE // CTRL-?

	KEY_F1  // ESC O P
	KEY_F2  // ESC O Q
	KEY_F3  // ESC O R
	KEY_F4  // ESC O S
	KEY_F5  // ESC [ 1 5 ~
	KEY_F6  // ESC [ 1 7 ~
	KEY_F7  // ESC [ 1 8 ~
	KEY_F8  // ESC [ 1 9 ~
	KEY_F9  // ESC [ 2 0 ~
	KEY_F10 // ESC [ 2 1 ~
	KEY_F11 // ESC [ 2 3 ~
	KEY_F12 // ESC [ 2 4 ~
	KEY_F13 // SHIFT-F1, ESC [ 1 ; 2 P
	KEY_F14 // SHIFT-F2, ESC [ 1 ; 2 Q
	KEY_F15 // SHIFT-F3, ESC [ 1 3 ; 2 ~
	KEY_F16 // SHIFT-F4, ESC [ 1 ; 2 S
	KEY_F17 // SHIFT-F5, ESC [ 1 5 ; 2 ~
	KEY_F18 // SHIFT-F6, ESC [ 1 7 ; 2 ~
	KEY_F19 // SHIFT-F7, ESC [ 1 8 ; 2 ~
	KEY_F20 // SHIFT-F8, ESC [ 1 9 ; 2 ~
	KEY_F21 // SHIFT-F9, ESC [ 2 0 ; 2 ~
	KEY_F22 // SHIFT-F10, ESC [ 2 1 ; 2 ~
	KEY_F23 // SHIFT-F11, ESC [ 2 3 ; 2 ~
	KEY_F24 // SHIFT-F12, ESC [ 2 4 ; 2 ~
	KEY_F25 // CTRL-F1, ESC [ 1 ; 5 P
	KEY_F26 // CTRL-F2, ESC [ 1 ; 5 Q
	KEY_F27 // CTRL-F3, ESC [ 1 ; 5 R
	KEY_F28 // CTRL-F4, ESC [ 1 ; 5 S
	KEY_F29 // CTRL-F5, ESC [ 1 5 ; 5 ~
	KEY_F30 // CTRL-F6, ESC [ 1 7 ; 5 ~
	KEY_F31 // CTRL-F7, ESC [ 1 8 ; 5 ~
	KEY_F32 // CTRL-F8, ESC [ 1 9 ; 5 ~
	KEY_F33 // CTRL-F9, ESC [ 2 0 ; 5 ~
	KEY_F34 // CTRL-F10, ESC [ 2 1 ; 5 ~
	KEY_F35 // CTRL-F11, ESC [ 2 3 ; 5 ~
	KEY_F36 // CTRL-F12, ESC [ 2 4 ; 5 ~
	KEY_F49 // ALT-F1, ESC [ 1 ; 3 P
	KEY_F50 // ALT-F2, ESC [ 1 ; 3 Q
	KEY_F51 // ALT-F3, ESC [ 1 3 ; 3 ~
	KEY_F52 // ALT-F4, ESC [ 1 ; 3 S
	KEY_F53 // ALT-F5, ESC [ 1 5 ; 3 ~
	KEY_F54 // ALT-F6, ESC [ 1 7 ; 3 ~
	KEY_F55 // ALT-F7, ESC [ 1 8 ; 3 ~
	KEY_F56 // ALT-F8, ESC [ 1 9 ; 3 ~
	KEY_F57 // ALT-F9, ESC [ 2 0 ; 3 ~
	KEY_F58 // ALT-F10, ESC [ 2 1 ; 3 ~
	KEY_F59 // ALT-F11, ESC [ 2 3 ; 3 ~
	KEY_F60 // ALT-F12, ESC [ 2 4 ; 3 ~

	KEY_UP          // ESC [ A
	KEY_DOWN        // ESC [ B
	KEY_RIGHT       // ESC [ C
	KEY_LEFT        // ESC [ D
	KEY_HOME        // ESC [ H
	KEY_DEL         // ESC [ P
	KEY_INS         // ESC [ 4 h
	KEY_END         // ESC [ 4 ~
	KEY_PGUP        // ESC [ 5 ~
	KEY_PGDN        // ESC [ 6 ~
	KEY_SHIFT_UP    // ESC [ 1 ; 2 A
	KEY_SHIFT_DOWN  // ESC [ 1 ; 2 B
	KEY_SHIFT_RIGHT // ESC [ 1 ; 2 C
	KEY_SHIFT_LEFT  // ESC [ 1 ; 2 D
	KEY_CTRL_UP     // ESC [ 1 ; 5 A
	KEY_CTRL_DOWN   // ESC [ 1 ; 5 B
	KEY_CTRL_RIGHT  // ESC [ 1 ; 5 C
	KEY_CTRL_LEFT   // ESC [ 1 ; 5 D
)

const (
	SS2 = "N"
	SS3 = "O"
	CSI = "["
)

var keySequences = map[string]KeyCode{
	SS3 + "P":     KEY_F1,
	SS3 + "Q":     KEY_F2,
	SS3 + "R":     KEY_F3,
	SS3 + "S":     KEY_F4,
	CSI + "15~":   KEY_F5,
	CSI + "17~":   KEY_F6,
	CSI + "18~":   KEY_F7,
	CSI + "19~":   KEY_F8,
	CSI + "20~":   KEY_F9,
	CSI + "21~":   KEY_F10,
	CSI + "23~":   KEY_F11,
	CSI + "24~":   KEY_F12,
	CSI + "1;2P":  KEY_F13,
	CSI + "1;2Q":  KEY_F14,
	CSI + "13;2~": KEY_F15,
	CSI + "1;2S":  KEY_F16,
	CSI + "15;2~": KEY_F17,
	CSI + "17;2~": KEY_F18,
	CSI + "18;2~": KEY_F19,
	CSI + "19;2~": KEY_F20,
	CSI + "20;2~": KEY_F21,
	CSI + "21;2~": KEY_F22,
	CSI + "23;2~": KEY_F23,
	CSI + "24;2~": KEY_F24,
	CSI + "1;5P":  KEY_F25,
	CSI + "1;5Q":  KEY_F26,
	CSI + "1;5R":  KEY_F27,
	CSI + "1;5S":  KEY_F28,
	CSI + "15;5~": KEY_F29,
	CSI + "17;5~": KEY_F30,
	CSI + "18;5~": KEY_F31,
	CSI + "19;5~": KEY_F32,
	CSI + "20;5~": KEY_F33,
	CSI + "21;5~": KEY_F34,
	CSI + "23;5~": KEY_F35,
	CSI + "24;5~": KEY_F36,
	CSI + "1;3P":  KEY_F49,
	CSI + "1;3Q":  KEY_F50,
	CSI + "13;3~": KEY_F51,
	CSI + "1;3S":  KEY_F52,
	CSI + "15;3~": KEY_F53,
	CSI + "17;3~": KEY_F54,
	CSI + "18;3~": KEY_F55,
	CSI + "19;3~": KEY_F56,
	CSI + "20;3~": KEY_F57,
	CSI + "21;3~": KEY_F58,
	CSI + "23;3~": KEY_F59,
	CSI + "24;3~": KEY_F60,
	CSI + "A":     KEY_UP,
	CSI + "B":     KEY_DOWN,
	CSI + "C":     KEY_RIGHT,
	CSI + "D":     KEY_LEFT,
	CSI + "F":     KEY_END,
	CSI + "H":     KEY_HOME,
	CSI + "P":     KEY_DEL,
	CSI + "4h":    KEY_INS,
	CSI + "4~":    KEY_END,
	CSI + "5~":    KEY_PGUP,
	CSI + "6~":    KEY_PGDN,
	CSI + "1;2A":  KEY_SHIFT_UP,
	CSI + "1;2B":  KEY_SHIFT_DOWN,
	CSI + "1;2C":  KEY_SHIFT_RIGHT,
	CSI + "1;2D":  KEY_SHIFT_LEFT,
	CSI + "1;5A":  KEY_CTRL_UP,
	CSI + "1;5B":  KEY_CTRL_DOWN,
	CSI + "1;5C":  KEY_CTRL_RIGHT,
	CSI + "1;5D":  KEY_CTRL_LEFT,
}

var oneByte = make([]byte, 1)

func ReadKey(reader *os.File, timeout time.Duration) (KeyCode, error) {
	if _, err := reader.Read(oneByte); err != nil {
		return KEY_UNKNOWN, err
	}

	b := oneByte[0]

	if b != 0x1b {
		if b > 0x7f {
			return KEY_UNKNOWN, nil
		}

		return KeyCode(b), nil
	}

	if err := readByteWithTimeout(reader, timeout); err != nil {
		if os.IsTimeout(err) {
			return KEY_ESC, nil
		}

		return KEY_UNKNOWN, err
	}

	b = oneByte[0]
	keys := []byte{b}

	if b != CSI[0] && b != SS2[0] && b != SS3[0] {
		return KEY_ESC, nil
	}

	for {
		if err := readByteWithTimeout(reader, timeout); err != nil {
			if os.IsTimeout(err) {
				return KEY_ESC, nil
			}

			return KEY_UNKNOWN, err
		}

		b = oneByte[0]
		keys = append(keys, b)

		if (b >= 'A' && b <= 'Z') || (b >= 'a' && b <= 'z') || b == '~' || b == ']' {
			break
		}
	}

	v, ok := keySequences[string(keys)]
	if !ok {
		return KEY_UNKNOWN, nil
	}

	return v, nil
}
