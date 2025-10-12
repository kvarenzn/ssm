// Copyright (C) 2024, 2025 kvarenzn
// SPDX-License-Identifier: GPL-3.0-or-later

package adb

import (
	"errors"
	"fmt"
	"io"
	"net"
	"time"
)

var ConnectionBroken = errors.New("connection broken")

type ADBError struct {
	Message string
}

func (e *ADBError) Error() string {
	return e.Message
}

type connection struct {
	conn    net.Conn
	timeout time.Duration
}

func NewConnection(addr string, timeout time.Duration) (*connection, error) {
	if conn, err := net.Dial("tcp", addr); err != nil {
		return nil, err
	} else {
		return &connection{
			conn:    conn,
			timeout: timeout,
		}, nil
	}
}

func send(w io.Writer, data []byte) error {
	length := len(data)
	sent := 0
	for sent < length {
		if l, err := w.Write(data[sent:]); err != nil {
			return err
		} else if l == 0 {
			return ConnectionBroken
		} else {
			sent += l
		}
	}

	return nil
}

func read(r io.Reader, size int) ([]byte, error) {
	data := make([]byte, size)
	if _, err := io.ReadFull(r, data); err != nil {
		return nil, err
	}

	return data, nil
}

func (c *connection) Send(what string) error {
	bytes := []byte(what)
	if err := send(c.conn, fmt.Appendf(nil, "%04x", len(bytes))); err != nil {
		return err
	}

	return send(c.conn, bytes)
}

func (c *connection) OpenSync() (*syncConnection, error) {
	if err := c.Run("sync:"); err != nil {
		return nil, err
	}

	return (*syncConnection)(c), nil
}

func digitToNumber(d byte) int {
	if d >= '0' && d <= '9' {
		return int(d - '0')
	} else if d >= 'a' && d <= 'f' {
		return int(d - 'a' + 10)
	} else if d >= 'A' && d <= 'F' {
		return int(d - 'A' + 10)
	} else {
		return 0
	}
}

func (c *connection) Read(size int) ([]byte, error) {
	return read(c.conn, size)
}

func (c *connection) ReadBytes() ([]byte, error) {
	size := 0

	if buf, err := read(c.conn, 4); err != nil {
		return nil, err
	} else {
		for i := range 4 {
			size <<= 4
			size += digitToNumber(buf[i])
		}
	}

	if size <= 0 {
		return nil, nil
	}

	return read(c.conn, size)
}

func (c *connection) ReadString() (string, error) {
	data, err := c.ReadBytes()
	return string(data), err
}

func (c *connection) Run(what string) error {
	if err := c.Send(what); err != nil {
		return err
	}

	resp, err := c.Read(4)
	if err != nil {
		return err
	}

	if string(resp) == "OKAY" {
		return nil
	}

	msg, err := c.ReadString()
	if err != nil {
		return err
	}

	return &ADBError{
		Message: msg,
	}
}

func (c *connection) Query(what string) (string, error) {
	if err := c.Run(what); err != nil {
		return "", err
	}

	return c.ReadString()
}

func (c *connection) Close() error {
	if c.conn == nil {
		return nil
	}

	return c.conn.Close()
}
