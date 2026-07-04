package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/bobby/stellaris-mods/generators/pkg/empire"
	"github.com/bobby/stellaris-mods/generators/pkg/gamedata"
	"github.com/bobby/stellaris-mods/generators/pkg/namelist"
	"github.com/bobby/stellaris-mods/generators/pkg/portrait"
	"github.com/bobby/stellaris-mods/pkg/log"
	"github.com/bobby/stellaris-mods/pkg/version"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"
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
		Name:  "create_empire_mod",
		Usage: "Generates a complete custom Stellaris empire mod from a YAML config.",
		Commands: []*cli.Command{
			{
				Name:      "generate",
				Usage:     "Generate an empire mod from a YAML config",
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
				Usage:     "Start web UI to edit an empire YAML config",
				ArgsUsage: "[config.yaml]",
				Action: func(c *cli.Context) error {
					yamlFile := "empire.yaml"
					if c.NArg() > 0 {
						yamlFile = c.Args().First()
					}

					// Setup HTTP endpoints for Web UI
					http.HandleFunc("/api/data", func(w http.ResponseWriter, r *http.Request) {
						if r.Method == http.MethodGet {
							cfg, err := empire.ParseConfig(yamlFile)
							if err != nil {
								if os.IsNotExist(err) {
									cfg = &empire.Config{} // Default empty config
								} else {
									http.Error(w, err.Error(), 500)
									return
								}
							}
							w.Header().Set("Content-Type", "application/json")
							json.NewEncoder(w).Encode(cfg)
						} else if r.Method == http.MethodPost {
							var cfg empire.Config
							if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
								http.Error(w, err.Error(), 400)
								return
							}

							data, err := yaml.Marshal(&cfg)
							if err != nil {
								http.Error(w, err.Error(), 500)
								return
							}

							if err := os.WriteFile(yamlFile, data, 0644); err != nil {
								http.Error(w, err.Error(), 500)
								return
							}
							w.WriteHeader(http.StatusOK)
						} else {
							http.Error(w, "Method not allowed", 405)
						}
					})

					http.HandleFunc("/api/gamedata", func(w http.ResponseWriter, r *http.Request) {
						if r.Method == http.MethodGet {
							gd, err := gamedata.Load(stellarisPath)
							if err != nil {
								http.Error(w, err.Error(), 500)
								return
							}
							w.Header().Set("Content-Type", "application/json")
							json.NewEncoder(w).Encode(gd)
						} else {
							http.Error(w, "Method not allowed", 405)
						}
					})

					http.HandleFunc("/api/image", func(w http.ResponseWriter, r *http.Request) {
						imagePath := r.URL.Query().Get("path")
						if imagePath == "" || strings.Contains(imagePath, "..") {
							http.Error(w, "Invalid path", 400)
							return
						}

						fullPath := filepath.Join(stellarisPath, filepath.Clean(imagePath))
						if !strings.HasSuffix(fullPath, ".dds") {
							http.Error(w, "Only .dds images are supported", 400)
							return
						}

						if _, err := os.Stat(fullPath); os.IsNotExist(err) {
							http.Error(w, "Image not found", 404)
							return
						}

						w.Header().Set("Content-Type", "image/png")
						
						// Run ImageMagick 'convert' to stream DDS to PNG
						cmd := exec.Command("convert", fullPath, "png:-")
						cmd.Stdout = w
						
						if err := cmd.Run(); err != nil {
							// Try 'magick' instead if 'convert' fails or isn't installed
							cmd2 := exec.Command("magick", fullPath, "png:-")
							cmd2.Stdout = w
							if err2 := cmd2.Run(); err2 != nil {
								log.Info("Failed to convert image: %v", err2)
							}
						}
					})

					http.HandleFunc("/api/upload", func(w http.ResponseWriter, r *http.Request) {
						if r.Method != http.MethodPost {
							http.Error(w, "Method not allowed", 405)
							return
						}
						
						file, header, err := r.FormFile("file")
						if err != nil {
							http.Error(w, "Failed to read file: "+err.Error(), 400)
							return
						}
						defer file.Close()

						// Secure the filename and save in the same directory as the yaml
						filename := filepath.Base(header.Filename)
						destPath := filepath.Join(filepath.Dir(yamlFile), filename)
						
						out, err := os.Create(destPath)
						if err != nil {
							http.Error(w, "Failed to save file: "+err.Error(), 500)
							return
						}
						defer out.Close()

						// We can't use io.Copy easily without importing "io", so we'll read it directly
						buffer := make([]byte, 1024)
						for {
							n, err := file.Read(buffer)
							if n > 0 {
								out.Write(buffer[:n])
							}
							if err != nil {
								break
							}
						}

						w.Header().Set("Content-Type", "application/json")
						json.NewEncoder(w).Encode(map[string]string{"filename": filename})
					})

					http.HandleFunc("/api/workspace_file", func(w http.ResponseWriter, r *http.Request) {
						filename := r.URL.Query().Get("file")
						if filename == "" || strings.Contains(filename, "..") || strings.Contains(filename, "/") || strings.Contains(filename, "\\") {
							http.Error(w, "Invalid filename", 400)
							return
						}
						
						fullPath := filepath.Join(filepath.Dir(yamlFile), filename)
						if _, err := os.Stat(fullPath); os.IsNotExist(err) {
							http.Error(w, "File not found", 404)
							return
						}
						http.ServeFile(w, r, fullPath)
					})

					http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
						if r.URL.Path == "/" {
							http.ServeFile(w, r, "generators/web/empire_generator/index.html")
						} else {
							http.ServeFile(w, r, filepath.Join("generators/web/empire_generator", filepath.Clean(r.URL.Path)))
						}
					})

					listener, err := net.Listen("tcp", "127.0.0.1:0")
					if err != nil {
						return errorExit("Failed to bind port: %v", err)
					}
					port := listener.Addr().(*net.TCPAddr).Port
					log.Info("Starting Web UI on http://127.0.0.1:%d", port)
					return http.Serve(listener, nil)
				},
			},
		},
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

	gd, err := gamedata.Load(stellarisPath)
	if err != nil {
		return errorExit("Failed to load game data: %v", err)
	}

	if err := cfg.Validate(gd); err != nil {
		return errorExit("Validation failed:\n%v", err)
	}

	modId := strings.ToLower(cfg.Name)
	reg, _ := regexp.Compile("[^a-z0-9]+")
	modId = reg.ReplaceAllString(modId, "_")
	modId = strings.Trim(modId, "_")
	if modId == "" {
		modId = "my_custom_empire"
	}
	
	modName := modId + "_mod"
	modDir := filepath.Join(modBasePath, modName)

	installedVersion, err := version.GetModCompatibilityVersion(stellarisPath)
	supportedVersion := "v3.*"
	if err != nil {
		log.Info("Warning: Could not detect Stellaris version (%v). Defaulting to %s", err, supportedVersion)
	} else {
		supportedVersion = "v" + installedVersion + ".*"
	}

	if err := os.MkdirAll(modDir, 0755); err != nil {
		return errorExit("Failed to create mod base dir: %v", err)
	}

	// Create .mod file
	modFileContent := fmt.Sprintf("name=\"%s\"\npath=\"mod/%s\"\nsupported_version=\"%s\"\n", cfg.Name+" Empire Mod", modName, supportedVersion)
	if err := os.WriteFile(filepath.Join(modBasePath, modName+".mod"), []byte(modFileContent), 0644); err != nil {
		return errorExit("Failed to write .mod file: %v", err)
	}

	// 1. Generate Empire Definition and Localisation
	prescriptedFile := filepath.Join(modDir, "common", "prescripted_countries", modId+".txt")
	if err := cfg.GenerateEmpireFile(prescriptedFile, modId); err != nil {
		return errorExit("Failed to write empire definition file: %v", err)
	}

	empireLocFile := filepath.Join(modDir, "localisation", "english", modId+"_l_english.yml")
	if err := cfg.GenerateLocalisationFile(empireLocFile, modId); err != nil {
		return errorExit("Failed to write empire localisation file: %v", err)
	}

	// 2. Generate Namelist (if specified)
	if cfg.Namelist.TSVPath != "" {
		tsvPath := cfg.Namelist.TSVPath
		if !filepath.IsAbs(tsvPath) {
			tsvPath = filepath.Join(filepath.Dir(yamlFile), tsvPath)
		}

		log.Info("Reading TSV \033[36m%s\033[0m...", tsvPath)
		nlData, err := namelist.ParseTSV(tsvPath)
		if err != nil {
			return errorExit("Failed to parse TSV: %v", err)
		}

		// Sync IDs
		nlData.ID = modId
		nlData.Name = cfg.Name

		modNameListsDir := filepath.Join(modDir, "common", "name_lists")
		modLocDir := filepath.Join(modDir, "localisation", "english")
		os.MkdirAll(modNameListsDir, 0755)
		os.MkdirAll(modLocDir, 0755)

		if err := namelist.WriteNamelistFile(filepath.Join(modNameListsDir, modId+"_names.txt"), nlData); err != nil {
			return errorExit("Failed to write namelist file: %v", err)
		}
		if err := namelist.WriteLocalisationFile(filepath.Join(modLocDir, modId+"_names_l_english.yml"), nlData); err != nil {
			return errorExit("Failed to write namelist localisation file: %v", err)
		}
	}

	// 3. Generate Portrait (if specified)
	if cfg.Species.PortraitImage != "" {
		inputPNG := cfg.Species.PortraitImage
		if !filepath.IsAbs(inputPNG) {
			inputPNG = filepath.Join(filepath.Dir(yamlFile), inputPNG)
		}

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
		targetClass := archetypeMap[cfg.Species.Class]
		if targetClass == "" {
			targetClass = cfg.Species.Class
		}

		// Use modId for portrait speciesName to keep the files named consistently
		if err := portrait.GenerateMod(modDir, inputPNG, modId, targetClass, stellarisPath); err != nil {
			return errorExit("Failed to generate portrait: %v", err)
		}
	}

	log.Success("Unified Empire Mod generated at \033[32m%s\033[0m", modDir)
	return nil
}
