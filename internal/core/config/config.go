package config

type Binary struct {
	Id          string `mapstructure:"id"`
	Name        string `mapstructure:"name"`
	Provider    string `mapstructure:"provider"`
	InstallPath string `mapstructure:"installPath"`
	Format      string `mapstructure:"format"`
}

type Config struct {
	Version  int      `mapstructure:"version"`
	Binaries []Binary `mapstructure:"binaries"`
}
