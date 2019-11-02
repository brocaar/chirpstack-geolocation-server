package cmd

import (
	"os"
	"text/template"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/brocaar/chirpstack-geolocation-server/internal/config"
)

const configTemplate = `[general]
# Log level
#
# debug=5, info=4, warning=3, error=2, fatal=1, panic=0
log_level={{ .General.LogLevel }}

# Geolocation-server configuration.
[geo_server]
  # Geolocation API.
  #
  # This is the geolocation API that can be used by ChirpStack Network Server.
  [geo_server.api]
  # ip:port to bind the api server
  bind="{{ .GeoServer.API.Bind }}"

  # CA certificate used by the api server (optional)
  ca_cert="{{ .GeoServer.API.CACert }}"

  # TLS certificate used by the api server (optional)
  tls_cert="{{ .GeoServer.API.TLSCert }}"

  # TLS key used by the api server (optional)
  tls_key="{{ .GeoServer.API.TLSKey }}"


  # Geolocation backend configuration.
  [geo_server.backend]
  # Type.
  #
  # The backend type to use.
  type="{{ .GeoServer.Backend.Type }}"

  # Request log directory.
  #
  # Logging requests can be used to "replay" geolocation requests and to compare
  # different geolocation backends. When left blank, logging will be disabled.
  request_log_dir="{{ .GeoServer.Backend.RequestLogDir }}"

    # Collos backend.
    [geo_server.backend.collos]
    # Collos subscription key.
    #
    # This key can be retrieved after creating a Collos account at:
    # http://preview.collos.org/
    subscription_key="{{ .GeoServer.Backend.Collos.SubscriptionKey }}"

    # Request timeout.
    #
    # This defines the request timeout when making calls to the Collos API.
    request_timeout="{{ .GeoServer.Backend.Collos.RequestTimeout }}"


    # LoRa Cloud backend.
    #
    # Please see https://www.loracloud.com/ for more information about this
    # geolocation service.
    [geo_server.backend.lora_cloud]
    # API URI.
    #
    # The URI of the Geolocation API. This URI can be found under
    # 'Token Management'.
    uri="{{ .GeoServer.Backend.LoRaCloud.URI }}"

    # API token.
    token="{{ .GeoServer.Backend.LoRaCloud.Token }}"

    # Request timeout.
    #
    # This defines the request timeout when making calls to the LoRa Cloud API.
    request_timeout="{{ .GeoServer.Backend.LoRaCloud.RequestTimeout }}"


# Prometheus metrics settings.
[metrics.prometheus]
# Enable Prometheus metrics endpoint.
endpoint_enabled={{ .Metrics.Prometheus.EndpointEnabled }}

# The ip:port to bind the Prometheus metrics server to for serving the
# metrics endpoint.
bind="{{ .Metrics.Prometheus.Bind }}"

# API timing histogram.
#
# By setting this to true, the API request timing histogram will be enabled.
# See also: https://github.com/grpc-ecosystem/go-grpc-prometheus#histograms
api_timing_histogram={{ .Metrics.Prometheus.APITimingHistogram }}
`

var configfileCmd = &cobra.Command{
	Use:   "configfile",
	Short: "Print the LoRa Geolocation Server configuration file",
	RunE: func(cmd *cobra.Command, args []string) error {
		t := template.Must(template.New("config").Parse(configTemplate))
		err := t.Execute(os.Stdout, &config.C)
		if err != nil {
			return errors.Wrap(err, "execute config template error")
		}
		return nil
	},
}
