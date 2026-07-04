package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/bobby/stellaris-mods/pkg/log"
	"github.com/urfave/cli/v2"

	"github.com/bobby/stellaris-mods/generators/pkg/empire"
	"github.com/bobby/stellaris-mods/generators/pkg/namelist"
)

func errorExit(format string, a ...interface{}) error {
	return cli.Exit(log.Errorf(format, a...), 1)
}

var (
	modBasePath = "mod" // relative to stellaris directory
)

func main() {
	app := &cli.App{
		Name:  "create_namelist_mod",
		Usage: "Generates a custom Stellaris namelist mod from a TSV file.",
		Commands: []*cli.Command{
			{
				Name:      "generate",
				Usage:     "Generate a namelist mod from a YAML empire config",
				ArgsUsage: "[config.yaml]",
				Action: func(c *cli.Context) error {
					if c.NArg() < 1 {
						return errorExit("Please provide an input YAML config file.")
					}
					yamlFile := c.Args().First()
					return generateMod(yamlFile)
				},
			},
			{
				Name:      "edit",
				Usage:     "Start web UI to edit a namelist TSV (creates template if missing)",
				ArgsUsage: "[input.tsv]",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "name",
						Aliases: []string{"n"},
						Usage:   "Display name for the mod (used when creating a new template)",
						Value:   "My Custom Namelist",
					},
				},
				Action: func(c *cli.Context) error {
					modName := c.String("name")
					
					modId := strings.ToLower(modName)
					reg, _ := regexp.Compile("[^a-z0-9]+")
					modId = reg.ReplaceAllString(modId, "_")
					modId = strings.Trim(modId, "_")
					if modId == "" {
						modId = "my_custom_namelist"
					}

					inputFile := filepath.Join("namelists", modId+".tsv")
					if c.NArg() > 0 {
						inputFile = c.Args().First()
					}
					
					// If the file doesn't exist, create it as a template
					if _, err := os.Stat(inputFile); os.IsNotExist(err) {
						if err := namelist.GenerateTemplate(inputFile, modName); err != nil {
							return err
						}
					}
					
					http.HandleFunc("/api/data", func(w http.ResponseWriter, r *http.Request) {
						if r.Method == http.MethodGet {
							data, err := namelist.ParseTSV(inputFile)
							if err != nil {
								if os.IsNotExist(err) {
									data = namelist.NewData()
								} else {
									http.Error(w, err.Error(), 500)
									return
								}
							}
							w.Header().Set("Content-Type", "application/json")
							json.NewEncoder(w).Encode(data)
						} else if r.Method == http.MethodPost {
							var data namelist.Data
							if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
								http.Error(w, err.Error(), 400)
								return
							}
							// Clean map structures if they are nil
							if data.ShipNames == nil { data.ShipNames = make(map[string][]string) }
							if data.FleetSeq == nil { data.FleetSeq = make(map[string]string) }
							if data.FleetRandom == nil { data.FleetRandom = make(map[string][]string) }
							if data.ArmySeq == nil { data.ArmySeq = make(map[string]string) }
							if data.PlanetNames == nil { data.PlanetNames = make(map[string][]string) }
							if data.CharFull == nil { data.CharFull = make(map[string][]string) }
							if data.CharRegnal == nil { data.CharRegnal = make(map[string][]string) }
							if data.Localisation == nil { data.Localisation = make(map[string]string) }

							if err := namelist.SaveTSV(inputFile, &data); err != nil {
								http.Error(w, err.Error(), 500)
								return
							}
							w.WriteHeader(http.StatusOK)
						} else {
							http.Error(w, "Method not allowed", 405)
						}
					})

					fs := http.FileServer(http.Dir("generators/web/namelist_editor"))
					http.Handle("/", fs)

					listener, err := net.Listen("tcp", "127.0.0.1:0")
					if err != nil {
						return errorExit("Failed to bind port: %v", err)
					}
					port := listener.Addr().(*net.TCPAddr).Port
					log.Info("Starting server on http://127.0.0.1:%d", port)
					return http.Serve(listener, nil)
				},
			},
		},
		Action: func(c *cli.Context) error {
			if c.NArg() < 1 {
				return cli.ShowAppHelp(c)
			}
			inputFile := c.Args().First()
			return generateMod(inputFile)
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

	tsvPath := cfg.Namelist.TSVPath
	if tsvPath == "" {
		return errorExit("No namelist.tsv_path specified in YAML")
	}

	if !filepath.IsAbs(tsvPath) {
		tsvPath = filepath.Join(filepath.Dir(yamlFile), tsvPath)
	}

	log.Info("Reading TSV \033[36m%s\033[0m...", tsvPath)
	data, err := namelist.ParseTSV(tsvPath)
	if err != nil {
		return errorExit("Failed to parse TSV: %v", err)
	}

	if data.ID == "" {
		data.ID = strings.ToLower(strings.ReplaceAll(cfg.Name, " ", "_"))
	}
	if data.Name == "" {
		data.Name = cfg.Name
	}

	modName := data.ID + "_namelist"
	modDir := filepath.Join(modBasePath, modName)
	modNameListsDir := filepath.Join(modDir, "common", "name_lists")
	modLocDir := filepath.Join(modDir, "localisation", "english")

	if err := os.MkdirAll(modNameListsDir, 0755); err != nil {
		return errorExit("Failed to create name_lists dir: %v", err)
	}
	if err := os.MkdirAll(modLocDir, 0755); err != nil {
		return errorExit("Failed to create localisation dir: %v", err)
	}

	supportedVersion := "v3.*"
	modFileContent := fmt.Sprintf("name=\"%s\"\npath=\"mod/%s\"\nsupported_version=\"%s\"\n", data.Name+" Namelist", modName, supportedVersion)
	if err := os.WriteFile(filepath.Join(modBasePath, modName+".mod"), []byte(modFileContent), 0644); err != nil {
		return errorExit("Failed to write .mod file: %v", err)
	}

	// Write namelist file
	if err := namelist.WriteNamelistFile(filepath.Join(modNameListsDir, data.ID+"_names.txt"), data); err != nil {
		return errorExit("Failed to write namelist file: %v", err)
	}

	// Write localisation file
	if err := namelist.WriteLocalisationFile(filepath.Join(modLocDir, data.ID+"_l_english.yml"), data); err != nil {
		return errorExit("Failed to write localisation file: %v", err)
	}

	log.Success("Namelist mod generated at \033[32m%s\033[0m", modDir)
	return nil
}
