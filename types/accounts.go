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

// OpenAccount sets initial balance on account.
func (ch *ChartOfAccounts) OpenAccount(accountID AccountID, balance AccountBalance) {
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
		verifyBalanceAndType(balance, account.accountType)
		if len(account.children) == 0 && (len(account.balances) > 0 || !account.openingBalance.IsZero()) {
			panic("cannot set opening balance on non-empty account")
		}
		account.openingBalance = account.openingBalance.Add(balance)
		accounts = account.children
	}

	if len(account.children) > 0 {
		panic("cannot set opening balance on non-leaf account")
	}
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

// OpeningBalance returns opening balance of the account.
func (ch *ChartOfAccounts) OpeningBalance(accountID AccountID) Denom {
	account := ch.getAccount(accountID)
	return account.accountType.balanceFn(account.openingBalance)
}

// DebitMonth returns debit balance on the account in month.
func (ch *ChartOfAccounts) DebitMonth(accountID AccountID, date time.Time) Denom {
	account := ch.getAccount(accountID)
	balance, exists := account.balances[newMonthKey(date)]
	if !exists {
		return BaseZero
	}
	return balance.Debit
}

// CreditMonth returns credit balance on the account in month.
func (ch *ChartOfAccounts) CreditMonth(accountID AccountID, date time.Time) Denom {
	account := ch.getAccount(accountID)
	balance, exists := account.balances[newMonthKey(date)]
	if !exists {
		return BaseZero
	}
	return balance.Credit
}

// BalanceMonth returns balance on the account in month.
func (ch *ChartOfAccounts) BalanceMonth(accountID AccountID, date time.Time) Denom {
	account := ch.getAccount(accountID)
	balance, exists := account.balances[newMonthKey(date)]
	if !exists {
		return BaseZero
	}
	return account.accountType.balanceFn(balance)
}

// DebitIncremental returns debit balance on the account in current month and all the previous ones.
func (ch *ChartOfAccounts) DebitIncremental(accountID AccountID, date time.Time) Denom {
	account := ch.getAccount(accountID)
	debit := account.openingBalance.Debit

	mKey := newMonthKey(date)
	for month := ch.period.Start; ; month = month.AddDate(0, 1, 0) {
		mKey2 := newMonthKey(month)
		if sum2, exists := account.balances[mKey2]; exists {
			debit = debit.Add(sum2.Debit)
		}
		if mKey2 == mKey {
			break
		}
	}

	return debit
}

// CreditIncremental returns credit balance on the account in current month and all the previous ones.
func (ch *ChartOfAccounts) CreditIncremental(accountID AccountID, date time.Time) Denom {
	account := ch.getAccount(accountID)
	credit := account.openingBalance.Credit

	mKey := newMonthKey(date)
	for month := ch.period.Start; ; month = month.AddDate(0, 1, 0) {
		mKey2 := newMonthKey(month)
		if sum2, exists := account.balances[mKey2]; exists {
			credit = credit.Add(sum2.Credit)
		}
		if mKey2 == mKey {
			break
		}
	}

	return credit
}

// BalanceIncremental returns balance on the account in current month and all the previous ones.
func (ch *ChartOfAccounts) BalanceIncremental(accountID AccountID, date time.Time) Denom {
	account := ch.getAccount(accountID)
	balance := account.openingBalance

	mKey := newMonthKey(date)
	for month := ch.period.Start; ; month = month.AddDate(0, 1, 0) {
		mKey2 := newMonthKey(month)
		if sum2, exists := account.balances[mKey2]; exists {
			balance = balance.Add(sum2)
		}
		if mKey2 == mKey {
			break
		}
	}

	return account.accountType.balanceFn(balance)
}

// Balance returns balance on the account.
func (ch *ChartOfAccounts) Balance(accountID AccountID) Denom {
	account := ch.getAccount(accountID)
	balance := account.openingBalance

	for _, b := range account.balances {
		balance = balance.Add(b)
	}

	return account.accountType.balanceFn(balance)
}

