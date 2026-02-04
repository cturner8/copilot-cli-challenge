package config

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/viper"
)

func ReadConfig(flagConfigPath string) Config {
	v := viper.New()

	v.SetConfigName("config")

	envConfigPath, isEnvConfigPathSet := os.LookupEnv("BINMATE_CONFIG_PATH")

	// set defaults
	v.SetDefault("version", 1)
	v.SetDefault("binaries", []Binary{})

	homeDir, _ := os.UserConfigDir()

	if flagConfigPath != "" {
		v.SetConfigFile(flagConfigPath)
	} else if isEnvConfigPathSet {
		v.SetConfigFile(envConfigPath)
	} else {
		// Add search paths to find the file
		v.AddConfigPath(fmt.Sprintf("%s/.binmate", homeDir))
	}

	// Find and read the config file
	v.ReadInConfig()

	var config Config

	// extract config
	err := v.Unmarshal(&config)
	if err != nil {
		log.Fatalf("unable to decode into struct, %v", err)
	}

	return config
}
