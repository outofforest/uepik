package documents

import (
	_ "embed"
	"text/template"
	"time"

	"github.com/outofforest/uepik/v2/accounts"
	"github.com/outofforest/uepik/v2/types"
)

var (
	_ types.SheetSource     = &CurrencyDiffDocument{}
	_ types.EntryDataSource = &CurrencyDiffDocument{}

	//go:embed currencydiff.tmpl.xml
	currencyDiffTmpl     string
	currencyDiffTemplate = template.Must(template.New("currencyDiff").Funcs(template.FuncMap{
		"date": date,
	}).Parse(currencyDiffTmpl))
)

// CurrencyDiffDocument represents currency diff documents.
type CurrencyDiffDocument struct {
	Document types.Document
	Company  types.Contractor
	Records  []CurrencyDiffRecord
	Summary  CurrencyDiffSummary
}

// GetDate returns date of currency diff.
func (d *CurrencyDiffDocument) GetDate() time.Time {
	return d.Document.Date
}

// GetDocument returns document.
func (d *CurrencyDiffDocument) GetDocument() types.Document {
	return d.Document
}

// GetContractor returns contractor.
func (d *CurrencyDiffDocument) GetContractor() types.Contractor {
	return d.Company
}

// GetNotes returns notes.
func (d *CurrencyDiffDocument) GetNotes() string {
	return "Różnice kursowe"
}

// GetSheet returns sheet to report.
func (d *CurrencyDiffDocument) GetSheet() types.Sheet {
	return types.Sheet{
		Date:     d.Document.Date,
		Template: currencyDiffTemplate,
		Data:     d,
		Config: types.SheetConfig{
			Name:       d.Document.SheetName,
			LockedRows: 8,
		},
	}
}

// CurrencyDiffRecord represents currency diff record.
type CurrencyDiffRecord struct {
	Date            time.Time
	Index           uint64
	DayOfMonth      uint8
	Document        types.Document
	PaymentDocument types.DocumentID
	Contractor      types.Contractor
	Amount          types.Denom
	DocumentRate    types.Number
	PaymentRate     types.Number
	Income          types.Denom
	Cost            types.Denom
}

// NewCurrencyDiffSummary returns new summary of currency diff.
func NewCurrencyDiffSummary() CurrencyDiffSummary {
	return CurrencyDiffSummary{
		Income: types.BaseZero,
		Cost:   types.BaseZero,
	}
}

// CurrencyDiffSummary is the summary of currency diff document.
type CurrencyDiffSummary struct {
	Income types.Denom
	Cost   types.Denom
}

// AddRecord adds record to the summary.
func (cds CurrencyDiffSummary) AddRecord(r CurrencyDiffRecord) CurrencyDiffSummary {
	cds.Income = cds.Income.Add(r.Income)
	cds.Cost = cds.Cost.Add(r.Cost)
	return cds
}

// NewCurrencyDiffDocument generates currency diff document.
func NewCurrencyDiffDocument(
	coa *types.ChartOfAccounts,
	document types.Document,
	company types.Contractor,
) *CurrencyDiffDocument {
	entries := coa.EntriesMonth(types.NewAccountID(accounts.RozniceKursowe), document.Date)
	if len(entries) == 0 {
		return nil
	}

	doc := &CurrencyDiffDocument{
		Document: document,
		Company:  company,
		Records:  make([]CurrencyDiffRecord, 0, len(entries)),
		Summary:  NewCurrencyDiffSummary(),
	}

	for i, e := range entries {
		data, ok := e.Data.(*types.CurrencyDiff)
		if !ok {
			panic("currency diff data source required")
		}

		r := CurrencyDiffRecord{
			Date:            data.GetDate(),
			Index:           uint64(i + 1),
			DayOfMonth:      uint8(data.GetDate().Day()),
			Document:        data.GetDocument(),
			PaymentDocument: data.BankRecord.Document,
			Contractor:      data.GetContractor(),
			Amount:          data.BankRecord.OriginalAmount.Abs(),
			DocumentRate:    data.DataRate,
			PaymentRate:     data.BankRecord.Rate,
			Income:          e.Amount.Credit,
			Cost:            e.Amount.Debit,
		}
		doc.Records = append(doc.Records, r)
		doc.Summary = doc.Summary.AddRecord(r)
	}

	return doc
}
