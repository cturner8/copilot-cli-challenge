# binmate SQLite Database Schema Implementation Plan

## Overview

This plan defines the database schema for binmate, a binary version manager CLI application. The schema supports installation tracking, version management, cache management, and comprehensive audit logging.

**Database Location**: `$HOME/.local/share/.binmate/user.db`

---

## Schema Design

### Entity Relationship Summary

```
binaries (1) ──< (N) installations
binaries (1) ──< (N) downloads
installations (1) ──< (1) versions
* ──> (N) logs (audit trail for all entities)
```

---

### Table Definitions

#### 1. `binaries`
Tracks binary definitions (synced from user `config.json`).

```sql
CREATE TABLE binaries (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id TEXT UNIQUE,                    -- aligns with user provided id, e.g., 'ghcp', 'gh', 'bun'
    name TEXT NOT NULL,                     -- e.g., 'copilot', 'gh'
    alias TEXT,                             -- optional override for bin name
    provider TEXT NOT NULL,                 -- 'github' (extensible for future providers)
    provider_path TEXT NOT NULL,            -- e.g., 'github/copilot-cli', 'cli/cli'
    install_path TEXT,                      -- custom install path override
    format TEXT NOT NULL,                   -- '.tar.gz', '.zip'
    asset_regex TEXT,                       -- regex to filter release assets
    release_regex TEXT,                     -- regex to filter releases
    created_at INTEGER NOT NULL,            -- Unix timestamp
    updated_at INTEGER NOT NULL,            -- Unix timestamp
    config_version INTEGER NOT NULL DEFAULT 1  -- track config.json version for sync detection
);

CREATE INDEX idx_binaries_user_id ON binaries(user_id);
CREATE INDEX idx_binaries_provider ON binaries(provider);
CREATE INDEX idx_binaries_name ON binaries(name);
CREATE INDEX idx_binaries_alias ON binaries(alias) WHERE alias IS NOT NULL;
```

**Rationale**:
- Primary key on `id` for internal references
- `user_id` maps to config.json `id` field for sync
- Indexes on `provider`, `name`, and `alias` for list/search operations
- `config_version` enables sync detection with config.json
- Partial index on `alias` only indexes non-null values
- All fields from current `Binary` struct preserved
- **No index on `config_version`**: Table is small (<100 rows typically), sync runs once at startup, and low cardinality (few distinct values) makes index selectivity poor. Full table scan is ~0.1ms; index write overhead isn't justified.

---

#### 2. `installations`
Tracks successfully installed binary versions.

```sql
CREATE TABLE installations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    binary_id INTEGER NOT NULL,             -- FK to binaries.id
    version TEXT NOT NULL,                  -- e.g., 'v1.2.3', '2.0.0'
    installed_path TEXT NOT NULL,           -- full path where binary is installed
    source_url TEXT NOT NULL,               -- download URL from provider
    file_size INTEGER NOT NULL,             -- size in bytes
    checksum TEXT NOT NULL,                 -- SHA256 checksum of installed binary
    checksum_algorithm TEXT NOT NULL DEFAULT 'SHA256', -- 'SHA256', 'MD5', etc.
    installed_at INTEGER NOT NULL,          -- Unix timestamp
    
    FOREIGN KEY (binary_id) REFERENCES binaries(id) ON DELETE CASCADE,
    UNIQUE(binary_id, version)              -- one installation per binary version
);

CREATE INDEX idx_installations_binary_id ON installations(binary_id);
CREATE INDEX idx_installations_installed_at ON installations(installed_at);
CREATE INDEX idx_installations_binary_installed ON installations(binary_id, installed_at DESC);
```

**Rationale**:
- Composite unique constraint prevents duplicate version installations
- `checksum` and `checksum_algorithm` support integrity verification
- Metadata fields (`source_url`, `file_size`) as specified
- `binary_id` is INTEGER to match `binaries.id` FK type
- Indexes optimise queries:
  - `idx_installations_binary_id`: List versions for binary
  - `idx_installations_version`: Get specific version (unique lookup)
  - `idx_installations_installed_at`: Recent installations across all binaries
  - `idx_installations_binary_installed`: Latest version per binary, rollback queries

