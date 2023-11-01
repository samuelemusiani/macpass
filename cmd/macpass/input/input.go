package input

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/musianisamuele/macpass/cmd/macpass/config"
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

	fmt.Print("Enter Password: ")
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println()

	password := string(bytePassword)
	return strings.TrimSpace(username), strings.TrimSpace(password)
}

func Mac() string {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter a MAC address: ")
	macAdd, err := reader.ReadString('\n')
	if err != nil {
		log.Fatal(err)
	}

	macAdd = strings.TrimSpace(macAdd)

	if _, err := net.ParseMAC(macAdd); err != nil {
		log.Fatal(err)
	}

	macAdd, err = macparse.ParseMac(macAdd, "linux")
	if err != nil {
		log.Fatal(err)
	}

	brdAdd := "ff:ff:ff:ff:ff:ff"
	if macAdd == brdAdd {
		log.Fatal("The brodcast address it is NOT a valide MAC address")
	}

	nullAdd := "00:00:00:00:00:00"
	if macAdd == nullAdd {
		log.Fatal("The null address it is NOT a valide MAC address")
	}

	// The mac address is valid
	return macAdd
}

func RegistrationTime() time.Duration {
	conf := config.Get()
	fmt.Printf("Enter the duration for the connection in hours (MAX %d): ",
		conf.MaxConnectionTime)
	var i int
	_, err := fmt.Scanf("%d", &i)
	if err != nil {
		log.Fatal(err)
	}

	if i > conf.MaxConnectionTime {
		i = conf.MaxConnectionTime
	} else if i <= 0 {
		i = 1
	}

	return time.Duration(i) * time.Hour
}
