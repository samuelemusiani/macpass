package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"internal/comunication"

	"github.com/go-ldap/ldap"
	"github.com/spf13/viper"
	"golang.org/x/term"
)

var ldapUri string = ""
var baseDN string = ""
var bindAdmin string = ""
var bindPassword string = ""
var userDNType string = ""
var socketPath string = ""
var maxTime int

func main() {
	setConfig()
	user := ldapAuth()
	macAdd := inputMac()
	time := timeRegistered()

	fmt.Print(macAdd + "\t" + user + "\t")
	fmt.Println(time)

	send(comunication.Request{User: user, Mac: macAdd, Duration: time})

}

func ldapAuth() string {
	l, err := ldap.DialURL(ldapUri)
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()

	// First bind with a read only user
	err = l.Bind(bindAdmin, bindPassword)
	if err != nil {
		log.Fatal(err)
	}

	username, password := credentials()

	err = l.Bind(userDNType+"="+username+","+baseDN, password)
	if err != nil {
		log.Fatal(err)
	}

	return username
}

func credentials() (string, string) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter Username: ")
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

func inputMac() string {
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

func setConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("toml")

	ex, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	exPath := filepath.Dir(ex)
	viper.AddConfigPath(exPath) // for now the config should be in the same directory

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Println("Config file not found")
			log.Fatal(err)
		} else {
			log.Println("Config file was found but another error was produced")
			log.Fatal(err)
		}
	}

	ldapUri = viper.GetString("ldapUri")
	baseDN = viper.GetString("baseDN")
	bindAdmin = viper.GetString("bindAdmin")
	bindPassword = viper.GetString("bindPassword")
	userDNType = viper.GetString("userDNType")
	socketPath = viper.GetString("socketPath")
	maxTime = viper.GetInt("maxConnectionTime")

	fmt.Println("Config parsed successfully")
}

func timeRegistered() time.Duration {
	fmt.Printf("Enter the duration for the connection in hours (MAX %d): ", maxTime)
	var i int
	_, err := fmt.Scanf("%d", &i)
	if err != nil {
		log.Fatal(err)
	}

	if i > maxTime {
		i = maxTime
	} else if i <= 0 {
		i = 1
	}

	return time.Duration(i) * time.Hour
}

func send(r comunication.Request) {
	// Connect to macpassd socket
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		log.Fatal(err)
	}

	jsonData, err := json.Marshal(r)
	if err != nil {
		log.Fatal(err)
	}

	_, err = conn.Write(jsonData)
	if err != nil {
		log.Fatal(err)
	}

	conn.Close()
}
