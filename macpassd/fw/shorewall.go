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

type maclistEntry struct {
	Disposition string
	Interface   string
	Mac         string
	IPAddr      string
	Comment     string
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

	entries, err := parseMAC4File()
	if err != nil {
		slog.With("err", err).Error("Unable to parse MACFile")
		return
	}

	entries = append(entries, maclistEntry{
		Disposition: "ACCEPT",
		Interface:   s.conf.Firewall.ShorewallIF,
		Mac:         r.Mac,
		IPAddr:      "",
		Comment:     fmt.Sprintf("#%s", r.User),
	})

	err = writeMAC4File(entries)
	if err != nil {
		slog.With("err", err).Error("Unable to write MACFile")
		return
	}

	// IPv6

	entries, err = parseMAC6File()
	if err != nil {
		slog.With("err", err).Error("Unable to parse MAC6File")
		return
	}

	entries = append(entries, maclistEntry{
		Disposition: "ACCEPT",
		Interface:   s.conf.Firewall.ShorewallIF,
		Mac:         r.Mac,
		IPAddr:      "",
		Comment:     fmt.Sprintf("#%s", r.User),
	})

	err = writeMAC6File(entries)
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
		if macs[i].Mac == r.Mac {
			macs = remove(macs, i)
			break
		}
	}

	err = writeMAC4File(macs)
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
		if macs[i].Mac == r.Mac {
			macs = remove(macs, i)
			break
		}
	}

	err = writeMAC6File(macs)
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

func parseMAC4File() ([]maclistEntry, error) {
	return parseMACFile(MACLIST4_PATH)
}

func parseMAC6File() ([]maclistEntry, error) {
	return parseMACFile(MACLIST6_PATH)
}

func parseMACFile(path string) ([]maclistEntry, error) {
	buff, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	buffs := string(buff)

	var entries []maclistEntry

	for _, l := range strings.Split(buffs, "\n") {
		fields := strings.Fields(l)
		if len(fields) == 0 || fields[0][0] == '#' {
			continue
		}

		if len(fields) < 3 {
			return nil, errors.New(fmt.Sprintf("Parsing MACFile. Line have less than 3 fields. Line: %s", l))
		}

		tmp := maclistEntry{
			Disposition: fields[0],
			Interface:   fields[1],
			Mac:         fields[2],
		}

		// I don't like this
		if len(fields) > 3 {
			if fields[3][0] == '#' {
				tmp.Comment = fields[3]
			} else {
				tmp.IPAddr = fields[3]
				if len(fields) > 4 {
					tmp.Comment = fields[4]
				}
			}
		}

		entries = append(entries, tmp)
	}

	return entries, nil
}

func writeMAC4File(entries []maclistEntry) error {
	return writeMACFile(MACLIST4_PATH, entries)
}

func writeMAC6File(entries []maclistEntry) error {
	return writeMACFile(MACLIST6_PATH, entries)
}

// Ethernet Interface as input
func writeMACFile(path string, entries []maclistEntry) error {
	var buff bytes.Buffer

	buff.Write([]byte("#DISPOSITION\tINTERFACE\tMAC\tIP\tCOMMENT\n"))
	for i := range entries {
		buff.Write(fmt.Appendf([]byte{}, "%s\t%s\t%s\t%s\t%s\n", entries[i].Disposition,
			entries[i].Interface, entries[i].Mac, entries[i].IPAddr, entries[i].Comment))
	}

	return os.WriteFile(path, buff.Bytes(), 0644)
}

func reload() {
	_, stderr, err := shellExec("shorewall reload")
	if err != nil {
		slog.With("err", err, "stderr", stderr).Error("Can't reload shorewall")
	}

	_, stderr, err = shellExec("shorewall -6 reload")
	if err != nil {
		slog.With("err", err, "stderr", stderr).Error("Can't reload shorewall6")
	}
}

func remove(s []maclistEntry, i int) []maclistEntry {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}
