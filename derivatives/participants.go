package derivatives

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/KartikeyaKotkar/nselib-go"
)

// parseParticipantCSV parses skiprows=1 participant files, dropping ragged lines
// (mirrors on_bad_lines="skip") and stripping tabs from headers.
func parseParticipantCSV(body []byte) (nselib.DataFrame, error) {
	df, err := nselib.ParseCSV(bytes.NewReader(body), nselib.SkipRows(1))
	if err != nil {
		return nil, err
	}
	if len(df) == 0 {
		return df, nil
	}
	width := len(df[0])
	out := make(nselib.DataFrame, 0, len(df))
	for _, r := range df {
		if len(r) != width {
			continue
		}
		rec := nselib.Record{}
		for k, v := range r {
			nk := strings.ReplaceAll(k, "\t", "")
			rec[nk] = v
		}
		out = append(out, rec)
	}
	// drop trailing summary row if non-numeric first column looks like a footer
	return out, nil
}

// ParticipantWiseOpenInterest fetches participant OI with archive fallback.
func ParticipantWiseOpenInterest(tradeDate string) (nselib.DataFrame, error) {
	t, err := time.Parse(nselib.LayoutDDMMYYYY, tradeDate)
	if err != nil {
		return nil, err
	}
	primary := fmt.Sprintf("https://nsearchives.nseindia.com/content/nsccl/fao_participant_oi_%s.csv",
		t.Format(nselib.LayoutDDMMYYYYCompact))
	body, err := defaultClient.FetchBytes(primary, "")
	if err != nil {
		fallback := fmt.Sprintf("https://archives.nseindia.com/content/nsccl/fao_participant_oi_%s.csv",
			t.Format(nselib.LayoutDDMMYYYYCompact))
		if body, err = defaultClient.FetchBytes(fallback, ""); err != nil {
			return nil, nselib.NewDataNotFoundError(fmt.Sprintf("No data available for : %s", tradeDate))
		}
	}
	return parseParticipantCSV(body)
}

// ParticipantWiseTradingVolume fetches participant trading volume.
func ParticipantWiseTradingVolume(tradeDate string) (nselib.DataFrame, error) {
	t, err := time.Parse(nselib.LayoutDDMMYYYY, tradeDate)
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("https://nsearchives.nseindia.com/content/nsccl/fao_participant_vol_%s.csv",
		t.Format(nselib.LayoutDDMMYYYYCompact))
	body, err := defaultClient.FetchBytes(url, "")
	if err != nil {
		return nil, nselib.NewDataNotFoundError(fmt.Sprintf("No data available for : %s", tradeDate))
	}
	return parseParticipantCSV(body)
}
