package repository

import (
	"testing"

	"cturner8/binmate/internal/database"
)

func TestLogsRepository_Log(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	logRepo := NewLogsRepository(db)

	entityType := "binary"
	entityID := "test-binary-1"
	message := "Test log message"
	errorDetails := "Test error details"
	metadata := `{"key":"value"}`
	durationMs := int64(100)
	userContext := "test-user"

	log := &database.Log{
		OperationType:   "install",
		OperationStatus: "success",
		EntityType:      &entityType,
		EntityID:        &entityID,
		Message:         &message,
		ErrorDetails:    &errorDetails,
		Metadata:        &metadata,
		DurationMs:      &durationMs,
		UserContext:     &userContext,
	}

	err := logRepo.Log(log)
	if err != nil {
		t.Fatalf("Failed to create log: %v", err)
	}

	if log.ID == 0 {
		t.Error("Expected ID to be set after creation")
	}

	if log.Timestamp == 0 {
		t.Error("Expected Timestamp to be set")
	}
}

func TestLogsRepository_LogStart(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	logRepo := NewLogsRepository(db)

	logID, err := logRepo.LogStart("install", "binary", "test-binary-1", "Starting installation")
	if err != nil {
		t.Fatalf("Failed to log start: %v", err)
	}

	if logID == 0 {
		t.Error("Expected logID to be set")
	}

	// Retrieve and verify
	logs, err := logRepo.GetRecent(1)
	if err != nil {
		t.Fatalf("Failed to get recent logs: %v", err)
	}

	if len(logs) != 1 {
		t.Fatalf("Expected 1 log, got %d", len(logs))
	}

	log := logs[0]

	if log.OperationType != "install" {
		t.Errorf("Expected OperationType install, got %s", log.OperationType)
	}

	if log.OperationStatus != "started" {
		t.Errorf("Expected OperationStatus started, got %s", log.OperationStatus)
	}

	if log.EntityType == nil || *log.EntityType != "binary" {
		t.Errorf("Expected EntityType binary, got %v", log.EntityType)
	}

	if log.EntityID == nil || *log.EntityID != "test-binary-1" {
		t.Errorf("Expected EntityID test-binary-1, got %v", log.EntityID)
	}

	if log.Message == nil || *log.Message != "Starting installation" {
		t.Errorf("Expected Message 'Starting installation', got %v", log.Message)
	}
}

func TestLogsRepository_LogEntity(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	logRepo := NewLogsRepository(db)

	// Create initial log
	logID, err := logRepo.LogStart("download", "none", "none", "Starting download")
	if err != nil {
		t.Fatalf("Failed to log start: %v", err)
	}

	// Update entity information
	err = logRepo.LogEntity(logID, "binary", "test-binary-2")
	if err != nil {
		t.Fatalf("Failed to log entity: %v", err)
	}

	// Retrieve and verify
	logs, err := logRepo.GetRecent(1)
	if err != nil {
		t.Fatalf("Failed to get recent logs: %v", err)
	}

	if len(logs) != 1 {
		t.Fatalf("Expected 1 log, got %d", len(logs))
	}

	log := logs[0]

	if log.EntityType == nil || *log.EntityType != "binary" {
		t.Errorf("Expected EntityType binary, got %v", log.EntityType)
	}

	if log.EntityID == nil || *log.EntityID != "test-binary-2" {
		t.Errorf("Expected EntityID test-binary-2, got %v", log.EntityID)
	}
}

func TestLogsRepository_LogSuccess(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	logRepo := NewLogsRepository(db)

	// Create initial log
	logID, err := logRepo.LogStart("install", "binary", "test-binary-3", "Starting installation")
	if err != nil {
		t.Fatalf("Failed to log start: %v", err)
	}

	// Mark as success
	err = logRepo.LogSuccess(logID, 250)
	if err != nil {
		t.Fatalf("Failed to log success: %v", err)
	}

	// Retrieve and verify
	logs, err := logRepo.GetRecent(1)
	if err != nil {
		t.Fatalf("Failed to get recent logs: %v", err)
	}

	if len(logs) != 1 {
		t.Fatalf("Expected 1 log, got %d", len(logs))
	}

	log := logs[0]

	if log.OperationStatus != "success" {
		t.Errorf("Expected OperationStatus success, got %s", log.OperationStatus)
	}

	if log.DurationMs == nil || *log.DurationMs != 250 {
		t.Errorf("Expected DurationMs 250, got %v", log.DurationMs)
	}
}

