package namelist

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/bobby/stellaris-mods/pkg/log"
)

type Data struct {
	ID           string              `json:"id"`
	Name         string              `json:"name"`
	Category     string              `json:"category"`
	ShipNames    map[string][]string `json:"shipNames"`
	FleetSeq     map[string]string   `json:"fleetSeq"`
	FleetRandom  map[string][]string `json:"fleetRandom"`
	ArmySeq      map[string]string   `json:"armySeq"`
	PlanetNames  map[string][]string `json:"planetNames"`
	CharFull     map[string][]string `json:"charFull"`
	CharRegnal   map[string][]string `json:"charRegnal"`
	Localisation map[string]string   `json:"localisation"`
}

func NewData() *Data {
	return &Data{
		ShipNames:    make(map[string][]string),
		FleetSeq:     make(map[string]string),
		FleetRandom:  make(map[string][]string),
		ArmySeq:      make(map[string]string),
		PlanetNames:  make(map[string][]string),
		CharFull:     make(map[string][]string),
		CharRegnal:   make(map[string][]string),
		Localisation: make(map[string]string),
	}
}

func GenerateTemplate(outputFile string, modName string) error {
	modId := strings.ToLower(modName)
	reg, _ := regexp.Compile("[^a-z0-9]+")
	modId = reg.ReplaceAllString(modId, "_")
	modId = strings.Trim(modId, "_")
	if modId == "" {
		modId = "my_custom_namelist"
	}

	content := fmt.Sprintf(`# Template for Stellaris Namelist Mod
# Columns: Type | Key | Value | Localisation (Optional)
# 
# Instructions:
# 1. Type 'Meta' defines mod information. 
#    - 'id' (Required) is the unique internal mod ID (no spaces). 
#    - 'name' is the display name. 
#    - 'category' is the UI category (e.g. Machine, Humanoid).
#    - Other optional keys: 'selectable', 'randomized', 'should_name_home_system_planets'.
# 2. Type 'Ship' defines ship names. 
#    - Keys can be: generic, corvette, destroyer, cruiser, battleship, titan, colossus, juggernaut, science, constructor, colonizer, transport, sponsored_colonizer, military_station_small, ion_cannon.
# 3. Type 'ShipClass' defines ship class design names (uses the same keys as Ship).
# 4. Type 'FleetSeq' defines sequential fleet names. Key is usually 'sequential_name'.
# 5. Type 'FleetRandom' defines random fleet names. Key is usually 'random_names'.
# 6. Type 'ArmySeq' defines sequential army names. 
#    - Keys can be: generic, defense_army, assault_army, occupation_army, machine_defense, machine_assault_1, slave_army, clone_army, undead_army, psionic_army, xenomorph_army, gene_warrior_army, etc.
# 7. Type 'ArmyRandom' defines random army names (uses the same keys as ArmySeq).
# 8. Type 'Planet' defines planet names. 
#    - Keys can be: generic, pc_desert, pc_tropical, pc_arid, pc_continental, pc_ocean, pc_tundra, pc_arctic, pc_savannah, pc_alpine.
# 9. Type 'CharFull', 'CharFirst', 'CharSecond', 'CharRegnal' defines character names. 
#    - Key is the culture group (e.g. 'default', 'male', 'female').
# 
# For sequential names, use the Localisation column to specify how it shows up in-game (e.g., "Task:$C$" where $C$ is the number).
Type	Key	Value	Localisation
Meta	id	%s	
Meta	name	%s	
Meta	category	Machine	
Ship	corvette	COR::001	
Ship	corvette	COR::002	
FleetSeq	sequential_name	MY_FLEET_NAMES	Task:$C$
FleetRandom	random_names	Alpha Group	
ArmySeq	machine_assault_1	MY_ARMY_ASSAULT	Surface:Correction $C$
Planet	pc_desert	LAYER:Silicate	
CharFull	default	Daemon	
CharRegnal	default	Integrator	
`, modId, modName)

	if err := os.MkdirAll(filepath.Dir(outputFile), 0755); err != nil {
		return fmt.Errorf("failed to create directory for template: %v", err)
	}
	if err := os.WriteFile(outputFile, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write template file: %v", err)
	}
	log.Success("Template generated at \033[32m%s\033[0m", outputFile)
	return nil
}

