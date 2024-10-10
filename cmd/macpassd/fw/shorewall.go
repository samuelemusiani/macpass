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

	"github.com/musianisamuele/macpass/cmd/macpassd/config"
	"github.com/musianisamuele/macpass/cmd/macpassd/registration"
)

const (
	SHELL        = "bash"
	MACLIST_PATH = "/etc/shorewall/maclist"
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

	slog.With("version", stdout).Info("Detected shorewall version")
}

func (s *Shorewall) Allow(r registration.Registration) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	macs, err := parseMACFile()
	if err != nil {
		slog.With("err", err).Error("Unable to parse MACFile")
		return
	}

	macs = append(macs, r.Mac)

	err = writeMACFile(macs, s.conf.Firewall.ShorewallIF)
	if err != nil {
		slog.With("err", err).Error("Unable to write MACFile")
		return
	}

	reload()
}

func (s *Shorewall) Delete(r registration.Registration) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	macs, err := parseMACFile()
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

	err = writeMACFile(macs, s.conf.Firewall.ShorewallIF)
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

func parseMACFile() ([]string, error) {
	buff, err := os.ReadFile(MACLIST_PATH)
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

// Ethernet Interface as input
func writeMACFile(macs []string, inter string) error {
	var buff bytes.Buffer

	buff.Write([]byte("#DISPOSITION\tINTERFACE\tMAC\n"))
	for i := range macs {
		buff.Write([]byte(fmt.Sprintf("ACCEPT\t%s\t%s\n", inter, macs[i])))
	}

	return os.WriteFile(MACLIST_PATH, buff.Bytes(), 0644)
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
