package empire

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bobby/stellaris-mods/generators/pkg/gamedata"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Name      string `yaml:"name"`
	Adjective string `yaml:"adjective"`
	Author    string `yaml:"author"`

	Species struct {
		Name          string   `yaml:"name"`
		Plural        string   `yaml:"plural"`
		Adjective     string   `yaml:"adjective"`
		Class         string   `yaml:"class"` // Archetype e.g. "humanoid"
		PortraitImage string   `yaml:"portrait_image"`
		Traits        []string `yaml:"traits"`
	} `yaml:"species"`

	Biography string `yaml:"biography"`

	Homeworld struct {
		Name       string `yaml:"name"`
		Class      string `yaml:"class"`
		SystemName string `yaml:"system_name"`
	} `yaml:"homeworld"`

	ShipPrefix string `yaml:"ship_prefix"`
	Shipset    string `yaml:"shipset"`
	Cityset    string `yaml:"cityset"`

	Government struct {
		Authority string   `yaml:"authority"`
		Ethics    []string `yaml:"ethics"`
		Civics    []string `yaml:"civics"`
		Origin    string   `yaml:"origin"`
	} `yaml:"government"`

	Namelist struct {
		TSVPath string `yaml:"tsv_path"`
	} `yaml:"namelist"`

	RoomBackground string `yaml:"room_background"`
	Flag           struct {
		Colors []string `yaml:"colors"`
		Icon   struct {
			Category string `yaml:"category"`
			File     string `yaml:"file"`
		} `yaml:"icon"`
		Pattern string `yaml:"pattern"`
	} `yaml:"flag"`
}

// ParseConfig reads the empire configuration from a YAML file.
func ParseConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// Validate checks the configuration against the provided GameData
func (c *Config) Validate(data *gamedata.GameData) error {
	var errs []string

	for _, trait := range c.Species.Traits {
		if !data.IsValidTrait(trait) {
			errs = append(errs, fmt.Sprintf("invalid trait: %s", trait))
		}
	}

	for _, ethic := range c.Government.Ethics {
		if !data.IsValidEthic(ethic) {
			errs = append(errs, fmt.Sprintf("invalid ethic: %s", ethic))
		}
	}

	for _, civic := range c.Government.Civics {
		if !data.IsValidCivic(civic) {
			errs = append(errs, fmt.Sprintf("invalid civic: %s", civic))
		}
	}

	if c.Government.Origin != "" && !data.IsValidOrigin(c.Government.Origin) {
		errs = append(errs, fmt.Sprintf("invalid origin: %s", c.Government.Origin))
	}

	if c.Government.Authority != "" && !data.IsValidAuthority(c.Government.Authority) {
		errs = append(errs, fmt.Sprintf("invalid authority: %s", c.Government.Authority))
	}

	if len(errs) > 0 {
		return fmt.Errorf("configuration errors:\n%s", strings.Join(errs, "\n"))
	}

	return nil
}

