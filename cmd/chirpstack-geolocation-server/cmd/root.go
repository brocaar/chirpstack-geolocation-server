package cmd

import (
	"bytes"
	"io/ioutil"
	"reflect"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/brocaar/chirpstack-geolocation-server/internal/config"
)

var cfgFile string
var version string

var rootCmd = &cobra.Command{
	Use:   "chirpstack-geolocation-server",
	Short: "ChirpStack Geolocation Server",
	Long: `ChirpStack Geolocation Server is an open-source Geolocation Server, part of the ChirpStack LoRaWAN Network Server stack.
	> documentation & support: https://www.chirpstack.io/geolocation-server/
	> source & copyright information: https://github.com/brocaar/chirpstack-geolocation-server/`,
	RunE: run,
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "path to configuration file (optional)")
	rootCmd.PersistentFlags().Int("log-level", 4, "debug=5, info=4, error=2, fatal=1, panic=0")

	viper.BindPFlag("general.log_level", rootCmd.PersistentFlags().Lookup("log-level"))

	viper.SetDefault("geo_server.api.bind", "0.0.0.0:8005")
	viper.SetDefault("geo_server.backend.type", "collos")
	viper.SetDefault("geo_server.backend.collos.request_timeout", time.Second)
	viper.SetDefault("geo_server.backend.lora_cloud.request_timeout", time.Second)

	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(configfileCmd)
	rootCmd.AddCommand(testResolveTDOA)
	rootCmd.AddCommand(testResolveMultiFrameTDOA)
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
		viper.SetConfigName("chirpstack-geolocation-server")
		viper.AddConfigPath(".")
		viper.AddConfigPath("$HOME/.config/chirpstack-geolocation-server")
		viper.AddConfigPath("/etc/chirpstack-geolocation-server")
		if err := viper.ReadInConfig(); err != nil {
			switch err.(type) {
			case viper.ConfigFileNotFoundError:
				log.Warning("No configuration file found, using defaults. See: https://www.chirpstack.io/geolocation-server/install/config/")
			default:
				log.WithError(err).Fatal("read configuration file error")
			}
		}
	}

	viperBindEnvs(config.C)

	if err := viper.Unmarshal(&config.C); err != nil {
		log.WithError(err).Fatal("unmarshal config error")
	}
}

func viperBindEnvs(iface interface{}, parts ...string) {
	ifv := reflect.ValueOf(iface)
	ift := reflect.TypeOf(iface)
	for i := 0; i < ift.NumField(); i++ {
		v := ifv.Field(i)
		t := ift.Field(i)
		tv, ok := t.Tag.Lookup("mapstructure")
		if !ok {
			tv = strings.ToLower(t.Name)
		}
		if tv == "-" {
			continue
		}

		switch v.Kind() {
		case reflect.Struct:
			viperBindEnvs(v.Interface(), append(parts, tv)...)
		default:
			key := strings.Join(append(parts, tv), ".")
			viper.BindEnv(key)
		}
	}
}
