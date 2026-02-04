package repository

// ConfigBinary represents a binary from config file for syncing
type ConfigBinary struct {
	ID           string
	Name         string
	Alias        string
	Provider     string
	Path         string
	InstallPath  string
	Format       string
	AssetRegex   string
	ReleaseRegex string
}
