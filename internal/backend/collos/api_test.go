package collos

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang/protobuf/ptypes"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/brocaar/loraserver/api/common"
	"github.com/brocaar/loraserver/api/geo"
	"github.com/brocaar/loraserver/api/gw"
)

type CollosTestSuite struct {
	suite.Suite

	apiResponse string
	apiRequest  string
	apiServer   *httptest.Server

	client geo.GeolocationServerServiceServer
}

func (ts *CollosTestSuite) SetupSuite() {
	log.SetLevel(log.ErrorLevel)

	ts.apiServer = httptest.NewServer(http.HandlerFunc(ts.apiHandler))

	ts.client = &Backend{
		requestTimeout: time.Second,
	}

	tdoaEndpoint = ts.apiServer.URL
	tdoaMultiFrameEndpoint = ts.apiServer.URL
}

func (ts *CollosTestSuite) TearDownSuite() {
	ts.apiServer.Close()
}

func (ts *CollosTestSuite) TestResolveTDOA() {
	ts.apiResponse = `
		{
			"result": {
			"latitude": 1.12345,
			"longitude": 1.22345,
			"altitude": 1.32345,
			"accuracy": 4.5,
			"algorithmType": "a-algorithm",
			"numberOfGatewaysReceived": 4,
			"numberOfGatewaysUsed": 3
			},
			"warnings": [
			],
			"errors": [
			],
			"correlationId": "abcde"
		}
	`

	now := time.Now()
	nowPB, _ := ptypes.TimestampProto(now)

	testTable := []struct {
		Name    string
		Request geo.ResolveTDOARequest

		ExpectedError    error
		ExpectedRequest  *tdoaRequest
		ExpectedResponse *geo.ResolveTDOAResponse
	}{
		{
			Name: "valid decrypted timestamp request",
			Request: geo.ResolveTDOARequest{
				DevEui: []byte{1, 2, 3, 4, 5, 6, 7, 8},
				FrameRxInfo: &geo.FrameRXInfo{
					RxInfo: []*gw.UplinkRXInfo{
						{
							GatewayId: []byte{1, 1, 1, 1, 1, 1, 1, 1},
							Location: &common.Location{
								Latitude:  1.1,
								Longitude: 1.2,
								Altitude:  1.3,
							},
							FineTimestampType: gw.FineTimestampType_PLAIN,
							FineTimestamp: &gw.UplinkRXInfo_PlainFineTimestamp{
								PlainFineTimestamp: &gw.PlainFineTimestamp{
									Time: nowPB,
								},
							},
						},
						{
							GatewayId: []byte{2, 1, 1, 1, 1, 1, 1, 1},
							Location: &common.Location{
								Latitude:  2.1,
								Longitude: 2.2,
								Altitude:  2.3,
							},
							FineTimestampType: gw.FineTimestampType_PLAIN,
							FineTimestamp: &gw.UplinkRXInfo_PlainFineTimestamp{
								PlainFineTimestamp: &gw.PlainFineTimestamp{
									Time: nowPB,
								},
							},
						},
						{
							GatewayId: []byte{3, 1, 1, 1, 1, 1, 1, 1},
							Location: &common.Location{
								Latitude:  3.1,
								Longitude: 3.2,
								Altitude:  3.3,
							},
							FineTimestampType: gw.FineTimestampType_PLAIN,
							FineTimestamp: &gw.UplinkRXInfo_PlainFineTimestamp{
								PlainFineTimestamp: &gw.PlainFineTimestamp{
									Time: nowPB,
								},
							},
						},
					},
				},
			},
			ExpectedRequest: &tdoaRequest{
				LoRaWAN: []loRaWANRX{
					{
						GatewayID: "0101010101010101",
						TOA:       now.Nanosecond(),
						AntennaLocation: antennaLocation{
							Latitude:  1.1,
							Longitude: 1.2,
							Altitude:  1.3,
						},
					},
					{
						GatewayID: "0201010101010101",
						TOA:       now.Nanosecond(),
						AntennaLocation: antennaLocation{
							Latitude:  2.1,
							Longitude: 2.2,
							Altitude:  2.3,
						},
					},
					{
						GatewayID: "0301010101010101",
						TOA:       now.Nanosecond(),
						AntennaLocation: antennaLocation{
							Latitude:  3.1,
							Longitude: 3.2,
							Altitude:  3.3,
						},
					},
				},
			},
			ExpectedResponse: &geo.ResolveTDOAResponse{
				Result: &geo.ResolveResult{
					Location: &common.Location{
						Latitude:  1.12345,
						Longitude: 1.22345,
						Altitude:  1.32345,
						Source:    common.LocationSource_GEO_RESOLVER,
						Accuracy:  4,
					},
				},
			},
		},
		{
			Name: "valid encrypted timestamp request",
			Request: geo.ResolveTDOARequest{
				DevEui: []byte{1, 2, 3, 4, 5, 6, 7, 8},
				FrameRxInfo: &geo.FrameRXInfo{
					RxInfo: []*gw.UplinkRXInfo{
						{
							GatewayId: []byte{1, 1, 1, 1, 1, 1, 1, 1},
							Location: &common.Location{
								Latitude:  1.1,
								Longitude: 1.2,
								Altitude:  1.3,
							},
							FineTimestampType: gw.FineTimestampType_ENCRYPTED,
							FineTimestamp: &gw.UplinkRXInfo_EncryptedFineTimestamp{
								EncryptedFineTimestamp: &gw.EncryptedFineTimestamp{
									FpgaId:      []byte{1},
									EncryptedNs: []byte{1, 1, 1, 1},
								},
							},
						},
						{
							GatewayId: []byte{2, 1, 1, 1, 1, 1, 1, 1},
							Location: &common.Location{
								Latitude:  2.1,
								Longitude: 2.2,
								Altitude:  2.3,
							},
							FineTimestampType: gw.FineTimestampType_ENCRYPTED,
							FineTimestamp: &gw.UplinkRXInfo_EncryptedFineTimestamp{
								EncryptedFineTimestamp: &gw.EncryptedFineTimestamp{
									FpgaId:      []byte{2},
									EncryptedNs: []byte{2, 1, 1, 1},
								},
							},
						},
						{
							GatewayId: []byte{3, 1, 1, 1, 1, 1, 1, 1},
							Location: &common.Location{
								Latitude:  3.1,
								Longitude: 3.2,
								Altitude:  3.3,
							},
							FineTimestampType: gw.FineTimestampType_ENCRYPTED,
							FineTimestamp: &gw.UplinkRXInfo_EncryptedFineTimestamp{
								EncryptedFineTimestamp: &gw.EncryptedFineTimestamp{
									FpgaId:      []byte{3},
									EncryptedNs: []byte{3, 1, 1, 1},
								},
							},
						},
					},
				},
			},
			ExpectedRequest: &tdoaRequest{
				LoRaWAN: []loRaWANRX{
					{
						GatewayID:    "0x01",
						EncryptedTOA: "AQEBAQ==",
						AntennaLocation: antennaLocation{
							Latitude:  1.1,
							Longitude: 1.2,
							Altitude:  1.3,
						},
					},
					{
						GatewayID:    "0x02",
						EncryptedTOA: "AgEBAQ==",
						AntennaLocation: antennaLocation{
							Latitude:  2.1,
							Longitude: 2.2,
							Altitude:  2.3,
						},
					},
					{
						GatewayID:    "0x03",
						EncryptedTOA: "AwEBAQ==",
						AntennaLocation: antennaLocation{
							Latitude:  3.1,
							Longitude: 3.2,
							Altitude:  3.3,
						},
					},
				},
			},
			ExpectedResponse: &geo.ResolveTDOAResponse{
				Result: &geo.ResolveResult{
					Location: &common.Location{
						Latitude:  1.12345,
						Longitude: 1.22345,
						Altitude:  1.32345,
						Source:    common.LocationSource_GEO_RESOLVER,
						Accuracy:  4,
					},
				},
			},
		},
	}

	for _, test := range testTable {
		ts.T().Run(test.Name, func(t *testing.T) {
			assert := require.New(t)

			resp, err := ts.client.ResolveTDOA(context.Background(), &test.Request)
			assert.Equal(test.ExpectedError, err)

			if test.ExpectedResponse != nil {
				assert.Equal(test.ExpectedResponse, resp)
			}

			if test.ExpectedRequest != nil {
				var req tdoaRequest
				assert.NoError(json.Unmarshal([]byte(ts.apiRequest), &req))
				assert.Equal(test.ExpectedRequest, &req)
			}
		})
	}
}

