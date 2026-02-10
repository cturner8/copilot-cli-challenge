package config

type Binary struct {
	Id   string `mapstructure:"id"`
	Name string `mapstructure:"name"`
	// TODO: implement as override of bin name, if provided
	Alias         string `mapstructure:"alias"`
	Provider      string `mapstructure:"provider"`
	Path          string `mapstructure:"path"`
	InstallPath   string `mapstructure:"installPath"`
	Format        string `mapstructure:"format"`
	AssetRegex    string `mapstructure:"assetRegex"`
	ReleaseRegex  string `mapstructure:"releaseRegex"`
	Authenticated bool   `mapstructure:"authenticated"`
}

// GlobalConfig represents global defaults that apply to all binaries
type GlobalConfig struct {
	InstallPath string                      `mapstructure:"installPath"` // Default install path for all binaries
	Providers   map[string]ProviderDefaults `mapstructure:"providers"`   // Provider-specific defaults (e.g., github.authenticated)
}

// ProviderDefaults represents provider-level configuration defaults
type ProviderDefaults struct {
	Authenticated bool `mapstructure:"authenticated"` // Whether to use authentication for API calls
}

type Config struct {
	Version    int          `mapstructure:"version"`
	Global     GlobalConfig `mapstructure:"global"` // Global configuration defaults
	Binaries   []Binary     `mapstructure:"binaries"`
	DateFormat string       `mapstructure:"dateFormat"` // Date format for display, e.g., "02/01/2006 15:04"
	LogLevel   string       `mapstructure:"logLevel"`
}
