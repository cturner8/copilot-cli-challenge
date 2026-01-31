package database

const InitialSchema = `
-- Enable foreign keys
PRAGMA foreign_keys = ON;

-- Schema migrations tracking
CREATE TABLE IF NOT EXISTS migrations (
    version INTEGER PRIMARY KEY,
    applied_at INTEGER NOT NULL,
    description TEXT NOT NULL
);

-- Binaries table
CREATE TABLE IF NOT EXISTS binaries (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id TEXT UNIQUE,
    name TEXT NOT NULL,
    alias TEXT,
    provider TEXT NOT NULL,
    provider_path TEXT NOT NULL,
    install_path TEXT,
    format TEXT NOT NULL,
    asset_regex TEXT,
    release_regex TEXT,
    config_digest TEXT NOT NULL DEFAULT '',
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,
    config_version INTEGER NOT NULL DEFAULT 1
);

CREATE INDEX IF NOT EXISTS idx_binaries_user_id ON binaries(user_id);
CREATE INDEX IF NOT EXISTS idx_binaries_provider ON binaries(provider);
CREATE INDEX IF NOT EXISTS idx_binaries_name ON binaries(name);
CREATE INDEX IF NOT EXISTS idx_binaries_alias ON binaries(alias) WHERE alias IS NOT NULL;

-- Installations table
CREATE TABLE IF NOT EXISTS installations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    binary_id INTEGER NOT NULL,
    version TEXT NOT NULL,
    installed_path TEXT NOT NULL,
    source_url TEXT NOT NULL,
    file_size INTEGER NOT NULL,
    checksum TEXT NOT NULL,
    checksum_algorithm TEXT NOT NULL DEFAULT 'SHA256',
    installed_at INTEGER NOT NULL,
    
    FOREIGN KEY (binary_id) REFERENCES binaries(id) ON DELETE CASCADE,
    UNIQUE(binary_id, version)
);

CREATE INDEX IF NOT EXISTS idx_installations_binary_id ON installations(binary_id);
CREATE INDEX IF NOT EXISTS idx_installations_installed_at ON installations(installed_at);
CREATE INDEX IF NOT EXISTS idx_installations_binary_installed ON installations(binary_id, installed_at DESC);

-- Active versions table
CREATE TABLE IF NOT EXISTS versions (
    binary_id INTEGER PRIMARY KEY,
    installation_id INTEGER NOT NULL,
    activated_at INTEGER NOT NULL,
    symlink_path TEXT NOT NULL,
    
    FOREIGN KEY (binary_id) REFERENCES binaries(id) ON DELETE CASCADE,
    FOREIGN KEY (installation_id) REFERENCES installations(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_versions_installation_id ON versions(installation_id);

-- Cached downloads table
CREATE TABLE IF NOT EXISTS downloads (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    binary_id INTEGER NOT NULL,
    version TEXT NOT NULL,
    cache_path TEXT NOT NULL UNIQUE,
    source_url TEXT NOT NULL,
    file_size INTEGER NOT NULL,
    checksum TEXT NOT NULL,
    checksum_algorithm TEXT NOT NULL DEFAULT 'SHA256',
    downloaded_at INTEGER NOT NULL,
    last_accessed_at INTEGER NOT NULL,
    is_complete INTEGER NOT NULL DEFAULT 1,
    
    FOREIGN KEY (binary_id) REFERENCES binaries(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_downloads_binary_id ON downloads(binary_id);
CREATE INDEX IF NOT EXISTS idx_downloads_last_accessed ON downloads(last_accessed_at);
CREATE INDEX IF NOT EXISTS idx_downloads_is_complete ON downloads(is_complete);
CREATE INDEX IF NOT EXISTS idx_downloads_binary_version ON downloads(binary_id, version);
CREATE INDEX IF NOT EXISTS idx_downloads_file_size ON downloads(file_size DESC);

-- Operation logs table
CREATE TABLE IF NOT EXISTS logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    operation_type TEXT NOT NULL,
    operation_status TEXT NOT NULL,
    entity_type TEXT,
    entity_id TEXT,
    message TEXT,
    error_details TEXT,
    metadata TEXT,
    timestamp INTEGER NOT NULL,
    duration_ms INTEGER,
    user_context TEXT
);

CREATE INDEX IF NOT EXISTS idx_logs_timestamp ON logs(timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_logs_operation_type ON logs(operation_type);
CREATE INDEX IF NOT EXISTS idx_logs_status ON logs(operation_status);
CREATE INDEX IF NOT EXISTS idx_logs_entity ON logs(entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_logs_entity_timestamp ON logs(entity_type, entity_id, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_logs_status_type_timestamp ON logs(operation_status, operation_type, timestamp DESC);

-- Record initial migration
INSERT OR IGNORE INTO migrations (version, applied_at, description)
VALUES (1, strftime('%s', 'now'), 'Initial schema');
`