func (ts *CollosTestSuite) TestResolveMultiFrameTDOA() {
	ts.apiResponse = `
		{
			"result": {
			"latitude": 1.12345,
			"longitude": 1.22345,
			"altitude": 1.32345,
			"accuracy": 4.5,
			"algorithmType": "a-algorithm",
			"numberOfGatewaysReceived": 4,
			"numberOfGatewaysUsed": 3
			},
			"warnings": [
			],
			"errors": [
			],
			"correlationId": "abcde"
		}
	`

	now := time.Now()
	nowPB, _ := ptypes.TimestampProto(now)

	testTable := []struct {
		Name    string
		Request geo.ResolveMultiFrameTDOARequest

		ExpectedError    error
		ExpectedRequest  *tdoaMultiFrameRequest
		ExpectedResponse *geo.ResolveMultiFrameTDOAResponse
	}{
		{
			Name: "valid decrypted timestamp request",
			Request: geo.ResolveMultiFrameTDOARequest{
				DevEui: []byte{1, 2, 3, 4, 5, 6, 7, 8},
				FrameRxInfoSet: []*geo.FrameRXInfo{
					{
						RxInfo: []*gw.UplinkRXInfo{
							{
								GatewayId: []byte{1, 1, 1, 1, 1, 1, 1, 1},
								Location: &common.Location{
									Latitude:  1.1,
									Longitude: 1.2,
									Altitude:  1.3,
								},
								FineTimestampType: gw.FineTimestampType_PLAIN,
								FineTimestamp: &gw.UplinkRXInfo_PlainFineTimestamp{
									PlainFineTimestamp: &gw.PlainFineTimestamp{
										Time: nowPB,
									},
								},
							},
							{
								GatewayId: []byte{2, 1, 1, 1, 1, 1, 1, 1},
								Location: &common.Location{
									Latitude:  2.1,
									Longitude: 2.2,
									Altitude:  2.3,
								},
								FineTimestampType: gw.FineTimestampType_PLAIN,
								FineTimestamp: &gw.UplinkRXInfo_PlainFineTimestamp{
									PlainFineTimestamp: &gw.PlainFineTimestamp{
										Time: nowPB,
									},
								},
							},
							{
								GatewayId: []byte{3, 1, 1, 1, 1, 1, 1, 1},
								Location: &common.Location{
									Latitude:  3.1,
									Longitude: 3.2,
									Altitude:  3.3,
								},
								FineTimestampType: gw.FineTimestampType_PLAIN,
								FineTimestamp: &gw.UplinkRXInfo_PlainFineTimestamp{
									PlainFineTimestamp: &gw.PlainFineTimestamp{
										Time: nowPB,
									},
								},
							},
						},
					},
				},
			},
			ExpectedRequest: &tdoaMultiFrameRequest{
				LoRaWAN: [][]loRaWANRX{
					{
						{
							GatewayID: "0101010101010101",
							TOA:       now.Nanosecond(),
							AntennaLocation: antennaLocation{
								Latitude:  1.1,
								Longitude: 1.2,
								Altitude:  1.3,
							},
						},
						{
							GatewayID: "0201010101010101",
							TOA:       now.Nanosecond(),
							AntennaLocation: antennaLocation{
								Latitude:  2.1,
								Longitude: 2.2,
								Altitude:  2.3,
							},
						},
						{
							GatewayID: "0301010101010101",
							TOA:       now.Nanosecond(),
							AntennaLocation: antennaLocation{
								Latitude:  3.1,
								Longitude: 3.2,
								Altitude:  3.3,
							},
						},
					},
				},
			},
			ExpectedResponse: &geo.ResolveMultiFrameTDOAResponse{
				Result: &geo.ResolveResult{
					Location: &common.Location{
						Latitude:  1.12345,
						Longitude: 1.22345,
						Altitude:  1.32345,
						Source:    common.LocationSource_GEO_RESOLVER,
						Accuracy:  4,
					},
				},
			},
		},
		{
			Name: "valid encrypted timestamp request",
			Request: geo.ResolveMultiFrameTDOARequest{
				DevEui: []byte{1, 2, 3, 4, 5, 6, 7, 8},
				FrameRxInfoSet: []*geo.FrameRXInfo{
					{
						RxInfo: []*gw.UplinkRXInfo{
							{
								GatewayId: []byte{1, 1, 1, 1, 1, 1, 1, 1},
								Location: &common.Location{
									Latitude:  1.1,
									Longitude: 1.2,
									Altitude:  1.3,
								},
								FineTimestampType: gw.FineTimestampType_ENCRYPTED,
								FineTimestamp: &gw.UplinkRXInfo_EncryptedFineTimestamp{
									EncryptedFineTimestamp: &gw.EncryptedFineTimestamp{
										FpgaId:      []byte{1},
										EncryptedNs: []byte{1, 1, 1, 1},
									},
								},
							},
							{
								GatewayId: []byte{2, 1, 1, 1, 1, 1, 1, 1},
								Location: &common.Location{
									Latitude:  2.1,
									Longitude: 2.2,
									Altitude:  2.3,
								},
								FineTimestampType: gw.FineTimestampType_ENCRYPTED,
								FineTimestamp: &gw.UplinkRXInfo_EncryptedFineTimestamp{
									EncryptedFineTimestamp: &gw.EncryptedFineTimestamp{
										FpgaId:      []byte{2},
										EncryptedNs: []byte{2, 1, 1, 1},
									},
								},
							},
							{
								GatewayId: []byte{3, 1, 1, 1, 1, 1, 1, 1},
								Location: &common.Location{
									Latitude:  3.1,
									Longitude: 3.2,
									Altitude:  3.3,
								},
								FineTimestampType: gw.FineTimestampType_ENCRYPTED,
								FineTimestamp: &gw.UplinkRXInfo_EncryptedFineTimestamp{
									EncryptedFineTimestamp: &gw.EncryptedFineTimestamp{
										FpgaId:      []byte{3},
										EncryptedNs: []byte{3, 1, 1, 1},
									},
								},
							},
						},
					},
				},
			},
			ExpectedRequest: &tdoaMultiFrameRequest{
				LoRaWAN: [][]loRaWANRX{
					{
						{
							GatewayID:    "0x01",
							EncryptedTOA: "AQEBAQ==",
							AntennaLocation: antennaLocation{
								Latitude:  1.1,
								Longitude: 1.2,
								Altitude:  1.3,
							},
						},
						{
							GatewayID:    "0x02",
							EncryptedTOA: "AgEBAQ==",
							AntennaLocation: antennaLocation{
								Latitude:  2.1,
								Longitude: 2.2,
								Altitude:  2.3,
							},
						},
						{
							GatewayID:    "0x03",
							EncryptedTOA: "AwEBAQ==",
							AntennaLocation: antennaLocation{
								Latitude:  3.1,
								Longitude: 3.2,
								Altitude:  3.3,
							},
						},
					},
				},
			},
			ExpectedResponse: &geo.ResolveMultiFrameTDOAResponse{
				Result: &geo.ResolveResult{
					Location: &common.Location{
						Latitude:  1.12345,
						Longitude: 1.22345,
						Altitude:  1.32345,
						Source:    common.LocationSource_GEO_RESOLVER,
						Accuracy:  4,
					},
				},
			},
		},
	}

	for _, tst := range testTable {
		ts.T().Run(tst.Name, func(t *testing.T) {
			assert := require.New(t)

			resp, err := ts.client.ResolveMultiFrameTDOA(context.Background(), &tst.Request)
			assert.Equal(tst.ExpectedError, err)

			if tst.ExpectedResponse != nil {
				assert.Equal(tst.ExpectedResponse, resp)
			}

			if tst.ExpectedRequest != nil {
				var req tdoaMultiFrameRequest
				assert.NoError(json.Unmarshal([]byte(ts.apiRequest), &req))
				assert.Equal(tst.ExpectedRequest, &req)
			}
		})
	}
}

func (ts *CollosTestSuite) apiHandler(w http.ResponseWriter, r *http.Request) {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}
	ts.apiRequest = string(b)
	w.Write([]byte(ts.apiResponse))
}

func TestCollos(t *testing.T) {
	suite.Run(t, new(CollosTestSuite))
}
