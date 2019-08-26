package loracloud

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"github.com/brocaar/lora-geo-server/internal/config"
	"github.com/brocaar/loraserver/api/common"
	"github.com/brocaar/loraserver/api/geo"
	"github.com/brocaar/lorawan"
)

// Backend implements the LoRa Cloud geolocation backend.
type Backend struct {
	uri            string
	token          string
	requestTimeout time.Duration
}

// NewBackend creates a new LoRa Cloud backend.
func NewBackend(c config.Config) (geo.GeolocationServerServiceServer, error) {
	return &Backend{
		uri:            c.GeoServer.Backend.LoRaCloud.URI,
		token:          c.GeoServer.Backend.LoRaCloud.Token,
		requestTimeout: c.GeoServer.Backend.LoRaCloud.RequestTimeout,
	}, nil
}

// ResolveTDOA resolves the location based on TDOA.
func (b *Backend) ResolveTDOA(ctx context.Context, req *geo.ResolveTDOARequest) (*geo.ResolveTDOAResponse, error) {
	lcReq, err := resolveTDOARequestToLoRaCloudRequest(req)
	if err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, err.Error())
	}

	tdoaResp, err := b.resolveTDOA(ctx, lcReq)
	if err != nil {
		return nil, grpc.Errorf(codes.Unknown, "geolocation error: %s", err)
	}

	var devEUI lorawan.EUI64
	copy(devEUI[:], req.DevEui)

	if len(tdoaResp.Warnings) != 0 {
		log.WithFields(log.Fields{
			"dev_eui":  devEUI,
			"warnings": tdoaResp.Warnings,
		}).Warning("backend/lora_cloud: backend returned warnings")
	}

	if len(tdoaResp.Errors) != 0 {
		log.WithFields(log.Fields{
			"dev_eui": devEUI,
			"errors":  tdoaResp.Errors,
		}).Error("backend/lora_cloud: backend returned errors")

		return nil, grpc.Errorf(codes.Internal, "backend returned errors: %v", tdoaResp.Errors)
	}

	return &geo.ResolveTDOAResponse{
		Result: &geo.ResolveResult{
			Location: &common.Location{
				Source:    common.LocationSource_GEO_RESOLVER,
				Accuracy:  uint32(tdoaResp.Result.Accuracy),
				Latitude:  tdoaResp.Result.Latitude,
				Longitude: tdoaResp.Result.Longitude,
				Altitude:  tdoaResp.Result.Altitude,
			},
		},
	}, nil
}

// ResolveMultiFrameTDOA resolves the location using TDOA, based on
// multiple frames.
func (b *Backend) ResolveMultiFrameTDOA(ctx context.Context, req *geo.ResolveMultiFrameTDOARequest) (*geo.ResolveMultiFrameTDOAResponse, error) {
	lcReq, err := resolveMutiFrameTDOARequestToLoRaCloudRequest(req)
	if err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, err.Error())
	}

	tdoaResp, err := b.resolveTDOAMultiFrame(ctx, lcReq)
	if err != nil {
		return nil, grpc.Errorf(codes.Unknown, "geolocation error: %s", err)
	}

	var devEUI lorawan.EUI64
	copy(devEUI[:], req.DevEui)

	if len(tdoaResp.Warnings) != 0 {
		log.WithFields(log.Fields{
			"dev_eui":  devEUI,
			"warnings": tdoaResp.Warnings,
		}).Warning("backend/lora_cloud: backend returned warnings")
	}

	if len(tdoaResp.Errors) != 0 {
		log.WithFields(log.Fields{
			"dev_eui": devEUI,
			"errors":  tdoaResp.Errors,
		}).Error("backend/lora_cloud: backend returned errors")

		return nil, grpc.Errorf(codes.Internal, "backend returned errors: %v", tdoaResp.Errors)
	}

	return &geo.ResolveMultiFrameTDOAResponse{
		Result: &geo.ResolveResult{
			Location: &common.Location{
				Source:    common.LocationSource_GEO_RESOLVER,
				Accuracy:  uint32(tdoaResp.Result.Accuracy),
				Latitude:  tdoaResp.Result.Latitude,
				Longitude: tdoaResp.Result.Longitude,
				Altitude:  tdoaResp.Result.Altitude,
			},
		},
	}, nil
}

func (b *Backend) resolveTDOA(ctx context.Context, tdoaReq tdoaRequest) (response, error) {
	d := loRaCloudAPIDuration("v2_tdoa")
	start := time.Now()
	resp, err := b.loRaCloudAPIRequest(ctx, tdoaEndpoint, tdoaReq)
	d.Observe(float64(time.Since(start)) / float64(time.Second))
	return resp, err
}

func (b *Backend) resolveTDOAMultiFrame(ctx context.Context, tdoaMultiFrameReq tdoaMultiFrameRequest) (response, error) {
	d := loRaCloudAPIDuration("v2_tdoa_multiframe")
	start := time.Now()
	resp, err := b.loRaCloudAPIRequest(ctx, tdoaMultiFrameEndpoint, tdoaMultiFrameReq)
	d.Observe(float64(time.Since(start)) / float64(time.Second))
	return resp, err
}

func (b *Backend) loRaCloudAPIRequest(ctx context.Context, endpoint string, v interface{}) (response, error) {
	endpoint = fmt.Sprintf(endpoint, b.uri)
	var resolveResp response

	bb, err := json.Marshal(v)
	if err != nil {
		return resolveResp, errors.Wrap(err, "marshal request error")
	}

	req, err := http.NewRequest("POST", endpoint, bytes.NewReader(bb))
	if err != nil {
		return resolveResp, errors.Wrap(err, "new request error")
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Ocp-Apim-Subscription-Key", b.token)

	reqCTX, cancel := context.WithTimeout(ctx, b.requestTimeout)
	defer cancel()

	req = req.WithContext(reqCTX)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return resolveResp, errors.Wrap(err, "http request error")
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bb, _ := ioutil.ReadAll(resp.Body)
		return resolveResp, fmt.Errorf("expected 200, got: %d (%s)", resp.StatusCode, string(bb))
	}

	if err = json.NewDecoder(resp.Body).Decode(&resolveResp); err != nil {
		return resolveResp, errors.Wrap(err, "unmarshal response error")
	}

	return resolveResp, nil
}
