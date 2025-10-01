package types

import (
	"sort"
	"text/template"
	"time"
)

// CostTaxType defines the tax type of the cost.
type CostTaxType string

// Cost tax types.
const (
	CostTaxTypeTaxable    CostTaxType = "taxable"
	CostTaxTypeNonTaxable CostTaxType = "nonTaxable"
)

// CostCategoryType defines the category type of the cost.
type CostCategoryType string

// Cost category types.
const (
	CostCategoryTypeFreeOfCharge CostCategoryType = "freeOfCharge"
	CostCategoryTypePaid         CostCategoryType = "paid"
)

// SellType is the sell type.
type SellType string

// Sell types.
const (
	SellTypeRecorded   = "recorded"
	SellTypeUnrecorded = "unrecorded"
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

// Months returns month dates belonging to the period.
func (p Period) Months() []time.Time {
	months := make([]time.Time, 0, 12)
	month := p.Start.AddDate(0, 0, -p.Start.Day()+1)
	for ; !p.End.Before(month); month = month.AddDate(0, 1, 0) {
		months = append(months, month)
	}
	return months
}

// DocumentID represents document ID.
type DocumentID string

// Document defines the document.
type Document struct {
	ID        DocumentID
	Date      time.Time
	SheetName string
}

// Contractor defines contractor.
type Contractor struct {
	Name    string
	Address string
	TaxID   string
}

// Due represents due.
type Due struct {
	Date   time.Time
	Amount Denom
}

// Payment defines payment.
type Payment struct {
	DocumentID DocumentID
	Date       time.Time
	Index      uint64
	Amount     Denom
}

// Operation defines operation which might bee accounted.
type Operation interface {
	BankRecords() []*BankRecord
	BookRecords(period Period, coa *ChartOfAccounts, bankRecords []*BankRecord, rates CurrencyRates) []ReportDocument
}

// ReportDocument represents a document in the report.
type ReportDocument struct {
	Date     time.Time
	Index    uint64
	Template *template.Template
	Data     any
	Config   SheetConfig
}

// Report is the full report.
type Report struct {
	Currencies []Currency
	Configs    []string
	Documents  []string
}

// BankRecord defines the properties of bank record.
type BankRecord struct {
	Date           time.Time
	Index          uint64
	Document       DocumentID
	DayOfMonth     uint8
	PaidDocument   Document
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

// FiscalYear defines fiscal year.
type FiscalYear struct {
	CompanyName    string
	CompanyAddress string
	CompanyTaxID   string
	Period         Period
	Init           Init
	Operations     []Operation
}

// BankReports returns bank reports.
func (fy *FiscalYear) BankReports(
	currencyRates CurrencyRates,
	years []*FiscalYear,
) (map[CurrencySymbol][]BankRecord, map[Operation][]*BankRecord) {
	opBankRecords := make(map[Operation][]*BankRecord, len(fy.Operations))
	for _, op := range fy.Operations {
		opBankRecords[op] = nil
	}
	for _, y := range years {
		bankRecords := []*BankRecord{}
		for _, op := range y.Operations {
			for _, br := range op.BankRecords() {
				if y.Period.Contains(br.Date) {
					bankRecords = append(bankRecords, br)
					if _, exists := opBankRecords[op]; exists {
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
func (fy *FiscalYear) BookRecords(
	coa *ChartOfAccounts,
	currencyRates CurrencyRates,
	bankRecords map[Operation][]*BankRecord,
) []ReportDocument {
	docs := []ReportDocument{}
	for _, o := range fy.Operations {
		docs = append(docs, o.BookRecords(fy.Period, coa, bankRecords[o], currencyRates)...)
	}
	for i := range docs {
		docs[i].Index = uint64(i)
	}
	sort.Slice(docs, func(i, j int) bool {
		d1 := docs[i]
		d2 := docs[j]
		return d1.Date.Before(d2.Date) || (d1.Date.Equal(d2.Date) && d1.Index < d2.Index)
	})

	return docs
}

func (fy *FiscalYear) bankReports(
	bankRecords []*BankRecord,
	currencyRates CurrencyRates,
) map[CurrencySymbol][]BankRecord {
	currencies := map[CurrencySymbol][]*BankRecord{}
	for _, br := range bankRecords {
		currencies[br.OriginalAmount.Currency] = append(currencies[br.OriginalAmount.Currency], br)
	}

	var zeroDenom Denom
	var zeroRate Number

	reports := map[CurrencySymbol][]BankRecord{}

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

		reports[currencySymbol] = records2
	}

	return reports
}

// PreviousDay computes the date of the previous day.
func PreviousDay(date time.Time) time.Time {
	return date.AddDate(0, 0, -1)
}

// MinDate returns the earliest date from set.
func MinDate(dates ...time.Time) time.Time {
	if len(dates) == 0 {
		panic("no dates")
	}
	date := dates[0]
	for _, d := range dates[1:] {
		if d.Before(date) {
			date = d
		}
	}
	return date
}

// MaxDate returns the latest date from set.
func MaxDate(dates ...time.Time) time.Time {
	if len(dates) == 0 {
		panic("no dates")
	}
	date := dates[0]
	for _, d := range dates[1:] {
		if d.After(date) {
			date = d
		}
	}
	return date
}

// NewVAT returns new VAT data source.
func NewVAT(date time.Time, data EntryDataSource) *VAT {
	return &VAT{
		Date: date,
		Data: data,
	}
}

// VAT is the data source related to VAT.
type VAT struct {
	Date time.Time
	Data EntryDataSource
}

// GetDate returns date.
func (v *VAT) GetDate() time.Time {
	return v.Date
}

// GetDocument returns document.
func (v *VAT) GetDocument() Document {
	return v.Data.GetDocument()
}

// GetContractor returns contractor.
func (v *VAT) GetContractor() Contractor {
	return v.Data.GetContractor()
}

// GetNotes returns notes.
func (v *VAT) GetNotes() string {
	return v.Data.GetNotes()
}

// NewCurrencyDiff creates new currency diff data source.
func NewCurrencyDiff(data EntryDataSource, dataRate Number, bankRecord *BankRecord) *CurrencyDiff {
	return &CurrencyDiff{
		Data:       data,
		DataRate:   dataRate,
		BankRecord: bankRecord,
	}
}

// CurrencyDiff is the data source of currency diff.
type CurrencyDiff struct {
	Data       EntryDataSource
	DataRate   Number
	BankRecord *BankRecord
}

// GetDate returns date.
func (cd *CurrencyDiff) GetDate() time.Time {
	return MaxDate(cd.Data.GetDate(), cd.BankRecord.Date)
}

// GetDocument returns document.
func (cd *CurrencyDiff) GetDocument() Document {
	return cd.Data.GetDocument()
}

// GetContractor returns contractor.
func (cd *CurrencyDiff) GetContractor() Contractor {
	return cd.Data.GetContractor()
}

// GetNotes returns notes.
func (cd *CurrencyDiff) GetNotes() string {
	return cd.Data.GetNotes()
}

// SheetConfig stores sheet config.
type SheetConfig struct {
	Name       string
	LockedRows uint64
}
