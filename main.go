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
	"golang.org/x/term"
)

const ldapUri string = "ldap://ldap.forumsys.com:389" //Server to test
const baseDN string = "dc=example,dc=com"
const bindAdmin string = "cn=read-only-admin," + baseDN
const bindPassword string = "password"
const userDNType string = "uid"

func main() {
	// TODO: Parse config file

	user, _ := ldapAuth()
	macAdd, _ := macRegistration()

	fmt.Println(macAdd + "\t" + user)
}

func ldapAuth() (string, error) {
	// Need to implement the auth part with ldap
	// https://pkg.go.dev/github.com/go-ldap/ldap/v3

	l, err := ldap.DialURL(ldapUri)
	if err != nil {
		log.Fatal(err)
	}

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

	defer l.Close()

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
