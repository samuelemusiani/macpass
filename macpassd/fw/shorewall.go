package fw

import (
	"bytes"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/musianisamuele/macpass/macpassd/config"
	"github.com/musianisamuele/macpass/macpassd/registration"
)

const (
	SHELL         = "bash"
	MACLIST4_PATH = "/etc/shorewall/maclist"
	MACLIST6_PATH = "/etc/shorewall6/maclist"
)

type Shorewall struct {
	// Lock for the shorewall mac file
	mutex sync.Mutex

	conf *config.Config
}

func (s *Shorewall) Init() {
	slog.Info("Initializing shorewall")

	s.conf = config.Get()

	stdout, stderr, err := shellExec("shorewall version")
	if err != nil {
		slog.With("err", err, "stderr", stderr).Error("Can't get shorewall version")
		os.Exit(3)
	}

	stdout6, stderr, err := shellExec("shorewall6 version")
	if err != nil {
		slog.With("err", err, "stderr", stderr).Error("Can't get shorewall version")
		os.Exit(3)
	}

	slog.With("version", stdout).Info("Detected shorewall version")
	slog.With("version6", stdout6).Info("Detected shorewall6 version")
}

func (s *Shorewall) Allow(r registration.Registration) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// IPv4

	macs, err := parseMAC4File()
	if err != nil {
		slog.With("err", err).Error("Unable to parse MACFile")
		return
	}

	macs = append(macs, r.Mac)

	err = writeMAC4File(macs, s.conf.Firewall.ShorewallIF)
	if err != nil {
		slog.With("err", err).Error("Unable to write MACFile")
		return
	}

	// IPv6

	macs, err = parseMAC6File()
	if err != nil {
		slog.With("err", err).Error("Unable to parse MAC6File")
		return
	}

	macs = append(macs, r.Mac)

	err = writeMAC6File(macs, s.conf.Firewall.ShorewallIF)
	if err != nil {
		slog.With("err", err).Error("Unable to write MAC6File")
		return
	}

	reload()
}

func (s *Shorewall) Delete(r registration.Registration) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// IPv4

	macs, err := parseMAC4File()
	if err != nil {
		slog.With("err", err).Error("Unable to parse MACFile")
		return
	}

	for i := range macs {
		if macs[i] == r.Mac {
			macs = remove(macs, i)
			break
		}
	}

	err = writeMAC4File(macs, s.conf.Firewall.ShorewallIF)
	if err != nil {
		slog.With("err", err).Error("Unable to write MACFile")
		return
	}

	// IPv6

	macs, err = parseMAC6File()
	if err != nil {
		slog.With("err", err).Error("Unable to parse MACFile")
		return
	}

	for i := range macs {
		if macs[i] == r.Mac {
			macs = remove(macs, i)
			break
		}
	}

	err = writeMAC6File(macs, s.conf.Firewall.ShorewallIF)
	if err != nil {
		slog.With("err", err).Error("Unable to write MACFile")
		return
	}

	reload()
}

func shellExec(command string) (string, string, error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := exec.Command(SHELL, "-c", command)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

func parseMAC4File() ([]string, error) {
	return parseMACFile(MACLIST4_PATH)
}

func parseMAC6File() ([]string, error) {
	return parseMACFile(MACLIST6_PATH)
}

func parseMACFile(path string) ([]string, error) {
	buff, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	buffs := string(buff)

	var macs []string

	for _, l := range strings.Split(buffs, "\n") {
		fields := strings.Fields(l)
		if len(fields) == 0 || fields[0][0] == '#' {
			continue
		}

		if len(fields) < 3 {
			return nil, errors.New(fmt.Sprintf("Parsing MACFile. Line have less than 3 fields. Line: %s", l))
		}

		macs = append(macs, fields[2])
	}

	return macs, nil
}

func writeMAC4File(macs []string, inter string) error {
	return writeMACFile(MACLIST4_PATH, macs, inter)
}

func writeMAC6File(macs []string, inter string) error {
	return writeMACFile(MACLIST6_PATH, macs, inter)
}

// Ethernet Interface as input
func writeMACFile(path string, macs []string, inter string) error {
	var buff bytes.Buffer

	buff.Write([]byte("#DISPOSITION\tINTERFACE\tMAC\n"))
	for i := range macs {
		buff.Write([]byte(fmt.Sprintf("ACCEPT\t%s\t%s\n", inter, macs[i])))
	}

	return os.WriteFile(path, buff.Bytes(), 0644)
}

func reload() {
	_, stderr, err := shellExec("shorewall reload")
	if err != nil {
		slog.With("err", err, "stderr", stderr).Error("Can't reload shorewall")
	}
}

func remove(s []string, i int) []string {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}
