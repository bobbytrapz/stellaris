package main

import (
	"log"
	"net"
	"net/http"
	"os/user"
	"path/filepath"

	"github.com/bobby/stellaris-mods/pkg/localisation"
)

func main() {
	usr, err := user.Current()
	if err != nil {
		log.Fatalf("Failed to get current user: %v", err)
	}

	// Try common Stellaris paths
	locDir := filepath.Join(usr.HomeDir, ".local/share/Steam/steamapps/common/Stellaris/localisation/english")
	
	log.Printf("Parsing localisation files in %s...", locDir)
	strs, err := localisation.ParseDirectory(locDir)
	if err != nil {
		log.Fatalf("Failed to parse directory: %v", err)
	}
	log.Printf("Parsed %d strings.", len(strs))

	store, err := localisation.NewStore()
	if err != nil {
		log.Fatalf("Failed to init store: %v", err)
	}
	
	if err := store.InsertStrings(strs); err != nil {
		log.Fatalf("Failed to insert strings into DB: %v", err)
	}
	log.Println("Strings loaded into database.")

	api := &localisation.API{Store: store}

	// Make sure this points to the right path when run from project root
	fs := http.FileServer(http.Dir("./generators/web/loc_editor"))
	http.Handle("/", fs)

	http.HandleFunc("/api/search", api.SearchHandler)
	http.HandleFunc("/api/categories", api.CategoriesHandler)
	http.HandleFunc("/api/update", api.UpdateHandler)
	http.HandleFunc("/api/generate", api.GenerateHandler)
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Fatalf("Failed to listen on a port: %v", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	log.Printf("Server listening on http://127.0.0.1:%d/", port)
	
	if err := http.Serve(listener, nil); err != nil {
		log.Fatal(err)
	}
}
