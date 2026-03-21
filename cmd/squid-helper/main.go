package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/tro3373/squid-brocker/internal/config"
	"github.com/tro3373/squid-brocker/internal/handler"
	"github.com/tro3373/squid-brocker/internal/tracker"
)

func main() {
	configPath := flag.String("config", "/etc/squid-brocker/rules.yaml", "path to rules.yaml")
	dataDir := flag.String("data", "/var/lib/squid-brocker", "path to data directory")
	flag.Parse()

	if err := run(*configPath, *dataDir); err != nil {
		log.Fatalf("squid-helper: %v", err)
	}
}

func run(configPath, dataDir string) error {
	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	store := tracker.NewFileStore(dataDir + "/state.json")
	tr, err := tracker.New(cfg, store)
	if err != nil {
		return fmt.Errorf("creating tracker: %w", err)
	}

	go handleSignals(tr)

	log.Printf("squid-helper started (config=%s, data=%s)", configPath, dataDir)
	return handler.Run(os.Stdin, os.Stdout, tr)
}

func handleSignals(_ *tracker.Tracker) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)
	<-sigCh
	log.Println("squid-helper shutting down")
	os.Exit(0)
}