---

#### 3. `versions`
Tracks which version is currently active (symlinked) for each binary.

```sql
CREATE TABLE versions (
    binary_id INTEGER PRIMARY KEY,          -- FK to binaries.id
    installation_id INTEGER NOT NULL,       -- FK to installations.id
    activated_at INTEGER NOT NULL,          -- Unix timestamp
    symlink_path TEXT NOT NULL,             -- path to the active symlink
    
    FOREIGN KEY (binary_id) REFERENCES binaries(id) ON DELETE CASCADE,
    FOREIGN KEY (installation_id) REFERENCES installations(id) ON DELETE CASCADE
);

CREATE INDEX idx_versions_installation_id ON versions(installation_id);
```

**Rationale**:
- One active version per binary (PRIMARY KEY on binary_id enforces 1:1)
- `binary_id` is INTEGER to match `binaries.id` FK type
- Supports quick lookups: "what version is active?" and "what binary uses this installation?"
- `symlink_path` stores the actual symlink location for verification

---

#### 4. `downloads`
Tracks cached downloaded files (before extraction/installation).

```sql
CREATE TABLE downloads (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    binary_id INTEGER NOT NULL,                -- FK to binaries.id
    version TEXT NOT NULL,                  -- version downloaded
    cache_path TEXT NOT NULL UNIQUE,        -- full path in cache directory
    source_url TEXT NOT NULL,               -- original download URL
    file_size INTEGER NOT NULL,             -- size in bytes
    checksum TEXT NOT NULL,                 -- SHA256 checksum of downloaded asset
    checksum_algorithm TEXT NOT NULL DEFAULT 'SHA256',
    downloaded_at INTEGER NOT NULL,         -- Unix timestamp
    last_accessed_at INTEGER NOT NULL,      -- Unix timestamp (for LRU cleanup)
    is_complete INTEGER NOT NULL DEFAULT 1, -- 0 = incomplete (resume support), 1 = complete
    
    FOREIGN KEY (binary_id) REFERENCES binaries(id) ON DELETE CASCADE
);

CREATE INDEX idx_downloads_binary_id ON downloads(binary_id);
CREATE INDEX idx_downloads_last_accessed ON downloads(last_accessed_at);
CREATE INDEX idx_downloads_is_complete ON downloads(is_complete);
CREATE INDEX idx_downloads_binary_version ON downloads(binary_id, version);
CREATE INDEX idx_downloads_file_size ON downloads(file_size DESC);
```

**Rationale**:
- Separate from `installations` to support cache cleanup without losing installation history
- `last_accessed_at` enables LRU cache eviction strategies
- `is_complete` flag supports download resumption (Phase 4 requirement)
- Indexes support cache management queries:
  - `idx_downloads_binary_id`: List cache entries for binary
  - `idx_downloads_last_accessed`: LRU eviction ordering
  - `idx_downloads_is_complete`: Find incomplete downloads for resume
  - `idx_downloads_binary_version`: Check if specific version is cached
  - `idx_downloads_file_size`: Size-based cache cleanup

---

#### 5. `logs`
Audit trail for all user and background operations.

```sql
CREATE TABLE logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    operation_type TEXT NOT NULL,           -- 'install', 'remove', 'switch', 'update', 'import', 'cleanup', 'sync'
    operation_status TEXT NOT NULL,         -- 'started', 'success', 'failed'
    entity_type TEXT,                       -- 'binary', 'installation', 'cache', 'config'
    entity_id TEXT,                         -- ID of the affected entity
    message TEXT,                           -- human-readable log message
    error_details TEXT,                     -- error stack trace or details (if failed)
    metadata TEXT,                          -- JSON blob for additional context
    timestamp INTEGER NOT NULL,             -- Unix timestamp
    duration_ms INTEGER,                    -- operation duration in milliseconds
    
    -- Optional: track user context (for multi-user systems)
    user_context TEXT                       -- username or UID
);

CREATE INDEX idx_logs_timestamp ON logs(timestamp DESC);
CREATE INDEX idx_logs_operation_type ON logs(operation_type);
CREATE INDEX idx_logs_status ON logs(operation_status);
CREATE INDEX idx_logs_entity ON logs(entity_type, entity_id);
CREATE INDEX idx_logs_entity_timestamp ON logs(entity_type, entity_id, timestamp DESC);
CREATE INDEX idx_logs_status_type_timestamp ON logs(operation_status, operation_type, timestamp DESC);
```

