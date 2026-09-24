package capitalmarket

import (
	"archive/zip"
	"bytes"
	"fmt"
	"time"

	"github.com/KartikeyaKotkar/nselib-go"
)

func parseTradeDate(tradeDate string) (time.Time, error) {
	return time.Parse(nselib.LayoutDDMMYYYY, tradeDate)
}

// BhavCopyWithDelivery fetches delivery bhavcopy CSV for a trade date.
func BhavCopyWithDelivery(tradeDate string) (nselib.DataFrame, error) {
	t, err := parseTradeDate(tradeDate)
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("https://nsearchives.nseindia.com/products/content/sec_bhavdata_full_%s.csv",
		t.Format(nselib.LayoutDDMMYYYYCompact))
	return defaultClient.FetchCSV(url, "")
}

// BhavCopyEquities fetches the zipped CM bhavcopy for a trade date.
func BhavCopyEquities(tradeDate string) (nselib.DataFrame, error) {
	t, err := parseTradeDate(tradeDate)
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("https://nsearchives.nseindia.com/content/cm/BhavCopy_NSE_CM_0_0_0_%s_F_0000.csv.zip",
		t.Format("20060102"))
	body, err := defaultClient.FetchBytes(url, "")
	if err != nil {
		return nil, err
	}
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

// BhavCopyIndices fetches the index closing bhavcopy (simple CSV).
func BhavCopyIndices(tradeDate string) (nselib.DataFrame, error) {
	t, err := parseTradeDate(tradeDate)
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("https://nsearchives.nseindia.com/content/indices/ind_close_all_%s.csv",
		t.Format(nselib.LayoutDDMMYYYYCompact))
	return defaultClient.FetchCSV(url, "")
}

// BhavCopySME fetches the SME bhavcopy.
func BhavCopySME(tradeDate string) (nselib.DataFrame, error) {
	t, err := parseTradeDate(tradeDate)
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("https://nsearchives.nseindia.com/archives/sme/bhavcopy/sme%s.csv",
		t.Format(nselib.LayoutDDMMYYCompact))
	return defaultClient.FetchCSV(url, "")
}

// SMEBhavCopy is an alias of BhavCopySME (matches Python duplicate).
func SMEBhavCopy(tradeDate string) (nselib.DataFrame, error) {
	return BhavCopySME(tradeDate)
}
