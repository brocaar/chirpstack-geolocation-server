package collos

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/golang/protobuf/ptypes"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"github.com/brocaar/loraserver/api/common"
	"github.com/brocaar/loraserver/api/geo"
	"github.com/brocaar/loraserver/api/gw"
	"github.com/brocaar/lorawan"
)

// API exposes the Collos geolocation- methods.
type API struct {
	config Config
}

// NewAPI creates a new API.
func NewAPI(c Config) geo.GeolocationServerServiceServer {
	return &API{
		config: c,
	}
}

// ResolveTDOA resolves the location based on TDOA.
func (a *API) ResolveTDOA(ctx context.Context, req *geo.ResolveTDOARequest) (*geo.ResolveTDOAResponse, error) {
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

	tdoaResp, err := resolveTDOA(ctx, a.config, todaReq)
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
