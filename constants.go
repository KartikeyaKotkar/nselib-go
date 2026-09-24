package nselib

// Date format layouts for Go's time.Parse / time.Format.
const (
	LayoutDDMMYYYY        = "02-01-2006"  // dd-mm-YYYY
	LayoutDDMMMYYYY       = "02-Jan-2006" // dd-Mon-YYYY
	LayoutDDMMYYYYCompact = "02012006"    // ddmmYYYY
	LayoutDDMMYYCompact   = "020106"      // ddmmyy
	LayoutMMMYY           = "Jan-06"      // Mon-yy
	LayoutYYYYMMDD        = "2006-01-02"  // YYYY-mm-dd (for get_month_from_date input)
	LayoutMon             = "Jan"         // Mon (for get_month_from_date output)
)

// Valid period values.
var EquityPeriods = []string{"1D", "1W", "1M", "3M", "6M", "1Y"}

// IndicesList for derivative routing.
var IndicesList = []string{"NIFTY", "FINNIFTY", "BANKNIFTY"}

// Column name slices mirroring nselib/constants.py.
var (
	PriceVolumeDeliverableColumns = []string{
		"Symbol", "Series", "Date", "PrevClose", "OpenPrice", "HighPrice",
		"LowPrice", "LastPrice", "ClosePrice", "AveragePrice", "TotalTradedQuantity",
		"TurnoverInRs", "No.ofTrades", "DeliverableQty", "%DlyQttoTradedQty",
	}
	PriceVolumeColumns = []string{
		"Symbol", "Series", "Date", "PrevClose", "OpenPrice", "HighPrice",
		"LowPrice", "LastPrice", "ClosePrice", "AveragePrice",
		"TotalTradedQuantity", "Turnover", "No.ofTrades",
	}
	DeliverableColumns = []string{
		"Symbol", "Series", "Date", "TradedQty", "DeliverableQty", "%DlyQttoTradedQty",
	}
	BulkDealColumns = []string{
		"Date", "Symbol", "SecurityName", "ClientName", "Buy/Sell", "QuantityTraded",
		"TradePrice/Wght.Avg.Price", "Remarks",
	}
	BlockDealsColumns = []string{
		"Date", "Symbol", "SecurityName", "ClientName", "Buy/Sell", "QuantityTraded",
		"TradePrice/Wght.Avg.Price", "Remarks",
	}
	ShortSellingColumns = []string{"Date", "Symbol", "SecurityName", "Quantity"}
	BhavcopyOldColumns  = []string{
		"TradDt", "ISIN", "TckrSymb", "SctySrs", "OpnPric", "HghPric", "LwPric",
		"ClsPric", "LastPric", "PrvsClsgPric", "TtlTradgVol", "TtlTrfVal", "TtlNbOfTxsExctd",
	}
	BhavcopyNewColumns = []string{
		"TIMESTAMP", "ISIN", "SYMBOL", "SERIES", "OPEN", "HIGH", "LOW",
		"CLOSE", "LAST", "PREVCLOSE", "TOTTRDQTY", "TOTTRDVAL", "TOTALTRADES",
	}
	FuturePriceVolumeColumns = []string{
		"TIMESTAMP", "INSTRUMENT", "SYMBOL", "EXPIRY_DT", "STRIKE_PRICE", "OPTION_TYPE",
		"MARKET_TYPE", "OPENING_PRICE", "TRADE_HIGH_PRICE", "TRADE_LOW_PRICE",
		"CLOSING_PRICE", "LAST_TRADED_PRICE", "PREV_CLS", "SETTLE_PRICE",
		"TOT_TRADED_QTY", "TOT_TRADED_VAL", "OPEN_INT", "CHANGE_IN_OI",
		"MARKET_LOT", "UNDERLYING_VALUE",
	}
	IndiaVIXColumns = []string{
		"TIMESTAMP", "INDEX_NAME", "OPEN_INDEX_VAL", "CLOSE_INDEX_VAL",
		"HIGH_INDEX_VAL", "LOW_INDEX_VAL", "PREV_CLOSE", "VIX_PTS_CHG", "VIX_PERC_CHG",
	}
	IndexDataColumns = []string{
		"INDEX_NAME", "OPEN_INDEX_VAL", "HIGH_INDEX_VAL", "CLOSE_INDEX_VAL",
		"LOW_INDEX_VAL", "TURN_OVER", "TRADED_QTY", "TIMESTAMP",
	}
	VarColumns = []string{
		"RecordType", "Symbol", "Series", "Isin", "SecurityVaR", "IndexVaR",
		"VaRMargin", "ExtremeLossRate", "AdhocMargin", "ApplicableMarginRate",
	}
)
