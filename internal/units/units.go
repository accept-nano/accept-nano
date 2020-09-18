package units

import (
	"github.com/shopspring/decimal"
)

const exp = 30

func NanoToRaw(nano decimal.Decimal) decimal.Decimal {
	return nano.Shift(exp)
}

func RawToNano(raw decimal.Decimal) decimal.Decimal {
	return raw.Shift(-exp)
}
