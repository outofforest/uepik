package types

import (
	"sort"
	"time"
)

// CostTaxType defines the tax type of the cost.
type CostTaxType string

// Cost tax types.
const (
	CostTaxTypeTaxable    CostTaxType = "taxable"
	CostTaxTypeNonTaxable CostTaxType = "nonTaxable"
)

// Period defines date range for fiscal year.
type Period struct {
	Start time.Time
	End   time.Time
}

// Contains checks if provided date fits into the range.
func (p Period) Contains(date time.Time) bool {
	return !date.Before(p.Start) && !date.After(p.End)
}

// Document defines the document.
type Document struct {
	ID   string
	Date time.Time
}

// CIT defines CIT properties.
type CIT struct {
	Date time.Time
}

// VAT defines VAT properties.
type VAT struct {
	Date time.Time
}

// Contractor defines contractor.
type Contractor struct {
	Name    string
	Address string
	TaxID   string
}

// Payment defines payment.
type Payment struct {
	Amount Denom
	Date   time.Time
	Index  uint64
}

// IsPaid tells if payment has been paid.
func (p Payment) IsPaid() bool {
	return !p.Date.IsZero()
}

// Operation defines operation which might bee accounted.
type Operation interface {
	BankRecords() []*BankRecord
	BookRecords(coa *ChartOfAccounts, bankRecords []*BankRecord, rates CurrencyRates)
}

// BookRecord defines the book record.
type BookRecord struct {
	Date            time.Time
	Index           uint64
	DayOfMonth      uint8
	Document        Document
	Contractor      Contractor
	Notes           string
	IncomeDonations Denom
	IncomeTrading   Denom
	IncomeOthers    Denom
	IncomeSum       Denom
	CostTaxed       Denom
	CostNotTaxed    Denom
}

// NewBookSummary creates new book summary.
func NewBookSummary() BookSummary {
	return BookSummary{
		IncomeDonations: BaseZero,
		IncomeTrading:   BaseZero,
		IncomeOthers:    BaseZero,
		IncomeSum:       BaseZero,
		CostTaxed:       BaseZero,
		CostNotTaxed:    BaseZero,
	}
}

// BookSummary is the page summary of the book report.
type BookSummary struct {
	IncomeDonations Denom
	IncomeTrading   Denom
	IncomeOthers    Denom
	IncomeSum       Denom
	CostTaxed       Denom
	CostNotTaxed    Denom
}

// AddRecord adds record to the summary.
func (bs BookSummary) AddRecord(r BookRecord) BookSummary {
	bs.IncomeDonations = bs.IncomeDonations.Add(r.IncomeDonations)
	bs.IncomeTrading = bs.IncomeTrading.Add(r.IncomeTrading)
	bs.IncomeOthers = bs.IncomeOthers.Add(r.IncomeOthers)
	bs.IncomeSum = bs.IncomeSum.Add(r.IncomeSum)
	bs.CostTaxed = bs.CostTaxed.Add(r.CostTaxed)
	bs.CostNotTaxed = bs.CostNotTaxed.Add(r.CostNotTaxed)
	return bs
}

// AddSummary adds another summary to this one.
func (bs BookSummary) AddSummary(bs2 BookSummary) BookSummary {
	bs.IncomeDonations = bs.IncomeDonations.Add(bs2.IncomeDonations)
	bs.IncomeTrading = bs.IncomeTrading.Add(bs2.IncomeTrading)
	bs.IncomeOthers = bs.IncomeOthers.Add(bs2.IncomeOthers)
	bs.IncomeSum = bs.IncomeSum.Add(bs2.IncomeSum)
	bs.CostTaxed = bs.CostTaxed.Add(bs2.CostTaxed)
	bs.CostNotTaxed = bs.CostNotTaxed.Add(bs2.CostNotTaxed)
	return bs
}

// VATRecord defines the VAT record.
type VATRecord struct {
	Date       time.Time
	Index      uint64
	DayOfMonth uint8
	Document   Document
	Contractor Contractor
	Notes      string
	Income     Denom
}

