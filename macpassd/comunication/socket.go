package comunication

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
	"github.com/musianisamuele/macpass/macpassd/requests"
)

type Socket struct {
	l net.Listener
}

func newSocket(conf *config.Socket) (*Socket, error) {
	// Create a socket for comunication between macpass and macpassd
	slog.With("socketPath", conf.Path).Debug("Creating socket")
	socket, err := net.Listen("unix", conf.Path)
	if err != nil {
		return nil, err
	}

	// Get user owner group
	group, err := user.Lookup(conf.User)
	if err != nil {
		return nil, err
	}
	uid, _ := strconv.Atoi(group.Uid)
	gid, _ := strconv.Atoi(group.Gid)

	if err := os.Chown(conf.Path, uid, gid); err != nil {
		return nil, err
	}

	// if err := os.Chmod(conf.LoggerPath, 0660); err != nil {
	// 	log.Fatal(err)
	// }

	// Cleanup the sockfile when macpassd is terminated
	closeChannel := make(chan os.Signal, 1)
	signal.Notify(closeChannel, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-closeChannel
		os.Remove(conf.Path)
	}()

	return &Socket{socket}, nil
}

func (s *Socket) Listen(fw fw.Firewall) {
	slog.With("socketAdd: ", s.l.Addr().String()).
		Info("Listening for new comunication on socket.")

	for {
		conn, err := s.l.Accept()
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

		requests.Handle(newEntry, fw)
	}
}
