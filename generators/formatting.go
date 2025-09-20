package generators

import (
	"strings"
	"time"

	"github.com/samber/lo"
	"github.com/xuri/excelize/v2"
)

var dateStyle = &excelize.Style{
	CustomNumFmt: lo.ToPtr(`yyyy-mm-dd`),
}

var textStyle = &excelize.Style{
	CustomNumFmt: lo.ToPtr("@"),
	Alignment: &excelize.Alignment{
		Horizontal: "left",
		Vertical:   "top",
		WrapText:   true,
	},
}

var intStyle = &excelize.Style{
	CustomNumFmt: lo.ToPtr("0"),
	Alignment: &excelize.Alignment{
		Horizontal: "right",
		Vertical:   "top",
	},
}

var headerStyle = &excelize.Style{
	Alignment: &excelize.Alignment{
		Horizontal: "center",
		Vertical:   "center",
		WrapText:   true,
	},
	Font: &excelize.Font{
		Size: 8,
		Bold: true,
	},
}

var columnIndexStyle = &excelize.Style{
	CustomNumFmt: lo.ToPtr("0"),
	Alignment: &excelize.Alignment{
		Horizontal: "center",
		Vertical:   "center",
	},
}

func excelNumberFormat(precision uint64) *excelize.Style {
	format := "0"
	if precision > 0 {
		format += "." + strings.Repeat("0", int(precision))
	}
	return &excelize.Style{
		CustomNumFmt: lo.ToPtr(format),
		Alignment: &excelize.Alignment{
			Horizontal: "right",
			Vertical:   "top",
		},
	}
}

func monthName(month time.Month) string {
	switch month {
	case time.January:
		return "styczeń"
	case time.February:
		return "luty"
	case time.March:
		return "marzec"
	case time.April:
		return "kwiecień"
	case time.May:
		return "maj"
	case time.June:
		return "czerwiec"
	case time.July:
		return "lipiec"
	case time.August:
		return "sierpień"
	case time.September:
		return "wrzesień"
	case time.October:
		return "październik"
	case time.November:
		return "listopad"
	case time.December:
		return "grudzień"
	default:
		panic("invalid month")
	}
}

func width(cms float64) float64 {
	return 5.11 * cms
}