func TestLogsRepository_LogFailure(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	logRepo := NewLogsRepository(db)

	// Create initial log
	logID, err := logRepo.LogStart("install", "binary", "test-binary-4", "Starting installation")
	if err != nil {
		t.Fatalf("Failed to log start: %v", err)
	}

	// Mark as failure
	err = logRepo.LogFailure(logID, "Connection timeout", 500)
	if err != nil {
		t.Fatalf("Failed to log failure: %v", err)
	}

	// Retrieve and verify
	logs, err := logRepo.GetRecent(1)
	if err != nil {
		t.Fatalf("Failed to get recent logs: %v", err)
	}

	if len(logs) != 1 {
		t.Fatalf("Expected 1 log, got %d", len(logs))
	}

	log := logs[0]

	if log.OperationStatus != "failed" {
		t.Errorf("Expected OperationStatus failed, got %s", log.OperationStatus)
	}

	if log.ErrorDetails == nil || *log.ErrorDetails != "Connection timeout" {
		t.Errorf("Expected ErrorDetails 'Connection timeout', got %v", log.ErrorDetails)
	}

	if log.DurationMs == nil || *log.DurationMs != 500 {
		t.Errorf("Expected DurationMs 500, got %v", log.DurationMs)
	}
}

func TestLogsRepository_GetRecent(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	logRepo := NewLogsRepository(db)

	// Create multiple logs
	for i := 1; i <= 5; i++ {
		entityType := "binary"
		entityID := "test-binary"
		message := "Test message"

		log := &database.Log{
			OperationType:   "install",
			OperationStatus: "success",
			EntityType:      &entityType,
			EntityID:        &entityID,
			Message:         &message,
		}

		err := logRepo.Log(log)
		if err != nil {
			t.Fatalf("Failed to create log %d: %v", i, err)
		}
	}

	// Get recent logs with limit
	logs, err := logRepo.GetRecent(3)
	if err != nil {
		t.Fatalf("Failed to get recent logs: %v", err)
	}

	if len(logs) != 3 {
		t.Errorf("Expected 3 logs, got %d", len(logs))
	}

	// Verify ordered by timestamp DESC
	for i := 0; i < len(logs)-1; i++ {
		if logs[i].Timestamp < logs[i+1].Timestamp {
			t.Error("Logs are not sorted by Timestamp DESC")
		}
	}
}

func TestLogsRepository_GetRecentEmpty(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	logRepo := NewLogsRepository(db)

	logs, err := logRepo.GetRecent(10)
	if err != nil {
		t.Fatalf("Failed to get recent logs: %v", err)
	}

	if len(logs) != 0 {
		t.Errorf("Expected 0 logs, got %d", len(logs))
	}
}

func TestLogsRepository_GetByType(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	logRepo := NewLogsRepository(db)

	entityType := "binary"
	entityID := "test-binary"
	message := "Test message"

	// Create logs with different operation types
	types := []string{"install", "install", "download", "uninstall"}
	for _, opType := range types {
		log := &database.Log{
			OperationType:   opType,
			OperationStatus: "success",
			EntityType:      &entityType,
			EntityID:        &entityID,
			Message:         &message,
		}

		err := logRepo.Log(log)
		if err != nil {
			t.Fatalf("Failed to create log: %v", err)
		}
	}

	// Get logs by type
	logs, err := logRepo.GetByType("install", 10)
	if err != nil {
		t.Fatalf("Failed to get logs by type: %v", err)
	}

	if len(logs) != 2 {
		t.Errorf("Expected 2 install logs, got %d", len(logs))
	}

	for _, log := range logs {
		if log.OperationType != "install" {
			t.Errorf("Expected OperationType install, got %s", log.OperationType)
		}
	}
}

func TestLogsRepository_GetByTypeEmpty(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	logRepo := NewLogsRepository(db)

	logs, err := logRepo.GetByType("nonexistent", 10)
	if err != nil {
		t.Fatalf("Failed to get logs by type: %v", err)
	}

	if len(logs) != 0 {
		t.Errorf("Expected 0 logs, got %d", len(logs))
	}
}

func TestLogsRepository_GetFailures(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	logRepo := NewLogsRepository(db)

	entityType := "binary"
	entityID := "test-binary"
	message := "Test message"

	// Create logs with different statuses
	statuses := []string{"success", "failed", "failed", "success", "started"}
	for _, status := range statuses {
		log := &database.Log{
			OperationType:   "install",
			OperationStatus: status,
			EntityType:      &entityType,
			EntityID:        &entityID,
			Message:         &message,
		}

		err := logRepo.Log(log)
		if err != nil {
			t.Fatalf("Failed to create log: %v", err)
		}
	}

	// Get failure logs
	logs, err := logRepo.GetFailures(10)
	if err != nil {
		t.Fatalf("Failed to get failure logs: %v", err)
	}

	if len(logs) != 2 {
		t.Errorf("Expected 2 failure logs, got %d", len(logs))
	}

	for _, log := range logs {
		if log.OperationStatus != "failed" {
			t.Errorf("Expected OperationStatus failed, got %s", log.OperationStatus)
		}
	}
}

