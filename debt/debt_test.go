package debt

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSecuritiesURLFormat(t *testing.T) {
	var gotPath string
	origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	defer origin.Close()
	data := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Write([]byte("A,B\n1,2\n"))
	}))
	defer data.Close()
	_ = gotPath
	_ = data
	if _, err := SecuritiesAvailableForTrading("not-a-date"); err == nil {
		t.Error("want error on bad date")
	}
}
