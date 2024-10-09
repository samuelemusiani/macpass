package fw

import (
	"github.com/musianisamuele/macpass/cmd/macpassd/registration"
)

type Shorewall struct {
}

func (s Shorewall) Init()                              { panic("Not implemented") }
func (s Shorewall) Allow(r registration.Registration)  { panic("Not implemented") }
func (s Shorewall) Delete(r registration.Registration) { panic("Not implemented") }