**Rationale**:
- Comprehensive audit trail for debugging and history tracking
- `operation_status` tracks lifecycle: started → success/failed
- `metadata` JSON field provides flexibility for operation-specific data
- `duration_ms` enables performance tracking
- Indexes support varied query patterns:
  - `idx_logs_timestamp`: Recent operations (dashboard, TUI)
  - `idx_logs_operation_type`: Filter by operation type
  - `idx_logs_status`: Filter by status (failures view)
  - `idx_logs_entity`: Operations on specific entity
  - `idx_logs_entity_timestamp`: Entity history with time ordering
  - `idx_logs_status_type_timestamp`: Combined filters (e.g., "failed installs")

---

## Query Patterns & Optimization

### Critical Query Patterns

#### 1. List all installed binaries with active versions
```sql
SELECT 
    b.id, b.name, b.provider,
    i.version, i.installed_path,
    av.activated_at, av.symlink_path
FROM binaries b
LEFT JOIN versions av ON b.id = av.binary_id
LEFT JOIN installations i ON av.installation_id = i.id
ORDER BY b.name;
```
**Optimization**: Indexes on `versions.binary_id` and `installations.id`

---

#### 2. Get all versions for a specific binary
```sql
SELECT 
    version, installed_path, file_size, installed_at,
    CASE WHEN av.installation_id = i.id THEN 1 ELSE 0 END as is_active
FROM installations i
LEFT JOIN versions av ON av.binary_id = i.binary_id
WHERE i.binary_id = ?
ORDER BY i.installed_at DESC;
```
**Optimization**: Index on `installations.binary_id`

---

#### 3. Get active version for a binary
```sql
SELECT i.version, i.installed_path, av.symlink_path
FROM versions av
JOIN installations i ON av.installation_id = i.id
WHERE av.binary_id = ?;
```
**Optimization**: Primary key lookup on `versions.binary_id`

---

#### 4. Find cached downloads for cleanup (LRU)
```sql
SELECT id, cache_path, file_size, last_accessed_at
FROM downloads
WHERE last_accessed_at < ?  -- cutoff timestamp
ORDER BY last_accessed_at ASC
LIMIT 10;
```
**Optimization**: Index on `downloads.last_accessed_at`

---

#### 5. Get recent operation logs with failures
```sql
SELECT operation_type, entity_type, entity_id, message, error_details, timestamp
FROM logs
WHERE operation_status = 'failed'
ORDER BY timestamp DESC
LIMIT 50;
```
**Optimization**: Composite index on `(operation_status, timestamp DESC)`

---

#### 6. Check if version is already installed
```sql
SELECT id, installed_path, checksum
FROM installations
WHERE binary_id = ? AND version = ?;
```
**Optimization**: Unique index on `(binary_id, version)`

---

#### 7. Verify binary integrity
```sql
SELECT checksum, checksum_algorithm
FROM installations
WHERE id = ?;
```
**Optimization**: Primary key lookup

---

#### 8. Get latest installed version for binary (update command)
```sql
SELECT version, installed_at 
FROM installations 
WHERE binary_id = ? 
ORDER BY installed_at DESC 
LIMIT 1;
```
**Optimization**: Index `idx_installations_binary_installed`

---

