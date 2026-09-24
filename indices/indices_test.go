package indices

import "testing"

func TestIndexListBroadMarket(t *testing.T) {
	list, err := IndexList("BroadMarketIndices")
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 18 {
		t.Errorf("want 18 broad market indices got %d", len(list))
	}
	found := false
	for _, n := range list {
		if n == "Nifty 50" {
			found = true
		}
	}
	if !found {
		t.Error("Nifty 50 missing")
	}
}

func TestCategoriesCoverage(t *testing.T) {
	if len(Categories) != 4 {
		t.Fatalf("want 4 categories got %d", len(Categories))
	}
	total := 0
	for name, cat := range Categories {
		if len(cat.IndicesList) == 0 {
			t.Errorf("%s empty", name)
		}
		total += len(cat.IndicesList)
		for _, idx := range cat.IndicesList {
			if _, ok := cat.IndexConstituentListURLs[idx]; !ok {
				t.Errorf("%s missing constituent URL for %s", name, idx)
			}
			if _, ok := cat.IndexFactsheetURLs[idx]; !ok {
				t.Errorf("%s missing factsheet URL for %s", name, idx)
			}
		}
	}
	if total != 107 {
		t.Errorf("want 107 total indices got %d", total)
	}
}

func TestInvalidCategory(t *testing.T) {
	if _, err := IndexList("Nope"); err == nil {
		t.Error("want error on bad category")
	}
	if err := ValidateIndexName("BroadMarketIndices", "Nope"); err == nil {
		t.Error("want error on bad index")
	}
	if err := ValidateIndexName("BroadMarketIndices", "Nifty 50"); err != nil {
		t.Errorf("valid index rejected: %v", err)
	}
}

func TestFactsheetURL(t *testing.T) {
	u, err := FactsheetURL("BroadMarketIndices", "Nifty 50")
	if err != nil || u == "" {
		t.Errorf("got %q %v", u, err)
	}
}
