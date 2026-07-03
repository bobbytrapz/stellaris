package gamedata

import (
	"testing"
)

func TestLoadGameData(t *testing.T) {
	stellarisPath := DefaultStellarisPath()
	data, err := Load(stellarisPath)
	if err != nil {
		t.Fatalf("Failed to load game data: %v", err)
	}

	if len(data.Traits) == 0 {
		t.Error("No traits loaded")
	}
	if len(data.Ethics) == 0 {
		t.Error("No ethics loaded")
	}
	if len(data.Civics) == 0 {
		t.Error("No civics loaded")
	}
	if len(data.Authorities) == 0 {
		t.Error("No authorities loaded")
	}

	// Spot checks for common built-ins
	if !data.IsValidTrait("trait_adaptive") {
		t.Error("Expected to find trait_adaptive")
	}
	if !data.IsValidEthic("ethic_fanatic_militarist") {
		t.Error("Expected to find ethic_fanatic_militarist")
	}
	if !data.IsValidCivic("civic_idealistic_foundation") {
		t.Error("Expected to find civic_idealistic_foundation")
	}
	if !data.IsValidAuthority("auth_democratic") {
		t.Error("Expected to find auth_democratic")
	}
	if !data.IsValidOrigin("origin_default") {
		t.Error("Expected to find origin_default")
	}
}