// GenerateEmpireFile writes the common/prescripted_countries/ txt file.
func (c *Config) GenerateEmpireFile(dst string, modID string) error {
	var sb strings.Builder

	// Add an identifier for the empire, usually customName_empire
	sb.WriteString(fmt.Sprintf("%s = {\n", modID))
	sb.WriteString(fmt.Sprintf("\tname = \"%s\"\n", c.Name))
	sb.WriteString(fmt.Sprintf("\tadjective = \"%s\"\n", c.Adjective))
	sb.WriteString("\tspawn_enabled = yes\n")

	if c.ShipPrefix != "" {
		sb.WriteString(fmt.Sprintf("\tship_prefix = \"%s\"\n", c.ShipPrefix))
	}

	// Species
	sb.WriteString("\tspecies = {\n")
	sb.WriteString(fmt.Sprintf("\t\tclass = \"%s\"\n", c.Species.Class))
	sb.WriteString(fmt.Sprintf("\t\tportrait = \"%s\"\n", modID)) // Assumes portrait mod gives it this ID
	sb.WriteString(fmt.Sprintf("\t\tname = \"%s\"\n", c.Species.Name))
	sb.WriteString(fmt.Sprintf("\t\tplural = \"%s\"\n", c.Species.Plural))
	sb.WriteString(fmt.Sprintf("\t\tadjective = \"%s\"\n", c.Species.Adjective))
	sb.WriteString(fmt.Sprintf("\t\tname_list = \"%s\"\n", modID)) // Assumes namelist mod gives it this ID
	
	for _, trait := range c.Species.Traits {
		sb.WriteString(fmt.Sprintf("\t\ttrait = \"%s\"\n", trait))
	}
	sb.WriteString("\t}\n")

	// Government
	sb.WriteString(fmt.Sprintf("\tauthority = \"%s\"\n", c.Government.Authority))
	sb.WriteString(fmt.Sprintf("\torigin = \"%s\"\n", c.Government.Origin))
	
	for _, ethic := range c.Government.Ethics {
		sb.WriteString(fmt.Sprintf("\tethic = \"%s\"\n", ethic))
	}
	for _, civic := range c.Government.Civics {
		sb.WriteString(fmt.Sprintf("\tcivic = \"%s\"\n", civic))
	}

	// Homeworld
	if c.Homeworld.Name != "" {
		sb.WriteString(fmt.Sprintf("\tplanet_name = \"%s\"\n", c.Homeworld.Name))
	}
	if c.Homeworld.Class != "" {
		sb.WriteString(fmt.Sprintf("\tplanet_class = \"%s\"\n", c.Homeworld.Class))
	}
	if c.Homeworld.SystemName != "" {
		sb.WriteString(fmt.Sprintf("\tsystem_name = \"%s\"\n", c.Homeworld.SystemName))
	}
	
	// Sets
	if c.Shipset != "" {
		sb.WriteString(fmt.Sprintf("\tgraphical_culture = \"%s\"\n", c.Shipset))
	}
	if c.Cityset != "" {
		sb.WriteString(fmt.Sprintf("\tcity_graphical_culture = \"%s\"\n", c.Cityset))
	}

	// Room & Flag
	if c.RoomBackground != "" {
		sb.WriteString(fmt.Sprintf("\troom = \"%s\"\n", c.RoomBackground))
	}
	
	sb.WriteString("\tflag = {\n")
	sb.WriteString(fmt.Sprintf("\t\ticon = {\n\t\t\tcategory = \"%s\"\n\t\t\tfile = \"%s\"\n\t\t}\n", c.Flag.Icon.Category, c.Flag.Icon.File))
	sb.WriteString(fmt.Sprintf("\t\tbackground = {\n\t\t\tcategory = \"backgrounds\"\n\t\t\tfile = \"%s\"\n\t\t}\n", c.Flag.Pattern))
	
	sb.WriteString("\t\tcolors = {\n")
	for _, color := range c.Flag.Colors {
		sb.WriteString(fmt.Sprintf("\t\t\t\"%s\"\n", color))
	}
	sb.WriteString("\t\t}\n")
	sb.WriteString("\t}\n")

	sb.WriteString("}\n")

	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}
	return os.WriteFile(dst, []byte(sb.String()), 0644)
}

// GenerateLocalisationFile writes the localisation for the empire name.
func (c *Config) GenerateLocalisationFile(dst string, modID string) error {
	var sb strings.Builder

	// Write UTF-8 BOM required by Stellaris
	sb.WriteString("\xef\xbb\xbf")
	sb.WriteString("l_english:\n")
	sb.WriteString(fmt.Sprintf("  %s: \"%s\"\n", modID, c.Name))
	sb.WriteString(fmt.Sprintf("  %s_ADJ: \"%s\"\n", modID, c.Adjective))
	
	if c.Biography != "" {
		// Clausewitz localization expects literal \n sequences to be preserved as actual newlines,
		// or as \n text inside the quotes. Since yaml might preserve actual newlines, we should replace them with \n
		bio := strings.ReplaceAll(c.Biography, "\n", "\\n")
		sb.WriteString(fmt.Sprintf("  %s_desc: \"%s\"\n", modID, bio))
	}

	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}
	return os.WriteFile(dst, []byte(sb.String()), 0644)
}
