package util

import (
	"log"

	"github.com/shopspring/decimal"
)

// x * y
func FloatMul(x, y float64) float64 {
	bx, by := convert(x, y)
	result := bx.Mul(by)
	floatRet, accuracy := result.Float64()
	log.Printf("%v * %v = %v , %v", bx, by, floatRet, accuracy)
	return floatRet
}

// x / y
func FloatDiv(x, y float64) float64 {
	bx, by := convert(x, y)
	result := bx.DivRound(by, 6)
	floatRet, accuracy := result.Float64()
	log.Printf("%v / %v = %v , %v", bx, by, floatRet, accuracy)
	return floatRet
}

// x + y
func FloatAdd(x, y float64) float64 {
	bx, by := convert(x, y)
	result := bx.Add(by)
	floatRet, accuracy := result.Float64()
	log.Printf("%v + %v = %v , %v", bx, by, floatRet, accuracy)
	return floatRet
}

// x - y
func FloatSub(x, y float64) float64 {
	bx, by := convert(x, y)
	result := bx.Sub(by)
	floatRet, accuracy := result.Float64()
	log.Printf("%v - %v = %v , %v", bx, by, floatRet, accuracy)
	return floatRet
}

func convert(x, y float64) (decimal.Decimal, decimal.Decimal) {
	bx := decimal.NewFromFloat(x)
	by := decimal.NewFromFloat(y)
	return bx, by
}