#### 9. Get previous version for rollback
```sql
SELECT i.id, i.version, i.installed_path
FROM installations i
JOIN versions v ON v.binary_id = i.binary_id
WHERE i.binary_id = ? AND i.id != v.installation_id
ORDER BY i.installed_at DESC
LIMIT 1;
```
**Optimization**: Index `idx_installations_binary_installed`

---

#### 10. Check if version is cached (resume check)
```sql
SELECT id, cache_path, is_complete, checksum
FROM downloads
WHERE binary_id = ? AND version = ?;
```
**Optimization**: Index `idx_downloads_binary_version`

---

#### 11. Get operation history for specific binary
```sql
SELECT operation_type, operation_status, message, timestamp, duration_ms
FROM logs
WHERE entity_type = 'binary' AND entity_id = ?
ORDER BY timestamp DESC
LIMIT 20;
```
**Optimization**: Index `idx_logs_entity_timestamp`

---

#### 12. Get failed installations (diagnostics)
```sql
SELECT entity_id, message, error_details, timestamp
FROM logs
WHERE operation_status = 'failed' AND operation_type = 'install'
ORDER BY timestamp DESC
LIMIT 50;
```
**Optimization**: Index `idx_logs_status_type_timestamp`

---

#### 13. Find binary by alias
```sql
SELECT id, name, provider, provider_path
FROM binaries
WHERE alias = ?;
```
**Optimization**: Partial index `idx_binaries_alias`

---

#### 14. Cache cleanup by size threshold
```sql
SELECT id, cache_path, file_size, last_accessed_at
FROM downloads
ORDER BY file_size DESC;
-- Application sums until threshold reached
```
**Optimization**: Index `idx_downloads_file_size`

---

## Migration Strategy

### Initial Schema Creation
```sql
-- migrations/001_initial_schema.sql
PRAGMA foreign_keys = ON;

-- Create all tables in dependency order
-- 1. binaries (no dependencies)
-- 2. installations (depends on binaries)
-- 3. versions (depends on binaries, installations)
-- 4. downloads (depends on binaries)
-- 5. logs (independent)

-- Create all indexes
```

### Schema Version Tracking
```sql
CREATE TABLE migrations (
    version INTEGER PRIMARY KEY,
    applied_at INTEGER NOT NULL,
    description TEXT NOT NULL
);

INSERT INTO migrations (version, applied_at, description)
VALUES (1, strftime('%s', 'now'), 'Initial schema');
```

---

## Implementation Work Plan

### Phase 1: Database Infrastructure
- [x] **Create database package structure** (`internal/database/`)
  - [x] `connection.go` - SQLite connection management
  - [x] `migrations.go` - Schema migration runner
  - [x] `schema.go` - Schema definitions and creation
- [x] **Implement database initialisation**
  - [x] Create database file at `$HOME/.local/share/.binmate/user.db`
  - [x] Run initial migration
  - [x] Enable foreign keys: `PRAGMA foreign_keys = ON`
  - [x] Set appropriate pragmas for performance (`PRAGMA journal_mode=WAL`)
- [ ] **Write database connection tests**
  - [ ] Test database creation
  - [ ] Test connection pooling
  - [ ] Test concurrent access

### Phase 2: Repository Layer (Data Access)
- [x] **Create repository interfaces** (`internal/database/repository/`)
  - [x] `binaries_repository.go` - CRUD for binaries
  - [x] `installations_repository.go` - CRUD for installations
  - [x] `versions_repository.go` - Active version management
  - [x] `downloads_repository.go` - Cache management
  - [x] `logs_repository.go` - Logging operations
- [x] **Implement binary repository methods**
  - [x] `Create(binary)` - Insert new binary
  - [x] `Update(binary)` - Update binary metadata
  - [x] `Get(id)` - Retrieve by ID
  - [x] `List()` - List all binaries
  - [x] `Delete(id)` - Remove binary (cascade deletes)
  - [x] `SyncFromConfig(config)` - Upsert from config.json
