package comunication

import (
	"errors"

	"github.com/musianisamuele/macpass/macpassd/config"
	"github.com/musianisamuele/macpass/macpassd/fw"
)

type ComunicationType string

const (
	TypeSocket ComunicationType = "socket"
	TypeHttp   ComunicationType = "http"
)

type Comunicator interface {
	Listen(fw fw.Firewall)
}

func New(conf *config.Server) (Comunicator, error) {
	switch ComunicationType(conf.Type) {
	case TypeSocket:
		return newSocket(&conf.Socket)
	case TypeHttp:
		return initHttp(&conf.Http)
	default:
		return nil, errors.New("Type not supported")
	}
}
