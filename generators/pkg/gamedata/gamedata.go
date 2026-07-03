package gamedata

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/bobby/stellaris-mods/generators/pkg/clausewitz"
)

type GameData struct {
	Traits            []string            `json:"Traits"`
	Ethics            []string            `json:"Ethics"`
	Civics            []string            `json:"Civics"`
	Authorities       []string            `json:"Authorities"`
	GraphicalCultures []string            `json:"GraphicalCultures"`
	RoomBackgrounds   []string            `json:"RoomBackgrounds"`
	FlagCategories    []string            `json:"FlagCategories"`
	FlagIcons         map[string][]string `json:"FlagIcons"`
	FlagPatterns      []string            `json:"FlagPatterns"`
	FlagColors        []string            `json:"FlagColors"`
}

// DefaultStellarisPath returns the default path to the Stellaris installation on Linux.
func DefaultStellarisPath() string {
	return filepath.Join(os.Getenv("HOME"), ".local", "share", "Steam", "steamapps", "common", "Stellaris")
}

func safeExtract(path string) []string {
	res, err := clausewitz.ExtractTopLevelKeysFromDir(path)
	if err != nil {
		return []string{}
	}
	return res
}

func listDDSFiles(dir string) []string {
	var files []string
	entries, err := os.ReadDir(dir)
	if err != nil {
		return files
	}
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".dds") {
			files = append(files, strings.TrimSuffix(e.Name(), ".dds"))
		}
	}
	return files
}

func listDirectories(dir string) []string {
	var dirs []string
	entries, err := os.ReadDir(dir)
	if err != nil {
		return dirs
	}
	for _, e := range entries {
		if e.IsDir() {
			dirs = append(dirs, e.Name())
		}
	}
	return dirs
}

// Load reads the Stellaris game files and extracts top-level identifiers.
func Load(stellarisPath string) (*GameData, error) {
	data := &GameData{
		FlagIcons: make(map[string][]string),
	}

	if _, err := os.Stat(stellarisPath); os.IsNotExist(err) {
		// Return empty data gracefully if game is not installed
		return data, nil
	}

	commonPath := filepath.Join(stellarisPath, "common")

	data.Traits = safeExtract(filepath.Join(commonPath, "traits"))
	data.Ethics = safeExtract(filepath.Join(commonPath, "ethics"))
	data.Civics = safeExtract(filepath.Join(commonPath, "governments", "civics"))
	data.Authorities = safeExtract(filepath.Join(commonPath, "governments", "authorities"))
	data.GraphicalCultures = safeExtract(filepath.Join(commonPath, "graphical_culture"))

	// Extract Colors from flags/colors.txt
	colorsBytes, err := os.ReadFile(filepath.Join(stellarisPath, "flags", "colors.txt"))
	if err == nil {
		re := regexp.MustCompile(`(?m)^\s*([a-zA-Z0-9_]+)\s*=\s*\{`)
		matches := re.FindAllStringSubmatch(string(colorsBytes), -1)
		for _, m := range matches {
			key := m[1]
			if key != "colors" {
				data.FlagColors = append(data.FlagColors, key)
			}
		}
	}

	// Rooms
	roomsDir := filepath.Join(stellarisPath, "gfx", "portraits", "city_sets")
	allRooms := listDDSFiles(roomsDir)
	for _, r := range allRooms {
		if strings.Contains(r, "room") && !strings.Contains(r, "_city_") {
			data.RoomBackgrounds = append(data.RoomBackgrounds, r)
		}
	}

	// Flags
	flagsDir := filepath.Join(stellarisPath, "flags")
	data.FlagPatterns = listDDSFiles(filepath.Join(flagsDir, "backgrounds"))
	
	cats := listDirectories(flagsDir)
	for _, c := range cats {
		if c != "backgrounds" && c != "colors" {
			data.FlagCategories = append(data.FlagCategories, c)
			data.FlagIcons[c] = listDDSFiles(filepath.Join(flagsDir, c))
		}
	}

	return data, nil
}

// Validation helpers
func (d *GameData) IsValidTrait(trait string) bool {
	return contains(d.Traits, trait)
}

func (d *GameData) IsValidEthic(ethic string) bool {
	return contains(d.Ethics, ethic)
}

func (d *GameData) IsValidCivic(civic string) bool {
	return contains(d.Civics, civic)
}

func (d *GameData) IsValidAuthority(authority string) bool {
	return contains(d.Authorities, authority)
}

func (d *GameData) IsValidOrigin(origin string) bool {
	// Origins are defined in the civics folder
	return contains(d.Civics, origin)
}

func contains(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}
