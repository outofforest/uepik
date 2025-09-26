package types

import (
	"sort"
	"time"
)

// NewChartOfAccounts creates new chart of accounts.
func NewChartOfAccounts(period Period, accounts ...*Account) *ChartOfAccounts {
	ch := &ChartOfAccounts{
		period:   period,
		accounts: map[AccountIDPart]*Account{},
	}

	for _, account := range accounts {
		if _, exists := ch.accounts[account.idPart]; exists {
			panic("account already registered")
		}
		ch.accounts[account.idPart] = account
	}

	return ch
}

// ChartOfAccounts represents chart of accounts.
type ChartOfAccounts struct {
	period   Period
	accounts map[AccountIDPart]*Account
	entryID  EntryID
}

// AddEntry adds entry to the account.
func (ch *ChartOfAccounts) AddEntry(accountID AccountID, entry *Entry) {
	if len(accountID) == 0 {
		panic("empty account ID")
	}

	if !ch.period.Contains(entry.Date) {
		return
	}

	entry.ID = ch.entryID
	ch.entryID++

	accounts := ch.accounts
	var account *Account
	for _, idPart := range accountID {
		var exists bool
		account, exists = accounts[idPart]
		if !exists {
			panic("account does not exist")
		}
		account.addEntry(entry)
		accounts = account.children
	}
	if len(account.children) > 0 {
		panic("entry must be added to a leaf account")
	}
}

// SumMonth returns sum of the entries on the account in month.
func (ch *ChartOfAccounts) SumMonth(accountID AccountID, date time.Time) Denom {
	sum, exists := ch.getAccount(accountID).sums[newMonthKey(date)]
	if !exists {
		return BaseZero
	}
	return sum
}

// SumIncremental returns sum of the entries on the account in current month and all the previous ones.
func (ch *ChartOfAccounts) SumIncremental(accountID AccountID, date time.Time) Denom {
	account := ch.getAccount(accountID)
	sum := BaseZero

	mKey := newMonthKey(date)
	for month := ch.period.Start; ; month = month.AddDate(0, 1, 0) {
		mKey2 := newMonthKey(month)
		if sum2, exists := account.sums[mKey2]; exists {
			sum = sum.Add(sum2)
		}
		if mKey2 == mKey {
			break
		}
	}

	return sum
}

// Amount returns amount of the entry on the account.
func (ch *ChartOfAccounts) Amount(accountID AccountID, entryID EntryID) Denom {
	entry, exists := ch.getAccount(accountID).entries[entryID]
	if !exists {
		return BaseZero
	}
	return entry.Amount
}

// Entries returns entries on the account.
func (ch *ChartOfAccounts) Entries(accountID AccountID) []*Entry {
	entries := ch.getAccount(accountID).entries
	results := make([]*Entry, 0, len(entries))
	for _, e := range entries {
		results = append(results, e)
	}

	sort.Slice(results, func(i, j int) bool {
		r1 := results[i]
		r2 := results[j]
		return r1.Date.Before(r2.Date) || (r1.Date.Equal(r2.Date) && (r1.Index < r2.Index ||
			(r1.Index == r2.Index && r1.ID < r2.ID)))
	})

	return results
}

func (ch *ChartOfAccounts) getAccount(accountID AccountID) *Account {
	if len(accountID) == 0 {
		panic("empty account ID")
	}

	accounts := ch.accounts
	var account *Account
	for _, idPart := range accountID {
		var exists bool
		account, exists = accounts[idPart]
		if !exists {
			panic("account does not exist")
		}
		accounts = account.children
	}
	return account
}

func newMonthKey(date time.Time) monthKey {
	return monthKey{
		year:  date.Year(),
		month: date.Month(),
	}
}

type monthKey struct {
	year  int
	month time.Month
}

// AccountIDPart is a part of the account ID.
type AccountIDPart uint64

// AccountID is an ID of account.
type AccountID []AccountIDPart

// NewAccountID builds account ID from parts.
func NewAccountID(parts ...AccountIDPart) AccountID {
	return parts
}

// NewAccount creates new account.
func NewAccount(idPart AccountIDPart, children ...*Account) *Account {
	a := &Account{
		idPart:   idPart,
		children: map[AccountIDPart]*Account{},
		entries:  map[EntryID]*Entry{},
		sums:     map[monthKey]Denom{},
	}
	for _, child := range children {
		if _, exists := a.children[child.idPart]; exists {
			panic("child account already registered")
		}
		a.children[child.idPart] = child
	}
	return a
}

// Account represents account in the chart of accounts.
type Account struct {
	children map[AccountIDPart]*Account
	idPart   AccountIDPart
	entries  map[EntryID]*Entry
	sums     map[monthKey]Denom
}

func (a *Account) addEntry(entry *Entry) {
	if _, exists := a.entries[entry.ID]; exists {
		panic("entry already exists")
	}
	a.entries[entry.ID] = entry
	mKey := newMonthKey(entry.Date)
	sumMonth, exists := a.sums[mKey]
	if !exists {
		sumMonth = BaseZero
	}
	a.sums[mKey] = sumMonth.Add(entry.Amount)
}

// EntryID represents entry ID.
type EntryID uint64

// NewEntry creates new entry.
func NewEntry(
	date time.Time,
	index uint64,
	document Document,
	contractor Contractor,
	amount Denom,
	notes string,
) *Entry {
	return &Entry{
		Date:       date,
		Index:      index,
		Document:   document,
		Contractor: contractor,
		Amount:     amount,
		Notes:      notes,
	}
}

// Entry represents entry on the account.
type Entry struct {
	ID         EntryID
	Date       time.Time
	Index      uint64
	Document   Document
	Contractor Contractor
	Amount     Denom
	Notes      string
}

// GetDate returns date of the entry.
func (e *Entry) GetDate() time.Time {
	return e.Date
}
