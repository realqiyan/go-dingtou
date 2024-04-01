package util

import (
	//"log"

	"github.com/shopspring/decimal"
)

// abs
func FloatAbs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// x * y
func FloatMul(x, y float64) float64 {
	bx, by := convert(x, y)
	result := bx.Mul(by)
	floatRet, _ := result.Float64()
	// log.Printf("%v * %v = %v , %v", bx, by, floatRet, accuracy)
	return floatRet
}

// x / y
func FloatDiv(x, y float64) float64 {
	bx, by := convert(x, y)
	result := bx.DivRound(by, 6)
	floatRet, _ := result.Float64()
	// log.Printf("%v / %v = %v , %v", bx, by, floatRet, accuracy)
	return floatRet
}

// x + y
func FloatAdd(x, y float64) float64 {
	bx, by := convert(x, y)
	result := bx.Add(by)
	floatRet, _ := result.Float64()
	// log.Printf("%v + %v = %v , %v", bx, by, floatRet, accuracy)
	return floatRet
}

// x - y
func FloatSub(x, y float64) float64 {
	bx, by := convert(x, y)
	result := bx.Sub(by)
	floatRet, _ := result.Float64()
	// log.Printf("%v - %v = %v , %v", bx, by, floatRet, accuracy)
	return floatRet
}

func convert(x, y float64) (decimal.Decimal, decimal.Decimal) {
	bx := decimal.NewFromFloat(x)
	by := decimal.NewFromFloat(y)
	return bx, by
}
