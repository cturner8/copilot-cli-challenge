package database

// Binary represents a binary definition
type Binary struct {
	ID            int64
	UserID        string
	Name          string
	Alias         *string
	Provider      string
	ProviderPath  string
	InstallPath   *string
	Format        string
	AssetRegex    *string
	ReleaseRegex  *string
	ConfigDigest  string // Digest of config fields for change detection (e.g., "sha256:abc123...")
	CreatedAt     int64
	UpdatedAt     int64
	ConfigVersion int
	Source        string // "config" for binaries from config.json, "manual" for user-added binaries
	Authenticated bool   // Whether to use GitHub token for authentication (for private repos or rate limit avoidance)
}

// Installation represents an installed binary version
type Installation struct {
	ID                int64
	BinaryID          int64
	Version           string
	InstalledPath     string
	SourceURL         string
	FileSize          int64
	Checksum          string
	ChecksumAlgorithm string
	InstalledAt       int64
}

// Version represents the active version of a binary
type Version struct {
	BinaryID       int64
	InstallationID int64
	ActivatedAt    int64
	SymlinkPath    string
}

// Download represents a cached download
type Download struct {
	ID                int64
	BinaryID          int64
	Version           string
	CachePath         string
	SourceURL         string
	FileSize          int64
	Checksum          string
	ChecksumAlgorithm string
	DownloadedAt      int64
	LastAccessedAt    int64
	IsComplete        bool
}

// Log represents an operation log entry
type Log struct {
	ID              int64
	OperationType   string
	OperationStatus string
	EntityType      *string
	EntityID        *string
	Message         *string
	ErrorDetails    *string
	Metadata        *string
	Timestamp       int64
	DurationMs      *int64
	UserContext     *string
}
