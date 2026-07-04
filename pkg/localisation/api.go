package localisation

import (
	"encoding/json"
	"net/http"
	"path/filepath"
)

type API struct {
	Store *Store
}

type LocItem struct {
	Key      string `json:"key"`
	Text     string `json:"text"`
	Version  string `json:"version"`
	Filename string `json:"filename"`
	Modified bool   `json:"modified"`
}

type LocEntity struct {
	BaseKey  string `json:"baseKey"`
	Category string `json:"category"`
	// Variant -> Suffix -> LocItem
	Variants map[string]map[string]LocItem `json:"variants"`
}

func groupLocStrings(strings []LocString) []LocEntity {
	// Group by BaseKey
	groups := make(map[string]*LocEntity)
	var orderedKeys []string

	for _, s := range strings {
		if _, exists := groups[s.BaseKey]; !exists {
			groups[s.BaseKey] = &LocEntity{
				BaseKey:  s.BaseKey,
				Category: s.Category,
				Variants: make(map[string]map[string]LocItem),
			}
			orderedKeys = append(orderedKeys, s.BaseKey)
		}
		
		entity := groups[s.BaseKey]
		if _, exists := entity.Variants[s.Variant]; !exists {
			entity.Variants[s.Variant] = make(map[string]LocItem)
		}
		
		entity.Variants[s.Variant][s.Suffix] = LocItem{
			Key:      s.Key,
			Text:     s.Text,
			Version:  s.Version,
			Filename: s.Filename,
			Modified: s.Modified,
		}
	}

	var results []LocEntity
	for _, k := range orderedKeys {
		results = append(results, *groups[k])
	}
	
	if results == nil {
		results = []LocEntity{}
	}
	return results
}

func (a *API) SearchHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	category := r.URL.Query().Get("cat")
	
	var strings []LocString
	var err error
	
	if category != "" {
		strings, err = a.Store.GetByCategory(category)
	} else {
		strings, err = a.Store.Search(query)
	}
	
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	entities := groupLocStrings(strings)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entities)
}

func (a *API) CategoriesHandler(w http.ResponseWriter, r *http.Request) {
	cats, err := a.Store.GetCategories()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cats)
}

func (a *API) UpdateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Key  string `json:"key"`
		Text string `json:"text"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := a.Store.Update(req.Key, req.Text); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *API) GenerateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	var req struct {
		Name string `json:"name"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	
	modName := req.Name
	if modName == "" {
		modName = "my_custom_localisation"
	}

	modified, err := a.Store.GetModified()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(modified) == 0 {
		http.Error(w, "No changes to generate", http.StatusBadRequest)
		return
	}

	// Group by filename to write back accurately
	byFile := make(map[string][]LocString)
	for _, m := range modified {
		byFile[m.Filename] = append(byFile[m.Filename], m)
	}

	outDir := filepath.Join(modName, "localisation", "english")
	for filename, strings := range byFile {
		if err := WriteModFile(outDir, filename, strings); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Generated successfully",
		"path": outDir,
	})
}
