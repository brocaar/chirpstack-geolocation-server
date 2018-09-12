package collos

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"
)

type tdoaRequest struct {
	LoRaWAN []loRaWANRX `json:"lorawan"`
}

type loRaWANRX struct {
	GatewayID       string          `json:"gatewayId"`
	AntennaID       int             `json:"antennaId"`
	RSSI            int             `json:"rssi"`
	SNR             float64         `json:"snr"`
	TOA             int             `json:"toa,omitempty"`
	EncryptedTOA    string          `json:"encryptedToa,omitempty"`
	AntennaLocation antennaLocation `json:"antennaLocation"`
}

type antennaLocation struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Altitude  float64 `json:"altitude"`
}

type response struct {
	Result        result   `json:"result"`
	Warnings      []string `json:"warnings"`
	Errors        []string `json:"errors"`
	CorrelationID string   `json:"correlationId"`
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

var tdoaEndpoint = "https://api.preview.collos.org/semtech-localization-algorithms/v2/tdoa"

func resolveTDOA(ctx context.Context, config Config, resolveReq tdoaRequest) (response, error) {
	var resolveResp response

	b, err := json.Marshal(resolveReq)
	if err != nil {
		return resolveResp, errors.Wrap(err, "marshal request error")
	}
	req, err := http.NewRequest("POST", tdoaEndpoint, bytes.NewReader(b))
	if err != nil {
		return resolveResp, errors.Wrap(err, "new request error")
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Ocp-Apim-Subscription-Key", config.SubscriptionKey)

	reqCTX, cancel := context.WithTimeout(ctx, config.RequestTimeout)
	defer cancel()

	req = req.WithContext(reqCTX)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return resolveResp, errors.Wrap(err, "http request error")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := ioutil.ReadAll(resp.Body)
		return resolveResp, fmt.Errorf("expected 200, got: %d (%s)", resp.StatusCode, string(b))
	}

	if err = json.NewDecoder(resp.Body).Decode(&resolveResp); err != nil {
		return resolveResp, errors.Wrap(err, "unmarshal response error")
	}

	return resolveResp, nil
}
