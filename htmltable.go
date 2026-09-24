package nselib

import (
	"strings"

	"golang.org/x/net/html"
)

// HTMLTable is a plain header + rows table extracted from HTML.
// Replaces pd.read_html for NSDL FPI and AMFI report parsing.
type HTMLTable struct {
	Headers []string
	Rows    [][]string
}

// ExtractHTMLTables parses all <table> elements in htmlText.
// First row becomes Headers, the rest Rows. Nested markup is flattened to text.
func ExtractHTMLTables(htmlText string) []HTMLTable {
	doc, err := html.Parse(strings.NewReader(htmlText))
	if err != nil {
		return nil
	}
	var out []HTMLTable
	var walk func(n *html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "table" {
			if t := parseTable(n); len(t.Headers) > 0 || len(t.Rows) > 0 {
				out = append(out, t)
			}
			return // don't descend into nested tables twice
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)
	return out
}

func parseTable(tbl *html.Node) HTMLTable {
	var grid [][]string
	var walkRows func(n *html.Node)
	walkRows = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "tr" {
			var row []string
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				if c.Type == html.ElementNode && (c.Data == "td" || c.Data == "th") {
					row = append(row, cellText(c))
				}
			}
			grid = append(grid, row)
			return
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walkRows(c)
		}
	}
	walkRows(tbl)
	// drop fully-empty rows
	var kept [][]string
	for _, r := range grid {
		empty := true
		for _, v := range r {
			if strings.TrimSpace(v) != "" {
				empty = false
				break
			}
		}
		if !empty {
			kept = append(kept, r)
		}
	}
	if len(kept) == 0 {
		return HTMLTable{}
	}
	return HTMLTable{Headers: kept[0], Rows: kept[1:]}
}

func cellText(n *html.Node) string {
	var sb strings.Builder
	var walk func(x *html.Node)
	walk = func(x *html.Node) {
		if x.Type == html.TextNode {
			sb.WriteString(x.Data)
		}
		for c := x.FirstChild; c != nil; c = c.NextSibling {
			if c.Type == html.ElementNode && (c.Data == "table") {
				continue
			}
			walk(c)
		}
	}
	walk(n)
	return strings.Join(strings.Fields(sb.String()), " ")
}
