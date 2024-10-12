package fw

import (
	"errors"

	"github.com/musianisamuele/macpass/macpassd/registration"
)

type Firewall interface {
	Init()
	Allow(r registration.Registration)
	Delete(r registration.Registration)
}

type FirewallType string

const (
	TYPE_IPTABLES  FirewallType = "iptables"
	TYPE_SHOREWALL FirewallType = "shorewall"
)

func New(t FirewallType) (Firewall, error) {
	switch t {
	case TYPE_IPTABLES:
		return &Iptables{}, nil
	case TYPE_SHOREWALL:
		return &Shorewall{}, nil
	default:
		return nil, errors.New("Type not supported")
	}
}
