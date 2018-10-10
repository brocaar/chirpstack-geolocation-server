package cmd

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	"github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/brocaar/lora-geo-server/internal/backends/collos"
	"github.com/brocaar/lora-geo-server/internal/config"
	"github.com/brocaar/loraserver/api/geo"
)

func run(cmd *cobra.Command, args []string) error {
	// set log-level
	log.SetLevel(log.Level(uint8(config.C.General.LogLevel)))

	// print start message
	log.WithFields(log.Fields{
		"version": version,
		"docs":    "https://www.loraserver.io/lora-geo-server/",
	}).Info("starting LoRa Geo Server")

	// start api server
	log.WithFields(log.Fields{
		"bind":     config.C.GeoServer.API.Bind,
		"ca_cert":  config.C.GeoServer.API.CACert,
		"tls_cert": config.C.GeoServer.API.TLSCert,
		"tls_key":  config.C.GeoServer.API.TLSKey,
	}).Info("starting api server")

	opts := gRPCLoggingServerOptions()
	if apiConf := config.C.GeoServer.API; apiConf.CACert != "" && apiConf.TLSCert != "" && apiConf.TLSKey != "" {
		creds := mustGetTransportCredentials(apiConf.CACert, apiConf.TLSCert, apiConf.TLSKey, true)
		opts = append(opts, grpc.Creds(creds))
	}
	gs := grpc.NewServer(opts...)
	geoAPI := collos.NewAPI(config.C.GeoServer.Backend.Collos)
	geo.RegisterGeolocationServerServiceServer(gs, geoAPI)

	ln, err := net.Listen("tcp", config.C.GeoServer.API.Bind)
	if err != nil {
		return errors.Wrap(err, "start api listener error")
	}
	go gs.Serve(ln)

	sigChan := make(chan os.Signal)
	exitChan := make(chan struct{})
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	log.WithField("signal", <-sigChan).Info("signal received")
	go func() {
		log.Warning("stopping lora-geo-server")
		// stop server
		exitChan <- struct{}{}
	}()
	select {
	case <-exitChan:
	case s := <-sigChan:
		log.WithField("signal", s).Info("signal received, stopping immediately")
	}

	return nil
}

func gRPCLoggingServerOptions() []grpc.ServerOption {
	logrusEntry := log.NewEntry(log.StandardLogger())
	logrusOpts := []grpc_logrus.Option{
		grpc_logrus.WithLevels(grpc_logrus.DefaultCodeToLevel),
	}

	return []grpc.ServerOption{
		grpc_middleware.WithUnaryServerChain(
			grpc_ctxtags.UnaryServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
			grpc_logrus.UnaryServerInterceptor(logrusEntry, logrusOpts...),
		),
		grpc_middleware.WithStreamServerChain(
			grpc_ctxtags.StreamServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
			grpc_logrus.StreamServerInterceptor(logrusEntry, logrusOpts...),
		),
	}
}

func mustGetTransportCredentials(caCert, tlsCert, tlsKey string, verifyClientCert bool) credentials.TransportCredentials {
	cert, err := tls.LoadX509KeyPair(tlsCert, tlsKey)
	if err != nil {
		log.WithFields(log.Fields{
			"cert": tlsCert,
			"key":  tlsKey,
		}).Fatalf("load key-pair error: %s", err)
	}

	rawCaCert, err := ioutil.ReadFile(caCert)
	if err != nil {
		log.WithField("ca", caCert).Fatalf("load ca cert error: %s", err)
	}

	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(rawCaCert) {
		log.WithField("ca_cert", caCert).Fatal("append ca certificate error")
	}

	if verifyClientCert {
		return credentials.NewTLS(&tls.Config{
			Certificates: []tls.Certificate{cert},
			ClientCAs:    caCertPool,
			ClientAuth:   tls.RequireAndVerifyClientCert,
		})
	}

	return credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caCertPool,
	})
}
