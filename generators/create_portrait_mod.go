package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bobby/stellaris-mods/pkg/log"
	"github.com/urfave/cli/v2"

	"github.com/bobby/stellaris-mods/generators/pkg/empire"
	"github.com/bobby/stellaris-mods/generators/pkg/portrait"
	"github.com/bobby/stellaris-mods/pkg/version"
)

func errorExit(format string, a ...interface{}) error {
	return cli.Exit(log.Errorf(format, a...), 1)
}

var (
	stellarisPath = filepath.Join(os.Getenv("HOME"), ".local", "share", "Steam", "steamapps", "common", "Stellaris")
	modBasePath   = "mod" // relative to stellaris directory
)

func main() {
	app := &cli.App{
		Name:      "create_portrait_mod",
		Usage:     "Generates a custom Stellaris portrait mod using the portrait sets method from a YAML empire config.",
		ArgsUsage: "[config.yaml]",
		Action: func(c *cli.Context) error {
			if c.NArg() < 1 {
				return cli.ShowAppHelp(c)
			}
			yamlFile := c.Args().First()
			return generateMod(yamlFile)
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal("%v", err)
	}
}

func generateMod(yamlFile string) error {
	log.Info("Reading YAML config \033[36m%s\033[0m...", yamlFile)
	cfg, err := empire.ParseConfig(yamlFile)
	if err != nil {
		return errorExit("Failed to parse YAML: %v", err)
	}

	inputPNG := cfg.Species.PortraitImage
	if inputPNG == "" {
		return errorExit("No species.portrait_image specified in YAML")
	}

	if !filepath.IsAbs(inputPNG) {
		inputPNG = filepath.Join(filepath.Dir(yamlFile), inputPNG)
	}

	speciesName := cfg.Species.Name
	speciesClass := cfg.Species.Class

	if speciesName == "" {
		return errorExit("No species.name specified in YAML")
	}
	if speciesClass == "" {
		return errorExit("No species.class specified in YAML")
	}

	// Map simple archetype class names to what generateMod expects (e.g. humanoid -> HUM)
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

	targetClass := archetypeMap[speciesClass]
	if targetClass == "" {
		targetClass = speciesClass // Assume they know what they are doing if they provided a raw target
	}

	customName := strings.ToLower(strings.ReplaceAll(speciesName, " ", "_"))
	modName := customName + "_portrait"
	modDir := filepath.Join(modBasePath, modName)

	installedVersion, err := version.GetModCompatibilityVersion(stellarisPath)
	supportedVersion := "v3.*"
	if err != nil {
		log.Info("Warning: Could not detect Stellaris version (%v). Defaulting to %s", err, supportedVersion)
	} else {
		supportedVersion = "v" + installedVersion + ".*"
	}

	// Create .mod file
	if err := os.MkdirAll(modBasePath, 0755); err != nil {
		return errorExit("Failed to create mod base dir: %v", err)
	}
	modFileContent := fmt.Sprintf("name=\"%s\"\npath=\"mod/%s\"\nsupported_version=\"%s\"\n", modName, modName, supportedVersion)
	if err := os.WriteFile(filepath.Join(modBasePath, modName+".mod"), []byte(modFileContent), 0644); err != nil {
		return errorExit("Failed to write .mod file: %v", err)
	}

	if err := portrait.GenerateMod(modDir, inputPNG, speciesName, targetClass, stellarisPath); err != nil {
		return errorExit("Failed to generate portrait mod: %v", err)
	}

	log.Info("Run \033[36m./install_mods\033[0m in the stellaris directory to install it.")
	return nil
}

