package main

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/url"
	"strings"

	"internal/comunication"

	"github.com/musianisamuele/macpass/cmd/macpass/config"
	"github.com/musianisamuele/macpass/cmd/macpass/db"
	"github.com/musianisamuele/macpass/cmd/macpass/input"

	krbclient "github.com/jcmturner/gokrb5/v8/client"
	krbconfig "github.com/jcmturner/gokrb5/v8/config"

	ldap "github.com/go-ldap/ldap/v3"
)

func main() {
	config.ParseConfig("./config.yaml")
	conf := config.Get()
	db.Connect(conf.DBPath)

	user := login()
	macAdd := input.Mac(user)
	time := input.RegistrationTime()

	fmt.Print(macAdd + "\t" + user + "\t")
	fmt.Println(time)

	db.InsertUser(user, macAdd)
	send(comunication.Request{User: user, Mac: macAdd, Duration: time})
}

func login() string {
	const krb5Conf = `[libdefaults]
  dns_lookup_realm = true
  dns_lookup_kdc = true
  `
	krbconf, err := krbconfig.NewFromString(krb5Conf)
	if err != nil {
		log.Fatal(err)
	}

	user, passwd := input.Credential()

	logins := config.Get().Login

	// test keberos logins
	for _, d := range logins.KerberosDomains {
		cl := krbclient.NewWithPassword(strings.Split(user, "@")[0], d.Realm,
			passwd, krbconf, krbclient.DisablePAFXFAST(d.DisablePAFXFAST))

		if err := cl.Login(); err == nil {
			// If login is succesful we return the user to bind the MAC address
			return user
		}
	}

	// test ldap logins
	for _, d := range logins.LdapDomains {
		l, err := establishLdapConnection(&d)
		if err != nil {
			continue
		}
		defer l.Close()

		// First bind with a read only user
		err = l.Bind(d.BindDN, d.BindPW)
		if err != nil {
			continue
		}

		err = l.Bind(d.UserDNType+"="+user+","+d.BaseDN, passwd)
		if err != nil {
			continue
		}

		return user
	}

	log.Fatal("The user can't be autenticated")
	return "" // unreachable code
}

func establishLdapConnection(d *config.LdapDomain) (*ldap.Conn, error) {

	dUrl, err := url.Parse(d.Address)
	if err != nil {
		return nil, err
	}

	tlsConfig := tls.Config{InsecureSkipVerify: d.InsecureSkipVerify}

	var c *ldap.Conn = nil

	switch dUrl.Scheme {
	case "ldap":
		{
			if d.StartTLS {
				c, err = ldap.DialURL(d.Address)
				if err != nil {
					return nil, err
				}
				err = c.StartTLS(&tlsConfig)
			} else {
				c, err = ldap.DialURL(d.Address)
				if err != nil {
					return nil, err
				}
			}
		}
	case "ldaps":
		{
			c, err = ldap.DialURL(d.Address, ldap.DialWithTLSConfig(&tlsConfig))
			if err != nil {
				return nil, err
			}
		}
	default:
		{
			return nil, errors.New("Invalid scheme in server URI: " + dUrl.Scheme)
		}
	}

	return c, nil
}

func send(r comunication.Request) {
	// Connect to macpassd socket
	conn, err := net.Dial("unix", config.Get().SocketPath)
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
