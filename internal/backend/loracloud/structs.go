package loracloud

import (
	"errors"

	"github.com/golang/protobuf/ptypes"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"github.com/brocaar/chirpstack-api/go/v3/geo"
	"github.com/brocaar/chirpstack-api/go/v3/gw"
	"github.com/brocaar/lorawan"
)

type tdoaRequest struct {
	LoRaWAN []loRaWANRX `json:"lorawan"`
}

type tdoaMultiFrameRequest struct {
	LoRaWAN [][]loRaWANRX `json:"lorawan"`
}

type loRaWANRX struct {
	GatewayID       string          `json:"gatewayId"`
	AntennaID       int             `json:"antennaId"`
	RSSI            int             `json:"rssi"`
	SNR             float64         `json:"snr"`
	TOA             int             `json:"toa,omitempty"`
	AntennaLocation antennaLocation `json:"antennaLocation"`
}

type antennaLocation struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Altitude  float64 `json:"altitude"`
}

type response struct {
	Result   result   `json:"result"`
	Warnings []string `json:"warnings"`
	Errors   []string `json:"errors"`
}

type result struct {
	Latitude                 float64 `json:"latitude"`
	Longitude                float64 `json:"longitude"`
	Altitude                 float64 `json:"altitude"`
	Accuracy                 float64 `json:"accuracy"`
	AlgorithmType            string  `json:"algorithmType"`
	NumberOfGatewaysReceived int     `json:"numberOfGatewaysReceived"`
	NumberOfGatewaysUsed     int     `json:"numberOfGatewaysUsed"`
}

const tdoaEndpoint = `%s/api/v2/tdoa`
const tdoaMultiFrameEndpoint = `%s/api/v2/tdoaMultiframe`

func resolveTDOARequestToLoRaCloudRequest(req *geo.ResolveTDOARequest) (tdoaRequest, error) {
	var tdoaReq tdoaRequest
	var err error

	if req.FrameRxInfo == nil {
		return tdoaReq, errors.New("frame_rx_info must not be nil")
	}

	var devEUI lorawan.EUI64
	copy(devEUI[:], req.DevEui)

	tdoaReq.LoRaWAN, err = rxInfoToLoRaCloud(devEUI, req.FrameRxInfo.RxInfo)
	if err != nil {
		return tdoaReq, err
	}

	return tdoaReq, nil
}

func resolveMutiFrameTDOARequestToLoRaCloudRequest(req *geo.ResolveMultiFrameTDOARequest) (tdoaMultiFrameRequest, error) {
	var tdoaReq tdoaMultiFrameRequest

	var devEUI lorawan.EUI64
	copy(devEUI[:], req.DevEui)

	for _, frame := range req.FrameRxInfoSet {
		lw, err := rxInfoToLoRaCloud(devEUI, frame.RxInfo)
		if err != nil {
			return tdoaReq, err
		}

		tdoaReq.LoRaWAN = append(tdoaReq.LoRaWAN, lw)
	}

	return tdoaReq, nil
}

func rxInfoToLoRaCloud(devEUI lorawan.EUI64, rxInfo []*gw.UplinkRXInfo) ([]loRaWANRX, error) {
	var out []loRaWANRX

	for _, rxInfo := range rxInfo {
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

			out = append(out, rx)
		}
	}

	return out, nil
}