- [ ] **Implement installations repository methods**
  - [x] `Create(installation)` - Record new installation
  - [x] `Get(binaryID, version)` - Retrieve specific version
  - [x] `ListByBinary(binaryID)` - Get all versions for binary
  - [x] `Delete(id)` - Remove installation record
  - [ ] `VerifyChecksum(id, expectedChecksum)` - Integrity check
- [x] **Implement active versions repository methods**
  - [x] `Set(binaryID, installationID, symlinkPath)` - Activate version
  - [x] `Get(binaryID)` - Get active version
  - [x] `Unset(binaryID)` - Remove active version
  - [x] `Switch(binaryID, installationID)` - Change active version
- [x] **Implement cached downloads repository methods**
  - [x] `Create(download)` - Record cached download
  - [x] `Get(binaryID, version)` - Check if cached
  - [x] `UpdateLastAccessed(id)` - Touch for LRU
  - [x] `MarkComplete(id)` - Complete download
  - [x] `ListForCleanup(cutoffTime, limit)` - LRU eviction
  - [x] `Delete(id)` - Remove cache entry
  - [x] `GetIncomplete()` - Find resumable downloads
- [x] **Implement operation logs repository methods**
  - [x] `Log(operation)` - Create log entry
  - [x] `LogStart(opType, entity)` - Log operation start
  - [x] `LogSuccess(id, duration)` - Mark operation success
  - [x] `LogFailure(id, error, duration)` - Mark operation failure
  - [x] `GetRecent(limit)` - Recent operations
  - [x] `GetByType(opType, limit)` - Filter by operation type
  - [x] `GetFailures(limit)` - Recent failures

### Phase 3: Integration with Existing Code
- [ ] **Update config sync logic**
  - [ ] Modify `internal/core/config/read_config.go` to sync with database
  - [ ] Call `binaries_repository.SyncFromConfig()` after reading config.json
  - [ ] Handle added/updated/removed binaries
- [ ] **Update installation flow**
  - [ ] Modify `internal/cli/install/command.go` to use database
  - [ ] Log operation start before download
  - [ ] Moved cached downloads from `/tmp` folder to `os.UserCacheDir`
  - [ ] Record cached download after successful download
  - [ ] Record installation after successful extraction
  - [ ] Update install path from `os.UserCacheDir` to `$HOME/.local/share/.binmate/versions/`. Implement OS aware path resolution strategy to overcome go limitation of lack of support for local share directory in `os` module.
  - [ ] Set active version after successful symlink
  - [ ] Log operation success/failure
- [ ] **Update version management**
  - [ ] Modify `internal/core/version/set_active_version.go` to use database
  - [ ] Update versions table when switching versions
  - [ ] Log version switch operations
- [ ] **Add cache management command**
  - [ ] Create `internal/cli/cache/command.go`
  - [ ] Implement `binmate cache list` - Show cached downloads
  - [ ] Implement `binmate cache clean` - Remove old cache (LRU)
  - [ ] Implement `binmate cache prune` - Remove all cache entries

### Phase 4: Query Optimisation & Performance
- [ ] **Benchmark critical queries**
  - [ ] List installed binaries with active versions
  - [ ] Get versions for binary
  - [ ] Latest version lookup (update command)
  - [ ] Rollback version lookup
  - [ ] Cache cleanup query (LRU and size-based)
  - [ ] Entity history lookup
- [ ] **Verify index usage with EXPLAIN QUERY PLAN**
  - [ ] Confirm `idx_installations_binary_installed` used for latest/rollback
  - [ ] Confirm `idx_downloads_binary_version` used for cache checks
  - [ ] Confirm `idx_logs_entity_timestamp` used for history queries
  - [ ] Confirm `idx_logs_status_type_timestamp` used for filtered logs
- [ ] **Implement query caching** (optional)
  - [ ] Cache binary list in memory
  - [ ] Invalidate on config sync
- [ ] **Configure SQLite pragmas**
  - [ ] `PRAGMA journal_mode=WAL` - Better concurrency
  - [ ] `PRAGMA synchronous=NORMAL` - Balance safety/performance
  - [ ] `PRAGMA cache_size=-64000` - 64MB cache
  - [ ] `PRAGMA temp_store=MEMORY` - Temp tables in memory

