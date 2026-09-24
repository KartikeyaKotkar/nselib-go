// Package indices ports nselib/indices to Go.
package indices

import (
	"github.com/KartikeyaKotkar/nselib-go"
)

var defaultClient = nselib.NewNSEClient()

// SetClient overrides the HTTP client (tests).
func SetClient(c *nselib.NSEClient) { defaultClient = c }

// ValidCategories lists the supported index categories.
var ValidCategories = []string{
	"SectoralIndices",
	"BroadMarketIndices",
	"ThematicIndices",
	"StrategyIndices",
}

// GetClass resolves an index category to its config.
// Python builds f"Nifty{index_category}" dynamically; Go looks up Categories.
func GetClass(indexCategory string) (*IndexCategory, error) {
	if cat, ok := Categories[indexCategory]; ok {
		return cat, nil
	}
	return nil, nselib.NewInvalidIndexCategoryError("'" + indexCategory + "': is an invalid Index Category")
}

// IndexList returns available indices for a category.
func IndexList(indexCategory string) ([]string, error) {
	cat, err := GetClass(indexCategory)
	if err != nil {
		return nil, err
	}
	return cat.IndicesList, nil
}

// ValidateIndexCategory checks the category name.
func ValidateIndexCategory(indexCategory string) error {
	if err := nselib.ValidateParamFromList(indexCategory, ValidCategories); err != nil {
		return nselib.NewInvalidIndexCategoryError("'" + indexCategory + "': is an invalid Index Category")
	}
	return nil
}

// ValidateIndexName checks that an index exists in a category.
func ValidateIndexName(indexCategory, indexName string) error {
	if err := ValidateIndexCategory(indexCategory); err != nil {
		return err
	}
	list, _ := IndexList(indexCategory)
	for _, n := range list {
		if n == indexName {
			return nil
		}
	}
	return nselib.NewInvalidIndexError("'" + indexName + "' is invalid index_name.")
}

// FactsheetURL returns the factsheet PDF URL for an index.
func FactsheetURL(indexCategory, indexName string) (string, error) {
	if err := ValidateIndexName(indexCategory, indexName); err != nil {
		return "", err
	}
	cat, _ := GetClass(indexCategory)
	return cat.IndexFactsheetURLs[indexName], nil
}

// ConstituentStockList returns constituent stocks for an index.
func ConstituentStockList(indexCategory, indexName string) (nselib.DataFrame, error) {
	if err := ValidateIndexName(indexCategory, indexName); err != nil {
		return nil, err
	}
	cat, _ := GetClass(indexCategory)
	url := cat.IndexConstituentListURLs[indexName]
	if url == "" {
		return nil, nselib.NewIndexDataNotFoundError("'" + indexName + "': No Data found for index")
	}
	df, err := defaultClient.FetchCSV(url, "")
	if err != nil {
		return nil, nselib.NewIndexDataNotFoundError("'" + indexName + "': No Data found for index, Kindly check the index category & name")
	}
	return df, nil
}

// LiveIndexPerformances returns live/last-traded performance for all indices.
func LiveIndexPerformances() (nselib.DataFrame, error) {
	var raw map[string]interface{}
	if err := defaultClient.FetchJSON("https://www.nseindia.com/api/allIndices",
		"https://www.nseindia.com/market-data/index-performances", &raw); err != nil {
		return nil, nselib.NewAPIError("Resource not available MSG: " + err.Error())
	}
	arr, _ := raw["data"].([]interface{})
	df := make(nselib.DataFrame, 0, len(arr))
	for _, item := range arr {
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		delete(m, "chartTodayPath")
		delete(m, "chart30dPath")
		delete(m, "chart365dPath")
		df = append(df, nselib.Record(m))
	}
	return df, nil
}
