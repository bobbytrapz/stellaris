package clausewitz

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExtractTopLevelKeys(t *testing.T) {
	stellarisPath := filepath.Join(os.Getenv("HOME"), ".local", "share", "Steam", "steamapps", "common", "Stellaris")
	traitsPath := filepath.Join(stellarisPath, "common", "traits", "04_species_traits.txt")
	
	if _, err := os.Stat(traitsPath); os.IsNotExist(err) {
		t.Skip("Stellaris installation not found, skipping test")
	}

	keys, err := ExtractTopLevelKeys(traitsPath)
	if err != nil {
		t.Fatalf("Failed to extract keys: %v", err)
	}

	if len(keys) == 0 {
		t.Fatal("Expected to find some keys, got 0")
	}

	foundAdaptive := false
	for _, key := range keys {
		if key == "trait_adaptive" {
			foundAdaptive = true
			break
		}
	}

	if !foundAdaptive {
		t.Errorf("Expected to find 'trait_adaptive' in keys. Found %d keys total. First few: %v", len(keys), keys[:10])
	}
}
