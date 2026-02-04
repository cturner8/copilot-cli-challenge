# TUI Implementation Plan - Binmate

## Problem Statement

The binmate application currently has a placeholder TUI that needs to be implemented with read-only operations and binary configuration management. The TUI should provide an interactive interface for viewing configured binaries, their installed versions, and adding new binary configurations by parsing GitHub release URLs.

## Approach

Build a Bubble Tea-based TUI with multiple views:
1. **Binaries List View** - Main view showing all configured binaries with metadata
2. **Versions View** - Detail view showing installed versions for a selected binary
3. **Add Binary View** - Form-based workflow for adding new binaries via URL parsing

The implementation will use Lipgloss for styling and layout, and integrate with the existing database layer for all data operations.

## User Confirmations

- ✅ New binaries added via TUI will be saved to database only (not config.json)
- ✅ Database is the source of truth for all binaries (config.json may be empty)
- ✅ Version ordering will use installation date (not semantic version parsing)
- ✅ URL parsing will only support GitHub release URLs (matching current provider)

## Workplan

### Phase 1: Setup and Dependencies

- [ ] Install Lipgloss dependency (`go get github.com/charmbracelet/lipgloss`)
- [ ] Create TUI package structure:
  - [ ] `internal/tui/styles.go` - Lipgloss style definitions
  - [ ] `internal/tui/views.go` - View constants and state management
  - [ ] `internal/tui/keys.go` - Key bindings
  - [ ] `internal/tui/messages.go` - Custom Bubble Tea messages

### Phase 2: Data Layer Integration

- [ ] Create `internal/tui/data.go` with helper functions:
  - [ ] `getBinariesWithMetadata()` - Fetch binaries with active version + install count
  - [ ] `getVersionsForBinary()` - Fetch all installations for a binary (ordered by date)
  - [ ] `getActiveVersion()` - Get currently active version for a binary
- [ ] Pass `DBService` and `Config` to TUI from root command

### Phase 3: Binaries List View

- [ ] Update `model.go` to include:
  - [ ] View state enum (list/versions/add)
  - [ ] Database service reference
  - [ ] Binaries list with metadata
  - [ ] Selected binary index
  - [ ] Error state
- [ ] Implement `init.go` to load initial binaries data
- [ ] Create `view_binaries_list.go`:
  - [ ] Render table/list of binaries with columns:
    - Binary name
    - Provider (GitHub)
    - Active version
    - Number of installed versions
  - [ ] Visual indicator for selected binary
  - [ ] Status bar with key hints
- [ ] Update `update.go` to handle:
  - [ ] Up/Down arrow navigation
  - [ ] Enter to view versions (transitions to versions view)
  - [ ] 'a' key to add new binary (transitions to add view)
  - [ ] 'q' to quit

### Phase 4: Versions View

- [ ] Create `view_versions.go`:
  - [ ] Header showing selected binary details
  - [ ] Table of installed versions with columns:
    - Version string
    - Installed date (formatted)
    - Active indicator (✓ for current version)
    - File size (human-readable)
    - Install path
  - [ ] Status bar with key hints
- [ ] Update `update.go` to handle:
  - [ ] Escape/Back to return to binaries list
  - [ ] Up/Down arrow navigation (for future write operations)

### Phase 5: Placeholder Views

While out of scope for the current implementation, to aid with layout and to avoid future refactoring/reorganising, add the following placeholder views for future functionality:

- [ ] Create "Downloads" view - interface that will allow to manage cached asset downloads
- [ ] Create "Configuration" view - interface that will allow management of global configuration/settings
- [ ] Create "Help" view - interface to provide user guidance on using the TUI/CLI application

### Phase 6: URL Parser

- [ ] Create `internal/core/url/parser.go`:
  - [ ] `ParseGitHubReleaseURL(url string)` function that extracts:
    - Owner (e.g., "github")
    - Repo (e.g., "copilot-cli")
    - Version (e.g., "v0.0.400")
    - Asset name (e.g., "copilot-linux-x64.tar.gz")
    - Format (detect from extension: .tar.gz or .zip)
  - [ ] Validation to ensure URL matches GitHub release pattern
  - [ ] Error handling for malformed URLs
- [ ] Create `internal/core/url/parser_test.go`:
  - [ ] Test valid GitHub release URLs
  - [ ] Test invalid URLs
  - [ ] Test edge cases (different formats, no version, etc.)

### Phase 7: Add Binary View - URL Input

- [ ] Create `view_add_binary_url.go`:
  - [ ] Text input for URL (using `textinput` from Bubble Tea libraries)
  - [ ] Instructions/help text
  - [ ] Real-time validation feedback
  - [ ] Status bar with key hints
