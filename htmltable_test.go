package nselib

import "testing"

func TestExtractHTMLTables(t *testing.T) {
	html := `<html><body><table>
<tr><th>Name</th><th>Val</th></tr>
<tr><td>A</td><td>1</td></tr>
<tr><td></td><td></td></tr>
<tr><td>B</td><td>2</td></tr>
</table></body></html>`
	tables := ExtractHTMLTables(html)
	if len(tables) != 1 {
		t.Fatalf("want 1 table got %d", len(tables))
	}
	tb := tables[0]
	if len(tb.Headers) != 2 || tb.Headers[0] != "Name" {
		t.Errorf("bad headers: %v", tb.Headers)
	}
	if len(tb.Rows) != 2 {
		t.Errorf("empty row not dropped: %v", tb.Rows)
	}
}

func TestExtractHTMLTablesNone(t *testing.T) {
	if got := ExtractHTMLTables("<html><body>hi</body></html>"); len(got) != 0 {
		t.Errorf("want none got %v", got)
	}
}
