package config

import (
	"github.com/brocaar/lora-geo-server/internal/backends/collos"
)

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
			Name   string        `mapstructure:"name"`
			Collos collos.Config `mapstructure:"collos"`
		} `mapstructure:"backend"`
	} `mapstructure:"geo_server"`
}

// C holds the global configufation.
var C Config
