package nselib

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNSEClientFetchLocal(t *testing.T) {
	origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{Name: "nse", Value: "1"})
		w.WriteHeader(200)
	}))
	defer origin.Close()
	data := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"ok":true}`))
	}))
	defer data.Close()

	c := NewNSEClient()
	var target map[string]bool
	if err := c.FetchJSON(data.URL, origin.URL, &target); err != nil {
		t.Fatal(err)
	}
	if !target["ok"] {
		t.Errorf("bad body: %v", target)
	}
}

func TestNSEClientFetchCSVLocal(t *testing.T) {
	origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	defer origin.Close()
	data := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("A,B\n1,2\n"))
	}))
	defer data.Close()

	c := NewNSEClient()
	df, err := c.FetchCSV(data.URL, origin.URL)
	if err != nil {
		t.Fatal(err)
	}
	if len(df) != 1 || df[0]["A"] != "1" {
		t.Errorf("bad df: %v", df)
	}
}

func TestNSEClientAPIError(t *testing.T) {
	origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	defer origin.Close()
	data := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	}))
	defer data.Close()

	c := NewNSEClient()
	if _, err := c.FetchBytes(data.URL, origin.URL); err == nil {
		t.Error("want error on 500")
	}
}
