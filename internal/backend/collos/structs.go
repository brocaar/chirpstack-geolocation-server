package collos

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
