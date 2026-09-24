package derivatives

import (
	"archive/zip"
	"bytes"
	"fmt"
	"time"

	"github.com/KartikeyaKotkar/nselib-go"
)

func extractBhavZip(body []byte) (nselib.DataFrame, error) {
	zr, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
	if err != nil {
		return nil, nselib.NewDataNotFoundError("open bhav zip: " + err.Error())
	}
	var out nselib.DataFrame
	for _, f := range zr.File {
		rc, err := f.Open()
		if err != nil {
			continue
		}
		df, err := nselib.ParseCSV(rc)
		_ = rc.Close()
		if err != nil {
			continue
		}
		out = append(out, df...)
	}
	return out, nil
}

// FNOBhavCopy fetches the F&O bhavcopy ZIP, falling back to the reports API on 403.
func FNOBhavCopy(tradeDate string) (nselib.DataFrame, error) {
	t, err := time.Parse(nselib.LayoutDDMMYYYY, tradeDate)
	if err != nil {
		return nil, err
	}
	primary := fmt.Sprintf("https://nsearchives.nseindia.com/content/fo/BhavCopy_NSE_FO_0_0_0_%s_F_0000.csv.zip",
		t.Format("20060102"))
	if body, err := defaultClient.FetchBytes(primary, ""); err == nil {
		return extractBhavZip(body)
	}
	fallback := "https://www.nseindia.com/api/reports?archives=" +
		"%5B%7B%22name%22%3A%22F%26O%20-%20Bhavcopy(csv)%22%2C%22type%22%3A%22archives%22%2C%22category%22" +
		fmt.Sprintf("%%3A%%22derivatives%%22%%2C%%22section%%22%%3A%%22equity%%22%%7D%%5D&date=%s",
			t.Format(nselib.LayoutDDMMMYYYY)) + "&type=equity&mode=single"
	body, err := defaultClient.FetchBytes(fallback, "")
	if err != nil {
		return nil, nselib.NewDataNotFoundError("Data not found, change the date...")
	}
	return extractBhavZip(body)
}
