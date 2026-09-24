package mutualfunds

import (
	"bytes"
	"strings"

	"github.com/ledongthuc/pdf"
)

// extractPDFText returns per-page text using ledongthuc/pdf (MIT, text-only).
// Pages are joined by "\n\x0c\n" for ParseReportContent.
func extractPDFText(content []byte) (string, error) {
	r, err := pdf.NewReader(bytes.NewReader(content), int64(len(content)))
	if err != nil {
		return "", err
	}
	var pages []string
	for i := 1; i <= r.NumPage(); i++ {
		rows, err := r.Page(i).GetTextByRow()
		if err != nil {
			return "", err
		}
		var lines []string
		for _, row := range rows {
			var parts []string
			for _, t := range row.Content {
				if s := strings.TrimSpace(t.S); s != "" {
					parts = append(parts, s)
				}
			}
			if len(parts) > 0 {
				lines = append(lines, strings.Join(parts, " "))
			}
		}
		pages = append(pages, strings.Join(lines, "\n"))
	}
	return strings.Join(pages, "\n\x0c\n"), nil
}
