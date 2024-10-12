package main

import (
	"encoding/json"
	"internal/comunication"
	"io"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"os/user"
	"strconv"
	"syscall"

	"github.com/musianisamuele/macpass/macpassd/config"
	"github.com/musianisamuele/macpass/macpassd/fw"
)

var socket net.Listener

func initComunication() {
	slog.Info("Initializing comunications with macpass")
	conf := config.Get()

	// Create a socket for comunication between macpass and macpassd
	slog.With("socketPath", conf.Socket.Path).Info("Creating socket")
	var err error
	socket, err = net.Listen("unix", conf.Socket.Path)
	if err != nil {
		slog.With("socketPath", conf.Socket.Path, "error", err).
			Error("Creating socket")
		os.Exit(4)
	}

	// Get user owner group
	group, err := user.Lookup(conf.Socket.User)
	if err != nil {
		slog.With("socketPath", conf.Socket.Path, "error", err).
			Error("Get user and group for socket")
		os.Exit(4)
	}
	uid, _ := strconv.Atoi(group.Uid)
	gid, _ := strconv.Atoi(group.Gid)

	if err := os.Chown(conf.Socket.Path, uid, gid); err != nil {
		slog.With("socketPath", conf.Socket.Path, "error", err).
			Error("Modify user and group permission for socket")
		os.Exit(4)
	}

	// if err := os.Chmod(conf.LoggerPath, 0660); err != nil {
	// 	log.Fatal(err)
	// }

	// Cleanup the sockfile when macpassd is terminated
	closeChannel := make(chan os.Signal, 1)
	signal.Notify(closeChannel, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-closeChannel
		os.Remove(conf.LoggerPath)
		os.Remove(conf.Socket.Path)
		os.Exit(0)
	}()
}

func handleComunication(fw fw.Firewall) {
	slog.With("socketAdd: ", socket.Addr().String()).
		Info("Listening for new comunication on socket.")

	for {
		conn, err := socket.Accept()
		if err != nil {
			slog.With("error", err).Error("Listening on socket")
			os.Exit(4)
		}

		buff := make([]byte, 4096)
		n, err := conn.Read(buff)

		if err != nil {
			if err == io.EOF {
				conn.Close()
				continue
			}

			slog.With("error", err).Error("Reading socket")
		}
		var newEntry comunication.Request
		if err := json.Unmarshal(buff[:n], &newEntry); err != nil {
			slog.With("error", err).Error("Decoding comunication from macpass")
		}

		handleRequest(newEntry, fw)
	}
}
