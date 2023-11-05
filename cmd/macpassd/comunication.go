package main

import (
	"encoding/json"
	"internal/comunication"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"os/user"
	"strconv"
	"syscall"

	"github.com/musianisamuele/macpass/cmd/macpassd/config"
)

var socket net.Listener

func initComunication() {
	log.Println("Initializing comunications with macpass")
	conf := config.Get()

	// Create a socket for comunication between macpass and macpassd
	log.Println("Creting socket in: ", conf.Socket.Path)
	var err error
	socket, err = net.Listen("unix", conf.Socket.Path)
	if err != nil {
		log.Fatal(err)
	}

	// Get user owner group
	group, err := user.Lookup(conf.Socket.User)
	if err != nil {
		log.Fatal(err)
	}
	uid, _ := strconv.Atoi(group.Uid)
	gid, _ := strconv.Atoi(group.Gid)

	if err := os.Chown(conf.Socket.Path, uid, gid); err != nil {
		log.Fatal(err)
	}

	if err := os.Chmod(conf.LoggerPath, 0660); err != nil {
		log.Fatal(err)
	}

	// Cleanup the sockfile when macpassd is terminated
	closeChannel := make(chan os.Signal, 1)
	signal.Notify(closeChannel, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-closeChannel
		os.Remove(conf.LoggerPath)
		os.Exit(0)
	}()
}

func handleComunication() {
	log.Println("Listening for new comunication on socket: ",
		socket.Addr().String())

	for {
		conn, err := socket.Accept()
		if err != nil {
			log.Fatal(err)
		}

		buff := make([]byte, 4096)
		n, err := conn.Read(buff)

		if err != nil {
			if err == io.EOF {
				conn.Close()
				continue
			}

			log.Println(err)
		}
		var newEntry comunication.Request
		if err := json.Unmarshal(buff[:n], &newEntry); err != nil {
			log.Println(err)
		}

		handleRequest(newEntry)
	}
}
