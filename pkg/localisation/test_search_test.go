package localisation

import (
	"fmt"
	"testing"
)

func TestSearch(t *testing.T) {
	store, err := NewStore()
	if err != nil {
		t.Fatal(err)
	}
	
	strings := []LocString{
		{Key: "leader_trait_politician", BaseKey: "leader_trait_politician", Variant: "default", Suffix: "name", Category: "Leader Trait", Text: "Politician", Filename: "test.yml"},
		{Key: "leader_trait_politician_desc", BaseKey: "leader_trait_politician", Variant: "default", Suffix: "desc", Category: "Leader Trait", Text: "A politician.", Filename: "test.yml"},
		{Key: "concept_hab_capital_desc", BaseKey: "concept_hab_capital", Variant: "default", Suffix: "desc", Category: "General", Text: "Has job_politician in it.", Filename: "test.yml"},
	}
	
	if err := store.InsertStrings(strings); err != nil {
		t.Fatal("Insert err:", err)
	}
	
	rows, err := store.db.Query(`
		SELECT base_key
		FROM strings_fts
		WHERE strings_fts MATCH ?
		ORDER BY bm25(strings_fts, 10.0, 10.0, 1.0)
	`, "politician")
	if err != nil {
		t.Fatal("Query err:", err)
	}
	defer rows.Close()
	
	fmt.Println("CTE results:")
	for rows.Next() {
		var bk string
		rows.Scan(&bk)
		fmt.Printf("bk:%s\n", bk)
	}

}