- [ ] Update `update.go` to handle:
  - [ ] Text input for URL
  - [ ] Enter to parse and continue (transitions to config form)
  - [ ] Escape to cancel and return to binaries list
  - [ ] Display parsing errors

### Phase 8: Add Binary View - Configuration Form

- [ ] Create `view_add_binary_form.go`:
  - [ ] Display parsed configuration fields (editable):
    - User ID (generated from name, editable)
    - Name (extracted from repo, editable)
    - Provider (fixed to "github")
    - Path (extracted as owner/repo, editable)
    - Format (detected from URL, editable)
    - InstallPath (optional, editable)
    - AssetRegex (optional, editable)
    - ReleaseRegex (optional, editable)
  - [ ] Form navigation (Tab/Shift+Tab between fields)
  - [ ] Field validation indicators
  - [ ] Status bar with key hints
- [ ] Update `update.go` to handle:
  - [ ] Tab/Shift+Tab for field navigation
  - [ ] Text input for each field
  - [ ] Ctrl+S to save (validates and creates binary in DB)
  - [ ] Escape to cancel and return to binaries list
- [ ] Implement save logic:
  - [ ] Validate required fields
  - [ ] Create database binary entry
  - [ ] Handle errors (duplicate ID, validation failures)
  - [ ] Return to binaries list on success

### Phase 9: Styling and Polish

- [ ] Define consistent colour scheme in `styles.go`:
  - [ ] Primary, secondary, accent colours
  - [ ] Selected item style
  - [ ] Error message style
  - [ ] Success message style
  - [ ] Border styles
- [ ] Apply styles to all views:
  - [ ] Consistent spacing and padding
  - [ ] Clear visual hierarchy
  - [ ] Responsive layout (within terminal bounds)
- [ ] Add loading states:
  - [ ] Spinner while fetching data
  - [ ] Loading message
- [ ] Add empty states:
  - [ ] "No binaries configured" message
  - [ ] "No versions installed" message
- [ ] Add help footer:
  - [ ] Context-sensitive key hints for each view

### Phase 10: Testing and Validation

- [ ] Manual testing:
  - [ ] Test with empty database
  - [ ] Test with populated database (sync config.json first)
  - [ ] Test navigation between all views
  - [ ] Test adding a new binary via URL
  - [ ] Test URL parsing with various formats
  - [ ] Test form validation and error handling
  - [ ] Test with narrow terminal widths
- [ ] Integration testing:
  - [ ] Verify database writes are correct
  - [ ] Verify symlink paths are properly displayed
  - [ ] Verify date formatting
- [ ] Error scenario testing:
  - [ ] Invalid URLs
  - [ ] Duplicate binary IDs
  - [ ] Database connection failures
  - [ ] Missing data edge cases

## Implementation Notes

### Database Integration
- The root command's `PersistentPreRun` hook already initializes the database
- Add a command level `PreRun` that invokes a full sync operation (similar to manually invoking the `sync` command)
- Need to pass `DBService` reference to TUI's `InitProgram()` function
- All data fetching should use repository methods from `internal/database/repository`

### URL Parsing Strategy
For a URL like `https://github.com/github/copilot-cli/releases/download/v0.0.400/copilot-linux-x64.tar.gz`:
1. Use the go `url` standard library
2. Validate it contains "github.com" domain
3. Extract owner (segment 3), repo (segment 4)
4. Extract version from segment 7 (after "download/")
5. Extract asset name from last segment
6. Detect format from file extension

### Generated Fields
- `User ID`: Default to asset prefix (e.g., "copilot-linux-x64.tar.gz" → "copilot")
- `Name`: Default to asset prefix (e.g., "copilot-linux-x64.tar.gz" → "copilot")
- `Path`: Format as "owner/repo" (e.g., "github/copilot-cli")
- `Format`: Detect from extension (".tar.gz" or ".zip")

### Key Bindings
- **Binaries List View**: ↑/↓ navigate, Enter view versions, 'a' add binary, 'q' quit
- **Versions View**: ↑/↓ navigate, Esc back to list
- **Add URL View**: Type URL, Enter parse, Esc cancel
- **Add Form View**: Tab/Shift+Tab navigate fields, Ctrl+S save, Esc cancel

### Lipgloss Usage
- Use `lipgloss.JoinVertical` and `lipgloss.JoinHorizontal` for layout
- Define reusable style presets in `styles.go`
- Use `lipgloss.Width` and `lipgloss.Height` for responsive sizing
- Apply borders using `lipgloss.Border` styles

## Future Enhancements (Out of Scope)
- Switch active version (write operation)
- Delete installed versions
- Search/filter binaries
- Sort by different columns
- Update binary configuration
- Delete binary
- Install new versions from TUI
