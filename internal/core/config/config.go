package config

type Binary struct {
	Id   string `mapstructure:"id"`
	Name string `mapstructure:"name"`
	// TODO: implement as override of bin name, if provided
	Alias        string `mapstructure:"alias"`
	Provider     string `mapstructure:"provider"`
	Path         string `mapstructure:"path"`
	InstallPath  string `mapstructure:"installPath"`
	Format       string `mapstructure:"format"`
	AssetRegex   string `mapstructure:"assetRegex"`
	ReleaseRegex string `mapstructure:"releaseRegex"`
}

type Config struct {
	Version  int      `mapstructure:"version"`
	Binaries []Binary `mapstructure:"binaries"`
}
