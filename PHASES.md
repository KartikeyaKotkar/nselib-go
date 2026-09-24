# nselib-go Build Phases

Decisions locked in Phase 0: `TotalTradedStocksResult{Summary, Details}` struct, phased delivery, PDF table gap flagged, module `github.com/KartikeyaKotkar/nselib-go`, root `package nselib`.

## Phase 0 — Decisions [DONE in plan]
- Q1 struct, Q2 phased, PDF `ledongthuc/pdf` vs `unidoc/unipdf` evaluation pending Phase 4.
- Refs: `python-to-go-conversion-plan.md:68,72,602`, `plan-review.md:52`.

## Phase 1 — Root foundation [DONE]
- Files: `go.mod, types.go, errors.go, constants.go, logger.go, httpclient.go, csv.go, dateutil.go, trading_calendar.go`
- Fixes: `3M → SubtractMonths(today,3)` (`nselib/libutil.py:120`), distinct `NewIndexDataNotFoundError` (`nselib/errors.py:79`).
- Tests: `errors_test.go, csv_test.go, dateutil_test.go, httpclient_test.go`.
- Exit: `go build ./...`, `go vet ./...`, `go test ./...` green.

## Phase 2 — capitalmarket + debt
- `capitalmarket/capitalmarket.go, fetchers.go, bhav.go` (5 funcs incl `bhav_copy_indices`), `reports.go, live.go, financials.go, turnover.go` (`extrame/xls`), `growth.go`. `debt/debt.go` bundled.
- Generic `fetchInChunks(365)` proven here.
- Exit: `TestEquityList`, `TestPriceVolumeData` parity.

## Phase 3 — derivatives + indices
- Derivatives 90-day chunks, `optionchain.go, participants.go`, FII `.xls`, F&O bhav ZIP fallback. Indices `config.go` ~300 lines, `Nifty` prefix vars + `Categories` map.
- Exit: `TestIndexList`, `TestFuturePriceVolumeData`.

## Phase 4 — mutualfunds + nsdlfpi + cashmarket
- `mutualfunds/amfi.go` Excel/HTML/PDF priority. `nsdlfpi/` types/client/parser/browser (`chromedp`). `cashmarket/` 7 delegates last.
- Exit: AMFI fetch, `TestFetchLatestBundle`, facade green.

## Phase 5 — Parity + release
- `go vet`, `golangci-lint`, `go doc`, Python vs Go shape compare, `PriceVolumeData SBIN 1Y` timing.
