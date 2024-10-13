package send

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"internal/comunication"
	"log"
	"net"
	"net/http"

	"github.com/musianisamuele/macpass/macpass/config"
)

type SendType string

const (
	SendSocketType SendType = "socket"
	SendHttpType   SendType = "http"
)

type Sender interface {
	Send(r comunication.Request)
}

type Socket struct {
	Path string
}

type HttpClient struct {
	Url  string
	Port uint16
}

func New(conf *config.Server) (Sender, error) {
	switch SendType(conf.Type) {
	case SendSocketType:
		return &Socket{
			Path: conf.Socket.Path,
		}, nil

	case SendHttpType:
		return &HttpClient{
			Url:  conf.Http.Url,
			Port: conf.Http.Port,
		}, nil
	default:
		return nil, errors.New("Type of sender not valid")
	}
}

func (s *Socket) Send(r comunication.Request) {
	// Connect to macpassd socket
	conn, err := net.Dial("unix", s.Path)
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

func (c *HttpClient) Send(r comunication.Request) {
	jsonData, err := json.Marshal(r)
	if err != nil {
		log.Fatal(err)
	}

	rr := bytes.NewReader(jsonData)

	url := fmt.Sprintf("http://%s:%d/", c.Url, c.Port)
	res, err := http.Post(url, "application/json", rr)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	if res.Status != "202 Accepted" {
		log.Fatal("Bad response from server")
	}
}
