package repository

import (
	"database/sql"
	"fmt"
	"time"

	"cturner8/binmate/internal/database"
)

type LogsRepository struct {
	db *database.DB
}

func NewLogsRepository(db *database.DB) *LogsRepository {
	return &LogsRepository{db: db}
}

// Log creates a log entry
func (r *LogsRepository) Log(log *database.Log) error {
	log.Timestamp = time.Now().Unix()

	result, err := r.db.Exec(`
INSERT INTO logs (operation_type, operation_status, entity_type, entity_id,
message, error_details, metadata, timestamp, duration_ms, user_context)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`, log.OperationType, log.OperationStatus, log.EntityType, log.EntityID,
		log.Message, log.ErrorDetails, log.Metadata, log.Timestamp,
		log.DurationMs, log.UserContext)

	if err != nil {
		return fmt.Errorf("failed to create log: %w", err)
	}

	log.ID, err = result.LastInsertId()
	return err
}

// LogStart creates a log entry for an operation start
func (r *LogsRepository) LogStart(opType, entityType, entityID, message string) (int64, error) {
	log := &database.Log{
		OperationType:   opType,
		OperationStatus: "started",
		EntityType:      &entityType,
		EntityID:        &entityID,
		Message:         &message,
	}

	if err := r.Log(log); err != nil {
		return 0, err
	}

	return log.ID, nil
}

// LogSuccess updates a log entry as successful
func (r *LogsRepository) LogEntity(id int64, entityType, entityID string) error {
	_, err := r.db.Exec(`
UPDATE logs 
SET entity_type = ?, entity_id = ?
WHERE id = ?
`, entityType, entityID, id)

	if err != nil {
		return fmt.Errorf("failed to log entity: %w", err)
	}

	return nil
}

// LogSuccess updates a log entry as successful
func (r *LogsRepository) LogSuccess(id int64, durationMs int64) error {
	_, err := r.db.Exec(`
UPDATE logs 
SET operation_status = 'success', duration_ms = ?
WHERE id = ?
`, durationMs, id)

	if err != nil {
		return fmt.Errorf("failed to log success: %w", err)
	}

	return nil
}

// LogFailure updates a log entry as failed
func (r *LogsRepository) LogFailure(id int64, errorDetails string, durationMs int64) error {
	_, err := r.db.Exec(`
UPDATE logs 
SET operation_status = 'failed', error_details = ?, duration_ms = ?
WHERE id = ?
`, errorDetails, durationMs, id)

	if err != nil {
		return fmt.Errorf("failed to log failure: %w", err)
	}

	return nil
}

// GetRecent retrieves recent log entries
func (r *LogsRepository) GetRecent(limit int) ([]*database.Log, error) {
	rows, err := r.db.Query(`
SELECT id, operation_type, operation_status, entity_type, entity_id,
message, error_details, metadata, timestamp, duration_ms, user_context
FROM logs
ORDER BY timestamp DESC
LIMIT ?
`, limit)

	if err != nil {
		return nil, fmt.Errorf("failed to get recent logs: %w", err)
	}
	defer rows.Close()

	return r.scanLogs(rows)
}

// GetByType retrieves logs filtered by operation type
func (r *LogsRepository) GetByType(opType string, limit int) ([]*database.Log, error) {
	rows, err := r.db.Query(`
SELECT id, operation_type, operation_status, entity_type, entity_id,
message, error_details, metadata, timestamp, duration_ms, user_context
FROM logs
WHERE operation_type = ?
ORDER BY timestamp DESC
LIMIT ?
`, opType, limit)

	if err != nil {
		return nil, fmt.Errorf("failed to get logs by type: %w", err)
	}
	defer rows.Close()

	return r.scanLogs(rows)
}

// GetFailures retrieves failed operations
func (r *LogsRepository) GetFailures(limit int) ([]*database.Log, error) {
	rows, err := r.db.Query(`
SELECT id, operation_type, operation_status, entity_type, entity_id,
message, error_details, metadata, timestamp, duration_ms, user_context
FROM logs
WHERE operation_status = 'failed'
ORDER BY timestamp DESC
LIMIT ?
`, limit)

	if err != nil {
		return nil, fmt.Errorf("failed to get failure logs: %w", err)
	}
	defer rows.Close()

	return r.scanLogs(rows)
}

// GetByEntity retrieves logs for a specific entity
func (r *LogsRepository) GetByEntity(entityType, entityID string, limit int) ([]*database.Log, error) {
	rows, err := r.db.Query(`
SELECT id, operation_type, operation_status, entity_type, entity_id,
message, error_details, metadata, timestamp, duration_ms, user_context
FROM logs
WHERE entity_type = ? AND entity_id = ?
ORDER BY timestamp DESC
LIMIT ?
`, entityType, entityID, limit)

	if err != nil {
		return nil, fmt.Errorf("failed to get logs by entity: %w", err)
	}
	defer rows.Close()

	return r.scanLogs(rows)
}

// scanLogs is a helper to scan log rows
func (r *LogsRepository) scanLogs(rows *sql.Rows) ([]*database.Log, error) {
	var logs []*database.Log
	for rows.Next() {
		log := &database.Log{}
		err := rows.Scan(&log.ID, &log.OperationType, &log.OperationStatus,
			&log.EntityType, &log.EntityID, &log.Message, &log.ErrorDetails,
			&log.Metadata, &log.Timestamp, &log.DurationMs, &log.UserContext)

		if err != nil {
			return nil, fmt.Errorf("failed to scan log: %w", err)
		}

		logs = append(logs, log)
	}

	return logs, rows.Err()
}
