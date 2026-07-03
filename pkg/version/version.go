package version

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// LauncherSettings represents the structure of Stellaris launcher-settings.json
type LauncherSettings struct {
	ModsCompatibilityVersion string `json:"modsCompatibilityVersion"`
}

// GetModCompatibilityVersion reads the launcher-settings.json file in the Stellaris
// root directory to automatically detect the currently installed game version's 
// compatibility target (e.g. "4.4").
func GetModCompatibilityVersion(stellarisPath string) (string, error) {
	settingsPath := filepath.Join(stellarisPath, "launcher-settings.json")
	
	fileData, err := os.ReadFile(settingsPath)
	if err != nil {
		return "", fmt.Errorf("failed to read launcher-settings.json: %w", err)
	}

	var settings LauncherSettings
	if err := json.Unmarshal(fileData, &settings); err != nil {
		return "", fmt.Errorf("failed to parse launcher-settings.json: %w", err)
	}

	if settings.ModsCompatibilityVersion == "" {
		// Fallback if the field is missing
		return "3", nil
	}

	return settings.ModsCompatibilityVersion, nil
}
