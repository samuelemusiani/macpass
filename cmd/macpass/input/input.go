package input

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"log/slog"
	"net"
	"net/mail"
	"os"
	"reflect"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/musianisamuele/macpass/cmd/macpass/config"
	"github.com/musianisamuele/macpass/cmd/macpass/db"
	"github.com/musianisamuele/macpass/pkg/macparse"
	"golang.org/x/term"
)

func Credential() (string, string) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter Email: ")
	username, err := reader.ReadString('\n')
	if err != nil {
		log.Fatal(err)
	}

	username = strings.TrimSpace(username)

	// Check if the email is valid
	_, err = mail.ParseAddress(username)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Print("Enter Password: ")
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println()

	password := string(bytePassword)
	return username, strings.TrimSpace(password)
}

func Mac(user string) string {
	// check if user has already enterd his mac in previous connections

	reader := bufio.NewReader(os.Stdin)

	if mac, isPresent := db.GetMac(user); isPresent {
		fmt.Println("In your previous connection you used this mac: ", mac)
		fmt.Print("Do you want to use it again? [Y/n]: ")

		var confirm bool

		for ok := true; ok; {
			response, err := reader.ReadString('\n')
			response = strings.ToLower(strings.TrimSpace(response))

			if err != nil {
				log.Println(err)
			}

			if response == "y" || response == "yes" || len(response) == 0 {
				ok = false
				confirm = true
			} else if response == "n" || response == "no" {
				ok = false
				confirm = false
			}
		}

		if confirm {
			return mac
		}
	}

	mac := extractMacFromSSH()
	if mac != "" {
		fmt.Println("In this connection you are using: ", mac)
		fmt.Print("Do you want to authenticate it? [Y/n]: ")

		var confirm bool

		for ok := true; ok; {
			response, err := reader.ReadString('\n')
			response = strings.ToLower(strings.TrimSpace(response))

			if err != nil {
				log.Println(err)
			}

			if response == "y" || response == "yes" || len(response) == 0 {
				ok = false
				confirm = true
			} else if response == "n" || response == "no" {
				ok = false
				confirm = false
			}
		}

		if confirm {
			return mac
		}
	}

	var macAdd string
	var err error

enterMac:

	fmt.Print("Enter a MAC address: ")
	macAdd, err = reader.ReadString('\n')
	if err != nil {
		log.Println(err)
		goto enterMac
	}

	macAdd = strings.TrimSpace(macAdd)

	if _, err := net.ParseMAC(macAdd); err != nil {
		log.Println(err)
		goto enterMac
	}

	macAdd, err = macparse.ParseMac(macAdd, "linux")
	if err != nil {
		log.Println(err)
		goto enterMac
	}

	brdAdd := "ff:ff:ff:ff:ff:ff"
	if macAdd == brdAdd {
		log.Println("The brodcast address it is NOT a valide MAC address")
		goto enterMac
	}

	nullAdd := "00:00:00:00:00:00"
	if macAdd == nullAdd {
		log.Println("The null address it is NOT a valide MAC address")
		goto enterMac
	}

	// The mac address is valid
	return macAdd
}

func RegistrationTime() time.Duration {
	conf := config.Get()
	reader := bufio.NewReader(os.Stdin)
	var i int

readAgain:
	fmt.Printf("Enter the duration for the connection in hours (MAX %d): ",
		conf.MaxConnectionTime)

	l, _, err := reader.ReadLine()
	if err != nil {
		log.Println(err)
		goto readAgain
	} else if len(l) == 0 {
		i = conf.MaxConnectionTime
	} else if i, err = strconv.Atoi(string(l)); err != nil {
		goto readAgain
	}

	if i > conf.MaxConnectionTime {
		i = conf.MaxConnectionTime
	} else if i <= 0 {
		i = 1
	}

	return time.Duration(i) * time.Hour
}

func extractMacFromSSH() string {
	sshClient, isPresent := os.LookupEnv("SSH_CLIENT")

	var ip string
	if !isPresent {
		slog.Warn("Could not fine the var SSH_CLIENT")
		return ""
	}

	index := strings.Index(sshClient, " ")
	if index == -1 {
		slog.With("SSH_CLIENT", sshClient).Error("Could not parse variable")
		return ""
	}

	ip = sshClient[0:index]
	slog.With("ip", ip).Debug("Found ip in ssh connection")

	f, err := os.ReadFile("/proc/net/arp")
	if err != nil {
		slog.With("err", err).Error("Could not read arp cache")
	}

	emptyMac, _ := net.ParseMAC("00:00:00:00:00:00")
	line := []byte{}
	hwPos := -1
	isFirstLine := true

	for _, data := range f {
		if !bytes.Equal([]byte{data}, []byte("\n")) {
			line = append(line, data)
		} else {
			slog.With("line", string(line)).Debug("Get arp file line")

			if isFirstLine {
				hwPos = strings.Index(string(line), "HW address")
				if hwPos == -1 {
					slog.With("line", string(line), "substring", "HW address").
						Error("Substring cannot be found. Arp tables is not right")
					break
				}
				slog.With("hwPos", hwPos).Debug("Found 'HW address' start")
				isFirstLine = false
			} else if strings.Contains(string(line), string(ip)) {
				mac, err := parseArpLine(line, hwPos)
				slog.With("ip", ip, "mac", mac, "err", err).Debug("Line parsed")

				if err != nil {
					slog.With("line", string(line), "err", err).
						Error("Error parsing arp line")
				} else if !reflect.DeepEqual(mac.String(), emptyMac.String()) {
					slog.With("mac", mac).Debug("Mac is not empty")
					return mac.String()
				}
			}
			line = line[:0]
		}
	}
	return ""
}

// line is the all line of the arp table
// hwPos the the number of the first byte that contains the mac address
func parseArpLine(line []byte, hwPos int) (net.HardwareAddr, error) {
	l := len(line)
	if hwPos >= l {
		return nil, fmt.Errorf("hwPos is greater than the line length")
	}
	if hwPos <= 0 {
		return nil, fmt.Errorf("hwPos is too small")
	}

	mac, err := net.ParseMAC(string(line[hwPos : hwPos+17]))

	return mac, err
}
