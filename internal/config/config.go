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
		} `mapstructure:"backend"`
	} `mapstructure:"geo_server"`
}

// C holds the global configufation.
var C Config
