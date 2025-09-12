// Copyright (C) 2024, 2025 kvarenzn
// SPDX-License-Identifier: GPL-3.0-or-later

package adb

import (
	"encoding/binary"
	"fmt"
	"net"
	"strings"
	"time"
)

const (
	ADBDefaultServerHost = "localhost"
	ADBDefaultServerPort = 5037
	ADBDefaultDaemonPort = 5555
	ADBDefaultTimeout    = 1 * time.Minute
)

type Client struct {
	addr string
}

func NewClient(host string, port int) *Client {
	return &Client{
		addr: net.JoinHostPort(host, fmt.Sprint(port)),
	}
}

func NewDefaultClient() *Client {
	return NewClient(ADBDefaultServerHost, ADBDefaultServerPort)
}

func (c *Client) Connect() (*connection, error) {
	return NewConnection(c.addr, ADBDefaultTimeout)
}

func (c *Client) Ping() error {
	conn, err := NewConnection(c.addr, ADBDefaultTimeout)
	if err != nil {
		return err
	}

	conn.Close()
	return nil
}

func (c *Client) Run(format string, a ...any) error {
	conn, err := c.Connect()
	if err != nil {
		return err
	}

	defer conn.Close()
	return conn.Run(fmt.Sprintf(format, a...))
}

func (c *Client) Query(format string, a ...any) (string, error) {
	conn, err := c.Connect()
	if err != nil {
		return "", err
	}

	defer conn.Close()
	return conn.Query(fmt.Sprintf(format, a...))
}

func (c *Client) QueryBytes(format string, a ...any) ([]byte, error) {
	conn, err := c.Connect()
	if err != nil {
		return nil, err
	}

	defer conn.Close()
	if err := conn.Run(fmt.Sprintf(format, a...)); err != nil {
		return nil, err
	}

	return conn.ReadBytes()
}

func (c *Client) Open(format string, a ...any) (*connection, error) {
	conn, err := c.Connect()
	if err != nil {
		return nil, err
	}

	return conn, conn.Run(fmt.Sprintf(format, a...))
}

func (c *Client) KillServer() error {
	conn, err := c.Connect()
	if err != nil {
		return err
	}

	defer conn.Close()
	return conn.Send("host:kill")
}

func (c *Client) Serials() ([]string, error) {
	res, err := c.Query("host:devices")
	if err != nil {
		return nil, err
	}

	devices := []string{}
	for l := range strings.Lines(res) {
		fields := strings.Fields(l)
		if len(fields) < 2 {
			continue
		}

		devices = append(devices, fields[0])
	}

	return devices, nil
}

func (c *Client) Devices() ([]*Device, error) {
	resp, err := c.Query("host:devices-l")
	if err != nil {
		return nil, err
	}

	devices := []*Device{}

	for l := range strings.Lines(resp) {
		l = strings.TrimSpace(l)
		if len(l) == 0 {
			continue
		}

		fields := strings.Fields(l)
		if len(fields) < 4 || len(fields[0]) == 0 {
			continue
		}

		attrs := map[string]string{}
		for _, f := range fields[2:] {
			kv := strings.Split(f, ":")
			if len(kv) < 2 {
				continue
			}

			attrs[kv[0]] = kv[1]
		}

		devices = append(devices, &Device{
			client: c,
			serial: fields[0],
			attrs:  attrs,
		})
	}

	return devices, nil
}

func (c *Client) ListForward(reverse bool) ([]Forward, error) {
	var resp string
	var err error
	if reverse {
		conn, err := c.Open("host:tport:any")
		if err != nil {
			return nil, err
		}
		buf := make([]byte, 8)
		conn.conn.Read(buf)
		binary.LittleEndian.Uint64(buf)
		resp, err = conn.Query("reverse:list-forward")
		conn.Close()
	} else {
		resp, err = c.Query("host:list-forward")
	}

	if err != nil {
		return nil, err
	}

	forwards := []Forward{}

	for l := range strings.Lines(resp) {
		l = strings.TrimSpace(l)
		if len(l) == 0 {
			continue
		}

		fields := strings.Fields(l)
		forwards = append(forwards, Forward{
			Serial: fields[0],
			Local:  fields[1],
			Remote: fields[2],
		})
	}

	return forwards, nil
}

func (c *Client) KillForward(local string, reverse bool) error {
	if reverse {
		conn, err := c.Open("host:tport:any")
		if err != nil {
			return err
		}
		defer conn.Close()
		buf := make([]byte, 8)
		conn.conn.Read(buf)
		binary.LittleEndian.Uint64(buf)
		return conn.Run(fmt.Sprintf("reverse:killforward:%s", local))
	} else {
		return c.Run("host:killforward:%s", local)
	}
}

func (c *Client) KillForwardAll(reverse bool) error {
	if reverse {
		conn, err := c.Open("host:tport:any")
		if err != nil {
			return err
		}
		defer conn.Close()
		buf := make([]byte, 8)
		conn.conn.Read(buf)
		return conn.Run("reverse:killforward-all")
	} else {
		return c.Run("host:killforward-all")
	}
}

func FirstAuthorizedDevice(devices []*Device) *Device {
	for _, d := range devices {
		if d.Authorized() {
			return d
		}
	}

	return nil
}
