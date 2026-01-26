package config

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/viper"
)

func ReadConfig() Config {
	v := viper.New()

	v.SetConfigName("config")

	configPath, configPathSet := os.LookupEnv("BINMATE_CONFIG_PATH")

	// set defaults
	v.SetDefault("version", 1)
	v.SetDefault("binaries", []Binary{})

	homeDir, _ := os.UserHomeDir()

	if configPathSet {
		v.SetConfigFile(configPath)
	} else {
		// Add search paths to find the file
		v.AddConfigPath("/etc/binmate/")
		v.AddConfigPath(fmt.Sprintf("%s/.binmate", homeDir))
	}

	// Find and read the config file
	err := v.ReadInConfig()
	if err != nil {
		log.Fatalf("unable to read config, %v", err)
	}

	// Watch the config for changes
	v.WatchConfig()

	var config Config

	// extract config
	err = v.Unmarshal(&config)
	if err != nil {
		log.Fatalf("unable to decode into struct, %v", err)
	}

	return config
}
