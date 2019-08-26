package test

import (
	"context"
	"encoding/csv"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/golang/protobuf/jsonpb"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/brocaar/lora-geo-server/internal/backend/collos"
	"github.com/brocaar/lora-geo-server/internal/backend/loracloud"
	"github.com/brocaar/lora-geo-server/internal/config"
	geo "github.com/brocaar/loraserver/api/geo"
)

// ResolveTDOA runs the given Resolve TDOA test-suite.
func ResolveTDOA(logDir string) error {
	var backend geo.GeolocationServerServiceServer
	var err error

	switch config.C.GeoServer.Backend.Type {
	case "collos":
		backend, err = collos.NewBackend(config.C)
	case "lora_cloud":
		backend, err = loracloud.NewBackend(config.C)
	}
	if err != nil {
		return errors.Wrap(err, "new backend error")
	}

	reportFilePath := filepath.Join(logDir, config.C.GeoServer.Backend.Type+"-report-"+time.Now().UTC().Format(time.RFC3339)+".csv")
	f, err := os.Create(reportFilePath)
	if err != nil {
		return errors.Wrap(err, "open report file error")
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	if err := w.Write([]string{
		"id",
		"latitude",
		"longitude",
		"altitude",
		"accuracy",
	}); err != nil {
		log.Fatal(err)
	}

	files, err := ioutil.ReadDir(logDir)
	if err != nil {
		return errors.Wrap(err, "read directory error")
	}

	for _, f := range files {
		if !strings.HasSuffix(f.Name(), ".request.json") {
			continue
		}

		req, err := loadResolveTDOARequest(logDir, f.Name())
		if err != nil {
			return errors.Wrap(err, "load ResolveTDOARequest error")
		}

		res, err := backend.ResolveTDOA(context.Background(), &req)
		if err != nil {
			log.WithField("file", f.Name()).WithError(err).Error("ResolveTDOA error")
			continue
		}

		if err := writeResolveTDOAResponse(logDir, strings.TrimRight(f.Name(), ".request.json")+"."+config.C.GeoServer.Backend.Type+".response.json", res); err != nil {
			return errors.Wrap(err, "write ResolveTDOAResponse error")
		}

		if res.Result == nil {
			log.WithField("file", f.Name()).Warning("nil result")
		}

		if res.Result.Location == nil {
			log.WithField("file", f.Name()).Warning("nil location")
		}

		if err := w.Write([]string{
			f.Name(),
			strconv.FormatFloat(res.Result.Location.Latitude, 'f', 6, 64),
			strconv.FormatFloat(res.Result.Location.Longitude, 'f', 6, 64),
			strconv.FormatFloat(res.Result.Location.Altitude, 'f', 6, 64),
			strconv.FormatInt(int64(res.Result.Location.Accuracy), 10),
		}); err != nil {
			return errors.Wrap(err, "csv write error")
		}
	}

	return nil
}

// ResolveMultiFrameTDOA runs the given Resolve multi-frame TDOA test-suite.
func ResolveMultiFrameTDOA(logDir string) error {
	var backend geo.GeolocationServerServiceServer
	var err error

	switch config.C.GeoServer.Backend.Type {
	case "collos":
		backend, err = collos.NewBackend(config.C)
	case "lora_cloud":
		backend, err = loracloud.NewBackend(config.C)
	}
	if err != nil {
		return errors.Wrap(err, "new backend error")
	}

	reportFilePath := filepath.Join(logDir, config.C.GeoServer.Backend.Type+"-report-"+time.Now().UTC().Format(time.RFC3339)+".csv")
	f, err := os.Create(reportFilePath)
	if err != nil {
		return errors.Wrap(err, "open report file error")
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	if err := w.Write([]string{
		"id",
		"latitude",
		"longitude",
		"altitude",
		"accuracy",
	}); err != nil {
		log.Fatal(err)
	}

	files, err := ioutil.ReadDir(logDir)
	if err != nil {
		return errors.Wrap(err, "read directory error")
	}

	for _, f := range files {
		if !strings.HasSuffix(f.Name(), ".request.json") {
			continue
		}

		req, err := loadResolveMultiFrameTDOARequest(logDir, f.Name())
		if err != nil {
			return errors.Wrap(err, "load ResolveTDOARequest error")
		}

		res, err := backend.ResolveMultiFrameTDOA(context.Background(), &req)
		if err != nil {
			log.WithField("file", f.Name()).WithError(err).Error("ResolveTDOA error")
			continue
		}

		if err := writeResolveMultiFrameTDOAResponse(logDir, strings.TrimRight(f.Name(), ".request.json")+"."+config.C.GeoServer.Backend.Type+".response.json", res); err != nil {
			return errors.Wrap(err, "write ResolveTDOAResponse error")
		}

		if res.Result == nil {
			log.WithField("file", f.Name()).Warning("nil result")
		}

		if res.Result.Location == nil {
			log.WithField("file", f.Name()).Warning("nil location")
		}

		if err := w.Write([]string{
			f.Name(),
			strconv.FormatFloat(res.Result.Location.Latitude, 'f', 6, 64),
			strconv.FormatFloat(res.Result.Location.Longitude, 'f', 6, 64),
			strconv.FormatFloat(res.Result.Location.Altitude, 'f', 6, 64),
			strconv.FormatInt(int64(res.Result.Location.Accuracy), 10),
		}); err != nil {
			return errors.Wrap(err, "csv write error")
		}
	}

	return nil
}

func loadResolveTDOARequest(logDir, fn string) (geo.ResolveTDOARequest, error) {
	var out geo.ResolveTDOARequest

	f, err := os.Open(filepath.Join(logDir, fn))
	if err != nil {
		return out, errors.Wrap(err, "open file error")
	}
	defer f.Close()

	m := jsonpb.Unmarshaler{
		AllowUnknownFields: true,
	}

	if err := m.Unmarshal(f, &out); err != nil {
		return out, errors.Wrap(err, "unmarshal error")
	}

	return out, nil
}

func loadResolveMultiFrameTDOARequest(logDir, fn string) (geo.ResolveMultiFrameTDOARequest, error) {
	var out geo.ResolveMultiFrameTDOARequest

	f, err := os.Open(filepath.Join(logDir, fn))
	if err != nil {
		return out, errors.Wrap(err, "open file error")
	}
	defer f.Close()

	m := jsonpb.Unmarshaler{
		AllowUnknownFields: true,
	}

	if err := m.Unmarshal(f, &out); err != nil {
		return out, errors.Wrap(err, "unmarshal error")
	}

	return out, nil
}

func writeResolveTDOAResponse(logDir, fn string, resp *geo.ResolveTDOAResponse) error {
	m := jsonpb.Marshaler{
		EmitDefaults: true,
	}

	f, err := os.Create(filepath.Join(logDir, fn))
	if err != nil {
		return errors.Wrap(err, "open file error")
	}
	defer f.Close()

	if err := m.Marshal(f, resp); err != nil {
		return errors.Wrap(err, "marshal error")
	}

	return nil
}

func writeResolveMultiFrameTDOAResponse(logDir, fn string, resp *geo.ResolveMultiFrameTDOAResponse) error {
	m := jsonpb.Marshaler{
		EmitDefaults: true,
	}

	f, err := os.Create(filepath.Join(logDir, fn))
	if err != nil {
		return errors.Wrap(err, "open file error")
	}
	defer f.Close()

	if err := m.Marshal(f, resp); err != nil {
		return errors.Wrap(err, "marshal error")
	}

	return nil
}
