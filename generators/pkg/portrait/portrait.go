package portrait

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/bobby/stellaris-mods/pkg/log"
)

// GenerateMod orchestrates the creation of a portrait mod within a given mod directory.
func GenerateMod(modDir, inputPNG, speciesName, speciesClass, stellarisPath string) error {
	// Generate a safe custom name from the species name (e.g. "My Custom Robot" -> "my_custom_robot")
	customName := strings.ToLower(strings.ReplaceAll(speciesName, " ", "_"))

	log.Info("Starting portrait generation for \033[36m%s\033[0m (Archetype \033[33m%s\033[0m)", customName, speciesClass)

	modGfxDir := filepath.Join(modDir, "gfx", "portraits", "portraits")
	modPortraitSetsDir := filepath.Join(modDir, "common", "portrait_sets")
	modPortraitCategoriesDir := filepath.Join(modDir, "common", "portrait_categories")
	modModelsDir := filepath.Join(modDir, "gfx", "models", "portraits", customName)

	if err := os.MkdirAll(modGfxDir, 0755); err != nil {
		return fmt.Errorf("failed to create mod gfx dir: %v", err)
	}
	if err := os.MkdirAll(modPortraitSetsDir, 0755); err != nil {
		return fmt.Errorf("failed to create mod portrait_sets dir: %v", err)
	}
	if err := os.MkdirAll(modPortraitCategoriesDir, 0755); err != nil {
		return fmt.Errorf("failed to create mod portrait_categories dir: %v", err)
	}
	if err := os.MkdirAll(modModelsDir, 0755); err != nil {
		return fmt.Errorf("failed to create models dir: %v", err)
	}

	// Convert PNG to DDS
	outputDDS := filepath.Join(modModelsDir, customName+".dds")
	log.Info("Converting \033[36m%s\033[0m to DDS format...", inputPNG)
	if err := convertPNGtoDDS(inputPNG, outputDDS); err != nil {
		return fmt.Errorf("ImageMagick conversion failed: %v\nMake sure ImageMagick is installed (convert/magick command)", err)
	}

	// Create the custom portrait definition
	customDefsFile := filepath.Join(modGfxDir, fmt.Sprintf("99_%s.txt", customName))
	if err := writePortraitDefinition(customDefsFile, customName); err != nil {
		return fmt.Errorf("failed to write custom portrait definitions: %v", err)
	}

	// Create the portrait set definition
	portraitSetFile := filepath.Join(modPortraitSetsDir, fmt.Sprintf("99_%s.txt", customName))
	if err := writePortraitSetDefinition(portraitSetFile, customName, speciesClass); err != nil {
		return fmt.Errorf("failed to write portrait set definitions: %v", err)
	}

	// Map archetype to category ID
	categoryMap := map[string]string{
		"MACHINE": "machines",
		"HUM":     "humanoids",
		"MAM":     "mammalians",
		"REP":     "reptilians",
		"AVI":     "avians",
		"ART":     "arthropoids",
		"MOL":     "molluscoids",
		"FUN":     "fungoids",
		"PLANT":   "plantoids",
		"LITHOID": "lithoids",
		"AQUATIC": "aquatics",
		"TOX":     "toxoids",
	}
	categoryID := categoryMap[speciesClass]
	if categoryID == "" {
		return fmt.Errorf("unknown category ID for species class: %s", speciesClass)
	}

	categoryFile := filepath.Join(modPortraitCategoriesDir, fmt.Sprintf("99_%s.txt", customName))
	if err := writePortraitCategoryDefinition(categoryFile, customName, categoryID); err != nil {
		return fmt.Errorf("failed to write portrait category definition: %v", err)
	}

	log.Success("Portrait mod generated at \033[32m%s\033[0m", modDir)
	return nil
}

func convertPNGtoDDS(src, dst string) error {
	// standard 512x512 size for stellaris static portraits
	args := []string{src, "-resize", "512x512!", "-define", "dds:compression=dxt5", dst}
	
	commands := []string{"convert", "magick"}
	var lastErr error
	for _, cmd := range commands {
		c := exec.Command(cmd, args...)
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		if err := c.Run(); err == nil {
			return nil
		} else {
			lastErr = err
		}
	}
	return lastErr
}

func writePortraitDefinition(dst, customName string) error {
	content := fmt.Sprintf(`portraits = {
	%s = {
		texturefile = "gfx/models/portraits/%s/%s.dds"
	}
}
`, customName, customName, customName)

	return os.WriteFile(dst, []byte(content), 0644)
}

func writePortraitSetDefinition(dst, customName, speciesClass string) error {
	content := fmt.Sprintf(`%s_set = {
	species_class = %s
	portraits = {
		"%s"
	}
	non_randomized_portraits = {
		"%s"
	}
}
`, customName, speciesClass, customName, customName)

	return os.WriteFile(dst, []byte(content), 0644)
}

func writePortraitCategoryDefinition(dst, customName, categoryID string) error {
	content := fmt.Sprintf(`%s = {
	sets = {
		%s_set
	}
}
`, categoryID, customName)

	return os.WriteFile(dst, []byte(content), 0644)
}
