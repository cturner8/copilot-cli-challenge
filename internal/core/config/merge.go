package config

// MergeBinaryWithGlobal merges global configuration with binary-specific configuration.
// Binary-specific settings always take precedence over global settings.
func MergeBinaryWithGlobal(binary Binary, global GlobalConfig) Binary {
	merged := binary

	// Apply global install path if binary doesn't have one
	if merged.InstallPath == "" && global.InstallPath != "" {
		merged.InstallPath = global.InstallPath
	}

	// Apply provider-level defaults if provider config exists
	if providerDefaults, exists := global.Providers[binary.Provider]; exists {
		// Only apply authenticated if binary hasn't explicitly set it
		// Note: We can't distinguish between false (explicit) and false (default zero value)
		// For now, we assume if it's false, it could be overridden by provider default
		if !merged.Authenticated && providerDefaults.Authenticated {
			merged.Authenticated = providerDefaults.Authenticated
		}
	}

	return merged
}