func SaveTSV(outputFile string, data *Data) error {
	if err := os.MkdirAll(filepath.Dir(outputFile), 0755); err != nil {
		return err
	}
	file, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	writer.Comma = '\t'

	// Write header
	writer.Write([]string{"Type", "Key", "Value", "Localisation"})

	// Meta
	if data.ID != "" { writer.Write([]string{"Meta", "id", data.ID, ""}) }
	if data.Name != "" { writer.Write([]string{"Meta", "name", data.Name, ""}) }
	if data.Category != "" { writer.Write([]string{"Meta", "category", data.Category, ""}) }

	// ShipNames
	for key, values := range data.ShipNames {
		for _, v := range values {
			writer.Write([]string{"Ship", key, v, ""})
		}
	}
	// FleetSeq
	for key, v := range data.FleetSeq {
		writer.Write([]string{"FleetSeq", key, v, data.Localisation[v]})
	}
	// FleetRandom
	for key, values := range data.FleetRandom {
		for _, v := range values {
			writer.Write([]string{"FleetRandom", key, v, ""})
		}
	}
	// ArmySeq
	for key, v := range data.ArmySeq {
		writer.Write([]string{"ArmySeq", key, v, data.Localisation[v]})
	}
	// PlanetNames
	for key, values := range data.PlanetNames {
		for _, v := range values {
			writer.Write([]string{"Planet", key, v, ""})
		}
	}
	// CharFull
	for key, values := range data.CharFull {
		for _, v := range values {
			writer.Write([]string{"CharFull", key, v, ""})
		}
	}
	// CharRegnal
	for key, values := range data.CharRegnal {
		for _, v := range values {
			writer.Write([]string{"CharRegnal", key, v, ""})
		}
	}

	writer.Flush()
	return writer.Error()
}

