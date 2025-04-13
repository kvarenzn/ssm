package adb

import (
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

type client struct {
	addr string
}

func NewClient(host string, port int) *client {
	return &client{
		addr: net.JoinHostPort(host, fmt.Sprint(port)),
	}
}

func NewDefaultClient() *client {
	return NewClient(ADBDefaultServerHost, ADBDefaultServerPort)
}

func (c *client) Connect() (*connection, error) {
	return NewConnection(c.addr, ADBDefaultTimeout)
}

func (c *client) Ping() error {
	conn, err := NewConnection(c.addr, ADBDefaultTimeout)
	if err != nil {
		return err
	}

	conn.Close()
	return nil
}

func (c *client) Run(format string, a ...any) error {
	conn, err := c.Connect()
	if err != nil {
		return err
	}

	defer conn.Close()
	return conn.Run(fmt.Sprintf(format, a...))
}

func (c *client) Query(format string, a ...any) (string, error) {
	conn, err := c.Connect()
	if err != nil {
		return "", err
	}

	defer conn.Close()
	return conn.Query(fmt.Sprintf(format, a...))
}

func (c *client) QueryBytes(format string, a ...any) ([]byte, error) {
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

func (c *client) Open(format string, a ...any) (*connection, error) {
	conn, err := c.Connect()
	if err != nil {
		return nil, err
	}

	return conn, conn.Run(fmt.Sprintf(format, a...))
}

func (c *client) KillServer() error {
	conn, err := c.Connect()
	if err != nil {
		return err
	}

	defer conn.Close()
	return conn.Send("host:kill")
}

func (c *client) Serials() ([]string, error) {
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

func (c *client) Devices() ([]*device, error) {
	resp, err := c.Query("host:devices-l")
	if err != nil {
		return nil, err
	}

	devices := []*device{}

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

		devices = append(devices, &device{
			client: c,
			serial: fields[0],
			attrs:  attrs,
		})
	}

	return devices, nil
}

func (c *client) ListForward(reverse bool) ([]Forward, error) {
	prefix := "host"
	if reverse {
		prefix = "reverse"
	}

	resp, err := c.Query("%s:list-forward", prefix)
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

func (c *client) KillForward(local string, reverse bool) error {
	prefix := "host"
	if reverse {
		prefix = "reverse"
	}

	return c.Run("%s:killforward:%s", prefix, local)
}

func (c *client) KillForwardAll(reverse bool) error {
	prefix := "host"
	if reverse {
		prefix = "reverse"
	}

	return c.Run("%s:killforward-all", prefix)
}