func TestLogsRepository_GetFailuresEmpty(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	logRepo := NewLogsRepository(db)

	entityType := "binary"
	entityID := "test-binary"
	message := "Test message"

	// Create only successful logs
	log := &database.Log{
		OperationType:   "install",
		OperationStatus: "success",
		EntityType:      &entityType,
		EntityID:        &entityID,
		Message:         &message,
	}

	err := logRepo.Log(log)
	if err != nil {
		t.Fatalf("Failed to create log: %v", err)
	}

	// Get failure logs
	logs, err := logRepo.GetFailures(10)
	if err != nil {
		t.Fatalf("Failed to get failure logs: %v", err)
	}

	if len(logs) != 0 {
		t.Errorf("Expected 0 failure logs, got %d", len(logs))
	}
}

func TestLogsRepository_GetByEntity(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	logRepo := NewLogsRepository(db)

	message := "Test message"

	// Create logs for different entities
	entities := []struct {
		entityType string
		entityID   string
	}{
		{"binary", "test-binary-1"},
		{"binary", "test-binary-1"},
		{"binary", "test-binary-2"},
		{"installation", "test-install-1"},
	}

	for _, entity := range entities {
		entityType := entity.entityType
		entityID := entity.entityID

		log := &database.Log{
			OperationType:   "install",
			OperationStatus: "success",
			EntityType:      &entityType,
			EntityID:        &entityID,
			Message:         &message,
		}

		err := logRepo.Log(log)
		if err != nil {
			t.Fatalf("Failed to create log: %v", err)
		}
	}

	// Get logs by entity
	logs, err := logRepo.GetByEntity("binary", "test-binary-1", 10)
	if err != nil {
		t.Fatalf("Failed to get logs by entity: %v", err)
	}

	if len(logs) != 2 {
		t.Errorf("Expected 2 logs for test-binary-1, got %d", len(logs))
	}

	for _, log := range logs {
		if log.EntityType == nil || *log.EntityType != "binary" {
			t.Errorf("Expected EntityType binary, got %v", log.EntityType)
		}

		if log.EntityID == nil || *log.EntityID != "test-binary-1" {
			t.Errorf("Expected EntityID test-binary-1, got %v", log.EntityID)
		}
	}
}

func TestLogsRepository_GetByEntityEmpty(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	logRepo := NewLogsRepository(db)

	logs, err := logRepo.GetByEntity("binary", "nonexistent", 10)
	if err != nil {
		t.Fatalf("Failed to get logs by entity: %v", err)
	}

	if len(logs) != 0 {
		t.Errorf("Expected 0 logs, got %d", len(logs))
	}
}

func TestLogsRepository_WithNullFields(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	logRepo := NewLogsRepository(db)

	// Create log with all optional fields as nil
	log := &database.Log{
		OperationType:   "test",
		OperationStatus: "started",
		EntityType:      nil,
		EntityID:        nil,
		Message:         nil,
		ErrorDetails:    nil,
		Metadata:        nil,
		DurationMs:      nil,
		UserContext:     nil,
	}

	err := logRepo.Log(log)
	if err != nil {
		t.Fatalf("Failed to create log with null fields: %v", err)
	}

	// Retrieve and verify
	logs, err := logRepo.GetRecent(1)
	if err != nil {
		t.Fatalf("Failed to get recent logs: %v", err)
	}

	if len(logs) != 1 {
		t.Fatalf("Expected 1 log, got %d", len(logs))
	}

	retrieved := logs[0]

	if retrieved.EntityType != nil {
		t.Errorf("Expected EntityType to be nil, got %v", retrieved.EntityType)
	}

	if retrieved.EntityID != nil {
		t.Errorf("Expected EntityID to be nil, got %v", retrieved.EntityID)
	}

	if retrieved.Message != nil {
		t.Errorf("Expected Message to be nil, got %v", retrieved.Message)
	}

	if retrieved.ErrorDetails != nil {
		t.Errorf("Expected ErrorDetails to be nil, got %v", retrieved.ErrorDetails)
	}

	if retrieved.Metadata != nil {
		t.Errorf("Expected Metadata to be nil, got %v", retrieved.Metadata)
	}

	if retrieved.DurationMs != nil {
		t.Errorf("Expected DurationMs to be nil, got %v", retrieved.DurationMs)
	}

	if retrieved.UserContext != nil {
		t.Errorf("Expected UserContext to be nil, got %v", retrieved.UserContext)
	}
}
