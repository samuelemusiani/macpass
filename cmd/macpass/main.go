package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"syscall"

	"github.com/go-ldap/ldap"
	"github.com/spf13/viper"
	"golang.org/x/term"
)

var ldapUri string = ""
var baseDN string = ""
var bindAdmin string = ""
var bindPassword string = ""
var userDNType string = ""
var outFilePath string = ""

func main() {
	setConfig()
	// user, _ := ldapAuth()
	user := "pippo"
	macAdd, _ := macRegistration()

	fmt.Println(macAdd + "\t" + user)

	f, err := os.OpenFile(outFilePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	if _, err = f.WriteString(macAdd + " " + user + "\n"); err != nil {
		log.Fatal(err)
	}
}

func ldapAuth() (string, error) {
	// Need to implement the auth part with ldap
	// https://pkg.go.dev/github.com/go-ldap/ldap/v3

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

	username, password, _ := credentials()

	err = l.Bind(userDNType+"="+username+","+baseDN, password)
	if err != nil {
		log.Fatal(err)
	}

	return username, nil
}

func credentials() (string, string, error) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter Username: ")
	username, err := reader.ReadString('\n')
	if err != nil {
		return "", "", err
	}

	fmt.Print("Enter Password: ")
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", "", err
	}
	fmt.Println()

	password := string(bytePassword)
	return strings.TrimSpace(username), strings.TrimSpace(password), nil
}

func macRegistration() (string, error) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter a MAC address: ")
	macAdd, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	macAdd = strings.TrimSpace(macAdd)

	if _, err := net.ParseMAC(macAdd); err != nil {
		log.Println(err)
	}

	return macAdd, nil
}

func setConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath(".") // for now the config should be in the same directory

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Println("Config file not found")
			log.Fatal(err)
		} else {
			log.Println("Config file was found but another error was produced")
			log.Fatal(err)
		}
	}

	log.Println("Config file found and successfully parsed")

	ldapUri = viper.GetString("ldapUri")
	baseDN = viper.GetString("baseDN")
	bindAdmin = viper.GetString("bindAdmin")
	bindPassword = viper.GetString("bindPassword")
	userDNType = viper.GetString("userDNType")
	outFilePath = viper.GetString("outFilePath")
}
