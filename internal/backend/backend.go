package backend

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_logrus "github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/brocaar/chirpstack-geolocation-server/internal/backend/collos"
	"github.com/brocaar/chirpstack-geolocation-server/internal/backend/logger"
	"github.com/brocaar/chirpstack-geolocation-server/internal/backend/loracloud"
	"github.com/brocaar/chirpstack-geolocation-server/internal/config"
	"github.com/brocaar/loraserver/api/geo"
)

func Setup(c config.Config) error {
	var b geo.GeolocationServerServiceServer
	var err error

	switch c.GeoServer.Backend.Type {
	case "collos":
		b, err = collos.NewBackend(c)
	case "lora_cloud":
		b, err = loracloud.NewBackend(c)
	default:
		return fmt.Errorf("unknown backend: %s", c.GeoServer.Backend.Type)
	}

	if err != nil {
		return errors.Wrap(err, "setup backend error")
	}

	b, err = logger.NewBackend(b, c)
	if err != nil {
		return errors.Wrap(err, "setup logging backend error")
	}

	log.WithFields(log.Fields{
		"backend":  c.GeoServer.Backend.Type,
		"bind":     c.GeoServer.API.Bind,
		"ca_cert":  c.GeoServer.API.CACert,
		"tls_cert": c.GeoServer.API.TLSCert,
		"tls_key":  c.GeoServer.API.TLSKey,
	}).Info("starting api server")

	if err := serveBackend(b); err != nil {
		return errors.Wrap(err, "serve backend error")
	}

	return nil
}

func serveBackend(b geo.GeolocationServerServiceServer) error {
	opts := gRPCLoggingServerOptions()
	if apiConf := config.C.GeoServer.API; apiConf.CACert != "" || apiConf.TLSCert != "" || apiConf.TLSKey != "" {
		creds, err := getTransportCredentials(apiConf.CACert, apiConf.TLSCert, apiConf.TLSKey, true)
		if err != nil {
			return errors.Wrap(err, "get transport credentials error")
		}
		opts = append(opts, grpc.Creds(creds))
	}

	gs := grpc.NewServer(opts...)
	geo.RegisterGeolocationServerServiceServer(gs, b)

	ln, err := net.Listen("tcp", config.C.GeoServer.API.Bind)
	if err != nil {
		return errors.Wrap(err, "start api listener error")
	}

	go gs.Serve(ln)

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
			grpc_prometheus.UnaryServerInterceptor,
		),
		grpc_middleware.WithStreamServerChain(
			grpc_ctxtags.StreamServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
			grpc_logrus.StreamServerInterceptor(logrusEntry, logrusOpts...),
			grpc_prometheus.StreamServerInterceptor,
		),
	}
}

func getTransportCredentials(caCert, tlsCert, tlsKey string, verifyClientCert bool) (credentials.TransportCredentials, error) {
	cert, err := tls.LoadX509KeyPair(tlsCert, tlsKey)
	if err != nil {
		return nil, errors.Wrap(err, "load key-pair error")
	}

	rawCaCert, err := ioutil.ReadFile(caCert)
	if err != nil {
		return nil, errors.Wrap(err, "load ca cert error")
	}

	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(rawCaCert) {
		return nil, errors.Wrap(err, "append ca certificate error")
	}

	if verifyClientCert {
		return credentials.NewTLS(&tls.Config{
			Certificates: []tls.Certificate{cert},
			ClientCAs:    caCertPool,
			ClientAuth:   tls.RequireAndVerifyClientCert,
		}), nil
	}

	return credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caCertPool,
	}), nil
}
