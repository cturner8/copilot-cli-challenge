package buildinfo

import (
	"runtime/debug"
	"strings"
)

const (
	DefaultVersion = "dev"
	DefaultUnknown = "unknown"
)

type Info struct {
	Version  string
	Commit   string
	Date     string
	Modified bool
}

var readBuildInfo = debug.ReadBuildInfo

// Resolve returns build metadata preferring linker-injected values and
// falling back to runtime build information when available.
func Resolve(version, commit, date string) Info {
	info := Info{
		Version: strings.TrimSpace(version),
		Commit:  strings.TrimSpace(commit),
		Date:    strings.TrimSpace(date),
	}

	if buildInfo, ok := readBuildInfo(); ok {
		if isUnsetVersion(info.Version) {
			mainVersion := strings.TrimSpace(buildInfo.Main.Version)
			if !isUnsetVersion(mainVersion) {
				info.Version = mainVersion
			}
		}

		for _, setting := range buildInfo.Settings {
			switch setting.Key {
			case "vcs.revision":
				if isUnsetUnknown(info.Commit) {
					info.Commit = strings.TrimSpace(setting.Value)
				}
			case "vcs.time":
				if isUnsetUnknown(info.Date) {
					info.Date = strings.TrimSpace(setting.Value)
				}
			case "vcs.modified":
				info.Modified = strings.EqualFold(strings.TrimSpace(setting.Value), "true")
			}
		}
	}

	if isUnsetVersion(info.Version) {
		info.Version = DefaultVersion
	}
	if isUnsetUnknown(info.Commit) {
		info.Commit = DefaultUnknown
	}
	if isUnsetUnknown(info.Date) {
		info.Date = DefaultUnknown
	}

	return info
}

func isUnsetVersion(version string) bool {
	normalised := strings.TrimSpace(strings.ToLower(version))
	return normalised == "" || normalised == "dev" || normalised == "(devel)" || normalised == "devel"
}

func isUnsetUnknown(value string) bool {
	normalised := strings.TrimSpace(strings.ToLower(value))
	return normalised == "" || normalised == DefaultUnknown
}
