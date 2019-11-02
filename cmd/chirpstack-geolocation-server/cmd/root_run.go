package cmd

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/brocaar/chirpstack-geolocation-server/internal/backend"
	"github.com/brocaar/chirpstack-geolocation-server/internal/config"
	"github.com/brocaar/chirpstack-geolocation-server/internal/metrics"
)

func run(cmd *cobra.Command, args []string) error {
	tasks := []func() error{
		setLogLevel,
		printStartMessage,
		setupMetrics,
		setupBackend,
	}

	for _, t := range tasks {
		if err := t(); err != nil {
			log.Fatal(err)
		}
	}

	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	log.WithField("signal", <-sigChan).Info("signal received")

	return nil
}

func setLogLevel() error {
	log.SetLevel(log.Level(uint8(config.C.General.LogLevel)))
	return nil
}

func printStartMessage() error {
	log.WithFields(log.Fields{
		"version": version,
		"docs":    "https://www.chirpstack.io/geolocation-server/",
	}).Info("starting ChirpStack Geolocation Server")
	return nil
}

func setupBackend() error {
	if err := backend.Setup(config.C); err != nil {
		return errors.Wrap(err, "setup backend error")
	}
	return nil
}

func setupMetrics() error {
	if err := metrics.Setup(config.C); err != nil {
		return errors.Wrap(err, "setup metrics error")
	}
	return nil
}
