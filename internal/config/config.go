package config

import "time"

// Version defines the LoRa Geo Server version.
var Version string

// Config defines the configuration structure.
type Config struct {
	General struct {
		LogLevel int `mapstructure:"log_level"`
	}

	GeoServer struct {
		API struct {
			Bind    string
			CACert  string `mapstructure:"ca_cert"`
			TLSCert string `mapstructure:"tls_cert"`
			TLSKey  string `mapstructure:"tls_key"`
		} `mapstructure:"api"`

		Backend struct {
			Type string `mapstructure:"type"`

			RequestLogDir string `mapstructure:"request_log_dir"`

			Collos struct {
				SubscriptionKey string        `mapstructure:"subscription_key"`
				RequestTimeout  time.Duration `mapstructure:"request_timeout"`
			} `mapstructure:"collos"`

			LoRaCloud struct {
				URI            string        `mapstructure:"uri"`
				Token          string        `mapstructure:"token"`
				RequestTimeout time.Duration `mapstructure:"request_timeout"`
			} `mapstructure:"lora_cloud"`
		} `mapstructure:"backend"`
	} `mapstructure:"geo_server"`

	Metrics struct {
		Prometheus struct {
			EndpointEnabled    bool   `mapstructure:"endpoint_enabled"`
			Bind               string `mapstructure:"bind"`
			APITimingHistogram bool   `mapstructure:"api_timing_histogram"`
		} `mapstructure:"prometheus"`
	} `mapstructure:"metrics"`
}

// C holds the global configufation.
var C Config