// Amount returns amount of the entry on the account.
func (ch *ChartOfAccounts) Amount(accountID AccountID, entryID EntryID) AccountBalance {
	entry, exists := ch.getAccount(accountID).entries[entryID]
	if !exists {
		return zeroAccountBalance
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
		return r1.Date.Before(r2.Date) || (r1.Date.Equal(r2.Date) && r1.ID < r2.ID)
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
func NewAccount(idPart AccountIDPart, accountType AccountType, children ...*Account) *Account {
	accountTypeDef, exists := accountTypes[accountType]
	if !exists {
		panic("account type does not exist")
	}
	a := &Account{
		idPart:         idPart,
		accountType:    accountTypeDef,
		children:       map[AccountIDPart]*Account{},
		entries:        map[EntryID]*Entry{},
		openingBalance: zeroAccountBalance,
		balances:       map[monthKey]AccountBalance{},
	}
	for _, child := range children {
		if _, exists := a.children[child.idPart]; exists {
			panic("child account already registered")
		}
		if child.accountType.allowDebit && !accountTypeDef.allowDebit {
			panic("debit not allowed on child account")
		}
		if child.accountType.allowCredit && !accountTypeDef.allowCredit {
			panic("credit not allowed on child account")
		}
		a.children[child.idPart] = child
	}
	return a
}

var zeroAccountBalance = AccountBalance{
	Debit:  BaseZero,
	Credit: BaseZero,
}

// DebitBalance creates account balance with debit.
func DebitBalance(amount Denom) AccountBalance {
	return AccountBalance{
		Debit:  amount,
		Credit: BaseZero,
	}
}

// CreditBalance creates account balance with credit.
func CreditBalance(amount Denom) AccountBalance {
	return AccountBalance{
		Debit:  BaseZero,
		Credit: amount,
	}
}

// AccountBalance stores balance of the account.
type AccountBalance struct {
	Debit  Denom
	Credit Denom
}

// IsZero returns if debit and credit balances are zero.
func (ab AccountBalance) IsZero() bool {
	return ab.Debit.EQ(BaseZero) && ab.Credit.EQ(BaseZero)
}

// Add adds balances.
func (ab AccountBalance) Add(balance AccountBalance) AccountBalance {
	ab.Debit = ab.Debit.Add(balance.Debit)
	ab.Credit = ab.Credit.Add(balance.Credit)
	return ab
}

func balanceDebitMinusCredit(balance AccountBalance) Denom {
	return balance.Debit.Sub(balance.Credit)
}

func balanceCreditMinusDebit(balance AccountBalance) Denom {
	return balance.Credit.Sub(balance.Debit)
}

// AccountType represent type of the account.
type AccountType uint

// Account types.
const (
	Assets AccountType = iota
	Liabilities
	Incomes
	Costs
)

// AccountTypeDefinition defines properties of the account.
type AccountTypeDefinition struct {
	allowDebit  bool
	allowCredit bool
	balanceFn   func(balance AccountBalance) Denom
}

var accountTypes = map[AccountType]AccountTypeDefinition{
	Assets: {
		allowDebit:  true,
		allowCredit: true,
		balanceFn:   balanceDebitMinusCredit,
	},
	Liabilities: {
		allowDebit:  true,
		allowCredit: true,
		balanceFn:   balanceCreditMinusDebit,
	},
	Incomes: {
		allowDebit:  false,
		allowCredit: true,
		balanceFn:   balanceCreditMinusDebit,
	},
	Costs: {
		allowDebit:  true,
		allowCredit: false,
		balanceFn:   balanceDebitMinusCredit,
	},
}

// Account represents account in the chart of accounts.
type Account struct {
	children       map[AccountIDPart]*Account
	idPart         AccountIDPart
	accountType    AccountTypeDefinition
	entries        map[EntryID]*Entry
	openingBalance AccountBalance
	balances       map[monthKey]AccountBalance
}

func (a *Account) addEntry(entry *Entry) {
	if _, exists := a.entries[entry.ID]; exists {
		panic("entry already exists")
	}
	verifyBalanceAndType(entry.Amount, a.accountType)

	a.entries[entry.ID] = entry
	mKey := newMonthKey(entry.Date)
	sumMonth, exists := a.balances[mKey]
	if !exists {
		sumMonth = zeroAccountBalance
	}
	a.balances[mKey] = sumMonth.Add(entry.Amount)
}

// EntryID represents entry ID.
type EntryID uint64

// NewEntry creates new entry.
func NewEntry(
	date time.Time,
	document Document,
	contractor Contractor,
	amount AccountBalance,
	notes string,
) *Entry {
	return &Entry{
		Date:       date,
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
	Document   Document
	Contractor Contractor
	Notes      string
	Amount     AccountBalance
}

// GetDate returns date of the entry.
func (e *Entry) GetDate() time.Time {
	return e.Date
}

func verifyBalanceAndType(balance AccountBalance, accountType AccountTypeDefinition) {
	if balance.Debit.NEQ(BaseZero) && !accountType.allowDebit {
		panic("debit not allowed on account")
	}
	if balance.Credit.NEQ(BaseZero) && !accountType.allowCredit {
		panic("debit not allowed on account")
	}
}
