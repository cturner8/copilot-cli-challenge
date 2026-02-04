package repository

import (
	"cturner8/binmate/internal/database"
)

// Service provides a convenient wrapper around all repositories
type Service struct {
	DB            *database.DB
	Binaries      *BinariesRepository
	Installations *InstallationsRepository
	Versions      *VersionsRepository
	Downloads     *DownloadsRepository
	Logs          *LogsRepository
}

// NewService creates a new database service
func NewService(db *database.DB) *Service {
	return &Service{
		DB:            db,
		Binaries:      NewBinariesRepository(db),
		Installations: NewInstallationsRepository(db),
		Versions:      NewVersionsRepository(db),
		Downloads:     NewDownloadsRepository(db),
		Logs:          NewLogsRepository(db),
	}
}

// Close closes the database connection
func (s *Service) Close() error {
	return s.DB.Close()
}
