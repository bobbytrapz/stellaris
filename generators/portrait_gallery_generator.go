package main

import (
	"html/template"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/bobby/stellaris-mods/pkg/log"
)

var (
	stellarisPath = filepath.Join(os.Getenv("HOME"), ".local", "share", "Steam", "steamapps", "common", "Stellaris")
	portraitsPath = filepath.Join(stellarisPath, "gfx", "portraits", "portraits")
	galleryDir    = "portrait_gallery" // relative to stellaris directory
	imagesDir     = filepath.Join(galleryDir, "images")
)

type Portrait struct {
	Name    string
	Image   string // path relative to portrait_gallery/ e.g. images/name.png
	DDSPath string // absolute path to source dds
}

func main() {
	log.Info("Starting Gallery Generator...")
	
	// Create output directories
	if err := os.MkdirAll(imagesDir, 0755); err != nil {
		log.Fatal("Failed to create gallery dir: %v", err)
	}

	portraits, err := parsePortraits()
	if err != nil {
		log.Fatal("Failed to parse portraits: %v", err)
	}

	log.Info("Found %d portraits. Converting to PNG...", len(portraits))

	var validPortraits []Portrait
	for i, p := range portraits {
		pngPath := filepath.Join(imagesDir, p.Name+".png")
		p.Image = "images/" + p.Name + ".png"
		
		if _, err := os.Stat(pngPath); os.IsNotExist(err) {
			if convertDDStoPNG(p.DDSPath, pngPath) {
				validPortraits = append(validPortraits, p)
			}
		} else {
			// Already converted
			validPortraits = append(validPortraits, p)
		}
		if (i+1)%50 == 0 {
			log.Info("Processed %d/%d portraits...", i+1, len(portraits))
		}
	}

	log.Info("Generated %d images. Building HTML...", len(validPortraits))
	if err := buildHTML(validPortraits); err != nil {
		log.Fatal("Failed to build HTML: %v", err)
	}

	log.Success("Gallery generation complete! Open portrait_gallery/index.html in your browser.")
}

func convertDDStoPNG(src, dst string) bool {
	commands := []string{"convert", "magick"}
	for _, cmd := range commands {
		// Use -scale to make them reasonably sized for a gallery if they are huge, 
		// but let's just convert them directly and let HTML resize.
		c := exec.Command(cmd, src, dst)
		err := c.Run()
		if err == nil {
			return true
		}
	}
	log.Warning("Failed to convert %s", src)
	return false
}

func parsePortraits() ([]Portrait, error) {
	var portraits []Portrait
	
	files, err := os.ReadDir(portraitsPath)
	if err != nil {
		return nil, err
	}

	// Regex to find texture paths. We just want the first one we find for a portrait.
	// Since the structure is nested, a simple line-by-line state machine works best.
	
	reBlockStart := regexp.MustCompile(`^\s*([a-zA-Z0-9_-]+)\s*=\s*\{`)
	reTextureFile := regexp.MustCompile(`texturefile\s*=\s*"([^"]+)"`)
	reDDS := regexp.MustCompile(`"([^"]+\.dds)"`)
	
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".txt") {
			continue
		}
		
		content, err := os.ReadFile(filepath.Join(portraitsPath, file.Name()))
		if err != nil {
			continue
		}
		
		lines := strings.Split(string(content), "\n")
		var currentPortrait string
		inPortraits := false
		bracketDepth := 0
		
		for _, line := range lines {
			// Strip comments
			if idx := strings.Index(line, "#"); idx != -1 {
				line = line[:idx]
			}
			
			// Track brackets
			for _, char := range line {
				if char == '{' {
					bracketDepth++
				} else if char == '}' {
					bracketDepth--
				}
			}
			
			if !inPortraits {
				if strings.Contains(line, "portraits = {") {
					inPortraits = true
					// Assuming "portraits = {" is on one line and increases depth
					// Actually the bracketDepth handles the nesting. Let's just track if we are in the main block.
				}
				continue
			}
			
			// We are inside portraits = {
			if bracketDepth == 0 {
				inPortraits = false
				continue
			}
			
			// If we are at depth 2, we are at a portrait definition: e.g. sd_hum_robot = {
			if bracketDepth == 2 || (bracketDepth == 1 && strings.Contains(line, "{")) {
				matches := reBlockStart.FindStringSubmatch(line)
				if len(matches) > 1 {
					currentPortrait = matches[1]
				}
			}
			
			if currentPortrait != "" {
				// Check for textures
				match := ""
				if m := reTextureFile.FindStringSubmatch(line); len(m) > 1 {
					match = m[1]
				} else if m := reDDS.FindStringSubmatch(line); len(m) > 1 {
					match = m[1]
				}
				
				if match != "" {
					ddsPath := filepath.Join(stellarisPath, match)
					if _, err := os.Stat(ddsPath); err == nil {
						// Only add if we haven't added this portrait yet
						alreadyAdded := false
						for _, p := range portraits {
							if p.Name == currentPortrait {
								alreadyAdded = true
								break
							}
						}
						if !alreadyAdded {
							portraits = append(portraits, Portrait{
								Name:    currentPortrait,
								DDSPath: ddsPath,
							})
						}
					}
				}
			}
			
			if bracketDepth <= 1 {
				currentPortrait = ""
			}
		}
	}
	
	return portraits, nil
}

func buildHTML(portraits []Portrait) error {
	tmplHTML := `
<!DOCTYPE html>
<html>
<head>
    <title>Stellaris Portrait Gallery</title>
    <style>
        body { font-family: sans-serif; background: #1e1e1e; color: #fff; }
        .gallery { display: flex; flex-wrap: wrap; gap: 20px; padding: 20px; justify-content: center; }
        .card { background: #2d2d2d; border-radius: 8px; padding: 10px; text-align: center; width: 250px; }
        .card img { max-width: 100%; height: auto; border-radius: 4px; }
        .card h3 { margin: 10px 0 5px 0; font-size: 16px; word-break: break-all; }
        .card input { width: 90%; padding: 5px; background: #444; color: #fff; border: 1px solid #555; text-align: center; }
    </style>
</head>
<body>
    <h1 style="text-align:center;">Stellaris Portrait Gallery</h1>
    <p style="text-align:center;">Use these portrait IDs for the target_vanilla argument.</p>
    <div class="gallery">
        {{range .}}
        <div class="card">
            <img src="{{.Image}}" alt="{{.Name}}" loading="lazy" />
            <h3>{{.Name}}</h3>
            <input type="text" value="{{.Name}}" readonly onclick="this.select();" />
        </div>
        {{end}}
    </div>
</body>
</html>
`
	t, err := template.New("gallery").Parse(tmplHTML)
	if err != nil {
		return err
	}

	f, err := os.Create(filepath.Join(galleryDir, "index.html"))
	if err != nil {
		return err
	}
	defer f.Close()

	return t.Execute(f, portraits)
}
