package collos

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"github.com/brocaar/lora-geo-server/internal/config"
	"github.com/brocaar/loraserver/api/common"
	"github.com/brocaar/loraserver/api/geo"
	"github.com/brocaar/loraserver/api/gw"
	"github.com/brocaar/lorawan"
)

// Backend implements the Collos geolocation backend.
type Backend struct {
	subscriptionKey string
	requestTimeout  time.Duration
}

// NewBackend creates a new Collos backend.
func NewBackend(c config.Config) (geo.GeolocationServerServiceServer, error) {
	return &Backend{
		subscriptionKey: c.GeoServer.Backend.Collos.SubscriptionKey,
		requestTimeout:  c.GeoServer.Backend.Collos.RequestTimeout,
	}, nil
}

// ResolveTDOA resolves the location based on TDOA.
func (b *Backend) ResolveTDOA(ctx context.Context, req *geo.ResolveTDOARequest) (*geo.ResolveTDOAResponse, error) {
	if req.FrameRxInfo == nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "frame_rx_info must not be nil")
	}

	var devEUI lorawan.EUI64
	var todaReq tdoaRequest
	copy(devEUI[:], req.DevEui)

	for _, rxInfo := range req.FrameRxInfo.RxInfo {
		var gatewayID lorawan.EUI64
		copy(gatewayID[:], rxInfo.GatewayId)

		if rxInfo.Location == nil {
			log.WithFields(log.Fields{
				"dev_eui":    devEUI,
				"gateway_id": gatewayID,
			}).Warning("location is nil, ignoring gateway")
			continue
		}

		if rxInfo.FineTimestampType == gw.FineTimestampType_NONE {
			log.WithFields(log.Fields{
				"dev_eui":             devEUI,
				"fine_timestamp_type": rxInfo.FineTimestampType,
				"gateway_id":          gatewayID,
			}).Warning("unsupported fine-typestamp type")
			continue
		}

		rx := loRaWANRX{
			AntennaID: int(rxInfo.Antenna),
			RSSI:      int(rxInfo.Rssi),
			SNR:       rxInfo.LoraSnr,
			AntennaLocation: antennaLocation{
				Latitude:  rxInfo.Location.Latitude,
				Longitude: rxInfo.Location.Longitude,
				Altitude:  rxInfo.Location.Altitude,
			},
		}

		if rxInfo.FineTimestampType == gw.FineTimestampType_PLAIN {
			plainTS := rxInfo.GetPlainFineTimestamp()
			if plainTS == nil {
				log.WithFields(log.Fields{
					"dev_eui":    devEUI,
					"gateway_id": gatewayID,
				}).Warning("plain_fine_timestamp must not be nil")
				continue
			}

			ts, err := ptypes.Timestamp(plainTS.Time)
			if err != nil {
				return nil, grpc.Errorf(codes.InvalidArgument, "timestamp error: %s", err)
			}

			rx.TOA = ts.Nanosecond()
			rx.GatewayID = gatewayID.String()
		}

		if rxInfo.FineTimestampType == gw.FineTimestampType_ENCRYPTED {
			encryptedTS := rxInfo.GetEncryptedFineTimestamp()
			if encryptedTS == nil {
				log.WithFields(log.Fields{
					"dev_eui":    devEUI,
					"gateway_id": gatewayID,
				}).Warning("encrypted_fine_timestamp must not be nil")
				continue
			}

			if len(encryptedTS.FpgaId) == 0 {
				log.WithFields(log.Fields{
					"dev_eui":    devEUI,
					"gateway_id": gatewayID,
				}).Warning("fpga_id must not be nil")
				continue
			}

			rx.GatewayID = fmt.Sprintf("%#x", encryptedTS.FpgaId)
			rx.EncryptedTOA = base64.StdEncoding.EncodeToString(encryptedTS.EncryptedNs)
		}

		todaReq.LoRaWAN = append(todaReq.LoRaWAN, rx)
	}

	if len(todaReq.LoRaWAN) < 3 {
		return nil, grpc.Errorf(codes.InvalidArgument, "not enough meta-data for geolocation")
	}

	tdoaResp, err := b.resolveTDOA(ctx, todaReq)
	if err != nil {
		return nil, grpc.Errorf(codes.Unknown, "geolocation error: %s", err)
	}

	if len(tdoaResp.Warnings) != 0 {
		log.WithFields(log.Fields{
			"dev_eui":  devEUI,
			"warnings": tdoaResp.Warnings,
		}).Warning("backend returned warnings")
	}

	if len(tdoaResp.Errors) != 0 {
		log.WithFields(log.Fields{
			"dev_eui": devEUI,
			"errors":  tdoaResp.Errors,
		}).Error("backend returned errors")
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

func (b *Backend) resolveTDOA(ctx context.Context, tdoaReq tdoaRequest) (response, error) {
	var resolveResp response

	bb, err := json.Marshal(tdoaReq)
	if err != nil {
		return resolveResp, errors.Wrap(err, "marshal request error")
	}

	req, err := http.NewRequest("POST", tdoaEndpoint, bytes.NewReader(bb))
	if err != nil {
		return resolveResp, errors.Wrap(err, "new request error")
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Ocp-Apim-Subscription-Key", b.subscriptionKey)

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