// NewVATSummary creates new VAT summary.
func NewVATSummary() VATSummary {
	return VATSummary{
		Income: BaseZero,
	}
}

// VATSummary is the page summary of the VAT report.
type VATSummary struct {
	Income Denom
}

// AddRecord adds record to the summary.
func (vs VATSummary) AddRecord(r VATRecord) VATSummary {
	vs.Income = vs.Income.Add(r.Income)
	return vs
}

// BankRecord defines the properties of bank record.
type BankRecord struct {
	Date           time.Time
	Index          uint64
	DayOfMonth     uint8
	Document       Document
	Contractor     Contractor
	OriginalAmount Denom
	BaseAmount     Denom
	Rate           Number
	OriginalSum    Denom
	BaseSum        Denom
	RateAverage    Number
}

// GetDate returns record's date.
func (r BankRecord) GetDate() time.Time {
	return r.Date
}

// Summary creates page summary from the record.
func (r BankRecord) Summary() BankSummary {
	return BankSummary{
		OriginalSum: r.OriginalSum,
		BaseSum:     r.BaseSum,
		RateAverage: r.RateAverage,
	}
}

// NewBankSummary creates new bank summary.
func NewBankSummary(currencyInit InitCurrency) BankSummary {
	return BankSummary{
		OriginalSum: currencyInit.OriginalSum,
		BaseSum:     currencyInit.BaseSum,
		RateAverage: currencyInit.BaseSum.Rate(currencyInit.OriginalSum),
	}
}

// BankSummary is the page summary of the bank record.
type BankSummary struct {
	OriginalSum Denom
	BaseSum     Denom
	RateAverage Number
}

// Report is the full report.
type Report struct {
	CompanyName    string
	CompanyAddress string
	Book           []BookReport
	Flow           []FlowReport
	VAT            []VATReport
	Bank           []BankCurrency
}

// BookReport stores book report.
type BookReport struct {
	Year                uint64
	Month               string
	Page                uint64
	Records             []BookRecord
	CurrentPageSummary  BookSummary
	PreviousPageSummary BookSummary
	MonthSummary        BookSummary
	YearSummary         BookSummary
}

// FlowReport is the flow report.
type FlowReport struct {
	Year  uint64
	Month string

	MonthIncome                Denom
	MonthCostsTaxed            Denom
	MonthProfit                Denom
	MonthCostsNotTaxedCurrent  Denom
	MonthCostsNotTaxedPrevious Denom

	TotalIncome                Denom
	TotalCostsTaxed            Denom
	TotalProfitYear            Denom
	TotalCostsNotTaxedCurrent  Denom
	TotalProfitPrevious        Denom
	TotalCostsNotTaxedPrevious Denom
	TotalProfit                Denom
}

// VATReport stores VAT report.
type VATReport struct {
	Year                uint64
	Month               string
	Page                uint64
	Records             []VATRecord
	PreviousPageSummary VATSummary
	CurrentPageSummary  VATSummary
}

// BankCurrency stores bank reports for single currency.
type BankCurrency struct {
	Currency Currency
	Reports  []BankReport
}

// BankReport is the bank report.
type BankReport struct {
	Year                uint64
	Month               string
	Page                uint64
	Records             []BankRecord
	PreviousPageSummary BankSummary
	CurrentPageSummary  BankSummary
}

// PreviousDay computes the date of the previous day.
func PreviousDay(date time.Time) time.Time {
	return date.AddDate(0, 0, -1)
}

// FiscalYear defines fiscal year.
type FiscalYear struct {
	CompanyName     string
	CompanyAddress  string
	ChartOfAccounts *ChartOfAccounts
	Period          Period
	Init            Init
	Operations      []Operation
}

