package cmd

import (
	"bytes"
	"io/ioutil"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/brocaar/lora-geo-server/internal/config"
)

var cfgFile string
var version string

var rootCmd = &cobra.Command{
	Use:   "lora-geo-server",
	Short: "LoRa Geo Server for LoRa Server",
	Long: `LoRa Geolocation Server provides geolocation services for LoRa Server
	> documentation & support: https://www.loraserver.io/lora-geo-server/
	> source & copyright information: https://github.com/brocaar/lora-geo-server/`,
	RunE: run,
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "path to configuration file (optional)")
	rootCmd.PersistentFlags().Int("log-level", 4, "debug=5, info=4, error=2, fatal=1, panic=0")

	viper.BindPFlag("general.log_level", rootCmd.PersistentFlags().Lookup("log-level"))

	viper.SetDefault("geo_server.api.bind", "0.0.0.0:8005")
	viper.SetDefault("geo_server.backend.name", "collos")
	viper.SetDefault("geo_server.backend.collos.request_timeout", time.Second)

	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(configfileCmd)
	rootCmd.AddCommand(testResolveTDOA)
}

// Execute executes the root command.
func Execute(v string) {
	version = v
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func initConfig() {
	config.Version = version

	if cfgFile != "" {
		b, err := ioutil.ReadFile(cfgFile)
		if err != nil {
			log.WithError(err).WithField("config", cfgFile).Fatal("error loading config file")
		}

		viper.SetConfigType("toml")
		if err := viper.ReadConfig(bytes.NewBuffer(b)); err != nil {
			log.WithError(err).WithField("config", cfgFile).Fatal("error loading config file")
		}
	} else {
		viper.SetConfigName("lora-geo-server")
		viper.AddConfigPath(".")
		viper.AddConfigPath("$HOME/.config/lora-geo-server")
		viper.AddConfigPath("/etc/lora-geo-server")
		if err := viper.ReadInConfig(); err != nil {
			switch err.(type) {
			case viper.ConfigFileNotFoundError:
				log.Warning("No configuration file found, using defaults. See: https://www.loraserver.io/lora-geo-server/install/config/")
			default:
				log.WithError(err).Fatal("read configuration file error")
			}
		}
	}

	if err := viper.Unmarshal(&config.C); err != nil {
		log.WithError(err).Fatal("unmarshal config error")
	}
}
