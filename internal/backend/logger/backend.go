package logger

import (
	"bytes"
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/brocaar/lora-geo-server/internal/config"
	"github.com/brocaar/loraserver/api/geo"
)

// Backend implements a logging backend.
type Backend struct {
	backend geo.GeolocationServerServiceServer
	logDir  string
}

// NewBackend creates a new logging backend, wrapping the given backend.
func NewBackend(b geo.GeolocationServerServiceServer, c config.Config) (geo.GeolocationServerServiceServer, error) {
	if b == nil {
		return nil, errors.New("the given backend must not be nil")
	}

	return &Backend{
		backend: b,
		logDir:  c.GeoServer.Backend.RequestLogDir,
	}, nil
}

// ResolveTDOA resolves the location based on TDOA.
func (b *Backend) ResolveTDOA(ctx context.Context, req *geo.ResolveTDOARequest) (*geo.ResolveTDOAResponse, error) {
	if err := b.logRequest("ResolveTDOA", req); err != nil {
		log.WithError(err).Error("backend/logger: log request error")
	}

	return b.backend.ResolveTDOA(ctx, req)
}

// ResolveMultiFrameTDOA resolves the location using TDOA, based on
// multiple frames.
func (b *Backend) ResolveMultiFrameTDOA(ctx context.Context, req *geo.ResolveMultiFrameTDOARequest) (*geo.ResolveMultiFrameTDOAResponse, error) {
	if err := b.logRequest("ResolveMultiFrameTDOA", req); err != nil {
		log.WithError(err).Error("backend/logger: log request error")
	}

	return b.backend.ResolveMultiFrameTDOA(ctx, req)
}

func (b *Backend) logRequest(prefix string, msg proto.Message) error {
	if b.logDir == "" {
		return nil
	}

	// in case it already exists, this does nothing
	if err := os.MkdirAll(filepath.Join(b.logDir, prefix), os.ModePerm); err != nil {
		return errors.Wrap(err, "make log directory error")
	}

	bb := bytes.NewBuffer(nil)
	m := jsonpb.Marshaler{
		EnumsAsInts:  false,
		EmitDefaults: true,
	}
	if err := m.Marshal(bb, msg); err != nil {
		return errors.Wrap(err, "marshal json error")
	}

	filePath := filepath.Join(b.logDir, prefix, time.Now().UTC().Format(time.RFC3339)+".request.json")
	if err := ioutil.WriteFile(filePath, bb.Bytes(), 0644); err != nil {
		return errors.Wrap(err, "write file error")
	}

	log.WithField("path", filePath).Info("backend/logger: log file created")

	return nil
}
