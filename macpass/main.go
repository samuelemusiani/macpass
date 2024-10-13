package main

import (
	"fmt"
	"log"
	"strings"

	"internal/comunication"

	"github.com/musianisamuele/macpass/macpass/config"
	"github.com/musianisamuele/macpass/macpass/db"
	"github.com/musianisamuele/macpass/macpass/input"
	"github.com/musianisamuele/macpass/macpass/send"

	krbclient "github.com/jcmturner/gokrb5/v8/client"
	krbconfig "github.com/jcmturner/gokrb5/v8/config"
)

func main() {
	config.ParseConfig("./config.yaml")
	conf := config.Get()
	db.Connect(conf.DBPath)

	sender, err := send.New(&conf.Server)
	if err != nil {
		log.Fatal(err)
	}

	user := login()
	macAdd := input.Mac(user)
	time := input.RegistrationTime()

	fmt.Print(macAdd + "\t" + user + "\t")
	fmt.Println(time)

	db.InsertUser(user, macAdd)
	sender.Send(comunication.Request{User: user, Mac: macAdd, Duration: time})
}

func login() string {
	conf := config.Get()

	mail := input.Mail()
	user := strings.Split(mail, "@")[0]

	for range 3 {
		passwd := input.Password()

		if conf.DummyLogin {
			fmt.Println("WARNING: Dummy login is on")
			if user != passwd {
				fmt.Println("ERROR: User or password are incorrect")
				continue
			}
			return mail
		} else {
			const krb5Conf = `[libdefaults]
  dns_lookup_realm = true
  dns_lookup_kdc = true
  `
			krbconf, err := krbconfig.NewFromString(krb5Conf)
			if err != nil {
				log.Fatal(err)
			}

			cl := krbclient.NewWithPassword(user, conf.Kerberos.Realm, passwd, krbconf,
				krbclient.DisablePAFXFAST(conf.Kerberos.DisablePAFXFAST))

			if err := cl.Login(); err != nil {
				fmt.Println("Error: ", err)
				continue
			}

			// If login is succesful we return the user to bind the MAC address
			return mail
		}
	}
	log.Fatal("Could not authenticate")
	return ""
}
