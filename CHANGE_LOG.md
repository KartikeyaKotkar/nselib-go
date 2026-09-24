# CHANGE LOG
#### All notable changes to nselib-go are listed here.
#### Python lineage lives upstream at [RuchiTanmay/nselib](https://github.com/RuchiTanmay/nselib/blob/main/CHANGE_LOG.md).

### Version: v0.1.0 [24/09/2026]
* Full Go port of nselib with complete API parity (~100 functions, 8 packages).
* Phase 1 — root foundation: `DataFrame`/`Record` types, error hierarchy (incl. `IndexDataNotFound`),
  cookie-priming HTTP client, CSV/date utilities, trading calendar.
* Phase 2 — `capitalmarket` (history, bhavcopies incl. `bhav_copy_indices`, VaR, live market,
  XBRL financials, `xls` turnover, CM growth) + `debt`.
* Phase 3 — `derivatives` (90-day history, option chain full/compact, participant OI/volume,
  FII stats, ban list) + `indices` (107 indices, generated config).
* Phase 4 — `mutualfunds` (AMFI xls/xlsx/html/pdf-text), `nsdlfpi` (HTTP + headless-Chrome fallback),
  `cashmarket` facade.
* Phase 5 — verification: `go vet`, `go test -race` green on linux/windows/macos, live NSE smoke green.
* Performance over Python: cookies prime once per origin; history chunks, XBRL filings, AMFI months,
  and expiries fetch concurrently. Measured NIFTY OPTIDX CE 1Y (208,906 rows): Go 7.8s vs Python 20.9s.
* Fixes ported from review: `3M` period handled in date derivation, `IndexDataNotFound` preserved.
* CI (build/vet/race on 3 OSes) + tag-triggered GitHub releases.
