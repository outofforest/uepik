package generators

import (
	"strings"

	"github.com/samber/lo"
	"github.com/xuri/excelize/v2"
)

const rowOffset = 2

var dateStyle = &excelize.Style{
	CustomNumFmt: lo.ToPtr(`yyyy-mm-dd`),
}

var textStyle = &excelize.Style{
	CustomNumFmt: lo.ToPtr("@"),
}

var intStyle = &excelize.Style{
	CustomNumFmt: lo.ToPtr("0"),
}

var styleHeader = &excelize.Style{
	Alignment: &excelize.Alignment{
		Horizontal: "center",
		Vertical:   "center",
		WrapText:   true,
	},
	Font: &excelize.Font{
		Bold: true,
	},
}

func excelNumberFormat(precision uint64) *excelize.Style {
	format := "0"
	if precision > 0 {
		format += "." + strings.Repeat("0", int(precision))
	}
	return &excelize.Style{
		CustomNumFmt: lo.ToPtr(format),
	}
}
