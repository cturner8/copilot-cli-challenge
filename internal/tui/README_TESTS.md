# TUI Package Tests

This directory contains comprehensive unit tests for the Terminal User Interface (TUI) package.

## Test Files

| File | Purpose | Lines | Test Count |
|------|---------|-------|------------|
| `helpers_test.go` | Text formatting and manipulation utilities | 251 | 36 |
| `model_test.go` | Model initialization and state management | 237 | 6 |
| `view_test.go` | View rendering functions | 467 | 18 |
| `update_test.go` | Update logic and user interaction handling | 549 | 24 |

## Quick Start

```bash
# Run all tests
go test ./internal/tui/...

# Run with verbose output
go test ./internal/tui/... -v

# Run with coverage
go test ./internal/tui/... -cover

# Run specific test
go test ./internal/tui/ -run TestFormatBytes
```

## Test Coverage

Current coverage: **41.0%** of statements

The tests focus on:
- ✅ View rendering logic
- ✅ User interaction flows
- ✅ State management
- ✅ Text formatting utilities
- ✅ Navigation and tab switching
- ✅ Error/success message display

## Test Organization

### helpers_test.go
Tests for helper functions in `text_utils.go` and `view_versions.go`:
- `formatBytes()` - Human-readable byte formatting
- `truncateText()` - Text truncation with ellipsis
- `truncatePathEnd()` - Path truncation keeping the end
- `padLeft()`, `padRight()`, `center()` - Text alignment

### model_test.go
Tests for model initialization and types:
- `initialModel()` - Default model initialization
- `viewState.String()` - View state string representations
- `createFormInputs()` - Binary form input creation
- `parsedBinaryConfig` - Configuration parsing

### view_test.go
Tests for view rendering functions:
- `renderBinariesList()` - Binary list with various states
- `renderVersions()` - Version list display
- `renderDownloads()` - Downloads placeholder view
- `renderConfiguration()` - Configuration view
- `renderHelp()` - Help documentation view
- `renderTabs()` - Tab bar rendering

### update_test.go
Tests for update and interaction logic:
- `Update()` - Main update dispatcher
- `updateBinariesList()` - Binary list interactions
- `updateVersions()` - Version list interactions
- Tab navigation (`getTabForKey`, `getNextTab`, `getPreviousTab`)
- Message handling (success, error, loading)

## Test Patterns

### Table-Driven Tests
```go
tests := []struct {
    name     string
    input    int64
    expected string
}{
    {"zero bytes", 0, "0 B"},
    {"kilobytes", 2048, "2.0 KB"},
}

for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        result := formatBytes(tt.input)
        if result != tt.expected {
            t.Errorf("got %q, want %q", result, tt.expected)
        }
    })
}
```

### State Verification
```go
m := initialModel(&repository.Service{}, &config.Config{})
m.currentView = viewBinariesList

updatedModel, _ := m.updateBinariesList(keyMsg)
m2 := updatedModel.(model)

if m2.currentView != viewVersions {
    t.Errorf("expected view change")
}
```

### Content Verification
```go
result := m.renderBinariesList()

expectedStrings := []string{"Binmate", "No binaries configured"}
for _, expected := range expectedStrings {
    if !strings.Contains(result, expected) {
        t.Errorf("missing expected string: %q", expected)
    }
}
```

## Design Decisions

1. **Mock Services**: Tests use minimal mocks since they focus on view logic and state management, not database operations.

2. **Proper Initialization**: All tests use `initialModel()` to ensure text inputs and other components are properly initialized.

3. **View-Only Testing**: Tests verify rendering output and state transitions, not the actual Bubble Tea program execution or database queries.

4. **Edge Case Coverage**: Tests include empty states, loading states, error conditions, and boundary cases.

## Adding New Tests

When adding new functionality to the TUI package:

1. **Helper Functions**: Add tests to `helpers_test.go`
2. **Model Changes**: Update `model_test.go`
3. **New Views**: Add tests to `view_test.go`
4. **Interaction Logic**: Add tests to `update_test.go`

Example:
```go
func TestNewFeature(t *testing.T) {
    m := initialModel(&repository.Service{}, &config.Config{})
    // Set up test state
    m.someField = testValue
    
    // Execute function
    result := m.newFeature()
    
    // Verify result
    if result != expected {
        t.Errorf("got %v, want %v", result, expected)
    }
}
```

## See Also

- [TEST_SUMMARY.md](TEST_SUMMARY.md) - Detailed test documentation
- [Bubble Tea Documentation](https://github.com/charmbracelet/bubbletea)
