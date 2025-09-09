package data

import (
	"testing"

	"github.com/samber/lo"
	"github.com/xuri/excelize/v2"

	"github.com/outofforest/uepik/generators"
)

func Test2025(t *testing.T) {
	f := excelize.NewFile()
	defer f.Close()

	lo.Must0(generators.Bank(f, R2025))
	lo.Must0(generators.Book(f, R2025))

	lo.Must0(f.DeleteSheet(f.GetSheetName(0)))
	lo.Must0(f.SaveAs("Data.xlsx"))
}
