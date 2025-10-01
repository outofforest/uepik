package operations

import (
	"time"

	"github.com/outofforest/uepik/accounts"
	"github.com/outofforest/uepik/types"
)

// Sell defines the income coming from goods or service sell.
type Sell struct {
	Date       time.Time
	Document   types.Document
	Contractor types.Contractor
	Dues       []types.Due
	Payments   []types.Payment
	Type       types.SellType
	Notes      string
}

// GetDate returns date of sell.
func (s *Sell) GetDate() time.Time {
	return s.Date
}

// GetDocument returns document.
func (s *Sell) GetDocument() types.Document {
	return s.Document
}

// GetContractor returns contractor.
func (s *Sell) GetContractor() types.Contractor {
	return s.Contractor
}

// GetNotes returns notes.
func (s *Sell) GetNotes() string {
	return s.Notes
}

// GetDues returns dues.
func (s *Sell) GetDues() []types.Due {
	return s.Dues
}

// GetPayments returns payments.
func (s *Sell) GetPayments() []types.Payment {
	return s.Payments
}

// BankRecords returns bank records for the sell.
func (s *Sell) BankRecords() []*types.BankRecord {
	records := []*types.BankRecord{}
	for _, payment := range s.Payments {
		records = append(records, &types.BankRecord{
			Date:           payment.Date,
			Index:          payment.Index,
			Document:       payment.DocumentID,
			PaidDocument:   s.Document,
			Contractor:     s.Contractor,
			OriginalAmount: payment.Amount,
		})
	}
	return records
}

// BookRecords returns book records for the sell.
func (s *Sell) BookRecords(
	period types.Period,
	coa *types.ChartOfAccounts,
	bankRecords []*types.BankRecord,
	rates types.CurrencyRates,
) []types.ReportDocument {
	if len(s.Dues) == 0 {
		panic("no dues")
	}
	amount := s.Dues[0].Amount
	for _, due := range s.Dues[1:] {
		amount = amount.Add(due.Amount)
	}

	incomeBase, incomeRate := rates.ToBase(amount, types.PreviousDay(s.Date))

	coa.AddEntry(s,
		types.NewEntryRecord(
			sellTypeToAccountID(s.Type),
			types.CreditBalance(incomeBase),
		),
		types.NewEntryRecord(
			types.NewAccountID(accounts.Odplatna),
			types.CreditBalance(incomeBase),
		),
		types.NewEntryRecord(
			types.NewAccountID(accounts.NiewydatkowanyDochod, accounts.WTrakcieRoku),
			types.CreditBalance(incomeBase),
		),
	)

	for _, br := range bankRecords {
		if br.Rate.EQ(incomeRate) {
			continue
		}

		var amount types.AccountBalance
		if incomeRate.GT(br.Rate) {
			amount = types.DebitBalance(br.OriginalAmount.ToBase(incomeRate.Sub(br.Rate)))
		} else {
			amount = types.CreditBalance(br.OriginalAmount.ToBase(br.Rate.Sub(incomeRate)))
		}

		coa.AddEntry(types.NewCurrencyDiff(s, incomeRate, br),
			types.NewEntryRecord(
				types.NewAccountID(accounts.RozniceKursowe, accounts.Odplatna),
				amount,
			),
		)

		vatDate := types.MinDate(s.Date, br.Date)
		vatBase, _ := rates.ToBase(br.OriginalAmount, types.PreviousDay(vatDate))
		coa.AddEntry(types.NewVAT(vatDate, s),
			types.NewEntryRecord(
				types.NewAccountID(accounts.VAT),
				types.CreditBalance(vatBase),
			),
		)
	}

	return nil
}

func sellTypeToAccountID(sellType types.SellType) types.AccountID {
	switch sellType {
	case types.SellTypeRecorded:
		return types.NewAccountID(accounts.PiK, accounts.Przychody, accounts.Operacyjne, accounts.Odplatna)
	case types.SellTypeUnrecorded:
		return types.NewAccountID(accounts.SprzedazNieewidencjonowana)
	default:
		panic("invalid sell type")
	}
}