// BankReports returns bank reports.
func (fy *FiscalYear) BankReports(
	currencyRates CurrencyRates,
	years []*FiscalYear,
) (map[CurrencySymbol]*[]BankRecord, map[Operation][]*BankRecord) {
	opBankRecords := make(map[Operation][]*BankRecord, len(fy.Operations))
	for _, op := range fy.Operations {
		opBankRecords[op] = nil
	}
	for _, y := range years {
		bankRecords := []*BankRecord{}
		for _, op := range y.Operations {
			brs := op.BankRecords()
			bankRecords = append(bankRecords, brs...)

			if _, exists := opBankRecords[op]; exists {
				for _, br := range brs {
					if y.Period.Contains(br.Date) {
						opBankRecords[op] = append(opBankRecords[op], br)
					}
				}
			}
		}
		report := y.bankReports(bankRecords, currencyRates)
		if y == fy {
			return report, opBankRecords
		}
	}
	panic("current fiscal year is not on the list")
}

// BookRecords generates book records.
func (fy *FiscalYear) BookRecords(currencyRates CurrencyRates, bankRecords map[Operation][]*BankRecord) {
	for _, o := range fy.Operations {
		o.BookRecords(fy.ChartOfAccounts, bankRecords[o], currencyRates)
	}
}

func (fy *FiscalYear) bankReports(
	bankRecords []*BankRecord,
	currencyRates CurrencyRates,
) map[CurrencySymbol]*[]BankRecord {
	currencies := map[CurrencySymbol][]*BankRecord{}
	for _, br := range bankRecords {
		currencies[br.OriginalAmount.Currency] = append(currencies[br.OriginalAmount.Currency], br)
	}

	var zeroDenom Denom
	var zeroRate Number

	reports := map[CurrencySymbol]*[]BankRecord{}

	for currencySymbol, records := range currencies {
		currency := Currencies.Currency(currencySymbol)

		sort.Slice(records, func(i, j int) bool {
			return records[i].Date.Before(records[j].Date) || (records[i].Date.Equal(records[j].Date) &&
				records[i].Index < records[j].Index)
		})

		records2 := make([]BankRecord, 0, len(records))

		total, exists := fy.Init.Currencies[currencySymbol]
		if !exists {
			panic("brak bilansu otwarcia waluty")
		}

		originalZero := Denom{
			Currency: currencySymbol,
			Amount:   NewNumber(0, 0, currency.AmountPrecision),
		}
		rate := total.BaseSum.Rate(total.OriginalSum)

		for i, br := range records {
			br.Index = uint64(i + 1)
			br.DayOfMonth = uint8(br.Date.Day())

			switch {
			case br.OriginalAmount != zeroDenom && br.BaseAmount == zeroDenom && br.Rate == zeroRate &&
				br.OriginalAmount.GT(originalZero):
				br.BaseAmount, br.Rate = currencyRates.ToBase(br.OriginalAmount, PreviousDay(br.Date))
			case br.OriginalAmount != zeroDenom && br.BaseAmount == zeroDenom && br.Rate == zeroRate:
				br.Rate = rate
				br.BaseAmount = br.OriginalAmount.ToBase(rate)
			case br.OriginalAmount != zeroDenom && br.BaseAmount == zeroDenom && br.Rate != zeroRate:
				br.BaseAmount = br.OriginalAmount.ToBase(br.Rate)
			case br.OriginalAmount != zeroDenom && br.BaseAmount != zeroDenom && br.Rate == zeroRate:
				br.Rate = br.BaseAmount.Rate(br.OriginalAmount)
			default:
				panic("invalid data in bank record")
			}

			total.OriginalSum = total.OriginalSum.Add(br.OriginalAmount)
			total.BaseSum = total.BaseSum.Add(br.BaseAmount)
			rate = total.BaseSum.Rate(total.OriginalSum)

			br.OriginalSum = total.OriginalSum
			br.BaseSum = total.BaseSum
			br.RateAverage = rate

			records2 = append(records2, *br)
		}

		reports[currencySymbol] = &records2
	}

	return reports
}