func ParseTSV(inputFile string) (*Data, error) {
	file, err := os.Open(inputFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = '\t'
	reader.Comment = '#'
	reader.FieldsPerRecord = -1
	reader.TrimLeadingSpace = true

	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	data := NewData()

	for i, record := range records {
		if len(record) == 0 || (len(record) > 0 && record[0] == "") {
			continue // skip empty lines
		}
		if i == 0 && record[0] == "Type" {
			continue // skip header if present
		}

		if len(record) < 3 {
			log.Warning("Line %d: insufficient columns, skipping.", i+1)
			continue
		}

		rowType := strings.TrimSpace(record[0])
		key := strings.TrimSpace(record[1])
		value := strings.TrimSpace(record[2])
		loc := ""
		if len(record) > 3 {
			loc = strings.TrimSpace(record[3])
		}

		switch rowType {
		case "Meta":
			if key == "id" {
				data.ID = value
			} else if key == "name" {
				data.Name = value
			} else if key == "category" {
				data.Category = value
			}
		case "Ship":
			data.ShipNames[key] = append(data.ShipNames[key], value)
		case "FleetSeq":
			data.FleetSeq[key] = value
			if loc != "" {
				data.Localisation[value] = loc
			}
		case "FleetRandom":
			data.FleetRandom[key] = append(data.FleetRandom[key], value)
		case "ArmySeq":
			data.ArmySeq[key] = value
			if loc != "" {
				data.Localisation[value] = loc
			}
		case "Planet":
			data.PlanetNames[key] = append(data.PlanetNames[key], value)
		case "CharFull":
			data.CharFull[key] = append(data.CharFull[key], value)
		case "CharRegnal":
			data.CharRegnal[key] = append(data.CharRegnal[key], value)
		default:
			log.Warning("Line %d: Unknown Type '%s', skipping.", i+1, rowType)
		}
	}

	return data, nil
}

func WriteNamelistFile(dst string, data *Data) error {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("%s = {\n", data.ID))
	if data.Category != "" {
		sb.WriteString(fmt.Sprintf("\tcategory = \"%s\"\n", data.Category))
	}
	sb.WriteString("\tshould_name_home_system_planets = no\n")
	sb.WriteString("\trandomized = no\n\n")

	// Ship Names
	if len(data.ShipNames) > 0 {
		sb.WriteString("\tship_names = {\n")
		sb.WriteString("\t\tgeneric = { }\n")
		for k, v := range data.ShipNames {
			sb.WriteString(fmt.Sprintf("\t\t%s = { ", k))
			for _, n := range v {
				sb.WriteString(fmt.Sprintf("\"%s\" ", n))
			}
			sb.WriteString("}\n")
		}
		sb.WriteString("\t}\n\n")
	}

	// Fleet Names
	if len(data.FleetSeq) > 0 || len(data.FleetRandom) > 0 {
		sb.WriteString("\tfleet_names = {\n")
		for k, v := range data.FleetRandom {
			sb.WriteString(fmt.Sprintf("\t\t%s = { ", k))
			for _, n := range v {
				sb.WriteString(fmt.Sprintf("\"%s\" ", n))
			}
			sb.WriteString("}\n")
		}
		for k, v := range data.FleetSeq {
			sb.WriteString(fmt.Sprintf("\t\t%s = %s\n", k, v))
		}
		sb.WriteString("\t}\n\n")
	}

	// Army Names
	if len(data.ArmySeq) > 0 {
		sb.WriteString("\tarmy_names = {\n")
		for k, v := range data.ArmySeq {
			sb.WriteString(fmt.Sprintf("\t\t%s = { sequential_name = %s }\n", k, v))
		}
		sb.WriteString("\t}\n\n")
	}

	// Planet Names
	if len(data.PlanetNames) > 0 {
		sb.WriteString("\tplanet_names = {\n")
		for k, v := range data.PlanetNames {
			sb.WriteString(fmt.Sprintf("\t\t%s = { names = { ", k))
			for _, n := range v {
				sb.WriteString(fmt.Sprintf("\"%s\" ", n))
			}
			sb.WriteString("} }\n")
		}
		sb.WriteString("\t}\n\n")
	}

	// Character Names
	if len(data.CharFull) > 0 || len(data.CharRegnal) > 0 {
		sb.WriteString("\tcharacter_names = {\n")
		// Gather all unique keys (usually just 'default')
		keys := make(map[string]bool)
		for k := range data.CharFull { keys[k] = true }
		for k := range data.CharRegnal { keys[k] = true }

		for k := range keys {
			sb.WriteString(fmt.Sprintf("\t\t%s = {\n", k))
			if full, ok := data.CharFull[k]; ok {
				sb.WriteString("\t\t\tfull_names = {\n\t\t\t\t")
				for _, n := range full {
					sb.WriteString(fmt.Sprintf("\"%s\" ", n))
				}
				sb.WriteString("\n\t\t\t}\n")
			}
			if regnal, ok := data.CharRegnal[k]; ok {
				sb.WriteString("\t\t\tregnal_full_names = {\n\t\t\t\t")
				for _, n := range regnal {
					sb.WriteString(fmt.Sprintf("\"%s\" ", n))
				}
				sb.WriteString("\n\t\t\t}\n")
			}
			sb.WriteString("\t\t}\n")
		}
		sb.WriteString("\t}\n")
	}

	sb.WriteString("}\n")

	return os.WriteFile(dst, []byte(sb.String()), 0644)
}

func WriteLocalisationFile(dst string, data *Data) error {
	var sb strings.Builder

	// Write UTF-8 BOM required by Stellaris
	sb.WriteString("\xef\xbb\xbf")
	sb.WriteString("l_english:\n")
	sb.WriteString(fmt.Sprintf("  name_list_%s: \"%s\"\n", data.ID, data.Name))

	for k, v := range data.Localisation {
		sb.WriteString(fmt.Sprintf("  %s: \"%s\"\n", k, v))
	}

	return os.WriteFile(dst, []byte(sb.String()), 0644)
}
