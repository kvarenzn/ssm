// Copyright (C) 2024, 2025 kvarenzn
// SPDX-License-Identifier: GPL-3.0-or-later

package adb

import (
	"bytes"
	"encoding/binary"
	"io"
)

type syncConnection connection

type SyncCommand string

const (
	SCList SyncCommand = "LIST"
	SCRecv SyncCommand = "RECV"
	SCSend SyncCommand = "SEND"
	SCData SyncCommand = "DATA"
	SCStat SyncCommand = "STAT"
	SCDone SyncCommand = "DONE"
)

func (s *syncConnection) Send(cmd SyncCommand, data []byte) error {
	msg := bytes.NewBufferString(string(cmd))
	if err := binary.Write(msg, binary.LittleEndian, int32(len(data))); err != nil {
		return err
	}

	m, err := msg.Write(data)
	if err != nil {
		return err
	}

	if m != len(data) {
		panic("?")
	}

	return send(s.conn, msg.Bytes())
}

const MaxChunkSize = 64 * 1024

func (s *syncConnection) Bulk(r io.Reader) error {
	buf := make([]byte, MaxChunkSize)
	for {
		n, err := r.Read(buf)
		if err == io.EOF {
			return nil
		} else if err != nil {
			return err
		}

		if n == 0 {
			panic("?")
		}

		if err := s.Send(SCData, buf[:n]); err != nil {
			return err
		}
	}
}

func (s *syncConnection) Done(status uint32) error {
	msg := bytes.NewBuffer([]byte(SCDone))
	if err := binary.Write(msg, binary.LittleEndian, status); err != nil {
		return err
	}

	return send(s.conn, msg.Bytes())
}

func (s *syncConnection) GetResponse() error {
	status, err := read(s.conn, 4)
	if err != nil {
		return err
	}

	var n uint32
	if err := binary.Read(s.conn, binary.LittleEndian, &n); err != nil {
		return err
	}

	msg, err := read(s.conn, int(n))
	if err != nil {
		return err
	}

	switch string(status) {
	case "OKAY":
		return nil
	case "FAIL":
		return &ADBError{
			Message: string(msg),
		}
	default:
		panic("?")
	}
}
