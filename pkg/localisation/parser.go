package localisation

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var lineRegex = regexp.MustCompile(`^\s*([\w\.\-]+):(\d*)\s+"(.*)"\s*$`)

type LocString struct {
	Key      string
	BaseKey  string
	Variant  string
	Suffix   string
	Version  string
	Category string
	Text     string
	Filename string
	Modified bool
}

func ParseFile(path string) ([]LocString, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	
	// Check for UTF-8 BOM
	bom := make([]byte, 3)
	n, err := io.ReadFull(reader, bom)
	if err == nil && n == 3 && bom[0] == 0xef && bom[1] == 0xbb && bom[2] == 0xbf {
		// BOM skipped
	} else if n > 0 {
		// No BOM, put bytes back
		file.Seek(0, 0)
		reader = bufio.NewReader(file)
	}

	var results []LocString
	filename := filepath.Base(path)

	for {
		line, err := reader.ReadString('\n')
		if err != nil && err != io.EOF {
			return nil, err
		}

		line = strings.TrimRight(line, "\r\n")

		// Skip comments and language headers
		if strings.HasPrefix(strings.TrimSpace(line), "#") || strings.HasPrefix(strings.TrimSpace(line), "l_") {
			if err == io.EOF {
				break
			}
			continue
		}

		matches := lineRegex.FindStringSubmatch(line)
		if len(matches) == 4 {
			key := matches[1]
			version := matches[2]
			text := matches[3]
			
			category := getCategory(key)
			baseKey, variant, suffix := decomposeKey(key)

			results = append(results, LocString{
				Key:      key,
				BaseKey:  baseKey,
				Variant:  variant,
				Suffix:   suffix,
				Version:  version,
				Category: category,
				Text:     text,
				Filename: filename,
			})
		}

		if err == io.EOF {
			break
		}
	}

	return results, nil
}

func getCategory(key string) string {
	if strings.Contains(key, ".") {
		return "Event"
	}

	prefixes := []struct {
		prefix string
		cat    string
	}{
		{"leader_trait_", "Leader Trait"},
		{"trait_", "Species Trait"},
		{"civic_", "Civic"},
		{"origin_", "Origin"},
		{"tech_", "Technology"},
		{"building_", "Building"},
		{"district_", "District"},
		{"edict_", "Edict"},
		{"policy_", "Policy"},
		{"tradition_", "Tradition"},
		{"ascension_perk_", "Ascension Perk"},
		{"ap_", "Ascension Perk"},
		{"job_", "Job"},
		{"decision_", "Planetary Decision"},
		{"planet_modifier_", "Planet Modifier"},
		{"pm_", "Planet Modifier"},
		{"relic_", "Relic"},
		{"resolution_", "Resolution"},
		{"megastructure_", "Megastructure"},
		{"casus_belli_", "Casus Belli"},
		{"cb_", "Casus Belli"},
		{"wargoal_", "Wargoal"},
		{"wg_", "Wargoal"},
		{"councilor_", "Councilor"},
		{"federation_", "Federation"},
		{"fed_", "Federation"},
		{"starbase_", "Starbase"},
		{"army_", "Army"},
		{"component_", "Ship Component"},
		{"ship_size_", "Ship Size"},
		{"anomaly_", "Anomaly"},
		{"situation_", "Situation"},
		{"estate_", "Estate"},
		{"message_", "Message"},
		{"diplo_", "Diplomacy"},
	}

	for _, p := range prefixes {
		if strings.HasPrefix(key, p.prefix) {
			return p.cat
		}
	}

	return "General"
}

func decomposeKey(key string) (baseKey string, variant string, suffix string) {
	if strings.Contains(key, ".") {
		parts := strings.Split(key, ".")
		if len(parts) > 1 {
			baseKey = strings.Join(parts[:len(parts)-1], ".")
			suffix = parts[len(parts)-1]
			variant = "default"
			return
		}
	}

	suffix = "name"
	variant = "default"
	baseKey = key

	suffixes := []string{"_desc", "_tooltip_delayed", "_tooltip", "_effects", "_plural"}
	for _, s := range suffixes {
		if strings.HasSuffix(baseKey, s) {
			suffix = strings.TrimPrefix(s, "_")
			baseKey = strings.TrimSuffix(baseKey, s)
			break
		}
	}

	variants := []string{"_machine", "_hive", "_corporate", "_synth", "_robot", "_cyborg", "_psionic", "_lithoid", "_plantoid", "_aquatic", "_toxoid", "_necrophage", "_gestalt"}
	for _, v := range variants {
		if strings.HasSuffix(baseKey, v) {
			// Failsafe: if removing variant makes it exactly the category prefix, it's not a variant (e.g. "trait_machine")
			testBase := strings.TrimSuffix(baseKey, v)
			catBefore := getCategory(baseKey)
			
			// Simple heuristic: if stripping the variant leaves us with just the prefix, we revert.
			// Actually, just checking if testBase == "trait" or similar is enough.
			if strings.HasSuffix(testBase, "_") {
				break
			}
			// If catBefore is not general, and testBase doesn't have the prefix anymore, don't strip
			if catBefore != "General" && getCategory(testBase) == "General" {
				break
			}

			variant = strings.TrimPrefix(v, "_")
			baseKey = testBase
			break
		}
	}

	return
}

func ParseDirectory(dir string) ([]LocString, error) {
	var allStrings []LocString
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".yml") {
			strs, parseErr := ParseFile(path)
			if parseErr != nil {
				fmt.Printf("Warning: Failed to parse %s: %v\n", path, parseErr)
			} else {
				allStrings = append(allStrings, strs...)
			}
		}
		return nil
	})
	return allStrings, err
}

func WriteModFile(dir string, filename string, strings []LocString) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	
	path := filepath.Join(dir, filename)
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err := file.Write([]byte{0xef, 0xbb, 0xbf}); err != nil {
		return err
	}

	writer := bufio.NewWriter(file)
	writer.WriteString("l_english:\n")
	for _, s := range strings {
		ver := s.Version
		if ver == "" {
			ver = "0"
		}
		writer.WriteString(fmt.Sprintf(" %s:%s \"%s\"\n", s.Key, ver, s.Text))
	}
	
	return writer.Flush()
}