### Phase 5: Testing
- [ ] **Unit tests for repositories**
  - [ ] Test CRUD operations for each repository
  - [ ] Test foreign key constraints
  - [ ] Test cascade deletes
  - [ ] Test unique constraints
- [ ] **Integration tests**
  - [ ] Test full installation flow with database
  - [ ] Test version switching with database
  - [ ] Test config sync with database
  - [ ] Test cache cleanup
- [ ] **Test data integrity**
  - [ ] Test checksum verification
  - [ ] Test concurrent access (WAL mode)
  - [ ] Test database recovery after crash

### Phase 6: Error Handling & Logging
- [ ] **Implement database error types**
  - [ ] `ErrNotFound` - Entity not found
  - [ ] `ErrDuplicate` - Unique constraint violation
  - [ ] `ErrForeignKey` - Foreign key constraint violation
- [ ] **Add database logging**
  - [ ] Log all database operations (at debug level)
  - [ ] Log slow queries (threshold: 100ms)
- [ ] **Implement retry logic**
  - [ ] Retry on `SQLITE_BUSY` errors
  - [ ] Exponential backoff for retries

### Phase 7: Documentation
- [ ] **Document database schema** (in code comments)
- [ ] **Create ER diagram** (using Mermaid)
- [ ] **Document query patterns** (with examples)
- [ ] **Write database migration guide**

---

## Relationships & Join Optimisation

### Foreign Key Relationships
```
binaries (1:N) installations
├── ON DELETE CASCADE - Remove all installations when binary removed
└── Indexed on installations.binary_id

binaries (1:N) downloads
├── ON DELETE CASCADE - Remove cache entries when binary removed
└── Indexed on downloads.binary_id

binaries (1:1) versions
├── ON DELETE CASCADE - Remove active version when binary removed
└── Primary key on versions.binary_id (enforces 1:1)

installations (1:1) versions
├── ON DELETE CASCADE - Unset active version if installation removed
└── Indexed on versions.installation_id
```

### Join Query Performance
- **List with active versions**: JOIN on indexed FK columns → O(1) per row
- **Version history**: Single table scan with index on binary_id → O(log n)
- **Cache cleanup**: Index scan on last_accessed_at → O(log n)

---

## Security Considerations

### Data Integrity
- ✅ Foreign keys enforced (`PRAGMA foreign_keys = ON`)
- ✅ Unique constraints prevent duplicate installations
- ✅ Checksums stored with algorithm for verification
- ✅ No credentials stored in database (config/env only)

### File System Security
- ✅ Database file permissions: `0600` (owner read/write only)
- ✅ Cache directory permissions: `0700`
- ✅ Validate paths before writing (prevent directory traversal)

### SQL Injection Prevention
- ✅ Use parameterised queries exclusively
- ✅ No string concatenation for SQL
- ✅ Validate entity IDs before queries

---

## Success Criteria

- ✅ All tables created with proper constraints
- ✅ Foreign keys enforced and cascade deletes work
- ✅ Indexes improve query performance (measured with EXPLAIN QUERY PLAN)
- ✅ Repository layer provides clean API for data access
- ✅ Config sync keeps database in sync with config.json
- ✅ Operation logs track all user actions
- ✅ Cache management works with LRU eviction
- ✅ Checksum verification prevents corrupted installations
- ✅ Test coverage >80% for database layer
- ✅ No N+1 query problems in list operations

---

## Future Enhancements (Out of Scope)

- **Full-text search**: Add FTS5 virtual table for searching binaries
- **Database backups**: Automatic backup before major operations
- **Statistics**: Track download counts, usage metrics
- **Multi-user support**: User-specific installations and permissions
- **Remote database**: Support for shared database (SQLite Cloud, PostgreSQL)
- **Database vacuum**: Automatic VACUUM on cleanup operations

