package config

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/viper"
)

type ConfigFlags struct {
	ConfigPath string
	LogLevel   string
}

func ReadConfig(flags ConfigFlags) Config {
	v := viper.New()

	v.SetConfigName("config")

	envConfigPath, isEnvConfigPathSet := os.LookupEnv("BINMATE_CONFIG_PATH")

	logLevel := flags.LogLevel
	if logLevel == "" {
		logLevel = "silent"
	}

	// set defaults
	v.SetDefault("version", 1)
	v.SetDefault("binaries", []Binary{})
	v.SetDefault("logLevel", logLevel)

	homeDir, _ := os.UserConfigDir()

	if flags.ConfigPath != "" {
		v.SetConfigFile(flags.ConfigPath)
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
