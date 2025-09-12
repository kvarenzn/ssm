// Copyright (C) 2024, 2025 kvarenzn
// SPDX-License-Identifier: GPL-3.0-or-later

package adb

import (
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"
	"time"
)

type Device struct {
	client *Client
	serial string
	attrs  map[string]string
}

func (d *Device) String() string {
	return fmt.Sprintf("Device(%s, %s)", d.serial, d.attrs)
}

var NoAttributeError = errors.New("no such attribute")

func (d *Device) getAttribute(attr string) (string, error) {
	if v, ok := d.attrs[attr]; ok {
		return v, nil
	} else {
		return "", NoAttributeError
	}
}

func (d *Device) Product() (string, error) {
	return d.getAttribute("product")
}

func (d *Device) Model() (string, error) {
	return d.getAttribute("model")
}

func (d *Device) USB() (string, error) {
	return d.getAttribute("usb")
}

func (d *Device) IsUSB() (bool, error) {
	if usb, err := d.USB(); err != nil {
		return false, err
	} else {
		return usb != "", nil
	}
}

func (d *Device) TransportID() (string, error) {
	return d.getAttribute("transport_id")
}

func (d *Device) Serial() string {
	return d.serial
}

func (d *Device) State() (string, error) {
	return d.client.Query("host-serial:%s:get-state", d.serial)
}

func (d *Device) DevPath() (string, error) {
	return d.client.Query("host-serial:%s:get-devpath", d.serial)
}

func (d *Device) Client() *Client {
	return d.client
}

func (d *Device) Authorized() bool {
	if state, err := d.State(); err == nil && state == "device" {
		return true
	}

	return false
}

type AndroidSocketNamespace string

const (
	NSNone     AndroidSocketNamespace = ""
	NSReserved AndroidSocketNamespace = "reserved"
	NSAbstract AndroidSocketNamespace = "abstract"
	FileSystem AndroidSocketNamespace = "filesystem"
)

func (d *Device) Open() (*connection, error) {
	return d.client.Open("host:transport:%s", d.serial)
}

func (d *Device) Forward(local, remote string, reverse, norebind bool) error {
	if reverse {
		conn, err := d.Open()
		if err != nil {
			return err
		}

		cmd := "reverse:forward"
		if norebind {
			cmd += ":norebind"
		}

		defer conn.Close()
		return conn.Run(fmt.Sprintf("%s:%s;%s", cmd, local, remote))
	}

	cmd := fmt.Sprintf("host-serial:%s:forward", d.serial)
	if norebind {
		cmd += ":norebind"
	}

	return d.client.Run("%s:%s;%s", cmd, local, remote)
}

func (d *Device) ListForward(reverse bool) ([]Forward, error) {
	forwards, err := d.client.ListForward(reverse)
	if err != nil {
		return nil, err
	}

	return slices.DeleteFunc(forwards, func(f Forward) bool {
		return f.Serial != d.serial
	}), nil
}

func (d *Device) KillForward(local string) error {
	return d.client.Run("host-serial:%s:killforward:%s", d.serial, local)
}

var EmptyCommandError = errors.New("shell command cannot be empty")

func (d *Device) RawSh(prog string, args ...string) ([]byte, error) {
	cmd := prog
	if len(args) > 0 {
		cmd += " " + strings.Join(args, " ")
	}

	if len(strings.TrimSpace(cmd)) == 0 {
		return nil, EmptyCommandError
	}

	conn, err := d.Open()
	if err != nil {
		return nil, err
	}

	defer conn.Close()

	if err := conn.Run(fmt.Sprintf("shell:%s", cmd)); err != nil {
		return nil, err
	}

	return io.ReadAll(conn.conn)
}

func (d *Device) Sh(prog string, args ...string) (string, error) {
	bytes, err := d.RawSh(prog, args...)
	return string(bytes), err
}

func (d *Device) push(data io.Reader, remote string, modTime time.Time, fileMode os.FileMode) error {
	conn, err := d.Open()
	if err != nil {
		return err
	}

	defer conn.Close()

	s, err := conn.OpenSync()
	if err != nil {
		return err
	}

	if err := s.Send(SCSend, fmt.Appendf(nil, "%s,%d", remote, fileMode)); err != nil {
		return err
	}

	if err := s.Bulk(data); err != nil {
		return err
	}

	if err := s.Done(uint32(modTime.Unix())); err != nil {
		return err
	}

	return s.GetResponse()
}

func (d *Device) Push(local *os.File, remote string) error {
	stat, err := local.Stat()
	if err != nil {
		return err
	}

	return d.push(local, remote, stat.ModTime(), os.FileMode(0o644))
}
