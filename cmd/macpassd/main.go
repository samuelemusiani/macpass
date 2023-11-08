package main

import (
	"log/slog"
	"os"
	"time"

	"github.com/musianisamuele/macpass/cmd/macpassd/config"
	"github.com/musianisamuele/macpass/cmd/macpassd/registration"
)

// var entriesLogger *log.Logger

func main() {

	// Config
	if len(os.Args) <= 1 {
		slog.Error("Please provide a config path")
		os.Exit(1)
	} else if len(os.Args) > 2 {
		slog.Error("Too many arguments provided")
		os.Exit(1)
	}

	config.ParseConfig(os.Args[1]) //tmp

	var logLevel slog.Level

	conf := config.Get()

	switch conf.LogLevel {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelWarn
	}
	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: logLevel})
	slog.SetDefault(slog.New(handler))

	// // Logger
	// f, err := os.OpenFile(config.Get().LoggerPath,
	// 	os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0660)
	// if err != nil {
	// 	slog.With("path", config.Get().LoggerPath, "error", err).
	// 		Error("Failed to open logger")
	// 	os.Exit(2)
	// }
	// entriesLogger = log.New(f, "", 3)

	startDaemon()
}

func startDaemon() {
	slog.Info("Starting macpassd daemon")

	initIptables()

	initComunication()
	go handleComunication()

	for {
		// checkIfStilConnected() TODO
		deleteOldEntries()
		scanNetwork()

		time.Sleep(10 * time.Second)
	}
}

func deleteOldEntries() {
	oldEntries := registration.GetOldEntries()

	slog.With("oldEntries", oldEntries).Debug("Removing old entries")
	for _, e := range oldEntries {
		deleteEntryFromFirewall(e)
		registration.Remove(e)
	}
}
