package github

import (
	"cturner8/binmate/internal/database"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"
)

// ReleaseInfo contains detailed information about a GitHub release
type ReleaseInfo struct {
	Name        string    `json:"name"`
	TagName     string    `json:"tag_name"`
	Body        string    `json:"body"`
	Draft       bool      `json:"draft"`
	Prerelease  bool      `json:"prerelease"`
	CreatedAt   time.Time `json:"created_at"`
	PublishedAt time.Time `json:"published_at"`
	HTMLURL     string    `json:"html_url"`
}

// RepositoryInfo contains basic repository information
type RepositoryInfo struct {
	Name            string `json:"name"`
	FullName        string `json:"full_name"`
	Description     string `json:"description"`
	StargazersCount int    `json:"stargazers_count"`
	ForksCount      int    `json:"forks_count"`
	HTMLURL         string `json:"html_url"`
}

// FetchReleaseNotes fetches the release notes for a specific version
func FetchReleaseNotes(binary *database.Binary, version string) (ReleaseInfo, error) {
	if binary.ProviderPath == "" {
		return ReleaseInfo{}, fmt.Errorf("path is required for binary config")
	}

	var url string
	if version == "latest" {
		url = fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", binary.ProviderPath)
	} else {
		// Need to fetch the release by tag name
		url = fmt.Sprintf("https://api.github.com/repos/%s/releases/tags/%s", binary.ProviderPath, version)
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return ReleaseInfo{}, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github+json")

	// Create HTTP client with optional authentication
	client, err := CreateHTTPClient(binary.Authenticated)
	if err != nil {
		// If authentication fails, fall back to unauthenticated
		client = &http.Client{}
	}

	resp, err := client.Do(req)
	if err != nil {
		return ReleaseInfo{}, fmt.Errorf("failed to fetch release info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return ReleaseInfo{}, fmt.Errorf("GitHub API returned %d: %s", resp.StatusCode, string(body))
	}

	var releaseInfo ReleaseInfo
	if err := json.NewDecoder(resp.Body).Decode(&releaseInfo); err != nil {
		return ReleaseInfo{}, fmt.Errorf("failed to decode release info: %w", err)
	}

	return releaseInfo, nil
}

// ListAvailableVersions fetches all available release versions for a binary
func ListAvailableVersions(binary *database.Binary, limit int) ([]ReleaseInfo, error) {
	if binary.ProviderPath == "" {
		return nil, fmt.Errorf("path is required for binary config")
	}

	if limit <= 0 {
		limit = 30 // Default limit
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/releases?per_page=%d", binary.ProviderPath, limit)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github+json")

	// Create HTTP client with optional authentication
	client, err := CreateHTTPClient(binary.Authenticated)
	if err != nil {
		// If authentication fails, fall back to unauthenticated
		client = &http.Client{}
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch releases: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API returned %d: %s", resp.StatusCode, string(body))
	}

	var releases []ReleaseInfo
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return nil, fmt.Errorf("failed to decode releases: %w", err)
	}

	// Filter out drafts
	var filteredReleases []ReleaseInfo
	for _, release := range releases {
		if !release.Draft {
			filteredReleases = append(filteredReleases, release)
		}
	}

	// Sort by published date (newest first)
	sort.Slice(filteredReleases, func(i, j int) bool {
		return filteredReleases[i].PublishedAt.After(filteredReleases[j].PublishedAt)
	})

	return filteredReleases, nil
}

// GetRepositoryInfo fetches basic repository information
func GetRepositoryInfo(binary *database.Binary) (RepositoryInfo, error) {
	if binary.ProviderPath == "" {
		return RepositoryInfo{}, fmt.Errorf("path is required for binary config")
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s", binary.ProviderPath)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return RepositoryInfo{}, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github+json")

	// Create HTTP client with optional authentication
	client, err := CreateHTTPClient(binary.Authenticated)
	if err != nil {
		// If authentication fails, fall back to unauthenticated
		client = &http.Client{}
	}

	resp, err := client.Do(req)
	if err != nil {
		return RepositoryInfo{}, fmt.Errorf("failed to fetch repository info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return RepositoryInfo{}, fmt.Errorf("GitHub API returned %d: %s", resp.StatusCode, string(body))
	}

	var repoInfo RepositoryInfo
	if err := json.NewDecoder(resp.Body).Decode(&repoInfo); err != nil {
		return RepositoryInfo{}, fmt.Errorf("failed to decode repository info: %w", err)
	}

	return repoInfo, nil
}

// StarRepository stars the configured repository for the authenticated GitHub user.
func StarRepository(binary *database.Binary) error {
	if binary.ProviderPath == "" {
		return fmt.Errorf("path is required for binary config")
	}

	pathParts := strings.Split(binary.ProviderPath, "/")
	if len(pathParts) != 2 || pathParts[0] == "" || pathParts[1] == "" {
		return fmt.Errorf("invalid GitHub repository path: %s", binary.ProviderPath)
	}

	url := fmt.Sprintf("https://api.github.com/user/starred/%s/%s", pathParts[0], pathParts[1])
	req, err := http.NewRequest("PUT", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github+json")
	client, err := CreateHTTPClient(true)
	if err != nil {
		return fmt.Errorf("failed to create authenticated HTTP client: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to star repository: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent || resp.StatusCode == http.StatusNotModified {
		return nil
	}

	body, _ := io.ReadAll(resp.Body)
	return fmt.Errorf("GitHub API returned %d: %s", resp.StatusCode, string(body))
}
