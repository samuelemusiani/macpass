package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"

	"internal/comunication"

	"github.com/musianisamuele/macpass/cmd/macpass/config"
	"github.com/musianisamuele/macpass/cmd/macpass/input"
)

func main() {
	config.ParseConfig("./config.yaml")

	user := login()
	macAdd := input.Mac()
	time := input.RegistrationTime()

	fmt.Print(macAdd + "\t" + user + "\t")
	fmt.Println(time)

	send(comunication.Request{User: user, Mac: macAdd, Duration: time})
}

func login() string {
	// user, passwd := input.Credential()
	// Should check user and passwd
	return ""
}

func send(r comunication.Request) {
	conf := config.Get()
	// Connect to macpassd socket
	conn, err := net.Dial("unix", conf.SocketPath)
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
