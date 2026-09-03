package main

import (
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/GioTld/aldea/internal/config"
)

func main() {
	cfgPath := flag.String("config", "tracker.yaml", "path to tracker config file")
	flag.Parse()

	cfg, err := config.LoadTrackerConfig(*cfgPath)
	if err != nil {
		slog.Error("loading config", "err", err)
		os.Exit(1)
	}

	srv, err := newTrackerServer(cfg)
	if err != nil {
		slog.Error("initializing tracker server", "err", err)
		os.Exit(1)
	}
	defer srv.close()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	httpSrv := &http.Server{
		Addr:    cfg.ListenAddr,
		Handler: srv.routes(),
	}

	go func() {
		slog.Info("trackerd starting", "addr", cfg.ListenAddr)
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server stopped unexpectedly", "err", err)
		}
	}()

	<-stop
	slog.Info("trackerd shutting down")
}
