package comunication

import (
	"encoding/json"
	"internal/comunication"
	"io"
	"log/slog"
	"net/http"

	"github.com/musianisamuele/macpass/macpassd/config"
	"github.com/musianisamuele/macpass/macpassd/fw"
	"github.com/musianisamuele/macpass/macpassd/requests"
)

type HttpServer struct {
	s http.ServeMux

	conf *config.HttpServer
}

var firewall fw.Firewall

func initHttp(conf *config.HttpServer) (*HttpServer, error) {
	var server HttpServer
	server.s.HandleFunc("/", rootHandler)
	server.conf = conf

	return &server, nil
}

func (s *HttpServer) Listen(fw fw.Firewall) {
	firewall = fw
	http.ListenAndServe(s.conf.Bind, &s.s)
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.Header().Set("Allow", "OPTIONS GET POST")
		return
	}

	if r.Method == http.MethodGet {
		w.Write([]byte("Hello from Macpassd!"))
		return
	}

	// POST

	body, err := io.ReadAll(r.Body)
	if err != nil {
		slog.With("err", err).Error("Reading body of http request")
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	var mReq comunication.Request
	err = json.Unmarshal(body, &mReq)
	if err != nil {
		slog.With("err", err, "body", body).Error("Reading unmarshaling body of http request")
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	requests.Handle(mReq, firewall)
	w.WriteHeader(http.StatusAccepted)
	return
}
