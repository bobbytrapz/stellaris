package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"github.com/bobby/stellaris-mods/pkg/log"
)

var verbose bool

func main() {
	flag.BoolVar(&verbose, "verbose", false, "Enable verbose output")
	flag.BoolVar(&verbose, "v", false, "Enable verbose output (shorthand)")
	flag.Parse()

	root, err := getModDirectoryRoot()
	if err != nil {
		log.Fatal("get_mod_root: %s", err)
	}
	
	if verbose {
		log.Info("Create mod directory: %s", root)
	}

	mods, err := listMods()
	if err != nil {
		log.Fatal("list_mods: %s", err)
	}

	for _, mod := range mods {
		existingModDirectory := filepath.Join(root, mod.Name)
		if verbose {
			log.Info("Removing existing directory: %s", existingModDirectory)
		}
		
		if err := os.RemoveAll(existingModDirectory); err != nil {
			log.Warning("failed to remove existing mod directory %q: %s", existingModDirectory, err)
		}

		for _, file := range mod.Files {
			from := filepath.Join("mod", file)
			to := filepath.Join(root, file)

			if verbose {
				log.Info("Copy: %q -> %q", from, to)
			}

			if err := copyFile(from, to); err != nil {
				log.Fatal("copy_file: %s", err)
			}
		}
		
		log.Success("Installed %s", mod.Name)
	}
	
	log.Success("All mods installed successfully!")
}

func getModDirectoryRoot() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", nil
	}
	// $HOME/.local/share/Paradox Interactive/Stellaris/mod
	return filepath.Join(home, ".local", "share", "Paradox Interactive", "Stellaris", "mod"), nil
}

func listMods() ([]Mod, error) {
	var mods []Mod
	userCreatedModDirectory := "mod"

	entries, err := os.ReadDir(userCreatedModDirectory)
	if err != nil {
		return mods, nil
	}

	for _, entry := range entries {
		if entry.IsDir() {
			root := filepath.Join(userCreatedModDirectory, entry.Name())
			rootMetadataFile := fmt.Sprintf("%s.mod", entry.Name())

			toAdd := Mod{
				Name:  entry.Name(),
				Path:  root,
				Files: []string{rootMetadataFile},
			}

			err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if !info.IsDir() {
					relPath, err := filepath.Rel(root, path)
					if err != nil {
						return err
					}
					toAdd.Files = append(toAdd.Files, filepath.Join(entry.Name(), relPath))
				}
				return nil
			})

			if err != nil {
				continue
			}

			mods = append(mods, toAdd)
		}
	}

	return mods, nil
}

type Mod struct {
	Name  string
	Path  string
	Files []string
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("error opening source file: %w", err)
	}
	defer sourceFile.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return fmt.Errorf("error creating directories: %w", err)
	}

	destFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("error creating destination file: %w", err)
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, sourceFile); err != nil {
		return fmt.Errorf("error copying file: %w", err)
	}

	return nil
}
