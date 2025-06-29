package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

func main() {
	root, err := getModDirectoryRoot()
	if err != nil {
		log.Fatalf("get_mod_root: %s", err)
	}
	log.Println("Create mod directory:", root)

	mods, err := listMods()
	if err != nil {
		log.Fatalf("list_mods: %s", err)
	}
	log.Println("Need to copy:", mods)

	for _, mod := range mods {
		log.Println("Copy:", mod.Name)

		existingModDirectory := filepath.Join(root, mod.Name)
		log.Println("Removing existing directory: ", existingModDirectory)
		if err := os.RemoveAll(existingModDirectory); err != nil {
			log.Printf("Warning: failed to remove existing mod directory %q: %s", existingModDirectory, err)
		}

		for _, file := range mod.Files {
			from := filepath.Join("mod", file)
			to := filepath.Join(root, file)

			log.Printf("Copy: %q -> %q", from, to)

			if err := copyFile(from, to); err != nil {
				log.Fatalf("copy_file: %s", err)
			}
		}
	}
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

func removeDirectory(path string) error {
	return os.RemoveAll(path)
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
