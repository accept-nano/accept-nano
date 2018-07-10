package main

import (
	"github.com/shopspring/decimal"
)

var rawMultiplier = decimal.RequireFromString("1000000000000000000000000000000")

func NanoToRaw(nano decimal.Decimal) decimal.Decimal {
	return nano.Mul(rawMultiplier)
}

func RawToNano(raw decimal.Decimal) decimal.Decimal {
	return raw.Div(rawMultiplier)
}
