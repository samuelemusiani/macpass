package main

import (
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net"
	"os"
	"strings"

	"internal/comunication"

	"github.com/musianisamuele/macpass/cmd/macpass/config"
	"github.com/musianisamuele/macpass/cmd/macpass/db"
	"github.com/musianisamuele/macpass/cmd/macpass/input"

	krbclient "github.com/jcmturner/gokrb5/v8/client"
	krbconfig "github.com/jcmturner/gokrb5/v8/config"
)

func main() {
	config.ParseConfig("./config.yaml")
	conf := config.Get()
	db.Connect(conf.DBPath)

	if conf.Debug {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}

	user := getUser()
	macAdd := input.Mac(user)
	time := input.RegistrationTime()

	fmt.Print(macAdd + "\t" + user + "\t")
	fmt.Println(time)

	db.InsertUser(user, macAdd)
	send(comunication.Request{User: user, Mac: macAdd, Duration: time})
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

func getSSHAuthKey() (string, error) {
	file, found := os.LookupEnv("SSH_USER_AUTH")

	if !found {
		return "", fmt.Errorf("Can't find SSH_USER_AUTH env variable")
	}

	key, err := os.ReadFile(file)
	if err != nil {
		return "", err
	}

	return string(key), nil
}

func getUser() string {
	var key_is_present = true
	var user string = ""

	key, err := getSSHAuthKey()
	if err != nil {
		slog.With("err", err).Debug("Can't find SSH auth key")
		key_is_present = false
	} else {
		user, err = db.GetUserFromKey(key)
		if err != nil {
			slog.With("key", key, "err", err).Debug("Can't get user from ssh key")
			user = ""
		}
	}

	if user == "" {
		user = login()
	}

	if key_is_present {
		// If the user had already authenticated we saved his ssh key
		db.AddKeyToUser(user, key)
	}

	return user
}
