package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/bobby/stellaris-mods/pkg/log"
	"github.com/bobby/stellaris-mods/pkg/version"
	"github.com/urfave/cli/v2"
)

func errorExit(format string, a ...interface{}) error {
	return cli.Exit(log.Errorf(format, a...), 1)
}

var (
	stellarisPath = filepath.Join(os.Getenv("HOME"), ".local", "share", "Steam", "steamapps", "common", "Stellaris")
	modBasePath   = "mod" // relative to stellaris directory
)

func main() {
	archetypeMap := map[string]string{
		"machine":   "MACHINE",
		"humanoid":  "HUM",
		"mammalian": "MAM",
		"avian":     "AVI",
		"reptilian": "REP",
		"fungoid":   "FUN",
		"plantoid":  "PLANT",
		"lithoid":   "LITHOID",
		"aquatic":   "AQUATIC",
		"toxoid":    "TOX",
	}

	commonFlags := []cli.Flag{
		&cli.StringFlag{
			Name:     "image",
			Aliases:  []string{"i"},
			Usage:    "Path to the input PNG image",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "name",
			Aliases:  []string{"n"},
			Usage:    "Species name to trigger the replacement (e.g. \"My Custom Robot\")",
			Required: true,
		},
	}

	commands := []*cli.Command{}
	for archetype, classStr := range archetypeMap {
		targetClass := classStr // capture for closure
		commands = append(commands, &cli.Command{
			Name:  archetype,
			Usage: fmt.Sprintf("Target the %s archetype", archetype),
			Flags: commonFlags,
			Action: func(c *cli.Context) error {
				return generateMod(c, targetClass)
			},
		})
	}

	customFlags := append([]cli.Flag{}, commonFlags...)
	customFlags = append(customFlags, &cli.StringFlag{
		Name:     "class",
		Usage:    "Species class to target (e.g. MACHINE)",
		Required: true,
	})

	commands = append(commands, &cli.Command{
		Name:  "custom",
		Usage: "Target a specific vanilla species class directly",
		Flags: customFlags,
		Action: func(c *cli.Context) error {
			return generateMod(c, c.String("class"))
		},
	})

	app := &cli.App{
		Name:     "create_portrait_mod",
		Usage:    "Generates a custom Stellaris portrait mod using the portrait sets method.",
		Commands: commands,
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal("%v", err)
	}
}

func generateMod(c *cli.Context, speciesClass string) error {
	inputPNG := c.String("image")
	speciesName := c.String("name")

	// Generate a safe custom name from the species name (e.g. "My Custom Robot" -> "my_custom_robot")
	customName := strings.ToLower(strings.ReplaceAll(speciesName, " ", "_"))
	modName := customName + "_portrait"

	log.Info("Starting portrait generation for \033[36m%s\033[0m (Archetype \033[33m%s\033[0m)", customName, speciesClass)

	// Create mod directory structure
	modDir := filepath.Join(modBasePath, modName)
	modGfxDir := filepath.Join(modDir, "gfx", "portraits", "portraits")
	modPortraitSetsDir := filepath.Join(modDir, "common", "portrait_sets")
	modPortraitCategoriesDir := filepath.Join(modDir, "common", "portrait_categories")
	modModelsDir := filepath.Join(modDir, "gfx", "models", "portraits", customName)

	if err := os.MkdirAll(modGfxDir, 0755); err != nil {
		return errorExit("Failed to create mod gfx dir: %v", err)
	}
	if err := os.MkdirAll(modPortraitSetsDir, 0755); err != nil {
		return errorExit("Failed to create mod portrait_sets dir: %v", err)
	}
	if err := os.MkdirAll(modPortraitCategoriesDir, 0755); err != nil {
		return errorExit("Failed to create mod portrait_categories dir: %v", err)
	}
	if err := os.MkdirAll(modModelsDir, 0755); err != nil {
		return errorExit("Failed to create models dir: %v", err)
	}

	installedVersion, err := version.GetModCompatibilityVersion(stellarisPath)
	supportedVersion := "v3.*"
	if err != nil {
		log.Info("Warning: Could not detect Stellaris version (%v). Defaulting to %s", err, supportedVersion)
	} else {
		supportedVersion = "v" + installedVersion + ".*"
	}

	// Create .mod file
	modFileContent := fmt.Sprintf("name=\"%s\"\npath=\"mod/%s\"\nsupported_version=\"%s\"\n", modName, modName, supportedVersion)
	if err := os.WriteFile(filepath.Join(modBasePath, modName+".mod"), []byte(modFileContent), 0644); err != nil {
		return errorExit("Failed to write .mod file: %v", err)
	}

	// Convert PNG to DDS
	outputDDS := filepath.Join(modModelsDir, customName+".dds")
	log.Info("Converting \033[36m%s\033[0m to DDS format...", inputPNG)
	if err := convertPNGtoDDS(inputPNG, outputDDS); err != nil {
		return errorExit("ImageMagick conversion failed: %v\nMake sure ImageMagick is installed (convert/magick command).", err)
	}

	// Create the custom portrait definition
	customDefsFile := filepath.Join(modGfxDir, fmt.Sprintf("99_%s.txt", customName))
	if err := writePortraitDefinition(customDefsFile, customName); err != nil {
		return errorExit("Failed to write custom portrait definitions: %v", err)
	}

	// Create the portrait set definition
	portraitSetFile := filepath.Join(modPortraitSetsDir, fmt.Sprintf("99_%s.txt", customName))
	if err := writePortraitSetDefinition(portraitSetFile, customName, speciesClass); err != nil {
		return errorExit("Failed to write portrait set definitions: %v", err)
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
		return errorExit("Unknown category ID for species class: %s", speciesClass)
	}

	categoryFile := filepath.Join(modPortraitCategoriesDir, fmt.Sprintf("99_%s.txt", customName))
	if err := writePortraitCategoryDefinition(categoryFile, customName, categoryID); err != nil {
		return errorExit("Failed to write portrait category definition: %v", err)
	}

	log.Success("Mod generated at \033[32m%s\033[0m", modDir)
	log.Info("Run \033[36m./install_mods\033[0m in the stellaris directory to install it.")
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

