# Phase 5 Future Implementation Plan
## Fully Functional Placeholder Views for Binmate TUI

**Document Version:** 1.0  
**Date:** February 2026  
**Status:** Planning Document  

---

## Executive Summary

This document outlines the detailed implementation plan for converting the three placeholder views (Downloads, Configuration, and Help) in the Binmate TUI from basic placeholders into fully functional, interactive views. These views are critical for providing users with complete control over their binary management workflow within the terminal interface.

**Current State:** Basic placeholder views exist with informational text describing future functionality.

**Target State:** Fully interactive views with CRUD operations, validation, and real-time updates.

---

## Table of Contents

1. [Downloads View Implementation](#downloads-view-implementation)
2. [Configuration View Implementation](#configuration-view-implementation)
3. [Help View Enhancement](#help-view-enhancement)
4. [Shared Infrastructure](#shared-infrastructure)
5. [Testing Strategy](#testing-strategy)
6. [Implementation Timeline](#implementation-timeline)
7. [Risk Assessment](#risk-assessment)

---

## Downloads View Implementation

### Overview

The Downloads view will provide comprehensive management of cached asset downloads, allowing users to:
- View all cached downloads with metadata
- Clear individual downloads to free space
- Clear all downloads with confirmation
- View cache statistics and total size
- Verify checksums for cached files

### Data Model Extensions

#### New Database Queries Required

```go
// In internal/database/repository/downloads.go

// GetAllDownloadsWithDetails returns all downloads with binary metadata
func (r *DownloadsRepository) GetAllDownloadsWithDetails() ([]DownloadWithBinary, error)

// GetCacheStatistics returns summary statistics for the download cache
func (r *DownloadsRepository) GetCacheStatistics() (*CacheStats, error)

// VerifyDownloadChecksum verifies the checksum of a cached download
func (r *DownloadsRepository) VerifyDownloadChecksum(downloadID int64) (bool, error)

// ClearAllDownloads removes all download records and optionally deletes files
func (r *DownloadsRepository) ClearAllDownloads(deleteFiles bool) error
```

#### New Data Structures

```go
type DownloadWithBinary struct {
    Download *database.Download
    Binary   *database.Binary
    FileExists bool
    VerifiedAt int64
}

type CacheStats struct {
    TotalDownloads  int
    TotalSize       int64
    OldestDownload  int64
    NewestDownload  int64
    VerifiedCount   int
    CorruptedCount  int
}
```

### UI Layout

```
┌─────────────────────────────────────────────────────────────────┐
│ 📦 Binmate - Binary Manager                                     │
│                                                                  │
│ [Binaries] [Versions] [Downloads] [Config] [Help]               │
│                                                                  │
│ 📥 Cached Downloads                                              │
│ ──────────────────────────────────────────────────────────────  │
│ Cache Statistics:                                                │
│ Total: 15 downloads | Size: 2.3 GB | Verified: 12/15           │
│                                                                  │
│ Binary           Version    Size      Downloaded    Status      │
│ ──────────────────────────────────────────────────────────────  │
│ > gh             2.40.1     45.2 MB   2 days ago    ✓ Verified  │
│   kubectl        1.28.3     52.1 MB   1 week ago    ✓ Verified  │
│   terraform      1.6.5      85.3 MB   2 weeks ago   ⚠ Unverified│
│   ...                                                            │
│                                                                  │
│ ↑/↓: navigate • v: verify • d: delete • D: delete all • q: quit │
└─────────────────────────────────────────────────────────────────┘
```

### Model State Extensions

```go
// In internal/tui/model.go

type model struct {
    // ... existing fields ...
    
    // Downloads view state
    downloads              []DownloadWithBinary
    cacheStats             *CacheStats
    selectedDownloadIdx    int
    verifyingDownload      bool
    confirmingClearAll     bool
    lastVerificationResult string
}
```

### Update Logic Extensions

```go
// In internal/tui/update.go

case viewDownloads:
    switch {
    case key.Matches(msg, keys.up):
        // Navigate up in downloads list
        if m.selectedDownloadIdx > 0 {
            m.selectedDownloadIdx--
        }
        
    case key.Matches(msg, keys.down):
        // Navigate down in downloads list
        if m.selectedDownloadIdx < len(m.downloads)-1 {
            m.selectedDownloadIdx++
        }
        
    case msg.String() == "v":
        // Verify selected download checksum
        return m, verifyDownloadChecksumCmd(
            m.downloads[m.selectedDownloadIdx].Download.ID,
        )
        
    case msg.String() == "d":
        // Delete selected download
        return m, deleteDownloadCmd(
            m.downloads[m.selectedDownloadIdx].Download.ID,
        )
        
    case msg.String() == "D":
        // Confirm clear all downloads
        m.confirmingClearAll = true
        return m, nil
        
    case msg.String() == "r":
        // Refresh download list and statistics
        return m, tea.Batch(
            loadDownloadsCmd(m.dbService),
            loadCacheStatsCmd(m.dbService),
        )
    }
```

### Bubble Tea Commands

```go
// New commands in internal/tui/messages.go

// loadDownloadsMsg contains loaded downloads data
type loadDownloadsMsg struct {
    downloads []DownloadWithBinary
    err       error
}

// loadCacheStatsMsg contains cache statistics
type loadCacheStatsMsg struct {
    stats *CacheStats
    err   error
}

// verifyDownloadMsg contains verification result
type verifyDownloadMsg struct {
    downloadID int64
    valid      bool
    err        error
}

// deleteDownloadMsg contains deletion result
type deleteDownloadMsg struct {
    downloadID int64
    success    bool
    err        error
}

// clearAllDownloadsMsg contains clear all result
type clearAllDownloadsMsg struct {
    count   int
    success bool
    err     error
}

// Commands
func loadDownloadsCmd(dbService *repository.Service) tea.Cmd
func loadCacheStatsCmd(dbService *repository.Service) tea.Cmd
func verifyDownloadChecksumCmd(downloadID int64) tea.Cmd
func deleteDownloadCmd(downloadID int64) tea.Cmd
func clearAllDownloadsCmd(dbService *repository.Service, deleteFiles bool) tea.Cmd
```

### View Rendering Implementation

Replace `renderDownloads()` in `internal/tui/view_placeholders.go` with full implementation in new file `internal/tui/view_downloads.go`:

```go
func (m model) renderDownloads() string {
    var b strings.Builder
    
    // Title and tabs
    b.WriteString(titleStyle.Render("📦 Binmate - Binary Manager"))
    b.WriteString("\n\n")
    b.WriteString(m.renderTabs())
    
    // Error/success messages
    if m.errorMessage != "" {
        b.WriteString(errorStyle.Render("Error: " + m.errorMessage))
        b.WriteString("\n\n")
    }
    if m.successMessage != "" {
        b.WriteString(successStyle.Render("✓ " + m.successMessage))
        b.WriteString("\n\n")
    }
    
    // Loading state
    if m.loading {
        b.WriteString(loadingStyle.Render("Loading downloads..."))
        b.WriteString("\n\n")
        b.WriteString(helpStyle.Render(getHelpText(m.currentView)))
        return b.String()
    }
    
    // Confirmation dialog for clear all
    if m.confirmingClearAll {
        b.WriteString(headerStyle.Render("Clear All Downloads?"))
        b.WriteString("\n\n")
        b.WriteString("This will remove all download records from the database.\n")
        b.WriteString("Files will remain on disk unless you choose to delete them.\n\n")
        b.WriteString("Press 'y' to clear database records only\n")
        b.WriteString("Press 'Y' (Shift+Y) to also delete cached files\n")
        b.WriteString("Press 'n' or Esc to cancel\n")
        return b.String()
    }
    
    // Cache statistics header
    b.WriteString(headerStyle.Render("📥 Cached Downloads"))
    b.WriteString("\n\n")
    
    if m.cacheStats != nil {
        stats := fmt.Sprintf(
            "Total: %d downloads | Size: %s | Verified: %d/%d",
            m.cacheStats.TotalDownloads,
            formatBytes(m.cacheStats.TotalSize),
            m.cacheStats.VerifiedCount,
            m.cacheStats.TotalDownloads,
        )
        b.WriteString(mutedStyle.Render(stats))
        b.WriteString("\n\n")
    }
    
    // Empty state
    if len(m.downloads) == 0 {
        b.WriteString(emptyStateStyle.Render("No cached downloads"))
        b.WriteString("\n\n")
        b.WriteString(helpStyle.Render(getHelpText(m.currentView)))
        return b.String()
    }
    
    // Table headers
    availableWidth := m.width
    if availableWidth == 0 {
        availableWidth = defaultTerminalWidth
    }
    
    totalWidth := availableWidth - columnPadding5
    binaryWidth := int(float64(totalWidth) * 0.25)
    versionWidth := int(float64(totalWidth) * 0.15)
    sizeWidth := int(float64(totalWidth) * 0.15)
    dateWidth := int(float64(totalWidth) * 0.20)
    statusWidth := int(float64(totalWidth) * 0.25)
    
    headers := []string{
        tableHeaderStyle.Width(binaryWidth).Render("Binary"),
        tableHeaderStyle.Width(versionWidth).Render("Version"),
        tableHeaderStyle.Width(sizeWidth).Render("Size"),
        tableHeaderStyle.Width(dateWidth).Render("Downloaded"),
        tableHeaderStyle.Width(statusWidth).Render("Status"),
    }
    b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, headers...))
    b.WriteString("\n")
    b.WriteString(strings.Repeat("─", totalWidth+columnPadding5))
    b.WriteString("\n")
    
    // Table rows
    for i, download := range m.downloads {
        style := normalStyle
        if i == m.selectedDownloadIdx {
            style = selectedStyle
        }
        
        binary := truncateText(download.Binary.Name, binaryWidth)
        version := truncateText(download.Download.Version, versionWidth)
        size := formatBytes(download.Download.FileSize)
        date := format.FormatTimestamp(download.Download.DownloadedAt, m.config.DateFormat)
        
        // Status indicator
        status := "⚠ Unverified"
        if download.VerifiedAt > 0 {
            status = "✓ Verified"
        }
        if !download.FileExists {
            status = "✗ Missing"
        }
        
        row := []string{
            style.Width(binaryWidth).Render(binary),
            style.Width(versionWidth).Render(version),
            style.Width(sizeWidth).Render(size),
            style.Width(dateWidth).Render(date),
            style.Width(statusWidth).Render(status),
        }
        
        b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, row...))
        b.WriteString("\n")
    }
    
    b.WriteString("\n")
    b.WriteString(helpStyle.Render(getHelpText(m.currentView)))
    
    return b.String()
}
```

### Implementation Checklist

- [ ] Add database queries for download management
- [ ] Create `DownloadWithBinary` and `CacheStats` data structures
- [ ] Add model state fields for downloads view
- [ ] Implement Bubble Tea commands for async operations
- [ ] Create `internal/tui/view_downloads.go` with full rendering
- [ ] Add update logic for navigation and actions
- [ ] Add confirmation dialog for bulk operations
- [ ] Implement checksum verification UI feedback
- [ ] Add loading and empty states
- [ ] Write unit tests for downloads view
- [ ] Write integration tests for download operations
- [ ] Update help text for downloads view
- [ ] Document downloads view functionality

---

## Configuration View Implementation

### Overview

The Configuration view will provide interactive management of binmate's global configuration, allowing users to:
- View current configuration settings
- Edit log level, date format, and other options
- Sync config.json to database
- Manage configuration file location
- Export database binaries back to config.json

### Data Model Extensions

#### New Configuration Operations

```go
// In internal/core/config/config.go

// UpdateConfig updates configuration values
func UpdateConfig(cfg *Config, updates ConfigUpdates) error

// ExportConfig exports binaries from database to config file
func ExportConfig(dbService *repository.Service, configPath string) error

// ValidateConfigValue validates a configuration setting
func ValidateConfigValue(key string, value interface{}) error
```

#### New Data Structures

```go
type ConfigUpdates struct {
    LogLevel   *string
    DateFormat *string
    Version    *int
}

type ConfigField struct {
    Name        string
    Description string
    Value       string
    Editable    bool
    Type        string // "string", "int", "bool", "select"
    Options     []string // For select type
}
```

### UI Layout

```
┌─────────────────────────────────────────────────────────────────┐
│ 📦 Binmate - Binary Manager                                     │
│                                                                  │
│ [Binaries] [Versions] [Downloads] [Config] [Help]               │
│                                                                  │
│ ⚙️  Configuration Settings                                       │
│ ──────────────────────────────────────────────────────────────  │
│                                                                  │
│ > Log Level:        info             [debug|info|warn|error]    │
│   Date Format:      2006-01-02       [editable]                 │
│   Config Version:   1                [read-only]                │
│   Config Path:      ~/.binmate/config.json                      │
│   Database Path:    ~/.binmate/binmate.db                       │
│                                                                  │
│ Binaries: 12 in config | 15 in database                         │
│                                                                  │
│ Actions:                                                         │
│ [s] Sync config → database                                      │
│ [e] Export database → config file                               │
│ [r] Reload configuration                                        │
│                                                                  │
│ ↑/↓: navigate • Enter: edit • s: sync • e: export • q: quit     │
└─────────────────────────────────────────────────────────────────┘
```

### Model State Extensions

```go
// In internal/tui/model.go

type model struct {
    // ... existing fields ...
    
    // Configuration view state
    configFields          []ConfigField
    selectedConfigIdx     int
    editingConfig         bool
    configEditInput       textinput.Model
    configSyncInProgress  bool
    configExportInProgress bool
}
```

### Update Logic Extensions

```go
// In internal/tui/update.go

case viewConfiguration:
    // If editing a field
    if m.editingConfig {
        switch {
        case key.Matches(msg, keys.enter):
            // Save the edited value
            return m, saveConfigValueCmd(
                m.configFields[m.selectedConfigIdx].Name,
                m.configEditInput.Value(),
            )
            
        case key.Matches(msg, keys.escape):
            // Cancel editing
            m.editingConfig = false
            m.configEditInput.Blur()
            return m, nil
        
        default:
            // Pass input to text input
            m.configEditInput, cmd = m.configEditInput.Update(msg)
            return m, cmd
        }
    }
    
    // Normal navigation
    switch {
    case key.Matches(msg, keys.up):
        if m.selectedConfigIdx > 0 {
            m.selectedConfigIdx--
        }
        
    case key.Matches(msg, keys.down):
        if m.selectedConfigIdx < len(m.configFields)-1 {
            m.selectedConfigIdx++
        }
        
    case key.Matches(msg, keys.enter):
        // Start editing selected field if editable
        field := m.configFields[m.selectedConfigIdx]
        if field.Editable {
            m.editingConfig = true
            m.configEditInput.SetValue(field.Value)
            m.configEditInput.Focus()
            return m, nil
        }
        
    case msg.String() == "s":
        // Sync config to database
        return m, syncConfigCmd(m.config, m.dbService)
        
    case msg.String() == "e":
        // Export database to config file
        return m, exportConfigCmd(m.dbService, m.config.Path)
        
    case msg.String() == "r":
        // Reload configuration
        return m, reloadConfigCmd()
    }
```

### Bubble Tea Commands

```go
// New commands in internal/tui/messages.go

// loadConfigFieldsMsg contains configuration fields
type loadConfigFieldsMsg struct {
    fields []ConfigField
    err    error
}

// saveConfigValueMsg contains save result
type saveConfigValueMsg struct {
    fieldName string
    success   bool
    err       error
}

// syncConfigMsg contains sync result
type syncConfigMsg struct {
    synced int
    err    error
}

// exportConfigMsg contains export result
type exportConfigMsg struct {
    exported int
    err      error
}

// reloadConfigMsg contains reloaded config
type reloadConfigMsg struct {
    config *config.Config
    err    error
}

// Commands
func loadConfigFieldsCmd(cfg *config.Config) tea.Cmd
func saveConfigValueCmd(fieldName, value string) tea.Cmd
func syncConfigCmd(cfg *config.Config, dbService *repository.Service) tea.Cmd
func exportConfigCmd(dbService *repository.Service, configPath string) tea.Cmd
func reloadConfigCmd() tea.Cmd
```

### View Rendering Implementation

Replace `renderConfiguration()` with full implementation in `internal/tui/view_configuration.go`:

```go
func (m model) renderConfiguration() string {
    var b strings.Builder
    
    // Title and tabs
    b.WriteString(titleStyle.Render("📦 Binmate - Binary Manager"))
    b.WriteString("\n\n")
    b.WriteString(m.renderTabs())
    
    // Error/success messages
    if m.errorMessage != "" {
        b.WriteString(errorStyle.Render("Error: " + m.errorMessage))
        b.WriteString("\n\n")
    }
    if m.successMessage != "" {
        b.WriteString(successStyle.Render("✓ " + m.successMessage))
        b.WriteString("\n\n")
    }
    
    // Loading state
    if m.loading || m.configSyncInProgress || m.configExportInProgress {
        msg := "Loading configuration..."
        if m.configSyncInProgress {
            msg = "Syncing configuration to database..."
        } else if m.configExportInProgress {
            msg = "Exporting database to config file..."
        }
        b.WriteString(loadingStyle.Render(msg))
        b.WriteString("\n\n")
        b.WriteString(helpStyle.Render(getHelpText(m.currentView)))
        return b.String()
    }
    
    // Header
    b.WriteString(headerStyle.Render("⚙️  Configuration Settings"))
    b.WriteString("\n\n")
    
    // Configuration fields
    for i, field := range m.configFields {
        // Field selection indicator
        indicator := "  "
        if i == m.selectedConfigIdx {
            indicator = "> "
        }
        
        // Field name
        fieldName := formLabelStyle.Render(field.Name + ":")
        
        // Field value (with edit mode handling)
        var fieldValue string
        if m.editingConfig && i == m.selectedConfigIdx {
            fieldValue = m.configEditInput.View()
        } else {
            valueStyle := normalStyle
            if !field.Editable {
                valueStyle = mutedStyle
            }
            fieldValue = valueStyle.Render(field.Value)
        }
        
        // Field metadata (type hints)
        var metadata string
        if field.Type == "select" {
            metadata = mutedStyle.Render(fmt.Sprintf("[%s]", strings.Join(field.Options, "|")))
        } else if field.Editable {
            metadata = mutedStyle.Render("[editable]")
        } else {
            metadata = mutedStyle.Render("[read-only]")
        }
        
        // Render field line
        fieldLine := fmt.Sprintf(
            "%s%-20s %-30s %s",
            indicator,
            fieldName,
            fieldValue,
            metadata,
        )
        
        b.WriteString(fieldLine)
        b.WriteString("\n")
        
        // Field description
        if field.Description != "" {
            desc := mutedStyle.Render("    " + field.Description)
            b.WriteString(desc)
            b.WriteString("\n")
        }
    }
    
    b.WriteString("\n")
    
    // Binary counts
    if m.config != nil {
        configCount := len(m.config.Binaries)
        dbCount := 0
        if binaries, err := m.dbService.Binaries().List(); err == nil {
            dbCount = len(binaries)
        }
        
        countsText := fmt.Sprintf(
            "Binaries: %d in config | %d in database",
            configCount,
            dbCount,
        )
        b.WriteString(mutedStyle.Render(countsText))
        b.WriteString("\n\n")
    }
    
    // Actions section
    b.WriteString(headerStyle.Render("Actions:"))
    b.WriteString("\n")
    b.WriteString("  [s] Sync config → database\n")
    b.WriteString("  [e] Export database → config file\n")
    b.WriteString("  [r] Reload configuration\n")
    b.WriteString("\n")
    
    b.WriteString(helpStyle.Render(getHelpText(m.currentView)))
    
    return b.String()
}
```

### Implementation Checklist

- [ ] Add configuration update operations
- [ ] Create `ConfigField` data structures
- [ ] Add model state fields for configuration editing
- [ ] Implement Bubble Tea commands for config operations
- [ ] Create `internal/tui/view_configuration.go` with full rendering
- [ ] Add update logic for field editing
- [ ] Implement inline editing with text input
- [ ] Add sync and export operations
- [ ] Add validation for configuration values
- [ ] Write unit tests for configuration view
- [ ] Write integration tests for config operations
- [ ] Update help text for configuration view
- [ ] Document configuration view functionality

---

## Help View Enhancement

### Overview

The Help view currently provides static help text. Enhancement will add:
- Interactive navigation through help topics
- Searchable help content
- Context-sensitive examples
- Keyboard shortcut reference
- Version and system information

### UI Layout

```
┌─────────────────────────────────────────────────────────────────┐
│ 📦 Binmate - Binary Manager                                     │
│                                                                  │
│ [Binaries] [Versions] [Downloads] [Config] [Help]               │
│                                                                  │
│ 📚 Help & Documentation                          Search: _______ │
│ ──────────────────────────────────────────────────────────────  │
│                                                                  │
│ Topics:                   │ Getting Started                      │
│ > Getting Started        │ ─────────────────────────────────── │
│   Managing Binaries      │ Binmate helps you install and manage │
│   Version Control        │ multiple versions of command-line    │
│   Downloads & Cache      │ binaries from GitHub releases.       │
│   Configuration          │                                      │
│   Keyboard Shortcuts     │ First Steps:                         │
│   Troubleshooting        │ 1. Add a binary: press 'a'          │
│   About                  │ 2. Install version: press 'i'        │
│                           │ 3. Switch version: Enter → 's'       │
│                           │                                      │
│                           │ Example:                             │
│                           │ $ binmate                            │
│                           │ (In TUI, press 'a')                  │
│                           │ URL: github.com/cli/cli/releases/... │
│                           │                                      │
│ ↑/↓: navigate topics • Enter: select • /: search • q: quit      │
└─────────────────────────────────────────────────────────────────┘
```

### Model State Extensions

```go
// In internal/tui/model.go

type HelpTopic struct {
    ID          string
    Title       string
    Content     string
    Keywords    []string
    Examples    []string
}

type model struct {
    // ... existing fields ...
    
    // Help view state
    helpTopics         []HelpTopic
    selectedTopicIdx   int
    helpSearchInput    textinput.Model
    helpSearchActive   bool
    filteredTopics     []HelpTopic
}
```

### Help Topics Structure

```go
// In internal/tui/help_topics.go

var helpTopics = []HelpTopic{
    {
        ID:    "getting-started",
        Title: "Getting Started",
        Content: `Binmate helps you install and manage multiple versions...`,
        Keywords: []string{"intro", "start", "begin", "first"},
        Examples: []string{
            "$ binmate",
            "(In TUI, press 'a' to add a binary)",
        },
    },
    {
        ID:    "managing-binaries",
        Title: "Managing Binaries",
        Content: `Add, update, and remove binaries...`,
        Keywords: []string{"add", "remove", "update", "install"},
        Examples: []string{
            "Add: press 'a' in Binaries view",
            "Install: press 'i' and enter version",
            "Update: press 'u' to update to latest",
        },
    },
    {
        ID:    "version-control",
        Title: "Version Control",
        Content: `Switch between installed versions...`,
        Keywords: []string{"switch", "version", "activate"},
        Examples: []string{
            "View versions: select binary and press Enter",
            "Switch version: navigate to version and press 's'",
        },
    },
    {
        ID:    "downloads-cache",
        Title: "Downloads & Cache",
        Content: `Manage cached downloads...`,
        Keywords: []string{"download", "cache", "disk", "space"},
        Examples: []string{
            "View cache: press '3' or navigate to Downloads tab",
            "Clear download: select and press 'd'",
        },
    },
    {
        ID:    "configuration",
        Title: "Configuration",
        Content: `Customize binmate settings...`,
        Keywords: []string{"config", "settings", "preferences"},
        Examples: []string{
            "Edit config: press '4' or navigate to Config tab",
            "Sync config: press 's' in Config view",
        },
    },
    {
        ID:    "keyboard-shortcuts",
        Title: "Keyboard Shortcuts",
        Content: `
Global:
  1-4        Switch tabs directly
  Tab        Cycle to next tab
  Shift+Tab  Cycle to previous tab
  q          Quit application
  Ctrl+C     Force quit

Binaries List View:
  ↑/↓        Navigate binaries
  Enter      View versions
  a          Add new binary
  i          Install specific version
  u          Update to latest
  U          Update all binaries
  r          Remove binary
  c          Check for updates
  m          Import binary

Versions View:
  ↑/↓        Navigate versions
  s/Enter    Switch to version
  d          Delete version
  Esc        Back to binaries list

Downloads View:
  ↑/↓        Navigate downloads
  v          Verify checksum
  d          Delete download
  D          Delete all downloads
  r          Refresh statistics

Configuration View:
  ↑/↓        Navigate fields
  Enter      Edit field
  s          Sync to database
  e          Export to config file
  r          Reload configuration
`,
        Keywords: []string{"keyboard", "shortcuts", "keys", "hotkeys"},
    },
    {
        ID:    "troubleshooting",
        Title: "Troubleshooting",
        Content: `
Common Issues:

Q: Binary not found after installation
A: Ensure install path is in your $PATH environment variable

Q: Permission denied errors
A: Check file permissions on install directory

Q: Failed to download release
A: Verify internet connection and GitHub API access

Q: Version shows as inactive after switching
A: Check symlink in install directory

Q: Config sync not working
A: Verify config.json syntax and file permissions
`,
        Keywords: []string{"error", "problem", "issue", "fix", "help"},
    },
    {
        ID:    "about",
        Title: "About Binmate",
        Content: `
Binmate - Binary Version Manager
Version: 0.1.0
Author: cturner8
License: MIT

A terminal-based tool for managing multiple versions of 
command-line binaries from GitHub releases.

Features:
✓ Install binaries from GitHub releases
✓ Manage multiple versions simultaneously
✓ Switch active versions with symlinks
✓ Cache downloads for offline access
✓ Interactive TUI and CLI modes
✓ SQLite-based state management

Repository: github.com/cturner8/binmate
Documentation: github.com/cturner8/binmate/wiki
`,
        Keywords: []string{"version", "about", "info", "author"},
    },
}
```

### Update Logic Extensions

```go
// In internal/tui/update.go

case viewHelp:
    // Search mode active
    if m.helpSearchActive {
        switch {
        case key.Matches(msg, keys.enter):
            // Apply search filter
            m.filteredTopics = filterTopics(m.helpTopics, m.helpSearchInput.Value())
            m.helpSearchActive = false
            m.helpSearchInput.Blur()
            return m, nil
            
        case key.Matches(msg, keys.escape):
            // Cancel search
            m.helpSearchActive = false
            m.helpSearchInput.Blur()
            m.filteredTopics = m.helpTopics
            return m, nil
        
        default:
            // Pass input to search
            m.helpSearchInput, cmd = m.helpSearchInput.Update(msg)
            return m, cmd
        }
    }
    
    // Normal navigation
    switch {
    case key.Matches(msg, keys.up):
        if m.selectedTopicIdx > 0 {
            m.selectedTopicIdx--
        }
        
    case key.Matches(msg, keys.down):
        topics := m.filteredTopics
        if len(topics) == 0 {
            topics = m.helpTopics
        }
        if m.selectedTopicIdx < len(topics)-1 {
            m.selectedTopicIdx++
        }
        
    case msg.String() == "/":
        // Activate search
        m.helpSearchActive = true
        m.helpSearchInput.Focus()
        return m, nil
    }
```

### View Rendering Implementation

Create enhanced `internal/tui/view_help.go`:

```go
func (m model) renderHelp() string {
    var b strings.Builder
    
    // Title
    b.WriteString(titleStyle.Render("📦 Binmate - Binary Manager"))
    b.WriteString("\n\n")
    
    // Tabs
    b.WriteString(m.renderTabs())
    
    // Header with search
    headerText := "📚 Help & Documentation"
    searchText := ""
    if m.helpSearchActive {
        searchText = "Search: " + m.helpSearchInput.View()
    } else {
        searchText = "Search: _______"
    }
    
    // Use lipgloss to create header row with search on right
    headerStyle := lipgloss.NewStyle().Width(40)
    searchStyle := lipgloss.NewStyle().Width(40).Align(lipgloss.Right)
    
    headerRow := lipgloss.JoinHorizontal(
        lipgloss.Top,
        headerStyle.Render(headerText),
        searchStyle.Render(mutedStyle.Render(searchText)),
    )
    b.WriteString(headerRow)
    b.WriteString("\n")
    b.WriteString(strings.Repeat("─", 80))
    b.WriteString("\n\n")
    
    // Two-column layout: topics list on left, content on right
    topics := m.filteredTopics
    if len(topics) == 0 {
        topics = m.helpTopics
    }
    
    // Left column: topics list
    var leftCol strings.Builder
    leftCol.WriteString(headerStyle.Render("Topics:"))
    leftCol.WriteString("\n")
    
    for i, topic := range topics {
        indicator := "  "
        style := normalStyle
        if i == m.selectedTopicIdx {
            indicator = "> "
            style = selectedStyle
        }
        topicLine := fmt.Sprintf("%s%s", indicator, topic.Title)
        leftCol.WriteString(style.Render(topicLine))
        leftCol.WriteString("\n")
    }
    
    // Right column: selected topic content
    var rightCol strings.Builder
    if m.selectedTopicIdx < len(topics) {
        selectedTopic := topics[m.selectedTopicIdx]
        rightCol.WriteString(headerStyle.Render(selectedTopic.Title))
        rightCol.WriteString("\n")
        rightCol.WriteString(strings.Repeat("─", 50))
        rightCol.WriteString("\n")
        rightCol.WriteString(selectedTopic.Content)
        rightCol.WriteString("\n")
        
        if len(selectedTopic.Examples) > 0 {
            rightCol.WriteString("\n")
            rightCol.WriteString(headerStyle.Render("Examples:"))
            rightCol.WriteString("\n")
            for _, example := range selectedTopic.Examples {
                rightCol.WriteString(mutedStyle.Render("  " + example))
                rightCol.WriteString("\n")
            }
        }
    }
    
    // Join columns
    leftColStyle := lipgloss.NewStyle().Width(25).Align(lipgloss.Left)
    rightColStyle := lipgloss.NewStyle().Width(55).Align(lipgloss.Left)
    
    content := lipgloss.JoinHorizontal(
        lipgloss.Top,
        leftColStyle.Render(leftCol.String()),
        rightColStyle.Render(rightCol.String()),
    )
    
    b.WriteString(content)
    b.WriteString("\n\n")
    b.WriteString(helpStyle.Render(getHelpText(m.currentView)))
    
    return b.String()
}

// filterTopics filters topics by search query (matches title or keywords)
func filterTopics(topics []HelpTopic, query string) []HelpTopic {
    if query == "" {
        return topics
    }
    
    query = strings.ToLower(query)
    var filtered []HelpTopic
    
    for _, topic := range topics {
        // Match against title
        if strings.Contains(strings.ToLower(topic.Title), query) {
            filtered = append(filtered, topic)
            continue
        }
        
        // Match against keywords
        for _, keyword := range topic.Keywords {
            if strings.Contains(strings.ToLower(keyword), query) {
                filtered = append(filtered, topic)
                break
            }
        }
    }
    
    return filtered
}
```

### Implementation Checklist

- [ ] Create `HelpTopic` data structure
- [ ] Create comprehensive help topics content
- [ ] Add model state fields for help view
- [ ] Implement search functionality
- [ ] Create `internal/tui/view_help.go` with enhanced rendering
- [ ] Add update logic for topic navigation and search
- [ ] Implement two-column layout
- [ ] Add keyword-based topic filtering
- [ ] Write unit tests for help view
- [ ] Write tests for search functionality
- [ ] Update help view key bindings
- [ ] Document help view functionality

---

## Shared Infrastructure

### Common Components Needed

#### 1. Confirmation Dialog Component

Reusable confirmation dialog for destructive operations:

```go
// In internal/tui/components/dialog.go

type ConfirmationDialog struct {
    Title       string
    Message     string
    Options     []ConfirmOption
    Visible     bool
    SelectedIdx int
}

type ConfirmOption struct {
    Key         string
    Label       string
    Description string
    Action      string // identifier for the action
}

func NewConfirmationDialog(title, message string, options []ConfirmOption) *ConfirmationDialog
func (d *ConfirmationDialog) View() string
func (d *ConfirmationDialog) Update(msg tea.Msg) (*ConfirmationDialog, tea.Cmd)
```

#### 2. Progress Indicator Component

For long-running operations:

```go
// In internal/tui/components/progress.go

type ProgressIndicator struct {
    Current int
    Total   int
    Message string
    Visible bool
}

func NewProgressIndicator(total int, message string) *ProgressIndicator
func (p *ProgressIndicator) Increment()
func (p *ProgressIndicator) View() string
```

#### 3. Inline Editor Component

For editing configuration values:

```go
// In internal/tui/components/editor.go

type InlineEditor struct {
    Label       string
    Value       string
    Placeholder string
    Validator   func(string) error
    Input       textinput.Model
    Active      bool
}

func NewInlineEditor(label, value, placeholder string) *InlineEditor
func (e *InlineEditor) Activate()
func (e *InlineEditor) Deactivate()
func (e *InlineEditor) View() string
func (e *InlineEditor) Update(msg tea.Msg) (*InlineEditor, tea.Cmd)
```

#### 4. Status Bar Component

Enhanced status bar for all views:

```go
// In internal/tui/components/statusbar.go

type StatusBar struct {
    LeftText   string
    CenterText string
    RightText  string
    Style      lipgloss.Style
}

func NewStatusBar() *StatusBar
func (s *StatusBar) SetError(msg string)
func (s *StatusBar) SetSuccess(msg string)
func (s *StatusBar) SetInfo(msg string)
func (s *StatusBar) View(width int) string
```

### Implementation Checklist

- [ ] Create `internal/tui/components/` directory
- [ ] Implement `ConfirmationDialog` component
- [ ] Implement `ProgressIndicator` component
- [ ] Implement `InlineEditor` component
- [ ] Implement `StatusBar` component
- [ ] Write unit tests for each component
- [ ] Document component APIs
- [ ] Create example usage for each component
- [ ] Integrate components into existing views

---

## Testing Strategy

### Unit Tests

Each view and component should have comprehensive unit tests:

```go
// Example: internal/tui/view_downloads_test.go

func TestRenderDownloads(t *testing.T) {
    tests := []struct {
        name     string
        model    model
        contains []string
    }{
        {
            name: "empty downloads",
            model: model{
                downloads: []DownloadWithBinary{},
                loading:   false,
            },
            contains: []string{"No cached downloads"},
        },
        {
            name: "with downloads",
            model: model{
                downloads: []DownloadWithBinary{
                    {
                        Download: &database.Download{
                            ID:      1,
                            Version: "1.0.0",
                        },
                        Binary: &database.Binary{
                            Name: "test-binary",
                        },
                    },
                },
                loading: false,
            },
            contains: []string{"test-binary", "1.0.0"},
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := tt.model.renderDownloads()
            for _, s := range tt.contains {
                if !strings.Contains(result, s) {
                    t.Errorf("expected %q in output", s)
                }
            }
        })
    }
}
```

### Integration Tests

Test actual database operations and user workflows:

```go
// Example: internal/tui/integration_test.go

func TestDownloadsWorkflow(t *testing.T) {
    // Setup in-memory database
    db := setupTestDB(t)
    defer db.Close()
    
    // Create test data
    createTestDownload(t, db, "gh", "2.40.1")
    
    // Initialize TUI model
    m := initialModel(db, testConfig())
    
    // Load downloads
    m, cmd := m.Update(loadDownloadsCmd(db))
    result := cmd()
    m, _ = m.Update(result)
    
    // Verify downloads loaded
    if len(m.downloads) != 1 {
        t.Errorf("expected 1 download, got %d", len(m.downloads))
    }
    
    // Test delete download
    m.selectedDownloadIdx = 0
    m, cmd = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
    
    // Verify download deleted
    downloads, _ := db.Downloads().List()
    if len(downloads) != 0 {
        t.Errorf("expected download to be deleted")
    }
}
```

### Manual Testing Checklist

Create a comprehensive manual testing checklist document:

```markdown
# Phase 5 Manual Testing Checklist

## Downloads View

- [ ] View loads with correct download count
- [ ] Cache statistics display correctly
- [ ] Navigate up/down through downloads
- [ ] Verify checksum shows success/failure correctly
- [ ] Delete single download removes from DB and optionally files
- [ ] Delete all downloads prompts confirmation
- [ ] Confirmation 'y' deletes DB records only
- [ ] Confirmation 'Y' deletes DB records and files
- [ ] Empty state displays when no downloads
- [ ] Loading state shows during operations
- [ ] Error messages display for failed operations
- [ ] Success messages display after operations
- [ ] Refresh updates statistics correctly

## Configuration View

- [ ] All config fields display correctly
- [ ] Navigate up/down through fields
- [ ] Enter activates editing for editable fields
- [ ] Enter does nothing for read-only fields
- [ ] Escape cancels editing
- [ ] Edited values save correctly
- [ ] Invalid values show error message
- [ ] Sync operation updates database
- [ ] Export operation creates/updates config file
- [ ] Reload operation refreshes display
- [ ] Binary counts match reality
- [ ] Help text shows correct shortcuts

## Help View

- [ ] All topics display in list
- [ ] Navigate up/down through topics
- [ ] Selected topic content shows on right
- [ ] Search activates with '/' key
- [ ] Search filters topics by title
- [ ] Search filters topics by keywords
- [ ] Escape clears search
- [ ] Enter applies search filter
- [ ] Empty search shows all topics
- [ ] Keyboard shortcuts are accurate
- [ ] Examples are clear and correct
- [ ] About info shows correct version

## Cross-View Testing

- [ ] Tab navigation works between all 5 views
- [ ] Direct tab keys (1-5) work
- [ ] Tab cycles forward correctly
- [ ] Shift+Tab cycles backward correctly
- [ ] Quit works from all views
- [ ] Error messages persist across views appropriately
- [ ] Success messages clear at appropriate times
- [ ] Window resize handles gracefully
- [ ] Narrow terminal widths work
```

---

## Implementation Timeline

### Phase 1: Foundation (Week 1-2)

**Goal:** Set up shared infrastructure and database layer

- Create shared components (dialog, progress, editor, status bar)
- Add database queries for downloads and configuration
- Create data structures and types
- Write unit tests for components

**Deliverables:**
- `internal/tui/components/` package with 4 components
- Database repository extensions
- Component unit tests (>80% coverage)

### Phase 2: Downloads View (Week 3-4)

**Goal:** Fully functional downloads management

- Implement downloads view rendering
- Add navigation and selection logic
- Implement checksum verification
- Implement delete operations (single and bulk)
- Add confirmation dialogs
- Write unit and integration tests

**Deliverables:**
- `internal/tui/view_downloads.go` (full implementation)
- Downloads update logic in `update.go`
- Downloads commands in `messages.go`
- Test suite with >75% coverage

### Phase 3: Configuration View (Week 5-6)

**Goal:** Fully functional configuration management

- Implement configuration view rendering
- Add field navigation and editing
- Implement sync and export operations
- Add validation for config values
- Write unit and integration tests

**Deliverables:**
- `internal/tui/view_configuration.go` (full implementation)
- Configuration update logic in `update.go`
- Configuration commands in `messages.go`
- Test suite with >75% coverage

### Phase 4: Help View Enhancement (Week 7)

**Goal:** Interactive and searchable help system

- Create comprehensive help topics content
- Implement two-column layout
- Add search functionality
- Add keyword filtering
- Write unit tests

**Deliverables:**
- `internal/tui/help_topics.go` (content)
- `internal/tui/view_help.go` (enhanced implementation)
- Help update logic with search
- Test suite with >80% coverage

### Phase 5: Polish and Testing (Week 8)

**Goal:** Quality assurance and documentation

- Run manual testing checklist
- Fix bugs discovered during testing
- Optimize performance
- Update documentation
- Create user guide

**Deliverables:**
- Bug fixes and optimizations
- Complete test coverage report
- Updated README with Phase 5 features
- User guide for new features

### Total Timeline: 8 weeks

---

## Risk Assessment

### High Risk Items

#### 1. Database Performance with Large Caches

**Risk:** Viewing downloads with thousands of cached files may be slow

**Mitigation:**
- Implement pagination for downloads list
- Add loading indicators for slow operations
- Consider caching statistics
- Add database indexes on frequently queried columns

#### 2. Configuration File Corruption

**Risk:** Editing config directly could corrupt JSON

**Mitigation:**
- Always validate before saving
- Create backup before writing
- Use atomic file writes (write to temp, then rename)
- Add rollback mechanism

#### 3. UI Rendering Performance

**Risk:** Complex layouts may be slow on large terminals

**Mitigation:**
- Benchmark rendering functions
- Cache rendered components where possible
- Use lazy rendering for off-screen content
- Set reasonable max dimensions

### Medium Risk Items

#### 4. Checksum Verification Time

**Risk:** Verifying large file checksums blocks UI

**Mitigation:**
- Run verification in background tea.Cmd
- Show progress indicator during verification
- Allow cancellation of long operations

#### 5. Search Performance

**Risk:** Searching help topics may be slow with many topics

**Mitigation:**
- Keep topic count reasonable (<50)
- Use simple string matching (not regex)
- Consider pre-building keyword index

### Low Risk Items

#### 6. Help Content Maintenance

**Risk:** Help content may become outdated

**Mitigation:**
- Store help in separate file for easy updates
- Add version to help content
- Include changelog reference

---

## Success Criteria

Phase 5 implementation will be considered complete when:

### Functional Requirements

- ✅ Downloads view displays all cached downloads with metadata
- ✅ Users can verify checksums and delete downloads
- ✅ Configuration view displays all editable settings
- ✅ Users can edit, save, sync, and export configuration
- ✅ Help view provides searchable, organized documentation
- ✅ All views handle loading, error, and empty states
- ✅ Tab navigation works seamlessly between all views

### Quality Requirements

- ✅ Unit test coverage >75% for all new code
- ✅ Integration tests cover key workflows
- ✅ Manual testing checklist 100% complete
- ✅ No critical or high-severity bugs
- ✅ Performance: all operations complete <2 seconds
- ✅ UI responsive on terminals 80x24 to 200x60

### Documentation Requirements

- ✅ All functions have godoc comments
- ✅ README updated with Phase 5 features
- ✅ User guide created for new functionality
- ✅ API documentation for new components
- ✅ Testing guide updated

---

## Appendix

### A. Related Files

Files that will be created or modified:

**New Files:**
- `internal/tui/components/dialog.go`
- `internal/tui/components/dialog_test.go`
- `internal/tui/components/progress.go`
- `internal/tui/components/progress_test.go`
- `internal/tui/components/editor.go`
- `internal/tui/components/editor_test.go`
- `internal/tui/components/statusbar.go`
- `internal/tui/components/statusbar_test.go`
- `internal/tui/view_downloads.go`
- `internal/tui/view_downloads_test.go`
- `internal/tui/view_configuration.go`
- `internal/tui/view_configuration_test.go`
- `internal/tui/view_help.go` (replaces placeholder)
- `internal/tui/view_help_test.go`
- `internal/tui/help_topics.go`
- `internal/tui/integration_test.go`
- `docs/PHASE_5_USER_GUIDE.md`
- `docs/PHASE_5_MANUAL_TESTING.md`

**Modified Files:**
- `internal/tui/model.go` (add state fields)
- `internal/tui/update.go` (add update logic)
- `internal/tui/messages.go` (add commands)
- `internal/database/repository/downloads.go` (add queries)
- `internal/core/config/config.go` (add config operations)
- `README.md` (document new features)

### B. Dependencies

No new external dependencies required. Uses existing:
- `github.com/charmbracelet/bubbletea`
- `github.com/charmbracelet/lipgloss`
- `github.com/charmbracelet/bubbles/textinput`

### C. Reference Implementation

See existing views for patterns:
- `internal/tui/view_binaries_list.go` - Table rendering
- `internal/tui/view_versions.go` - Detail view
- `internal/tui/view_add_binary.go` - Form input
- `internal/tui/update.go` - Key handling and state transitions

---

## Conclusion

This implementation plan provides a comprehensive roadmap for completing Phase 5 of the Binmate TUI implementation. By following this plan, the Downloads, Configuration, and Help views will be transformed from basic placeholders into fully functional, interactive interfaces that complete the TUI experience.

The plan emphasizes:
- **Incremental development** with clear phases
- **Quality assurance** through comprehensive testing
- **User experience** with loading states, confirmations, and help
- **Maintainability** through shared components and clear structure
- **Documentation** for future developers and users

Total estimated effort: **8 weeks** (1 developer)

For questions or clarifications, refer to the existing TUI implementation patterns or consult the Bubble Tea documentation.
